package uast

import (
	"testing"

	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/stretchr/testify/require"
)

func TestAllImportPaths(t *testing.T) {
	root := nodes.Array{
		toNode(RuntimeImport{
			Path: Identifier{Name: "a"},
		}),
		toNode(InlineImport{
			Path: String{Value: "a"},
		}),
		toNode(RuntimeReImport{
			Path: QualifiedIdentifier{Names: []Identifier{
				{Name: "a"},
				{Name: "b"},
			}},
		}),
		toNode(Import{
			Path: Alias{
				Node: Identifier{Name: "c"},
			},
		}),
	}
	paths := AllImportPaths(root)
	require.Equal(t, []string{"a", "a/b", "c"}, paths)
}
