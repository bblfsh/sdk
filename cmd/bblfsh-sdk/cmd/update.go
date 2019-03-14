package cmd

import (
	"gopkg.in/bblfsh/sdk.v2/build"
	"gopkg.in/bblfsh/sdk.v2/cmd"
)

const UpdateCommandDescription = "updates an already initialized driver"

type UpdateCommand struct {
	DryRun bool `long:"dry-run" description:"don't writes nothing just checks if something should be written"`

	cmd.Command
}

func (c *UpdateCommand) Options() *build.UpdateOptions {
	opt := &build.UpdateOptions{
		DryRun:   c.DryRun,
		Noticef:  cmd.Notice.Printf,
		Warningf: cmd.Warning.Printf,
	}
	if c.Verbose {
		opt.Debugf = cmd.Debug.Printf
	}
	return opt
}

func (c *UpdateCommand) Execute(args []string) error {
	return build.SDKUpdate(c.Root, c.Options())
}
