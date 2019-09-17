package main

import (
	"fmt"
	"os"

	"github.com/bblfsh/sdk/v3/cmd/bblfsh-sdk/cmd"

	"github.com/jessevdk/go-flags"
)

var version string
var build string

func main() {
	parser := flags.NewNamedParser("bblfsh-sdk", flags.Default)
	parser.AddCommand("info", cmd.InfoCommandDescription, "", &cmd.InfoCommand{})
	parser.AddCommand("update", cmd.UpdateCommandDescription, "", &cmd.UpdateCommand{})
	parser.AddCommand("init", cmd.InitCommandDescription, "", &cmd.InitCommand{})
	parser.AddCommand("build", cmd.BuildCommandDescription, "", &cmd.BuildCommand{})
	parser.AddCommand("test", cmd.TestCommandDescription, "", &cmd.TestCommand{})
	parser.AddCommand("tag", cmd.TagCommandDescription, "", &cmd.TagCommand{})
	parser.AddCommand("release", cmd.ReleaseCommandDescription, "", &cmd.ReleaseCommand{})
	parser.AddCommand("push", cmd.PushCommandDescription, "", &cmd.PushCommand{})
	parser.AddCommand("ast2gv", cmd.Ast2GraphvizCommandDescription, "", &cmd.Ast2GraphvizCommand{})
	parser.AddCommand("request", cmd.RequestCommandDescription, "", &cmd.RequestCommand{})

	if _, err := parser.Parse(); err != nil {
		if _, ok := err.(*flags.Error); ok {
			parser.WriteHelp(os.Stdout)
			fmt.Printf("\nBuild information\n  commit: %s\n  date:%s\n", version, build)
		}

		os.Exit(1)
	}
}
