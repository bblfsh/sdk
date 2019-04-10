package build

import (
	"os"
)

func (d *Driver) modTidy() error {
	if _, err := os.Stat(d.path("go.sum")); os.IsNotExist(err) {
		err = execIn(d.root, nil, "go", "mod", "tidy")
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if _, err := os.Stat(d.path("vendor")); os.IsNotExist(err) {
		err = execIn(d.root, nil, "go", "mod", "vendor")
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
