package manifest

import (
	"io"

	"github.com/BurntSushi/toml"
)

const Filename = "manifest.toml"

type DevelopmentStatus string

const (
	Planning DevelopmentStatus = "planning"
	PreAlpha DevelopmentStatus = "pre-alpha"
	Alpha    DevelopmentStatus = "alpha"
	Beta     DevelopmentStatus = "beta"
	Stable   DevelopmentStatus = "stable"
	Mature   DevelopmentStatus = "mature"
	Inactive DevelopmentStatus = "inactive"
)

type Manifest struct {
	Language string            `toml:"language,omitempty"`
	Version  string            `toml:"version,omitempty"`
	Status   DevelopmentStatus `toml:"status"`
	Runtime  struct {
		NativeVersion []string `toml:"native_version,omitempty"`
		GoVersion     string   `toml:"go_version,omitempty"`
	} `toml:"runtime,omitempty"`
}

func (m *Manifest) Encode(w io.Writer) error {
	e := toml.NewEncoder(w)
	return e.Encode(m)
}

func (m *Manifest) Decode(r io.Reader) error {
	if _, err := toml.DecodeReader(r, m); err != nil {
		return err
	}

	return nil
}
