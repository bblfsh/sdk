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

type Node interface {
	Handle() Handle
	Kind() nodes.Kind

	AsValue() nodes.Value

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

func loadNode(n Node) nodes.Node {
	if n == nil {
		return nil
	}
	if nd, ok := n.(*goNode); ok {
		return nd.n
	}
	switch kind := n.Kind(); kind {
	case nodes.KindNil:
		return nil
	case nodes.KindObject:
		sz := n.Size()
		m := make(nodes.Object, sz)
		for i := 0; i < sz; i++ {
			k, v := n.KeyAt(i), n.ValueAt(i)
			m[k] = loadNode(v)
		}
		return m
	case nodes.KindArray:
		sz := n.Size()
		arr := make(nodes.Array, 0, sz)
		for i := 0; i < sz; i++ {
			v := n.ValueAt(i)
			arr = append(arr, loadNode(v))
		}
		return arr
	default:
		if IsValue(n) {
			return n.AsValue()
		}
		panic(fmt.Errorf("unknown kind: %v", kind))
	}
}
