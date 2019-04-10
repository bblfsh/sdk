package build

import (
	"bytes"
	"fmt"
	"os/exec"
)

func gitRev(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = path
	data, err := cmd.Output()
	if err != nil {
		// maybe we don't have any commits yet
		cmd = exec.Command("git", "rev-list", "-n", "1", "--all")
		cmd.Dir = path
		data2, err2 := cmd.Output()
		data2 = bytes.TrimSpace(data2)
		if err2 != nil || len(data2) != 0 {
			// return original errors
			return "", fmt.Errorf("cannot determine git hash: %v", err)
		}
		// no commits yet
		return "", nil
	}
	return string(bytes.TrimSpace(data)), nil
}

func gitHasChanges(path string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain", "-uno")
	cmd.Dir = path
	data, err := cmd.Output()
	if err != nil {
		return false, err
	}
	data = bytes.TrimSpace(data)
	return len(data) != 0, nil
}

func gitInit(path string) error {
	if out, err := exec.Command("git", "init", path).CombinedOutput(); err != nil {
		return fmt.Errorf("cannot init a repository: %v\n%s\n", err, string(out))
	}
	return nil
}

func gitAdd(root, file string) error {
	git := exec.Command("git", "add", file)
	git.Dir = root
	if out, err := git.CombinedOutput(); err != nil {
		return fmt.Errorf("cannot add a file to git: %v\n%s\n", err, string(out))
	}
	return nil
}
