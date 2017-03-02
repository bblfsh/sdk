package cmd

import (
	"fmt"

	"github.com/bblfsh/sdk/cli"
	"github.com/bblfsh/sdk/manifest"
)

const InitCommandDescription = "initializes a driver for a given language and OS"

type InitCommand struct {
	Args struct {
		Language string `positional-arg-name:"language"  description:"target langunge of the driver"`
		OS       string `positional-arg-name:"os" description:"distribution used to run the runtime. (Values: alpine or debian)"`
	} `positional-args:"yes"`

	UpdateCommand
}

func (c *InitCommand) Execute(args []string) error {
	if err := c.processManifest(); err != nil {
		return err
	}

	return c.UpdateCommand.Execute(args)
}

func (c *InitCommand) processManifest() error {
	if c.Args.Language == "" || c.Args.OS == "" {
		return fmt.Errorf("`language` and `os` arguments are mandatory")
	}

	cli.Notice.Printf("initializing driver %q, creating new manifest\n", c.Args.Language)
	if _, err := c.readManifest(); err == nil {
		cli.Warning.Printf("driver already initialized. %q detected\n", manifest.Filename)
	}

	return c.processTemplateAsset(manifestTpl, c.Args, false)
}
