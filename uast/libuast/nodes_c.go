package main

/*
#include "uast.h"
#include <stdlib.h>

// Start of Go helpers
NodeKind callKind(NodeIface* iface, Uast* ctx, NodeHandle node);

const char * callAsString(const NodeIface* iface, Uast* ctx, NodeHandle node);
int64_t      callAsInt(const NodeIface* iface, Uast* ctx, NodeHandle node);
uint64_t     callAsUint(const NodeIface* iface, Uast* ctx, NodeHandle node);
double       callAsFloat(const NodeIface* iface, Uast* ctx, NodeHandle node);
bool         callAsBool(const NodeIface* iface, Uast* ctx, NodeHandle node);

size_t       callSize(const NodeIface* iface, Uast* ctx, NodeHandle node);
const char * callKeyAt(const NodeIface* iface, Uast* ctx, NodeHandle node, size_t i);
NodeHandle   callValueAt(const NodeIface* iface, Uast* ctx, NodeHandle node, size_t i);

NodeHandle callNewObject(const NodeIface* iface, Uast* ctx, size_t size);
NodeHandle callNewArray(const NodeIface* iface, Uast* ctx, size_t size);
NodeHandle callNewString(const NodeIface* iface, Uast* ctx, const char * v);
NodeHandle callNewInt(const NodeIface* iface, Uast* ctx, int64_t v);
NodeHandle callNewUint(const NodeIface* iface, Uast* ctx, uint64_t v);
NodeHandle callNewFloat(const NodeIface* iface, Uast* ctx, double v);
NodeHandle callNewBool(const NodeIface* iface, Uast* ctx, bool v);

void callSetValue(const NodeIface* iface, Uast* ctx, NodeHandle node, size_t i, NodeHandle v);
void callSetKeyValue(const NodeIface* iface, Uast* ctx, NodeHandle node, const char * k, NodeHandle v);
// End of Go helpers
*/
import "C"

import (
	"fmt"
	"sort"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

var _ NodeIface = (*cNodes)(nil)

// cNodes implements a Go nodes interface for node provided by libuast clients.
// In this case Go side only passes opaque handles to objects and calls a C interface implementation to access nodes.
type cNodes struct {
	impl *C.NodeIface
	ctx  *C.Uast
}

func (c *cNodes) Free() {
	// noting to do - nodes are managed by client code
}

var kindToGo = map[C.NodeKind]nodes.Kind{
	C.NODE_NULL:   nodes.KindNil,
	C.NODE_OBJECT: nodes.KindObject,
	C.NODE_ARRAY:  nodes.KindArray,
	C.NODE_STRING: nodes.KindString,
	C.NODE_INT:    nodes.KindInt,
	C.NODE_UINT:   nodes.KindUint,
	C.NODE_FLOAT:  nodes.KindFloat,
	C.NODE_BOOL:   nodes.KindBool,
}

func (c *cNodes) asNode(h C.NodeHandle) Node {
	if h == 0 {
		return nil
	}
	ckind := C.callKind(c.impl, c.ctx, h)
	kind, ok := kindToGo[ckind]
	if !ok || kind == nodes.KindNil {
		return nil
	}
	switch kind {
	case nodes.KindObject:
		return &cObject{c: c, h: h}
	case nodes.KindArray:
		return &cArray{c: c, h: h, sz: -1}
	}
	return &cValue{c: c, h: h, kind: kind}
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

var _ Object = (*cObject)(nil)

// offsSort sorts keys while updating an offsets array accordingly.
type offsSort struct {
	keys []string // sorted keys
	ind  []int    // original indexes
}

func (arr offsSort) Len() int {
	return len(arr.keys)
}

func (arr offsSort) Less(i, j int) bool {
	return arr.keys[i] < arr.keys[j]
}

func (arr offsSort) Swap(i, j int) {
	arr.keys[i], arr.keys[j] = arr.keys[j], arr.keys[i]
	arr.ind[i], arr.ind[j] = arr.ind[j], arr.ind[i]
}

type cObject struct {
	c    *cNodes
	h    C.NodeHandle
	keys []string // sorted keys
	ind  []int    // original indexes
}

func (n *cObject) Handle() Handle {
	return Handle(n.h)
}

func (n *cObject) Kind() nodes.Kind {
	return nodes.KindObject
}

func (n *cObject) Value() nodes.Value {
	return nil
}

func (n *cObject) size() int {
	v := C.callSize(n.c.impl, n.c.ctx, n.h)
	return int(v)
}
func (n *cObject) Size() int {
	if n.keys != nil {
		return len(n.keys)
	}
	return n.size()
}

func (n *cObject) keyAt(i int) string {
	cstr := C.callKeyAt(n.c.impl, n.c.ctx, n.h, C.size_t(i))
	s := C.GoString(cstr)
	//C.free(unsafe.Pointer(cstr))
	return s
}

func (n *cObject) Keys() []string {
	if n.keys != nil {
		return n.keys
	}
	sz := n.size()
	n.keys = make([]string, sz)
	for i := 0; i < sz; i++ {
		k := n.keyAt(i)
		n.keys[i] = k
	}
	if !sort.StringsAreSorted(n.keys) {
		n.ind = make([]int, len(n.keys))
		for i := range n.ind {
			n.ind[i] = i
		}
		sort.Sort(offsSort{
			keys: n.keys, ind: n.ind,
		})
	}
	return n.keys
}

func (n *cObject) valueAt(i int) Node {
	if i < 0 || i >= len(n.keys) {
		return nil
	}
	if n.ind != nil {
		i = n.ind[i]
	}
	v := C.callValueAt(n.c.impl, n.c.ctx, n.h, C.size_t(i))
	return n.c.asNode(v)
}

func (n *cObject) ValueAt(key string) (nodes.External, bool) {
	for i, k := range n.Keys() {
		if k == key {
			v := n.valueAt(i)
			return v, true
		}
	}
	return nil, false
}

var _ Array = (*cArray)(nil)

type cArray struct {
	c  *cNodes
	h  C.NodeHandle
	sz int // cached size, -1 means it was not yet cached
}

func (n *cArray) Handle() Handle {
	return Handle(n.h)
}

func (n *cArray) Kind() nodes.Kind {
	return nodes.KindArray
}

func (n *cArray) Value() nodes.Value {
	return nil
}

func (n *cArray) Size() int {
	if n.sz < 0 {
		// cache the size
		sz := C.callSize(n.c.impl, n.c.ctx, n.h)
		n.sz = int(sz)
	}
	return n.sz
}

func (n *cArray) valueAt(i int) Node {
	v := C.callValueAt(n.c.impl, n.c.ctx, n.h, C.size_t(i))
	return n.c.asNode(v)
}

func (n *cArray) ValueAt(i int) nodes.External {
	return n.valueAt(i)
}

var _ Node = (*cValue)(nil)

type cValue struct {
	c    *cNodes
	h    C.NodeHandle
	kind nodes.Kind  // cached kind
	val  nodes.Value // cached Go value
}

func (n *cValue) Handle() Handle {
	return Handle(n.h)
}

func (n *cValue) Kind() nodes.Kind {
	return n.kind
}

func (n *cValue) Value() nodes.Value {
	if n.val != nil {
		return n.val
	}
	switch n.kind {
	case nodes.KindString:
		cstr := C.callAsString(n.c.impl, n.c.ctx, n.h)
		s := C.GoString(cstr)
		//C.free(unsafe.Pointer(cstr))
		n.val = nodes.String(s)
	case nodes.KindInt:
		v := C.callAsInt(n.c.impl, n.c.ctx, n.h)
		n.val = nodes.Int(v)
	case nodes.KindUint:
		v := C.callAsUint(n.c.impl, n.c.ctx, n.h)
		n.val = nodes.Uint(v)
	case nodes.KindFloat:
		v := C.callAsFloat(n.c.impl, n.c.ctx, n.h)
		n.val = nodes.Float(v)
	case nodes.KindBool:
		v := C.callAsBool(n.c.impl, n.c.ctx, n.h)
		n.val = nodes.Bool(v)
	default:
		return nil
	}
	return n.val
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
