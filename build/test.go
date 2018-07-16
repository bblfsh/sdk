package build

import (
	"os"
	"os/user"
	"path/filepath"
)

func Test(path string) error {
	id, err := Build(path, "")
	if err != nil {
		return err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	return execIn(path, os.Stdout, "docker", "run", "--rm",
		"-u", u.Uid+":"+u.Gid,
		"-e", "BBLFSH_TEST_LOCAL=true",
		"-v", filepath.Join(path, "fixtures")+":/opt/fixtures",
		"--workdir", "/opt/driver/bin",
		"--entrypoint", "./fixtures.test",
		id,
	)
}
