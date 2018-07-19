package build

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/bblfsh/sdk.v2/internal/docker"
	"gopkg.in/bblfsh/sdk.v2/protocol"
)

const (
	integrationTestName = "_integration"
	fixturesDir         = "fixtures"
)

func (d *Driver) Test(image string) error {
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
	if err := d.testIntegration(image); err != nil {
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
	return docker.RunAndWait(cli, os.Stdout, os.Stderr, docker.CreateContainerOptions{
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
}

func (d *Driver) testIntegration(image string) error {
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

	srv, err := RunWithDriver(lang, image)
	if err != nil {
		return err
	}
	defer srv.Close()

	const timeout = time.Minute * 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cli, err := srv.ClientV1(ctx)
	if err != nil {
		return err
	}

	for _, name := range tests {
		data, err := ioutil.ReadFile(name)
		if err != nil {
			return err
		}
		resp, err := cli.Parse(ctx, &protocol.ParseRequest{
			Language: lang,
			Content:  string(data),
		})
		if err != nil {
			srv.DumpLogs(os.Stderr)
			return err
		} else if resp.Status != protocol.Ok {
			srv.DumpLogs(os.Stderr)
			return fmt.Errorf("parse error: %v", resp.Errors)
		}
		buf := bytes.NewBuffer(nil)
		err = protocol.Pretty(resp.UAST, buf, protocol.IncludeAll)
		if err != nil {
			return err
		}
		expName := name + ".legacy"
		data, err = ioutil.ReadFile(expName)
		if err == nil {
			if !bytes.Equal(data, buf.Bytes()) {
				ioutil.WriteFile(expName+"_got", buf.Bytes(), 0644)
				return fmt.Errorf("test %q failed", name)
			}
		} else if os.IsNotExist(err) {
			ioutil.WriteFile(expName, buf.Bytes(), 0644)
		} else if err != nil {
			return err
		}
	}

	return nil
}
