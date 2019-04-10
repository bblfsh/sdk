package cmd

import (
	"github.com/bblfsh/sdk/v3/build"
	"github.com/bblfsh/sdk/v3/cmd"
)

const UpdateCommandDescription = "updates an already initialized driver"

type UpdateCommand struct {
	DryRun bool `long:"dry-run" description:"don't writes nothing just checks if something should be written"`

	cmd.Command
}

func (c *UpdateCommand) Options() *build.UpdateOptions {
	opt := &build.UpdateOptions{
		DryRun:  c.DryRun,
		Notice:  cmd.Notice.Printf,
		Warning: cmd.Warning.Printf,
	}
	if c.Verbose {
		opt.Debug = cmd.Debug.Printf
	}
	return opt
}

func (c *UpdateCommand) Execute(args []string) error {
	return build.UpdateSDK(c.Root, c.Options())
}
