package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bblfsh/sdk/manifest"
)

type ManifestCommand struct {
	SetEnv   bool `long:"set-env" description:"sets the variables in the environment"`
	PrintEnv bool `long:"print-env" description:"prints the manifest as env variables"`
	Args     struct {
		Root string `positional-arg-name:"project-root" default:"."`
	} `positional-args:"yes"`
}

func (c *ManifestCommand) Execute(args []string) error {
	m, err := c.readManifest()
	if err != nil {
		return err
	}

	return c.processManifest(m)
}

func (c *ManifestCommand) readManifest() (*manifest.Manifest, error) {
	return manifest.Load(filepath.Join(c.Args.Root, manifest.Filename))
}

func (c *ManifestCommand) processManifest(m *manifest.Manifest) error {
	if err := c.processValue("LANGUAGE", m.Language); err != nil {
		return err
	}

	nv := strings.Join(m.Runtime.NativeVersion, ":")
	if err := c.processValue("RUNTIME_NATIVE_VERSION", nv); err != nil {
		return err
	}

	if err := c.processValue("RUNTIME_GO_VERSION", m.Runtime.GoVersion); err != nil {
		return err
	}

	return nil
}

func (c *ManifestCommand) processValue(key, value string) error {
	if c.PrintEnv {
		fmt.Printf("%s=%s\n", key, value)
	}

	if c.SetEnv {
		return os.Setenv(key, value)
	}

	return nil
}
