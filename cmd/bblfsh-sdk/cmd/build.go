package cmd

import (
	"path/filepath"

	"gopkg.in/bblfsh/sdk.v2/assets/build"
	"gopkg.in/bblfsh/sdk.v2/cmd"
)

const sdkPath = ".sdk"

const PrepareBuildCommandDescription = "installs locally the build system for a driver"

type PrepareBuildCommand struct {
	cmd.Command
}

func (c *PrepareBuildCommand) Execute(args []string) error {
	sdkFolder := filepath.Join(c.Root, sdkPath)
	return build.RestoreAssets(sdkFolder, "")
}
