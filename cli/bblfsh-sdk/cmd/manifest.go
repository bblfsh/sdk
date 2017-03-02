package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bblfsh/sdk/manifest"
)

type ManifestCommand struct {
	command
}

func (c *ManifestCommand) Execute(args []string) error {
	m, err := c.readManifest()
	if err != nil {
		return err
	}

	c.processManifest(m)
	return nil
}

func (c *ManifestCommand) readManifest() (*manifest.Manifest, error) {
	return manifest.Load(filepath.Join(c.Root, manifest.Filename))
}

func (c *ManifestCommand) processManifest(m *manifest.Manifest) {
	c.processValue("LANGUAGE", m.Language)
	c.processValue("RUNTIME_OS", string(m.Runtime.OS))

	nv := strings.Join(m.Runtime.NativeVersion, ":")
	c.processValue("RUNTIME_NATIVE_VERSION", nv)
	c.processValue("RUNTIME_GO_VERSION", m.Runtime.GoVersion)
}

func (c *ManifestCommand) processValue(key, value string) error {
	_, err := fmt.Printf("%s=%s\n", key, value)
	return err
}
