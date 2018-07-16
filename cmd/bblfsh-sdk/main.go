package main

import (
	"fmt"
	"os"

	"gopkg.in/bblfsh/sdk.v2/cmd/bblfsh-sdk/cmd"

	"github.com/jessevdk/go-flags"
)

var version string
var build string

func main() {
	parser := flags.NewNamedParser("bblfsh-sdk", flags.Default)
	parser.AddCommand("prepare-build", cmd.PrepareBuildCommandDescription, "", &cmd.PrepareBuildCommand{})
	parser.AddCommand("update", cmd.UpdateCommandDescription, "", &cmd.UpdateCommand{})
	parser.AddCommand("init", cmd.InitCommandDescription, "", &cmd.InitCommand{})
	parser.AddCommand("prepare", cmd.PrepareCommandDescription, "", &cmd.PrepareCommand{})
	parser.AddCommand("build", cmd.BuildCommandDescription, "", &cmd.BuildCommand{})
	parser.AddCommand("test", cmd.TestCommandDescription, "", &cmd.TestCommand{})

	if _, err := parser.Parse(); err != nil {
		if _, ok := err.(*flags.Error); ok {
			parser.WriteHelp(os.Stdout)
			fmt.Printf("\nBuild information\n  commit: %s\n  date:%s\n", version, build)
		}

		os.Exit(1)
	}
}
