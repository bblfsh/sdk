package main

import (
	"fmt"
	"os"

	"github.com/bblfsh/sdk/v3/cmd/bblfsh-drivers-release/cmd"

	"github.com/jessevdk/go-flags"
)

func main() {
	parser := flags.NewNamedParser("bblfsh-drivers-release", flags.Default)
	parser.AddCommand("collect", cmd.CollectCommandDescription, "", &cmd.CollectCommand{})
	parser.AddCommand("apply", cmd.ApplyCommandDescription, "", &cmd.ApplyCommand{})

	if _, err := parser.Parse(); err != nil {
		if _, ok := err.(*flags.Error); ok {
			parser.WriteHelp(os.Stdout)
		}
		fmt.Println()
		os.Exit(1)
	}
}
