package build

import (
	"os"
	"path/filepath"
)

func depEnsure(path string) error {
	if _, err := os.Stat(filepath.Join(path, "Gopkg.toml")); os.IsNotExist(err) {
		err = execIn(path, nil, "dep", "init")
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(path, "vendor")); os.IsNotExist(err) {
		err = execIn(path, nil, "dep", "ensure", "--vendor-only")
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
