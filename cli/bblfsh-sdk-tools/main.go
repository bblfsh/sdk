package main

import (
	"fmt"
	"os"

	"github.com/bblfsh/sdk/cli/bblfsh-sdk-tools/cmd"

	"github.com/jessevdk/go-flags"
)

var version string
var build string

func main() {
	parser := flags.NewNamedParser("bblfsh-sdk-tools", flags.Default)
	parser.AddCommand("manifest", cmd.ManifestCommandDescription, "", &cmd.ManifestCommand{})

	if _, err := parser.Parse(); err != nil {
		if _, ok := err.(*flags.Error); ok {
			parser.WriteHelp(os.Stdout)
			fmt.Printf("\nBuild information\n  commit: %s\n  date:%s\n", version, build)
		}

		os.Exit(1)
	}
}
