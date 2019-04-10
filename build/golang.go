package build

import (
	"os"
	"os/exec"
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

func goExec(dir string, args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	return cmd.Run()
}

func modExec(dir string, args ...string) error {
	args = append([]string{"mod"}, args...)
	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	return cmd.Run()
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
