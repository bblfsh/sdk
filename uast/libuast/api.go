package main

/*
#include "uast.h"
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"unsafe"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes/nodesproto"
)

func main() {}

func newUast(iface C.NodeIface, h Handle) *C.Uast {
	sz := unsafe.Sizeof(C.Uast{})
	u := (*C.Uast)(C.malloc(C.size_t(sz)))
	u.iface = iface
	u.handle = C.uint64_t(h)
	u.root = 0
	u.ctx = 0
	return u
}

//export UastNew
// Uast needs a node implementation in order to work. This is needed
// because the data structure of the node itself is not defined by this
// library, instead it provides an interface that is expected to be satisfied by
// the binding providers.
//
// This architecture allows libuast to work with every language's native node
// data structures.
//
// Returns NULL and sets LastError if the Uast couldn't initialize.
func UastNew(iface C.NodeIface, ctx C.UastHandle) *C.Uast {
	c := newContext(&cNodes{
		impl: &iface,
		ctx:  ctx,
	})

	u := newUast(iface, c.Handle())
	u.ctx = ctx
	return u
}

//export UastDecode
func UastDecode(p unsafe.Pointer, sz C.size_t) *C.Uast {
	data := C.GoBytes(p, C.int(sz))

	nd := &goNodes{}

	c := newContext(nd)
	u := newUast(goImpl, c.Handle())

	n, err := nodesproto.ReadTree(bytes.NewReader(data))
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

//export UastFree
// Releases Uast resources.
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

//export LastError
// Return last encountered error, if any.
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
func UastFilter(ctx *C.Uast, node C.NodeHandle, query *C.char) C.NodeHandle {
	c := getContextFrom(ctx)
	if c == nil {
		return 0
	}
	if node == 0 {
		node = ctx.root
	}
	h := Handle(node)

	root := c.impl.AsNode(h)

	qu := C.GoString(query)
	res, _ := c.Filter(root, qu)
	if res == nil {
		return 0
	}
	return C.NodeHandle(res.Handle())
}

//export UastIteratorNew
func UastIteratorNew(ctx *C.Uast, node C.NodeHandle, order C.TreeOrder) *C.UastIterator {
	panic("not implemented") // FIXME
}

//export UastIteratorFree
func UastIteratorFree(it *C.UastIterator) {
	panic("not implemented") // FIXME
}

//export UastIteratorNext
func UastIteratorNext(it *C.UastIterator) C.NodeHandle {
	panic("not implemented") // FIXME
}
