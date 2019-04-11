package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func printCommand(cmd string, args ...string) {
	fmt.Fprintln(os.Stderr, "+", cmd, strings.Join(args, " "))
}

func execIn(wd string, out io.Writer, cmd string, args ...string) error {
	printCommand(cmd, args...)
	c := exec.Command(cmd, args...)
	c.Dir = wd
	c.Stderr = os.Stderr
	c.Stdout = out
	return c.Run()
}

func cmdIn(wd string, cmd string, args ...string) Cmd {
	printCommand(cmd, args...)
	c := exec.Command(cmd, args...)
	c.Dir = wd
	return Cmd{c}
}

type Cmd struct {
	*exec.Cmd
}

func (c Cmd) Run() error {
	err := c.Cmd.Run()
	if e, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("%s\n%s", e.ProcessState, string(e.Stderr))
	}
	return err
}
