package xpath

import (
	"github.com/antchfx/xpath"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/query"
)

func isValue(n nodes.External) bool {
	if n == nil {
		return false
	}
	k := n.Kind()
	return k.In(nodes.KindsValues)
}

func New() query.Interface {
	return &index{}
}

type index struct{}

func (t *index) newNavigator(n nodes.External) xpath.NodeNavigator {
	// TODO: zero copy
	d := conv(n)
	return newNavigator(d)
}

func (t *index) Prepare(query string) (query.Query, error) {
	exp, err := xpath.Compile(query)
	if err != nil {
		return nil, err
	}
	return &xQuery{idx: t, exp: exp}, nil
}

func (t *index) Execute(root nodes.External, query string) (query.Iterator, error) {
	q, err := t.Prepare(query)
	if err != nil {
		return nil, err
	}
	return q.Execute(root)
}

type xQuery struct {
	idx *index
	exp *xpath.Expr
}

func (q *xQuery) Execute(root nodes.External) (query.Iterator, error) {
	nav := q.idx.newNavigator(root)
	it := q.exp.Select(nav)
	return &iterator{it: it}, nil
}

type iterator struct {
	it *xpath.NodeIterator
}

func (it *iterator) Next() bool {
	return it.it.MoveNext()
}

func (it *iterator) Node() nodes.External {
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
