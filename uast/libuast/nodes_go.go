package main

/*
#include "uast_go.h"
*/
import "C"

import (
	"fmt"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

var goImpl *C.NodeIface

func init() {
	goImpl = C.uastImpl()
}

func getContextGo(h C.UastHandle) *Context {
	if h == 0 {
		return nil
	}
	return getContext(Handle(h))
}

func setContextErrorGo(h C.UastHandle, err error) {
	c := getContextGo(h)
	if c == nil {
		panic(err)
	}
	c.setError(err)
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
	v, _ := nd.Value().(nodes.String)
	return C.CString(string(v))
}

//export uastAsInt
func uastAsInt(ctx C.UastHandle, node C.NodeHandle) C.int64_t {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return 0
	}
	v, _ := nd.Value().(nodes.Int)
	return C.int64_t(v)
}

//export uastAsUint
func uastAsUint(ctx C.UastHandle, node C.NodeHandle) C.uint64_t {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return 0
	}
	v, _ := nd.Value().(nodes.Uint)
	return C.uint64_t(v)
}

//export uastAsFloat
func uastAsFloat(ctx C.UastHandle, node C.NodeHandle) C.double {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return 0
	}
	v, _ := nd.Value().(nodes.Float)
	return C.double(v)
}

//export uastAsBool
func uastAsBool(ctx C.UastHandle, node C.NodeHandle) C.bool {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return false
	}
	v, _ := nd.Value().(nodes.Bool)
	return C.bool(v)
}

//export uastSize
func uastSize(ctx C.UastHandle, node C.NodeHandle) C.size_t {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return 0
	}
	var sz int
	switch nd.Kind() {
	case nodes.KindObject:
		if o, ok := nd.(Object); ok {
			sz = o.Size()
		}
	case nodes.KindArray:
		if o, ok := nd.(Array); ok {
			sz = o.Size()
		}
	}
	return C.size_t(sz)
}

//export uastKeyAt
func uastKeyAt(ctx C.UastHandle, node C.NodeHandle, i C.size_t) *C.char {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return nil
	}
	o, ok := nd.(Object)
	if !ok {
		err := fmt.Errorf("expected object, got: %T", nd)
		setContextErrorGo(ctx, err)
		return nil
	}
	keys := o.Keys()
	ind := int(i)
	if ind < 0 || ind >= len(keys) {
		err := fmt.Errorf("index out of bounds: %d, %d", ind, len(keys))
		setContextErrorGo(ctx, err)
		return nil
	}
	v := keys[ind]
	return C.CString(v)
}

//export uastValueAt
func uastValueAt(ctx C.UastHandle, node C.NodeHandle, i C.size_t) C.NodeHandle {
	nd := getNodeGo(ctx, node)
	if nd == nil {
		return 0
	}
	var (
		ind = int(i)
		v   Node
	)
	switch nd := nd.(type) {
	case Object:
		c := getContextGo(ctx)

		keys := nd.Keys()
		if ind < 0 || ind >= len(keys) {
			return 0
		}
		key := keys[ind]
		val, ok := nd.ValueAt(key)
		if !ok {
			err := fmt.Errorf("cannot fetch key: %q", key)
			c.setError(err)
			return 0
		}
		v = c.toNode(val)
	case Array:
		c := getContextGo(ctx)

		sz := nd.Size()
		if ind < 0 || ind >= sz {
			err := fmt.Errorf("index out of bounds: %d, %d", ind, sz)
			c.setError(err)
			return 0
		}
		val := nd.ValueAt(ind)
		v = c.toNode(val)
	}
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
	s := C.GoString(v)
	h := c.impl.NewValue(nodes.String(s))
	return C.NodeHandle(h.Handle())
}

//export uastNewInt
func uastNewInt(ctx C.UastHandle, v C.int64_t) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewValue(nodes.Int(v))
	return C.NodeHandle(h.Handle())
}

//export uastNewUint
func uastNewUint(ctx C.UastHandle, v C.uint64_t) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewValue(nodes.Uint(v))
	return C.NodeHandle(h.Handle())
}

//export uastNewFloat
func uastNewFloat(ctx C.UastHandle, v C.double) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewValue(nodes.Float(v))
	return C.NodeHandle(h.Handle())
}

//export uastNewBool
func uastNewBool(ctx C.UastHandle, v C.bool) C.NodeHandle {
	c := getContextGo(ctx)
	if c == nil {
		return 0
	}
	h := c.impl.NewValue(nodes.Bool(v))
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

type Native interface {
	Node
	Native() nodes.Node
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
func (m *goNodes) newNode(h Handle, n nodes.Node) Node {
	switch n := n.(type) {
	case nil:
		return nil
	case nodes.Object:
		return &goObject{c: m, h: h, obj: n}
	case nodes.Array:
		return &goArray{c: m, h: h, arr: n}
	}
	return &goValue{c: m, h: h, val: n.(nodes.Value)}
}
func (m *goNodes) toNode(n nodes.Node) Node {
	if n == nil {
		return nil
	}
	h := m.toHandle(n)
	return m.newNode(h, n)
}
func (m *goNodes) AsNode(h Handle) Node {
	n := m.nodes[h]
	if n == nil {
		return nil
	}
	return m.newNode(h, n)
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

func (m *goNodes) NewValue(v nodes.Value) Node {
	if v == nil {
		return nil
	}
	return m.toNode(v)
}

var (
	_ Native = (*goObject)(nil)
	_ Object = (*goObject)(nil)
)

type goObject struct {
	c    *goNodes
	h    Handle
	obj  nodes.Object
	keys []string
}

func (n *goObject) Native() nodes.Node {
	return n.obj
}

func (n *goObject) Handle() Handle {
	return n.h
}

func (n *goObject) Kind() nodes.Kind {
	return nodes.KindObject
}

func (n *goObject) Value() nodes.Value {
	return nil
}

func (n *goObject) Size() int {
	return n.obj.Size()
}

func (n *goObject) Keys() []string {
	if n.keys == nil {
		n.keys = n.obj.Keys()
	}
	return n.keys
}

func (n *goObject) ValueAt(key string) (nodes.External, bool) {
	return n.obj.ValueAt(key)
}

var (
	_ Native = (*goArray)(nil)
	_ Array  = (*goArray)(nil)
)

type goArray struct {
	c   *goNodes
	h   Handle
	arr nodes.Array
}

func (n *goArray) Native() nodes.Node {
	return n.arr
}

func (n *goArray) Handle() Handle {
	return n.h
}

func (n *goArray) Kind() nodes.Kind {
	return nodes.KindArray
}

func (n *goArray) Value() nodes.Value {
	return nil
}

func (n *goArray) Size() int {
	return n.arr.Size()
}

func (n *goArray) ValueAt(i int) nodes.External {
	return n.arr.ValueAt(i)
}

var _ Native = (*goValue)(nil)

type goValue struct {
	c   *goNodes
	h   Handle
	val nodes.Value
}

func (n *goValue) Native() nodes.Node {
	return n.val
}

func (n *goValue) Handle() Handle {
	if n == nil {
		return 0
	}
	return n.h
}

func (n *goValue) Kind() nodes.Kind {
	if n == nil {
		return nodes.KindNil
	}
	return nodes.KindOf(n.val)
}

func (n *goValue) Value() nodes.Value {
	return n.val
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
		n.arr[i] = v.(Native).Native()
	}
}

func (n *goTmpNode) SetKeyValue(k string, v Node) {
	if n.obj == nil {
		panic("not an object")
	}
	if v == nil {
		n.obj[k] = nil
	} else {
		n.obj[k] = v.(Native).Native()
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
