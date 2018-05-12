package protouast

import (
	"fmt"

	"gopkg.in/bblfsh/sdk.v2/uast"
)

//go:generate protoc --proto_path=$GOPATH/src:. --gogo_out=. uast.proto

func ToProto(n uast.Node) *Node {
	var nd isNode_Node
	switch n := n.(type) {
	case nil:
		// don't set the field
	case uast.Value:
		v := ValueToProto(n)
		nd = &Node_Value{Value: v}
	case uast.Object:
		v := ObjectToProto(n)
		nd = &Node_Object{Object: v}
	case uast.Array:
		v := ArrayToProto(n)
		nd = &Node_Array{Array: v}
	default:
		panic(fmt.Errorf("unexpected type: %T", n))
	}
	return &Node{Node: nd}
}

func ValueToProto(v uast.Value) *Value {
	var pv isValue_Value
	switch v := v.(type) {
	case nil:
		pv = &Value_Null{Null: true}
	case uast.String:
		pv = &Value_Str{Str: string(v)}
	case uast.Int:
		pv = &Value_Int{Int: int64(v)}
	case uast.Float:
		pv = &Value_Float{Float: float64(v)}
	case uast.Bool:
		pv = &Value_Bool{Bool: bool(v)}
	default:
		panic(fmt.Errorf("unexpected value type: %T", v))
	}
	return &Value{Value: pv}
}

func ObjectToProto(m uast.Object) *Object {
	obj := &Object{
		Fields: make(map[string]*Node, len(m)),
	}
	for k, v := range m {
		obj.Fields[k] = ToProto(v)
	}
	return obj
}

func ArrayToProto(l uast.Array) *Array {
	arr := &Array{
		Array: make([]*Node, 0, len(l)),
	}
	for _, v := range l {
		arr.Array = append(arr.Array, ToProto(v))
	}
	return arr
}

var (
	_ pbNode = (*Node_Value)(nil)
	_ pbNode = (*Node_Object)(nil)
	_ pbNode = (*Node_Array)(nil)
)

type pbNode interface {
	isNode_Node
	ToNative() uast.Node
}

func (m *Node) ToNative() uast.Node {
	if m == nil || m.Node == nil {
		return nil
	}
	nd, ok := m.Node.(pbNode)
	if !ok {
		panic(fmt.Errorf("unexpected type: %T", m.Node))
	}
	return nd.ToNative()
}

func (m *Node_Object) ToNative() uast.Node {
	if m == nil {
		return nil
	}
	return m.Object.ToNative()
}

func (m *Object) ToNative() uast.Object {
	if m == nil {
		return nil
	}
	o := make(uast.Object, len(m.Fields))
	for k, v := range m.Fields {
		o[k] = v.ToNative()
	}
	return o
}

func (m *Node_Array) ToNative() uast.Node {
	if m == nil {
		return nil
	}
	return m.Array.ToNative()
}

func (m *Array) ToNative() uast.Array {
	if m == nil {
		return nil
	}
	l := make(uast.Array, 0, len(m.Array))
	for _, v := range m.Array {
		l = append(l, v.ToNative())
	}
	return l
}

func (m *Node_Value) ToNative() uast.Node {
	if m == nil {
		return nil
	}
	return m.Value.ToNative()
}

func (m *Value) ToNative() uast.Value {
	if m == nil {
		return nil
	}
	pv, ok := m.Value.(pbValue)
	if !ok {
		panic(fmt.Errorf("unexpected type: %T", m.Value))
	}
	return pv.ToNative()
}

var (
	_ pbValue = (*Value_Null)(nil)
	_ pbValue = (*Value_Str)(nil)
	_ pbValue = (*Value_Int)(nil)
	_ pbValue = (*Value_Float)(nil)
	_ pbValue = (*Value_Bool)(nil)
)

type pbValue interface {
	isValue_Value
	ToNative() uast.Value
}

func (_ *Value_Null) ToNative() uast.Value {
	return nil
}

func (m *Value_Str) ToNative() uast.Value {
	if m == nil {
		return nil
	}
	return uast.String(m.Str)
}

func (m *Value_Int) ToNative() uast.Value {
	if m == nil {
		return nil
	}
	return uast.Int(m.Int)
}

func (m *Value_Float) ToNative() uast.Value {
	if m == nil {
		return nil
	}
	return uast.Float(m.Float)
}

func (m *Value_Bool) ToNative() uast.Value {
	if m == nil {
		return nil
	}
	return uast.Bool(m.Bool)
}
