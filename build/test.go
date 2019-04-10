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

	"github.com/bblfsh/sdk/v3/driver"
	"github.com/bblfsh/sdk/v3/internal/docker"
	"github.com/bblfsh/sdk/v3/uast/uastyaml"
	protocol1 "gopkg.in/bblfsh/sdk.v1/protocol"
	uast1 "gopkg.in/bblfsh/sdk.v1/uast"
)

const (
	integrationTestName = "_integration"
	syntaxErrTestName   = "_syntax_error"
	fixturesDir         = "fixtures"
)

func (d *Driver) Test(bblfshdVers, image string, bench bool) error {
	if image == "" {
		id, err := d.Build("")
		if err != nil {
			return err
		}
		image = id
	}
	if err := d.testFixtures(image, bench); err != nil {
		return err
	}
	if err := d.testIntegration(bblfshdVers, image); err != nil {
		return err
	}
	return nil
}

func (d *Driver) testFixtures(image string, bench bool) error {
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
	)
	printCommand(
		"docker", "run", "--rm",
		"-u", usr,
		"-v", mnt,
		"--workdir", wd,
		"--entrypoint", bin,
		image,
	)
	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			User:         usr,
			Image:        image,
			AttachStdout: true,
			AttachStderr: true,
			WorkingDir:   wd,
			Entrypoint:   []string{bin},
		},
		HostConfig: &docker.HostConfig{
			AutoRemove: true,
			Binds:      []string{mnt},
		},
	}
	// Run tests.
	//
	// Write output to buffer and show it only if the test fails.
	if err := d.runContainer(cli, nil, opts); err != nil {
		return err
	}
	if !bench {
		return nil
	}
	// Run benchmarks.
	//
	// We do it in two separate passes to isolate benchmark output from test output.
	outPath := filepath.Join(d.root, "bench.txt")
	fmt.Fprintf(os.Stderr, "running benchmarks (%s)\n", outPath)
	benchLog, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer benchLog.Close()

	printCommand(
		"docker", "run", "--rm",
		"-u", usr,
		"-v", mnt,
		"--workdir", wd,
		"--entrypoint", `'`+bin+` -test.run=NONE -test.bench=.'`,
		image,
	)

	opts.Config.Entrypoint = append(opts.Config.Entrypoint, "-test.run=NONE", "-test.bench=.")
	if err := d.runContainer(cli, benchLog, opts); err != nil {
		return err
	}
	if err = benchLog.Close(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) runContainer(cli *docker.Client, sout io.Writer, opts docker.CreateContainerOptions) error {
	buf := bytes.NewBuffer(nil)
	var stdout io.Writer = buf
	if Verbose() {
		stdout = io.MultiWriter(buf, os.Stderr)
	}
	var stderr = stdout
	if sout != nil {
		stdout = io.MultiWriter(stdout, sout)
	}
	err := docker.RunAndWait(cli, stdout, stderr, opts)
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
		exp, err := uastyaml.Marshal(ast)
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

	pref = d.path(fixturesDir, syntaxErrTestName)
	names, err = filepath.Glob(pref + ".*")
	if err != nil {
		return err
	}
	tests = nil
	for _, name := range names {
		suff := strings.TrimPrefix(name, pref)
		if strings.Count(suff, ".") > 1 {
			// we want files with a single extension
			continue
		}
		tests = append(tests, name)
	}
	if len(tests) == 0 {
		fmt.Fprintf(os.Stderr, "WARNING: expected at least one test called './%s/%s.xxx'\n", fixturesDir, syntaxErrTestName)
		return nil
	}

	for _, name := range tests {
		data, err := ioutil.ReadFile(name)
		if err != nil {
			return err
		}
		content := string(data)

		// test v2 protocol
		_, err = cli2.Parse(ctx, content, &driver.ParseOptions{
			Mode:     driver.ModeSemantic,
			Language: lang,
		})
		if !driver.ErrSyntax.Is(err) {
			srv.DumpLogs(os.Stderr)
			return fmt.Errorf("expected syntax error, got: %v", err)
		}
	}

	return nil
}
