package build

import (
	"bytes"
	"os/exec"
)

func gitRev(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = path
	data, err := cmd.Output()
	if err != nil {
		return "", err
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
