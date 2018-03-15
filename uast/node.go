package uast

import (
	"fmt"
	"sort"

	"gopkg.in/bblfsh/sdk.v1/uast/role"
	"gopkg.in/src-d/go-errors.v1"
)

var (
	// ErrUnsupported is returned for features that are not supported by an implementation.
	ErrUnsupported = errors.NewKind("unsupported: %s")
)

const applySort = false

// Special field keys for Object
const (
	KeyType  = "@type"  // InternalType
	KeyToken = "@token" // Token
	KeyRoles = "@role"  // Roles, for representations see RoleList
	// TODO: a single @pos field with "start" and "end" fields?
	KeyStart = "@start" // StartPosition
	KeyEnd   = "@end"   // EndPosition
)

// NewNode creates a default AST node with Unannotated role.
func NewNode() Object {
	return Object{KeyRoles: RoleList(role.Unannotated)}
}

// EmptyNode creates a new empty node with no fields.
func EmptyNode() Object {
	return Object{}
}

// Node is a generic interface for structures used in AST.
//
// Can be one of:
//	* Object
//	* Array
//	* Value
type Node interface {
	// Clone creates a deep copy of the node.
	Clone() Node
	Native() interface{}
	isNode() // to limit possible types
}

// Value is a generic interface for values of AST node fields.
//
// Can be one of:
//	* String
//	* Int
//	* Float
//	* Bool
type Value interface {
	Node
	isValue() // to limit possible types
}

// Object is a representation of generic AST node with fields.
type Object map[string]Node

func (Object) isNode() {}

// Native converts an object to a generic Go map type (map[string]interface{}).
func (m Object) Native() interface{} {
	if m == nil {
		return nil
	}
	o := make(map[string]interface{}, len(m))
	for k, v := range m {
		if v != nil {
			o[k] = v.Native()
		} else {
			o[k] = nil
		}
	}
	return o
}

// Keys returns a sorted list of node keys.
func (m Object) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Clone returns a deep copy of an Object.
func (m Object) Clone() Node {
	out := make(Object, len(m))
	for k, v := range m {
		if v != nil {
			out[k] = v.Clone()
		} else {
			out[k] = nil
		}
	}
	return out
}

// CloneObject clones this AST node only, without deep copy of field values.
func (m Object) CloneObject() Object {
	out := make(Object, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// CloneProperties returns an object containing all field that are values.
func (m Object) CloneProperties() Object {
	out := make(Object)
	for k, v := range m {
		if v, ok := v.(Value); ok {
			out[k] = v
		}
	}
	return out
}

// Children returns a list of all internal nodes of type Object and Array.
func (m Object) Children() []Node {
	out := make([]Node, 0, len(m))
	// order should be predictable
	for _, k := range m.Keys() {
		v := m[k]
		if _, ok := v.(Value); !ok {
			out = append(out, v)
		}
	}
	return out
}

// Properties returns a map containing all field of object that are values.
func (m Object) Properties() map[string]Value {
	out := make(map[string]Value)
	for k, v := range m {
		if v, ok := v.(Value); ok {
			out[k] = v
		}
	}
	return out
}

// SetProperty is a helper for setting node properties.
func (m Object) SetProperty(k, v string) Object {
	m[k] = String(v)
	return m
}

// Type is a helper for getting node type (see KeyType).
func (m Object) Type() string {
	s, _ := m[KeyType].(String)
	return string(s)
}

// SetType is a helper for setting node type (see KeyType).
func (m Object) SetType(typ string) Object {
	return m.SetProperty(KeyType, typ)
}

// Token is a helper for getting node token (see KeyToken).
func (m Object) Token() string {
	t := m[KeyToken]
	s, ok := t.(String)
	if ok {
		return string(s)
	}
	v, _ := t.(Value)
	if v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

// SetToken is a helper for setting node type (see KeyToken).
func (m Object) SetToken(tok string) Object {
	return m.SetProperty(KeyToken, tok)
}

// Roles is a helper for getting node UAST roles (see KeyRoles).
func (m Object) Roles() []role.Role {
	arr, ok := m[KeyRoles].(Array)
	if !ok {
		return nil
	}
	out := make([]role.Role, 0, len(arr))
	for _, v := range arr {
		if r, ok := v.(String); ok {
			out = append(out, role.FromString(string(r)))
		}
	}
	return out
}

// SetRoles is a helper for setting node UAST roles (see KeyRoles).
func (m Object) SetRoles(roles ...role.Role) Object {
	m[KeyRoles] = RoleList(roles...)
	return m
}

// StartPosition returns start position of the node in source file.
func (m Object) StartPosition() *Position {
	o, _ := m[KeyStart].(Object)
	return AsPosition(o)
}

// EndPosition returns start position of the node in source file.
func (m Object) EndPosition() *Position {
	o, _ := m[KeyEnd].(Object)
	return AsPosition(o)
}

// Array is an ordered list of AST nodes.
type Array []Node

func (Array) isNode() {}

// Native converts an array to a generic Go slice type ([]interface{}).
func (m Array) Native() interface{} {
	if m == nil {
		return nil
	}
	o := make([]interface{}, 0, len(m))
	for _, v := range m {
		if v != nil {
			o = append(o, v.Native())
		} else {
			o = append(o, nil)
		}
	}
	return o
}

// Clone returns a deep copy of an Array.
func (m Array) Clone() Node {
	out := make(Array, 0, len(m))
	for _, v := range m {
		out = append(out, v.Clone())
	}
	return out
}

// CloneList creates a copy of an Array without copying it's elements.
func (m Array) CloneList() Array {
	out := make(Array, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

// String is a string value used in AST fields.
type String string

func (String) isNode()  {}
func (String) isValue() {}

// Native converts the value to a string.
func (v String) Native() interface{} {
	return string(v)
}

// Clone returns a copy of the value.
func (v String) Clone() Node {
	return v
}

// Int is a integer value used in AST fields.
type Int int64

func (Int) isNode()  {}
func (Int) isValue() {}

// Native converts the value to an int64.
func (v Int) Native() interface{} {
	return int64(v)
}

// Clone returns a copy of the value.
func (v Int) Clone() Node {
	return v
}

// Float is a floating point value used in AST fields.
type Float float64

func (Float) isNode()  {}
func (Float) isValue() {}

// Native converts the value to a float64.
func (v Float) Native() interface{} {
	return float64(v)
}

// Clone returns a copy of the value.
func (v Float) Clone() Node {
	return v
}

// Bool is a boolean value used in AST fields.
type Bool bool

func (Bool) isNode()  {}
func (Bool) isValue() {}

// Native converts the value to a bool.
func (v Bool) Native() interface{} {
	return bool(v)
}

// Clone returns a copy of the value.
func (v Bool) Clone() Node {
	return v
}

// ToNode converts objects returned by schema-less encodings such as JSON to Node objects.
func ToNode(o interface{}) (Node, error) {
	switch o := o.(type) {
	case nil:
		return nil, nil
	case Node:
		return o, nil
	case map[string]interface{}:
		n := make(Object, len(o))
		for k, v := range o {
			nv, err := ToNode(v)
			if err != nil {
				return nil, err
			}
			n[k] = nv
		}
		return n, nil
	case []interface{}:
		n := make(Array, 0, len(o))
		for _, v := range o {
			nv, err := ToNode(v)
			if err != nil {
				return nil, err
			}
			n = append(n, nv)
		}
		return n, nil
	case string:
		return String(o), nil
	case int:
		return Int(o), nil
	case int64:
		return Int(o), nil
	case float64:
		if float64(int64(o)) != o {
			return Float(o), nil
		}
		return Int(o), nil
	case bool:
		return Bool(o), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", o)
	}
}

// WalkPreOrder visits all nodes of the tree in pre-order.
func WalkPreOrder(root Node, walk func(Node) bool) {
	if !walk(root) {
		return
	}
	switch n := root.(type) {
	case Object:
		for _, k := range n.Keys() {
			WalkPreOrder(n[k], walk)
		}
	case Array:
		for _, s := range n {
			WalkPreOrder(s, walk)
		}
	}
}

// Apply takes a root node and applies callback to each node of the tree recursively.
// Apply returns an old or a new node and a flag that indicates if node was changed or not.
// If callback returns true and a new node, Apply will make a copy of parent node and
// will replace an old value with a new one. It will make a copy of all parent
// nodes recursively in this case.
func Apply(root Node, apply func(n Node) (Node, bool)) (Node, bool) {
	if root == nil {
		return nil, false
	}
	var changed bool
	switch n := root.(type) {
	case Object:
		var nn Object
		if applySort {
			for _, k := range n.Keys() {
				v := n[k]
				if nv, ok := Apply(v, apply); ok {
					if nn == nil {
						nn = n.CloneObject()
					}
					nn[k] = nv
				}
			}
		} else {
			for k, v := range n {
				if nv, ok := Apply(v, apply); ok {
					if nn == nil {
						nn = n.CloneObject()
					}
					nn[k] = nv
				}
			}
		}
		if nn != nil {
			changed = true
			root = nn
		}
	case Array:
		var nn Array
		for i, v := range n {
			if nv, ok := Apply(v, apply); ok {
				if nn == nil {
					nn = n.CloneList()
				}
				nn[i] = nv
			}
		}
		if nn != nil {
			changed = true
			root = nn
		}
	}
	nn, changed2 := apply(root)
	return nn, changed || changed2
}
