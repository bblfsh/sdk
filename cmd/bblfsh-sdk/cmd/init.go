package cmd

import (
	"gopkg.in/bblfsh/sdk.v2/build"
	"gopkg.in/bblfsh/sdk.v2/cmd"
)

const InitCommandDescription = "initializes a driver for a given language and OS"

type InitCommand struct {
	Args struct {
		Language string `positional-arg-name:"language"  description:"target language of the driver"`
	} `positional-args:"yes"`

	cmd.Command
}

func (c *InitCommand) Execute(args []string) error {
	opt := &build.InitOptions{
		Notice:  cmd.Notice.Printf,
		Warning: cmd.Warning.Printf,
	}
	if c.Verbose {
		opt.Debug = cmd.Debug.Printf
	}
	return build.InitDriver(c.Root, c.Args.Language, opt)
}
