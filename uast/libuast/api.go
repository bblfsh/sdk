package main

/*
#include "uast.h"
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"unsafe"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes/nodesproto"
	"gopkg.in/bblfsh/sdk.v2/uast/query"
	"gopkg.in/bblfsh/sdk.v2/uast/yaml"
)

func main() {}

func newUAST(iface *C.NodeIface, h Handle) *C.Uast {
	sz := unsafe.Sizeof(C.Uast{})
	u := (*C.Uast)(C.malloc(C.size_t(sz)))
	u.iface = iface
	u.handle = C.uint64_t(h)
	u.root = 0
	u.ctx = 0
	return u
}

//export UastNew
// UastNew initializes a new UAST context that will use a provided node interface as an implementation.
// This allows libuast to work with every language's native node data structures. Client can pass
// an additional UastHandle to distinguish between different UAST context instances.
//
// The returned context pointer is guaranteed to be not NULL. Client should check LastError before
// using the context and deallocate it with UastFree in case of an error occurs, or when the context
// is no longer needed.
func UastNew(iface *C.NodeIface, ctx C.UastHandle) *C.Uast {
	cn := &cNodes{
		impl: iface,
	}
	c := newContext(cn)

	u := newUAST(iface, c.Handle())
	u.ctx = ctx
	cn.ctx = u
	return u
}

//export UastDecode
// UastDecode accepts a pointer to a buffer with a specified size and decodes the content into
// a new UAST structure.
//
// The new UAST context will use internal node interface implementation and all the nodes will
// be managed by libuast.
//
// The returned context pointer is guaranteed to be not NULL. Client should check LastError before
// using the context and deallocate it with UastFree in case of an error occurs, or when the context
// is no longer needed.
func UastDecode(p unsafe.Pointer, sz C.size_t, format C.UastFormat) *C.Uast {
	if format == 0 {
		format = C.UAST_BINARY
	}
	data := C.GoBytes(p, C.int(sz))

	nd := &goNodes{}

	c := newContext(nd)
	u := newUAST(goImpl, c.Handle())

	var (
		n   nodes.Node
		err error
	)
	switch format {
	case C.UAST_BINARY:
		n, err = nodesproto.ReadTree(bytes.NewReader(data))
	case C.UAST_YAML:
		n, err = uastyml.Unmarshal(data)
	default:
		err = fmt.Errorf("unknown format: %v", format)
	}
	if err != nil {
		c.last = err
		return u
	}
	if n != nil {
		u.root = C.NodeHandle(nd.toNode(n).Handle())
	}
	u.ctx = C.UastHandle(c.Handle())
	return u
}

//export UastEncode
// UastEncode encodes a given node and all it's children to a specified encoding. It returns a pointer to the bytes buffer
// and writes the size of the buffer to the size pointer.
func UastEncode(ctx *C.Uast, node C.NodeHandle, size *C.size_t, format C.UastFormat) unsafe.Pointer {
	if format == 0 {
		format = C.UAST_BINARY
	}
	c := getContextFrom(ctx)
	if c == nil {
		return nil
	}
	if node == 0 {
		node = ctx.root
	}

	h := Handle(node)
	root := c.impl.AsNode(h)
	n, err := loadNode(root)
	if err != nil {
		c.setError(err)
		return nil
	}

	buf := bytes.NewBuffer(nil)
	switch format {
	case C.UAST_BINARY:
		err = nodesproto.WriteTo(buf, n)
	case C.UAST_YAML:
		err = uastyml.NewEncoder(buf).Encode(n)
	default:
		err = fmt.Errorf("unknown format: %v", format)
	}
	if err != nil {
		c.setError(err)
		return nil
	}
	sz := buf.Len()
	if size != nil {
		*size = C.size_t(sz)
	}
	return C.CBytes(buf.Bytes())
}

//export UastFree
// UastFree releases resources associated with this UAST context.
func UastFree(ctx *C.Uast) {
	if ctx == nil {
		return
	}
	h := Handle(ctx.handle)
	C.free(unsafe.Pointer(ctx))
	getContext(h).free()
}

func getContextFrom(p *C.Uast) *Context {
	if p == nil {
		return nil
	}
	return getContext(Handle(p.handle))
}

func setContextError(h *C.Uast, err error) {
	c := getContextFrom(h)
	if c == nil {
		panic(err)
	}
	c.setError(err)
}

//export LastError
// LastError return the last encountered error in this context, if any.
func LastError(ctx *C.Uast) *C.char {
	c := getContextFrom(ctx)
	if c == nil {
		return nil
	}
	err := c.Error()
	if err == nil {
		return nil
	}
	return C.CString(err.Error())
}

//export UastFilter
// UastFilter filters UAST starting from a given node with XPath query and returns an iterator for results.
// Caller should free an iterator by calling UastIteratorFree.
func UastFilter(ctx *C.Uast, node C.NodeHandle, query *C.char) *C.UastIterator {
	c := getContextFrom(ctx)
	if c == nil {
		return nil
	}
	if node == 0 {
		node = ctx.root
	}
	h := Handle(node)

	root := c.impl.AsNode(h)

	qu := C.GoString(query)
	res, err := c.Filter(root, qu)
	if err != nil {
		c.setError(err)
		return nil
	} else if res == nil {
		return nil
	}
	return newIterator(ctx, res.Handle())
}

//export SetError
// SetError sets an error state for this context. It can be used inside interface functions to indicate an error to the caller.
func SetError(ctx *C.Uast, str *C.char) {
	c := getContextFrom(ctx)
	if c == nil {
		return
	}
	s := C.GoString(str)
	c.setError(errors.New(s))
}

var toOrder = map[C.TreeOrder]query.IterOrder{
	C.PRE_ORDER:      query.PreOrder,
	C.POST_ORDER:     query.PostOrder,
	C.LEVEL_ORDER:    query.LevelOrder,
	C.POSITION_ORDER: query.PositionOrder,
}

func newIterator(ctx *C.Uast, h Handle) *C.UastIterator {
	sz := unsafe.Sizeof(C.UastIterator{})
	it := (*C.UastIterator)(C.malloc(C.size_t(sz)))
	it.ctx = ctx
	it.handle = C.uintptr_t(h)
	return it
}

//export UastIteratorNew
// UastIteratorNew starts an iterator from a given node in a specified order. Caller should free an iterator by calling UastIteratorFree.
func UastIteratorNew(ctx *C.Uast, node C.NodeHandle, order C.TreeOrder) *C.UastIterator {
	ord, ok := toOrder[order]
	if !ok {
		ord = -1
	}
	c := getContextFrom(ctx)
	if c == nil {
		return nil
	}
	root := c.impl.AsNode(Handle(node))
	it := c.Iterate(root, ord)
	return newIterator(ctx, it.h)
}

//export UastIteratorFree
// UastIteratorFree release all resources associated with an iterator.
func UastIteratorFree(it *C.UastIterator) {
	if it == nil {
		return
	}
	c := getContextFrom(it.ctx)
	if c != nil {
		h := Handle(it.handle)
		c.AsIterator(h).Close()
	}
	C.free(unsafe.Pointer(it))
}

//export UastIteratorNext
// UastIteratorNext return the next node on this iterator or 0 if iterator is empty.
func UastIteratorNext(it *C.UastIterator) C.NodeHandle {
	if it == nil {
		return 0
	}
	c := getContextFrom(it.ctx)
	if c == nil {
		return 0
	}
	n := c.AsIterator(Handle(it.handle)).Next()
	if n == nil {
		return 0
	}
	return C.NodeHandle(n.Handle())
}
