package main

/*
#include "uast.h"
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"unsafe"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes/nodesproto"
	"gopkg.in/bblfsh/sdk.v2/uast/yaml"
)

func main() {}

func newUast(iface *C.NodeIface, h Handle) *C.Uast {
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
	c := newContext(&cNodes{
		impl: iface,
		ctx:  ctx,
	})

	u := newUast(iface, c.Handle())
	u.ctx = ctx
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
	u := newUast(goImpl, c.Handle())

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
