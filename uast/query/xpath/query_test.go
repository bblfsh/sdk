package xpath

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/query"
)

func toNode(o interface{}) nodes.Node {
	n, err := uast.ToNode(o)
	if err != nil {
		panic(err)
	}
	return n
}

func TestFilter(t *testing.T) {
	var root = nodes.Array{
		toNode(uast.Identifier{Name: "Foo"}),
	}

	idx := New()
	it, err := idx.Execute(root, "//uast:Identifier[Name='Foo']")
	require.NoError(t, err)
	expect(t, it, root[0])

	it, err = idx.Execute(root, "//Identifier")
	require.NoError(t, err)
	expect(t, it)
}

func expect(t testing.TB, it query.Iterator, exp ...nodes.Node) {
	var out []nodes.Node
	for it.Next() {
		out = append(out, it.Node().(nodes.Node))
	}
	require.Equal(t, exp, out)
}
