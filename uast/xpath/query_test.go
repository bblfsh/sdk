package xpath

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
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

	n := &goNode{n: root}
	tree := Index(n)
	it, err := tree.Filter("//uast:Identifier[Name='Foo']")
	require.NoError(t, err)
	expect(t, it, root[0])

	it, err = tree.Filter("//Identifier")
	require.NoError(t, err)
	expect(t, it)
}

func expect(t testing.TB, it *Iterator, exp ...nodes.Node) {
	var out []nodes.Node
	for it.Next() {
		out = append(out, it.Node().(*goNode).n)
	}
	require.Equal(t, exp, out)
}

var _ Node = (*goNode)(nil)

type goNode struct {
	n    nodes.Node
	keys []string
}

func (n *goNode) Kind() nodes.Kind {
	if n == nil {
		return nodes.KindNil
	}
	return nodes.KindOf(n.n)
}

func (n *goNode) AsValue() nodes.Value {
	return n.n.(nodes.Value)
}

func (n *goNode) Size() int {
	switch v := n.n.(type) {
	case nodes.Object:
		return len(v)
	case nodes.Array:
		return len(v)
	}
	return 0
}

func (n *goNode) cacheKeys() {
	if n.keys != nil {
		return
	}
	obj := n.n.(nodes.Object)
	n.keys = obj.Keys()
}

func (n *goNode) KeyAt(i int) string {
	n.cacheKeys()
	if i < 0 || i >= len(n.keys) {
		return ""
	}
	return n.keys[i]
}

func (n *goNode) ValueAt(i int) Node {
	if arr, ok := n.n.(nodes.Array); ok {
		if i < 0 || i >= len(arr) {
			return nil
		}
		return &goNode{n: arr[i]}
	}
	n.cacheKeys()
	if i < 0 || i >= len(n.keys) {
		return nil
	}
	obj := n.n.(nodes.Object)
	v := obj[n.keys[i]]
	return &goNode{n: v}
}
