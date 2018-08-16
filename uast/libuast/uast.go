package main

import (
	"fmt"
	"sync"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/xpath"
)

type Handle uintptr

type Node interface {
	Handle() Handle
	Kind() nodes.Kind

	AsString() nodes.String
	AsInt() nodes.Int
	AsUint() nodes.Uint
	AsFloat() nodes.Float
	AsBool() nodes.Bool

	Size() int

	KeyAt(i int) string
	ValueAt(i int) Node
}

type NodeIface interface {
	Free()
	AsNode(h Handle) Node
	AsTmpNode(h Handle) TmpNode

	NewObject(sz int) Handle
	NewArray(sz int) Handle

	NewString(v string) Node
	NewInt(v int64) Node
	NewUint(v uint64) Node
	NewFloat(v float64) Node
	NewBool(v bool) Node
}

type TmpNode interface {
	SetValue(i int, v Node)
	SetKeyValue(k string, v Node)
	Build() Node
}

var (
	mu    sync.RWMutex
	last  Handle
	ctxes = make(map[Handle]*Context)
)

func newContext(impl NodeIface) *Context {
	mu.Lock()
	defer mu.Unlock()

	last++
	h := last

	ctx := &Context{
		h: h, impl: impl,
	}
	ctxes[h] = ctx
	return ctx
}

func getContext(h Handle) *Context {
	mu.RLock()
	ctx := ctxes[h]
	mu.RUnlock()
	return ctx
}

type Context struct {
	h    Handle
	last error
	impl NodeIface
}

func (c *Context) Handle() Handle {
	return c.h
}
func (c *Context) Error() error {
	return c.last
}
func (c *Context) free() {
	if c == nil {
		return
	}
	mu.Lock()
	delete(ctxes, c.h)
	mu.Unlock()
	c.impl.Free()
}

type xpathNode struct {
	Node
}

func (n xpathNode) ValueAt(i int) xpath.Node {
	v := n.Node.ValueAt(i)
	if v == nil {
		return nil
	}
	return xpathNode{v}
}

func (c *Context) setError(err error) {
	if c.last == nil {
		c.last = err
	}
}
func (c *Context) Filter(root Node, query string) (Node, error) {
	ind := xpath.Index(xpathNode{root})
	it, err := ind.Filter(query)
	if err != nil {
		c.setError(err)
		return nil, err
	}
	var nodes []Node
	for it.Next() {
		n := it.Node()
		if n == nil {
			nodes = append(nodes, nil)
		} else {
			nodes = append(nodes, n.(xpathNode).Node)
		}
	}
	// TODO: it can be a single Bool node, for example
	res := c.impl.NewArray(len(nodes))
	tmp := c.impl.AsTmpNode(res)
	if tmp == nil {
		err = fmt.Errorf("cannot create a result node")
		c.setError(err)
		return nil, err
	}
	for i, v := range nodes {
		tmp.SetValue(i, v)
	}
	return tmp.Build(), nil
}
