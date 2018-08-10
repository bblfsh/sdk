package build

import (
	"os"
)

func (d *Driver) depEnsure() error {
	if _, err := os.Stat(d.path("Gopkg.toml")); os.IsNotExist(err) {
		err = execIn(d.root, nil, "dep", "init")
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if _, err := os.Stat(d.path("Gopkg.lock")); os.IsNotExist(err) {
		err = execIn(d.root, nil, "dep", "ensure")
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if _, err := os.Stat(d.path("vendor")); os.IsNotExist(err) {
		err = execIn(d.root, nil, "dep", "ensure", "--vendor-only")
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
