package main

/*
#include "uast_go.h"
*/
import "C"

import "gopkg.in/bblfsh/sdk.v2/uast/nodes"

var goImpl C.NodeIface

func init() {
	goImpl = C.uastImpl()
}

func getContextGo(h C.UastHandle) *Context {
	if h == 0 {
		return nil
	}
	return getContext(Handle(h))
}

func getNodeGo(ctx C.UastHandle, node C.NodeHandle) Node {
	c := getContextGo(ctx)
	if c == nil {
		return nil
	}
	return c.impl.AsNode(Handle(node))
}

//export uastKind
func uastKind(ctx C.UastHandle, node C.NodeHandle) C.NodeKind {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return C.NODE_NULL
	}
	kind := nd.Kind()
	switch kind {
	case nodes.KindNil:
		return C.NODE_NULL
	case nodes.KindObject:
		return C.NODE_OBJECT
	case nodes.KindArray:
		return C.NODE_ARRAY
	case nodes.KindString:
		return C.NODE_STRING
	case nodes.KindInt:
		return C.NODE_INT
	case nodes.KindUint:
		return C.NODE_UINT
	case nodes.KindFloat:
		return C.NODE_FLOAT
	case nodes.KindBool:
		return C.NODE_BOOL
	}
	return C.NODE_NULL
}

//export uastAsString
func uastAsString(ctx C.UastHandle, node C.NodeHandle) *C.char {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return nil
	}
	v := nd.AsString()
	return C.CString(string(v))
}

//export uastAsInt
func uastAsInt(ctx C.UastHandle, node C.NodeHandle) C.int64_t {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return 0
	}
	v := nd.AsInt()
	return C.int64_t(v)
}

//export uastAsUint
func uastAsUint(ctx C.UastHandle, node C.NodeHandle) C.uint64_t {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return 0
	}
	v := nd.AsUint()
	return C.uint64_t(v)
}

//export uastAsFloat
func uastAsFloat(ctx C.UastHandle, node C.NodeHandle) C.double {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return 0
	}
	v := nd.AsFloat()
	return C.double(v)
}

//export uastAsBool
func uastAsBool(ctx C.UastHandle, node C.NodeHandle) C.bool {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return false
	}
	v := nd.AsBool()
	return C.bool(v)
}

//export uastSize
func uastSize(ctx C.UastHandle, node C.NodeHandle) C.size_t {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return 0
	}
	v := nd.Size()
	return C.size_t(v)
}

//export uastKeyAt
func uastKeyAt(ctx C.UastHandle, node C.NodeHandle, i C.size_t) *C.char {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return nil
	}
	v := nd.KeyAt(int(i))
	return C.CString(v)
}

//export uastValueAt
func uastValueAt(ctx C.UastHandle, node C.NodeHandle, i C.size_t) C.NodeHandle {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return 0
	}
	v := nd.ValueAt(int(i))
	if v == nil {
		return 0
	}
	return C.NodeHandle(v.Handle())
}

//export uastNewObject
func uastNewObject(ctx C.UastHandle, sz C.size_t) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewObject(int(sz))
	return C.NodeHandle(h)
}

//export uastNewArray
func uastNewArray(ctx C.UastHandle, sz C.size_t) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewArray(int(sz))
	return C.NodeHandle(h)
}

//export uastNewString
func uastNewString(ctx C.UastHandle, v *C.char) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewString(C.GoString(v))
	return C.NodeHandle(h.Handle())
}

//export uastNewInt
func uastNewInt(ctx C.UastHandle, v C.int64_t) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewInt(int64(v))
	return C.NodeHandle(h.Handle())
}

//export uastNewUint
func uastNewUint(ctx C.UastHandle, v C.uint64_t) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewUint(uint64(v))
	return C.NodeHandle(h.Handle())
}

//export uastNewFloat
func uastNewFloat(ctx C.UastHandle, v C.double) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewFloat(float64(v))
	return C.NodeHandle(h.Handle())
}

//export uastNewBool
func uastNewBool(ctx C.UastHandle, v C.bool) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewBool(bool(v))
	return C.NodeHandle(h.Handle())
}

//export uastSetValue
func uastSetValue(ctx C.UastHandle, node C.NodeHandle, i C.size_t, val C.NodeHandle) {
	c := getContextGo(ctx)
	if c == nil {
		return
	}
	n := c.impl.AsTmpNode(Handle(node))
	if n == nil {
		return
	}
	v := c.impl.AsNode(Handle(val))
	n.SetValue(int(i), v)
}

//export uastSetKeyValue
func uastSetKeyValue(ctx C.UastHandle, node C.NodeHandle, key *C.char, val C.NodeHandle) {
	c := getContextGo(ctx)
	if c == nil {
		return
	}
	n := c.impl.AsTmpNode(Handle(node))
	if n == nil {
		return
	}
	k := C.GoString(key)
	v := c.impl.AsNode(Handle(val))
	n.SetKeyValue(k, v)
}

var _ NodeIface = (*goNodes)(nil)

type goNodes struct {
	last  Handle
	nodes map[Handle]nodes.Node
	tmp   map[Handle]*goTmpNode
}

func (m *goNodes) next() Handle {
	m.last++
	h := m.last
	return h
}

func (m *goNodes) Free() {
	m.nodes = nil
}

func (m *goNodes) toHandle(n nodes.Node) Handle {
	if n == nil {
		return 0
	}
	h := m.next()
	if m.nodes == nil {
		m.nodes = make(map[Handle]nodes.Node)
	}
	m.nodes[h] = n
	return h
}
func (m *goNodes) toNode(n nodes.Node) Node {
	if n == nil {
		return nil
	}
	h := m.toHandle(n)
	return &goNode{c: m, h: h, n: n}
}
func (m *goNodes) AsNode(h Handle) Node {
	n := m.nodes[h]
	if n == nil {
		return nil
	}
	return &goNode{c: m, h: h, n: n}
}
func (m *goNodes) AsTmpNode(h Handle) TmpNode {
	n := m.tmp[h]
	if n == nil {
		return nil
	}
	return n
}

func (m *goNodes) newTmp(n *goTmpNode) Handle {
	h := m.next()
	n.c, n.h = m, h
	if m.tmp == nil {
		m.tmp = make(map[Handle]*goTmpNode)
	}
	m.tmp[h] = n
	return h
}

func (m *goNodes) NewObject(sz int) Handle {
	return m.newTmp(&goTmpNode{obj: make(nodes.Object, sz)})
}

func (m *goNodes) NewArray(sz int) Handle {
	return m.newTmp(&goTmpNode{arr: make(nodes.Array, sz)})
}

func (m *goNodes) NewString(v string) Node {
	return m.toNode(nodes.String(v))
}

func (m *goNodes) NewInt(v int64) Node {
	return m.toNode(nodes.Int(v))
}

func (m *goNodes) NewUint(v uint64) Node {
	return m.toNode(nodes.Uint(v))
}

func (m *goNodes) NewFloat(v float64) Node {
	return m.toNode(nodes.Float(v))
}

func (m *goNodes) NewBool(v bool) Node {
	return m.toNode(nodes.Bool(v))
}

var _ Node = (*goNode)(nil)

type goNode struct {
	c    *goNodes
	h    Handle
	n    nodes.Node
	keys []string
}

func (n *goNode) Handle() Handle {
	if n == nil {
		return 0
	}
	return n.h
}

func (n *goNode) Kind() nodes.Kind {
	if n == nil {
		return nodes.KindNil
	}
	return nodes.KindOf(n.n)
}

func (n *goNode) AsString() nodes.String {
	return n.n.(nodes.String)
}

func (n *goNode) AsInt() nodes.Int {
	return n.n.(nodes.Int)
}

func (n *goNode) AsUint() nodes.Uint {
	return n.n.(nodes.Uint)
}

func (n *goNode) AsFloat() nodes.Float {
	return n.n.(nodes.Float)
}

func (n *goNode) AsBool() nodes.Bool {
	return n.n.(nodes.Bool)
}

func (n *goNode) Size() int {
	switch v := n.n.(type) {
	case nodes.Object:
		return len(v)
	case nodes.Array:
		return len(v)
	}
	return 0
}

func (n *goNode) cacheKeys() {
	if n.keys != nil {
		return
	}
	obj := n.n.(nodes.Object)
	n.keys = obj.Keys()
}

func (n *goNode) KeyAt(i int) string {
	n.cacheKeys()
	if i < 0 || i >= len(n.keys) {
		return ""
	}
	return n.keys[i]
}

func (n *goNode) ValueAt(i int) Node {
	if arr, ok := n.n.(nodes.Array); ok {
		if i < 0 || i >= len(arr) {
			return nil
		}
		return n.c.toNode(arr[i])
	}
	n.cacheKeys()
	if i < 0 || i >= len(n.keys) {
		return nil
	}
	obj := n.n.(nodes.Object)
	v := obj[n.keys[i]]
	return n.c.toNode(v)
}

var _ TmpNode = (*goTmpNode)(nil)

type goTmpNode struct {
	c   *goNodes
	h   Handle
	obj nodes.Object
	arr nodes.Array
}

func (n *goTmpNode) SetValue(i int, v Node) {
	if n.arr == nil {
		panic("not an array")
	}
	if v != nil {
		n.arr[i] = v.(*goNode).n
	}
}

func (n *goTmpNode) SetKeyValue(k string, v Node) {
	if n.obj == nil {
		panic("not an object")
	}
	if v == nil {
		n.obj[k] = nil
	} else {
		n.obj[k] = v.(*goNode).n
	}
}

func (n *goTmpNode) Build() Node {
	var nd Node
	if n.obj != nil {
		nd = n.c.toNode(n.obj)
		n.obj = nil
	} else if n.arr != nil {
		nd = n.c.toNode(n.arr)
		n.arr = nil
	}
	return nd
}
