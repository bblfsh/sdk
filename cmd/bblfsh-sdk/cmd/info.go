package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/bblfsh/sdk/v3/cmd"
	"github.com/bblfsh/sdk/v3/driver/manifest"
)

const InfoCommandDescription = "prints info about the driver"

type InfoCommand struct {
	cmd.Command
}

func (c *InfoCommand) Execute(args []string) error {
	m, err := manifest.Load(filepath.Join(c.Root, manifest.Filename))
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}
