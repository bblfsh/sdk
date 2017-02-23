package cmd

import "fmt"

type BootstrapCommand struct {
}

func (c *BootstrapCommand) Execute(args []string) error {
	fmt.Println("not implemented")
	return nil
}
