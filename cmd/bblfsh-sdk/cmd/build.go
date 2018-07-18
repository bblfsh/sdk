package cmd

import (
	"fmt"
	"path/filepath"

	assets "gopkg.in/bblfsh/sdk.v2/assets/build"
	"gopkg.in/bblfsh/sdk.v2/build"
	"gopkg.in/bblfsh/sdk.v2/cmd"
)

const sdkPath = ".sdk"

const PrepareBuildCommandDescription = "installs locally the build system for a driver"

type PrepareBuildCommand struct {
	cmd.Command
}

func (c *PrepareBuildCommand) Execute(args []string) error {
	sdkFolder := filepath.Join(c.Root, sdkPath)
	return assets.RestoreAssets(sdkFolder, "")
}

const PrepareCommandDescription = "installs locally the build system for a driver"

type PrepareCommand struct {
	cmd.Command
}

func (c *PrepareCommand) Execute(args []string) error {
	d, err := build.NewDriver(c.Root)
	if err != nil {
		return err
	}
	return d.Prepare()
}

const BuildCommandDescription = "builds the driver"

type BuildCommand struct {
	cmd.Command
}

func (c *BuildCommand) Execute(args []string) error {
	name := ""
	if len(args) != 0 {
		name = args[0]
	}
	d, err := build.NewDriver(c.Root)
	if err != nil {
		return err
	}
	id, err := d.Build(name)
	if err != nil {
		return err
	}
	fmt.Println(id)
	return nil
}

const TestCommandDescription = "tests the driver using fixtures"

type TestCommand struct {
	cmd.Command
}

func (c *TestCommand) Execute(args []string) error {
	d, err := build.NewDriver(c.Root)
	if err != nil {
		return err
	}
	return d.Test()
}
