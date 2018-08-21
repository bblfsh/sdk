package main

import (
	"fmt"
	"sync"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/xpath"
)

type Handle uintptr

func IsValue(n Node) bool {
	if n == nil {
		return false
	}
	k := n.Kind()
	return k.In(nodes.KindsValues)
}

type Base interface {
	Handle() Handle
}

type Node interface {
	Base
	nodes.External
}

type Object interface {
	Base
	nodes.ExternalObject
}

type Array interface {
	Base
	nodes.ExternalArray
}

type NodeIface interface {
	Free()
	AsNode(h Handle) Node
	AsTmpNode(h Handle) TmpNode

	NewObject(sz int) Handle
	NewArray(sz int) Handle
	NewValue(v nodes.Value) Node
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

func (c *Context) setError(err error) {
	if c.last == nil {
		c.last = err
	}
}
func (c *Context) toNode(n nodes.External) Node {
	if n == nil {
		return nil
	} else if nd, ok := n.(Node); ok {
		return nd
	}
	// TODO: find a better way to convert these nodes
	return c.impl.(*goNodes).toNode(n.(nodes.Node))
}
func (c *Context) Filter(root Node, query string) (Node, error) {
	ind := xpath.Index(root)
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
			nodes = append(nodes, n.(Node))
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

func loadNode(n Node) (nodes.Node, error) {
	if n == nil {
		return nil, nil
	}
	if nd, ok := n.(Native); ok {
		return nd.Native(), nil
	}
	return nodes.ToNode(n, nil)
}
