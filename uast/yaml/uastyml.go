package uastyml

import (
	"io"

	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/uastyaml"
)

// Marshal encode the UAST to a human-readable YAML.
//
// Deprecated: use uastyaml.Marshal instead
func Marshal(n nodes.Node) ([]byte, error) {
	return uastyaml.Marshal(n)
}

// NewEncoder creates a YAML encoder for UAST.
//
// Deprecated: use uastyaml.NewEncoder instead
func NewEncoder(w io.Writer) *uastyaml.Encoder {
	return uastyaml.NewEncoder(w)
}

// Encoder is a YAML encoder for UAST.
//
// Deprecated: use uastyaml.Encoder instead
type Encoder = uastyaml.Encoder

// Unmarshal decodes YAML to a UAST.
//
// Deprecated: use uastyaml.Unmarshal instead
func Unmarshal(data []byte) (nodes.Node, error) {
	return uastyaml.Unmarshal(data)
}
