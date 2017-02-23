package cmd

import (
	"path/filepath"

	"github.com/bblfsh/sdk/assets/build"
)

type BuildSystemCommand struct {
	Args struct {
		Root string `positional-arg-name:"project-root" default:"."`
	} `positional-args:"yes"`
}

func (c *BuildSystemCommand) Execute(args []string) error {
	sdkFolder := filepath.Join(c.Args.Root, ".sdk")
	return build.RestoreAssets(sdkFolder, "")
}
