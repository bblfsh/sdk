package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func create(path string) (*os.File, error) {
	fmt.Fprintln(os.Stderr, "w", path)
	return os.Create(path)
}

func execIn(wd string, out io.Writer, cmd string, args ...string) error {
	fmt.Fprintln(os.Stderr, "+", cmd, strings.Join(args, " "))
	c := exec.Command(cmd, args...)
	c.Dir = wd
	c.Stderr = os.Stderr
	c.Stdout = out
	return c.Run()
}
