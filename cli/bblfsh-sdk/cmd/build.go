package cmd

import (
	"path/filepath"

	"github.com/bblfsh/sdk/assets/build"
)

const sdkPath = ".sdk"

type PrepareBuildCommand struct {
	Args struct {
		Root string `positional-arg-name:"project-root" default:"."`
	} `positional-args:"yes"`
}

func (c *PrepareBuildCommand) Execute(args []string) error {
	sdkFolder := filepath.Join(c.Args.Root, sdkPath)
	return build.RestoreAssets(sdkFolder, "")
}
