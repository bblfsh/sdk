package transformer

import (
	"sort"

	"gopkg.in/bblfsh/sdk.v1/uast"
)

const allowUnusedFields = false

func noNode(n uast.Node) error {
	if n == nil {
		return nil
	}
	return ErrUnexpectedNode.New(n)
}

// Is checks if the current node is a primitive and is equal to a given value.
// Reversal changes the type of the node to primitive and assigns given value to the node.
func Is(v uast.Value) Op {
	return opIs{v: v}
}

type opIs struct {
	v uast.Value
}

func (op opIs) Check(st *State, n uast.Node) (bool, error) {
	v2, ok := n.(uast.Value)
	if !ok {
		return op.v == nil && n == nil, nil
	}
	return op.v == v2, nil
}

func (op opIs) Construct(st *State, n uast.Node) (uast.Node, error) {
	nv := op.v
	return nv, nil
}

// Var stores current node as a value to a named variable in the shared state.
// Reversal replaces current node with the one from named variable. Variables can store subtrees.
func Var(name string) Op {
	return opVar{name: name}
}

type opVar struct {
	name string
}

func (op opVar) Check(st *State, n uast.Node) (bool, error) {
	if err := st.SetVar(op.name, n); err != nil {
		return false, err
	}
	return true, nil
}

func (op opVar) Construct(st *State, n uast.Node) (uast.Node, error) {
	if err := noNode(n); err != nil {
		return nil, err
	}
	val, ok := st.GetVar(op.name)
	if !ok {
		return nil, ErrVariableNotDefined.New(op.name)
	}
	// TODO: should we clone it?
	return val, nil
}

// Any matches any node and throws it away. Reversal will create a node with create op.
func Any(create Mod) Op {
	return opAny{create: create}
}

type opAny struct {
	create Mod
}

func (op opAny) Check(st *State, n uast.Node) (bool, error) {
	return true, nil // always succeeds
}

func (op opAny) Construct(st *State, n uast.Node) (uast.Node, error) {
	return op.create.Construct(st, n)
}

// AnyVal accept any value and aways creates a node with a provided one.
func AnyVal(val uast.Value) Op {
	return Any(Is(val))
}

// And checks current node with all ops and fails if any of them fails.
// Reversal applies all modifications from ops to the current node.
// Typed ops should be at the beginning of the list to make sure that `Construct`
// creates a correct node type before applying specific changes to it.
func And(ops ...Op) Op {
	if len(ops) == 1 {
		return ops[0]
	}
	return opAnd(ops)
}

type opAnd []Op

func (op opAnd) Check(st *State, n uast.Node) (bool, error) {
	for i, sub := range op {
		if ok, err := sub.Check(st, n); err != nil {
			return false, errAnd.Wrap(err, i, sub)
		} else if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (op opAnd) Construct(st *State, n uast.Node) (uast.Node, error) {
	for i, sub := range op {
		var err error
		n, err = sub.Construct(st, n)
		if err != nil {
			return nil, errAnd.Wrap(err, i, sub)
		}
	}
	return n, nil
}

func Obj(ops map[string]Op) Op {
	return Object(ops)
}

var _ Op = Object{}

// Obj verifies that current node is an object and checks it with provided ops.
// Reversal changes node type to object and applies a provided operations to it.
// This operation will populate a list of unprocessed keys for current object,
// so the transformation code can verify that transform was complete.
// FIXME: update docs here
// Out checks specific object field with an op.
// Reversal creates a field in an object using provided op. It will also
// remove the key from the list of unprocessed keys for this specific node.
type Object map[string]Op

func (o Object) Check(st *State, n uast.Node) (bool, error) {
	return opObject{fields: o}.Check(st, n)
}

func (o Object) Construct(st *State, n uast.Node) (uast.Node, error) {
	return opObject{fields: o}.Construct(st, n)
}

func Part(vr string, o Object) Op {
	return opObject{fields: o, other: vr}
}

type opObject struct {
	fields map[string]Op
	other  string // preserve other fields
}

func (o opObject) Keys() []string {
	keys := make([]string, 0, len(o.fields))
	for k := range o.fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (o opObject) Check(st *State, n uast.Node) (bool, error) {
	cur, ok := n.(uast.Object)
	if !ok {
		return false, nil
	}
	for _, key := range o.Keys() {
		n, ok = cur[key]
		if !ok {
			return false, nil
		}
		sub := o.fields[key]
		ok, err := sub.Check(st, n)
		if err != nil {
			return false, errKey.Wrap(err, key)
		} else if !ok {
			return false, nil
		}
	}
	if o.other == "" {
		if !allowUnusedFields {
			for k := range cur {
				if _, ok := o.fields[k]; !ok {
					return false, ErrUnusedField.New(k)
				}
			}
		}
		return true, nil
	}
	// TODO: consider throwing an error if a transform is defined as partial, but in fact it's not
	left := make(uast.Object)
	for k, v := range cur {
		if _, ok := o.fields[k]; !ok {
			left[k] = v
		}
	}
	err := st.SetVar(o.other, left)
	return err == nil, err
}

func (o opObject) Construct(st *State, old uast.Node) (uast.Node, error) {
	if err := noNode(old); err != nil {
		return nil, err
	}
	obj := make(uast.Object, len(o.fields))
	for _, key := range o.Keys() {
		sub := o.fields[key]
		v, err := sub.Construct(st, nil)
		if err != nil {
			return obj, errKey.Wrap(err, key)
		}
		obj[key] = v
	}
	if o.other == "" {
		return obj, nil
	}
	v, ok := st.GetVar(o.other)
	if !ok {
		return obj, ErrVariableNotDefined.New(o.other)
	}
	left, ok := v.(uast.Object)
	if !ok {
		return obj, ErrExpectedObject.New(v)
	}
	for k, v := range left {
		obj[k] = v
	}
	return obj, nil
}

// String asserts that value equals a specific string value.
func String(val string) Op {
	return Is(uast.String(val))
}

// Int asserts that value equals a specific integer value.
func Int(val int) Op {
	return Is(uast.Int(val))
}

// TypedObj is a shorthand for an object with a specific type
// and multiples operations on it.
func TypedObj(typ string, ops map[string]Op) Op {
	obj := Object(ops)
	obj[uast.KeyType] = String(typ)
	return obj
}

// Arr checks if the current object is a list with a number of elements
// matching a number of ops, and applies ops to corresponding elements.
// Reversal creates a list of the size that matches the number of ops
// and creates each element with the corresponding op.
func Arr(ops ...Op) Op {
	return opArr(ops)
}

type opArr []Op

func (op opArr) Check(st *State, n uast.Node) (bool, error) {
	arr, ok := n.(uast.List)
	if !ok {
		return false, nil
	} else if len(arr) != len(op) {
		return false, nil
	}
	for i, sub := range op {
		if ok, err := sub.Check(st, arr[i]); err != nil {
			return false, errElem.Wrap(err, i, sub)
		} else if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (op opArr) Construct(st *State, n uast.Node) (uast.Node, error) {
	if err := noNode(n); err != nil {
		return nil, err
	}
	arr := make(uast.List, 0, len(op))
	for i, sub := range op {
		nn, err := sub.Construct(st, n)
		if err != nil {
			return nil, errElem.Wrap(err, i, sub)
		}
		arr = append(arr, nn)
	}
	return arr, nil
}

// One is a shorthand for a list with one element.
func One(op Op) Op {
	return Arr(op)
}

// Lookup uses a value of current node to find a replacement for it
// in the map and checks result with op.
// The reverse step will use a reverse map to lookup value created by
// op and will assign it to the current node.
// Since reversal transformation needs to build a reverse map,
// the mapping should not be ambiguous in reverse direction (no duplicate values).
func Lookup(op Op, m map[uast.Value]uast.Value) Op {
	rev := make(map[uast.Value]uast.Value, len(m))
	for k, v := range m {
		if _, ok := rev[v]; ok {
			panic(ErrAmbiguousValue.New("map has ambigous value %v", v))
		}
		rev[v] = k
	}
	return opLookup{op: op, fwd: m, rev: rev}
}

type opLookup struct {
	op       Op
	fwd, rev map[uast.Value]uast.Value
}

func (op opLookup) Check(st *State, n uast.Node) (bool, error) {
	v, ok := n.(uast.Value)
	if !ok {
		return false, ErrExpectedValue.New(n)
	}
	vn, ok := op.fwd[v]
	if !ok {
		return false, ErrUnhandledValue.New(v)
	}
	return op.op.Check(st, vn)
}

func (op opLookup) Construct(st *State, n uast.Node) (uast.Node, error) {
	if err := noNode(n); err != nil {
		return nil, err
	}
	nn, err := op.op.Construct(st, nil)
	if err != nil {
		return nil, err
	}
	v, ok := nn.(uast.Value)
	if !ok {
		return nil, ErrExpectedValue.New(n)
	}
	vn, ok := op.rev[v]
	if !ok {
		return nil, ErrUnhandledValue.New(v)
	}
	return vn, nil
}

// LookupVar is a shorthand to lookup value stored in variable.
func LookupVar(vr string, m map[uast.Value]uast.Value) Op {
	return Lookup(Var(vr), m)
}
