package build

import (
	"io"
	"os"
)

func (d *Driver) modPrepare() error {
	if _, err := os.Stat(d.path("go.sum")); os.IsNotExist(err) {
		err = modTidy(d.root)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if _, err := os.Stat(d.path("vendor")); os.IsNotExist(err) {
		err = modVendor(d.root)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func goCmd(dir string, stdout io.Writer, args ...string) Cmd {
	cmd := cmdIn(dir, "go", args...)
	cmd.Stdout = stdout
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	return cmd
}

func modExec(dir string, args ...string) error {
	return goCmd(dir, nil, append([]string{"mod"}, args...)...).Run()
}

func modTidy(dir string) error {
	return modExec(dir, "tidy")
}

func modInit(dir, pkg string) error {
	return modExec(dir, "init", pkg)
}

func modVendor(dir string) error {
	return modExec(dir, "vendor")
}
