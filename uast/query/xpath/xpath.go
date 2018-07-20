package xpath

import (
	"github.com/antchfx/xpath"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

func Index(n nodes.Node) *Tree {
	return &Tree{doc: conv(n)}
}

type Tree struct {
	doc *node
}

func (t *Tree) Filter(query string) (*Iterator, error) {
	exp, err := xpath.Compile(query)
	if err != nil {
		return nil, err
	}
	it := exp.Select(newNavigator(t.doc))
	return &Iterator{it: it}, nil
}

type Iterator struct {
	it *xpath.NodeIterator
}

func (it *Iterator) Next() bool {
	return it.it.MoveNext()
}
func (it *Iterator) Node() nodes.Node {
	c := it.it.Current()
	if c == nil {
		return nil
	}
	nav := c.(*nodeNavigator)
	if nav.cur == nil {
		return nil
	}
	return nav.cur.Node
}
