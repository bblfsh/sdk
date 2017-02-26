package cmd

import (
	"path/filepath"

	"github.com/bblfsh/sdk/assets/build"
)

const sdkPath = ".sdk"

type PrepareBuildCommand struct {
	command
}

func (c *PrepareBuildCommand) Execute(args []string) error {
	sdkFolder := filepath.Join(c.Root, sdkPath)
	return build.RestoreAssets(sdkFolder, "")
}
