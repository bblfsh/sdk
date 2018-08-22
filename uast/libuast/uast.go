package main

import (
	"sync"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/query"
	"gopkg.in/bblfsh/sdk.v2/uast/query/xpath"
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

	xpath query.Interface

	lasth Handle
	iters map[Handle]*Iterator
}

func (c *Context) next() Handle {
	c.lasth++
	h := c.lasth
	return h
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
	c.iters = nil
}

type Iterator struct {
	c  *Context
	h  Handle
	it query.Iterator
}

func (it *Iterator) Handle() Handle {
	if it == nil {
		return 0
	}
	return it.h
}

func (it *Iterator) Next() Node {
	if it == nil || it.it == nil {
		return nil
	}
	if !it.it.Next() {
		return nil
	}
	return it.c.toNode(it.it.Node())
}

func (it *Iterator) Close() error {
	if it == nil || it.c == nil {
		return nil
	}
	delete(it.c.iters, it.h)
	it.it = nil
	return nil
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

func (c *Context) AsIterator(h Handle) *Iterator {
	return c.iters[h]
}

func (c *Context) NewIterator(it query.Iterator) *Iterator {
	h := c.next()
	cit := &Iterator{c: c, h: h, it: it}
	if c.iters == nil {
		c.iters = make(map[Handle]*Iterator)
	}
	c.iters[h] = cit
	return cit
}

func (c *Context) Iterate(root Node, order query.IterOrder) *Iterator {
	it := query.NewIterator(root, order)
	return c.NewIterator(it)
}

func (c *Context) Filter(root Node, query string) (*Iterator, error) {
	if c.xpath == nil {
		c.xpath = xpath.New()
	}
	it, err := c.xpath.Execute(root, query)
	if err != nil {
		c.setError(err)
		return nil, err
	}
	return c.NewIterator(it), nil
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
