package main

/*
#include "uast.h"
#include <stdlib.h>

// Start of Go helpers
NodeKind callKind(NodeIface* iface, UastHandle ctx, NodeHandle node);

const char * callAsString(const NodeIface* iface, UastHandle ctx, NodeHandle node);
int64_t      callAsInt(const NodeIface* iface, UastHandle ctx, NodeHandle node);
uint64_t     callAsUint(const NodeIface* iface, UastHandle ctx, NodeHandle node);
double       callAsFloat(const NodeIface* iface, UastHandle ctx, NodeHandle node);
bool         callAsBool(const NodeIface* iface, UastHandle ctx, NodeHandle node);

size_t       callSize(const NodeIface* iface, UastHandle ctx, NodeHandle node);
const char * callKeyAt(const NodeIface* iface, UastHandle ctx, NodeHandle node, size_t i);
NodeHandle   callValueAt(const NodeIface* iface, UastHandle ctx, NodeHandle node, size_t i);

NodeHandle callNewObject(const NodeIface* iface, UastHandle ctx, size_t size);
NodeHandle callNewArray(const NodeIface* iface, UastHandle ctx, size_t size);
NodeHandle callNewString(const NodeIface* iface, UastHandle ctx, const char * v);
NodeHandle callNewInt(const NodeIface* iface, UastHandle ctx, int64_t v);
NodeHandle callNewUint(const NodeIface* iface, UastHandle ctx, uint64_t v);
NodeHandle callNewFloat(const NodeIface* iface, UastHandle ctx, double v);
NodeHandle callNewBool(const NodeIface* iface, UastHandle ctx, bool v);

void callSetValue(const NodeIface* iface, UastHandle ctx, NodeHandle node, size_t i, NodeHandle v);
void callSetKeyValue(const NodeIface* iface, UastHandle ctx, NodeHandle node, const char * k, NodeHandle v);
// End of Go helpers
*/
import "C"

import (
	"fmt"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

var _ NodeIface = (*cNodes)(nil)

type cNodes struct {
	impl *C.NodeIface // allocated by Go
	ctx  C.UastHandle
}

func (c *cNodes) Free() {
	// noting to do - nodes are managed by client code
}

func (c *cNodes) asNode(h C.NodeHandle) Node {
	if h == 0 {
		return nil
	}
	return &cNode{c: c, h: h}
}
func (c *cNodes) AsNode(h Handle) Node {
	return c.asNode(C.NodeHandle(h))
}
func (c *cNodes) AsTmpNode(h Handle) TmpNode {
	if h == 0 {
		return nil
	}
	return &cTmpNode{c: c, h: C.NodeHandle(h)}
}

func (c *cNodes) NewObject(sz int) Handle {
	h := C.callNewObject(c.impl, c.ctx, C.size_t(sz))
	return Handle(h)
}

func (c *cNodes) NewArray(sz int) Handle {
	h := C.callNewArray(c.impl, c.ctx, C.size_t(sz))
	return Handle(h)
}

func (c *cNodes) NewValue(v nodes.Value) Node {
	if v == nil {
		return nil
	}
	var n C.NodeHandle
	switch v := v.(type) {
	case nodes.String:
		n = C.callNewString(c.impl, c.ctx, C.CString(string(v)))
	case nodes.Int:
		n = C.callNewInt(c.impl, c.ctx, C.int64_t(v))
	case nodes.Uint:
		n = C.callNewUint(c.impl, c.ctx, C.uint64_t(v))
	case nodes.Float:
		n = C.callNewFloat(c.impl, c.ctx, C.double(v))
	case nodes.Bool:
		n = C.callNewBool(c.impl, c.ctx, C.bool(v))
	default:
		panic(fmt.Errorf("unknown value type: %T", v))
	}
	return c.asNode(n)
}

var _ Node = (*cNode)(nil)

type cNode struct {
	c *cNodes
	h C.NodeHandle
}

func (n *cNode) Handle() Handle {
	return Handle(n.h)
}

func (n *cNode) kind() C.NodeKind {
	return C.callKind(n.c.impl, n.c.ctx, n.h)
}
func (n *cNode) Kind() nodes.Kind {
	switch n.kind() {
	case C.NODE_NULL:
		return nodes.KindNil
	case C.NODE_OBJECT:
		return nodes.KindObject
	case C.NODE_ARRAY:
		return nodes.KindArray
	case C.NODE_STRING:
		return nodes.KindString
	case C.NODE_INT:
		return nodes.KindInt
	case C.NODE_UINT:
		return nodes.KindUint
	case C.NODE_FLOAT:
		return nodes.KindFloat
	case C.NODE_BOOL:
		return nodes.KindBool
	default:
		return nodes.KindNil
	}
}

func (n *cNode) AsValue() nodes.Value {
	switch n.kind() {
	case C.NODE_STRING:
		cstr := C.callAsString(n.c.impl, n.c.ctx, n.h)
		s := C.GoString(cstr)
		//C.free(unsafe.Pointer(cstr))
		return nodes.String(s)
	case C.NODE_INT:
		v := C.callAsInt(n.c.impl, n.c.ctx, n.h)
		return nodes.Int(v)
	case C.NODE_UINT:
		v := C.callAsUint(n.c.impl, n.c.ctx, n.h)
		return nodes.Uint(v)
	case C.NODE_FLOAT:
		v := C.callAsFloat(n.c.impl, n.c.ctx, n.h)
		return nodes.Float(v)
	case C.NODE_BOOL:
		v := C.callAsBool(n.c.impl, n.c.ctx, n.h)
		return nodes.Bool(v)
	default:
		return nil
	}
}

func (n *cNode) Size() int {
	v := C.callSize(n.c.impl, n.c.ctx, n.h)
	return int(v)
}

func (n *cNode) KeyAt(i int) string {
	cstr := C.callKeyAt(n.c.impl, n.c.ctx, n.h, C.size_t(i))
	s := C.GoString(cstr)
	//C.free(unsafe.Pointer(cstr))
	return s
}

func (n *cNode) ValueAt(i int) Node {
	v := C.callValueAt(n.c.impl, n.c.ctx, n.h, C.size_t(i))
	return n.c.asNode(v)
}

type cTmpNode struct {
	c *cNodes
	h C.NodeHandle
}

func (n *cTmpNode) SetValue(i int, v Node) {
	var h C.NodeHandle
	if v != nil {
		h = C.NodeHandle(v.Handle())
	}
	C.callSetValue(n.c.impl, n.c.ctx, n.h, C.size_t(i), h)
}

func (n *cTmpNode) SetKeyValue(k string, v Node) {
	var h C.NodeHandle
	if v != nil {
		h = C.NodeHandle(v.Handle())
	}
	C.callSetKeyValue(n.c.impl, n.c.ctx, n.h, C.CString(k), h)
}

func (n *cTmpNode) Build() Node {
	return n.c.asNode(n.h)
}
