package build

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	protocol1 "gopkg.in/bblfsh/sdk.v1/protocol"
	uast1 "gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v2/driver"
	"gopkg.in/bblfsh/sdk.v2/internal/docker"
	"gopkg.in/bblfsh/sdk.v2/uast/yaml"
)

const (
	integrationTestName = "_integration"
	fixturesDir         = "fixtures"
)

func (d *Driver) Test(bblfshdVers, image string) error {
	if image == "" {
		id, err := d.Build("")
		if err != nil {
			return err
		}
		image = id
	}
	if err := d.testFixtures(image); err != nil {
		return err
	}
	if err := d.testIntegration(bblfshdVers, image); err != nil {
		return err
	}
	return nil
}

func (d *Driver) testFixtures(image string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	cli, err := docker.Dial()
	if err != nil {
		return err
	}
	usr := u.Uid + ":" + u.Gid
	mnt := filepath.Join(d.root, fixturesDir) + ":/opt/fixtures"
	const (
		wd  = "/opt/driver/bin"
		bin = "./fixtures.test"
		env = "BBLFSH_TEST_LOCAL=true"
	)
	printCommand(
		"docker", "run", "--rm",
		"-u", usr,
		"-e", env,
		"-v", mnt,
		"--workdir", wd,
		"--entrypoint", bin,
		image,
	)
	buf := bytes.NewBuffer(nil)
	var out io.Writer = buf
	if Verbose() {
		out = io.MultiWriter(buf, os.Stderr)
	}
	err = docker.RunAndWait(cli, out, out, docker.CreateContainerOptions{
		Config: &docker.Config{
			User:         usr,
			Image:        image,
			AttachStdout: true,
			AttachStderr: true,
			WorkingDir:   wd,
			Entrypoint:   []string{bin},
			Env: []string{
				env,
			},
		},
		HostConfig: &docker.HostConfig{
			AutoRemove: true,
			Binds:      []string{mnt},
		},
	})
	if err != nil {
		if !Verbose() {
			buf.WriteTo(os.Stderr)
		}
		return err
	} else if bytes.Contains(buf.Bytes(), []byte("FAIL")) {
		buf.WriteTo(os.Stderr)
		return fmt.Errorf("tests failed")
	}
	return nil
}

func (d *Driver) testIntegration(bblfshdVers, image string) error {
	m, err := d.readBuildManifest()
	if err != nil {
		return err
	}
	lang := m.Language

	pref := d.path(fixturesDir, integrationTestName)
	names, err := filepath.Glob(pref + ".*")
	if err != nil {
		return err
	}
	var tests []string
	for _, name := range names {
		suff := strings.TrimPrefix(name, pref)
		if strings.Count(suff, ".") > 1 {
			// we want files with a single extension
			continue
		}
		tests = append(tests, name)
	}
	if len(tests) == 0 {
		return fmt.Errorf("expected at least one test called './%s/%s.xxx'", fixturesDir, integrationTestName)
	}

	srv, err := RunWithDriver(bblfshdVers, lang, image)
	if err != nil {
		return err
	}
	defer srv.Close()

	const timeout = time.Minute * 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cli1, err := srv.ClientV1(ctx)
	if err != nil {
		return err
	}
	cli2, err := srv.ClientV2(ctx)
	if err != nil {
		return err
	}

	for _, name := range tests {
		data, err := ioutil.ReadFile(name)
		if err != nil {
			return err
		}
		content := string(data)

		// test v1 protocol
		resp, err := cli1.Parse(ctx, &protocol1.ParseRequest{
			Language: lang,
			Content:  content,
		})
		if err != nil {
			srv.DumpLogs(os.Stderr)
			return err
		} else if resp.Status != protocol1.Ok {
			srv.DumpLogs(os.Stderr)
			return fmt.Errorf("parse error: %v", resp.Errors)
		}
		buf := bytes.NewBuffer(nil)
		err = uast1.Pretty(resp.UAST, buf, uast1.IncludeAll)
		if err != nil {
			return err
		}
		expName := name + ".legacy"
		data, err = ioutil.ReadFile(expName)
		if err == nil {
			if !bytes.Equal(data, buf.Bytes()) {
				ioutil.WriteFile(expName+"_got", buf.Bytes(), 0644)
				return fmt.Errorf("v1 test %q failed", name)
			}
			_ = os.Remove(expName + "_got")
		} else if os.IsNotExist(err) {
			ioutil.WriteFile(expName, buf.Bytes(), 0644)
		} else if err != nil {
			return err
		}

		// test v2 protocol
		ast, err := cli2.Parse(ctx, content, &driver.ParseOptions{
			Mode:     driver.ModeSemantic,
			Language: lang,
		})
		if err != nil {
			srv.DumpLogs(os.Stderr)
			return err
		}
		buf.Reset()
		exp, err := uastyml.Marshal(ast)
		if err != nil {
			return err
		}
		expName = name + ".sem.uast"
		got, err := ioutil.ReadFile(expName)
		if err != nil {
			return err
		} else if !bytes.Equal(exp, got) {
			ioutil.WriteFile(expName+"_got", buf.Bytes(), 0644)
			return fmt.Errorf("v2 test %q failed", name)
		}
	}

	return nil
}
