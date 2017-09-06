package cmd

import (
	"path/filepath"

	"gopkg.in/bblfsh/sdk.v0/assets/build"
	"gopkg.in/bblfsh/sdk.v0/cli"
)

const sdkPath = ".sdk"

const PrepareBuildCommandDescription = "installs locally the build system for a driver"

type PrepareBuildCommand struct {
	cli.Command
}

func (c *PrepareBuildCommand) Execute(args []string) error {
	sdkFolder := filepath.Join(c.Root, sdkPath)
	return build.RestoreAssets(sdkFolder, "")
}
