package transformer

import (
	"fmt"
	"sort"

	"gopkg.in/bblfsh/sdk.v1/uast"
)

const (
	allowUnusedFields  = false
	errorOnFilterCheck = false
)

func noNode(n uast.Node) error {
	if n == nil {
		return nil
	}
	return ErrUnexpectedNode.New(n)
}

func filtered(format string, args ...interface{}) (bool, error) {
	if !errorOnFilterCheck {
		return false, nil
	}
	return false, fmt.Errorf(format, args...)
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
	val, err := st.MustGetVar(op.name)
	if err != nil {
		return nil, err
	}
	// TODO: should we clone it?
	return val, nil
}

// AnyNode matches any node and throws it away. Reversal will create a node with create op.
func AnyNode(create Mod) Op {
	return opAnyNode{create: create}
}

type opAnyNode struct {
	create Mod
}

func (op opAnyNode) Check(st *State, n uast.Node) (bool, error) {
	return true, nil // always succeeds
}

func (op opAnyNode) Construct(st *State, n uast.Node) (uast.Node, error) {
	return op.create.Construct(st, n)
}

// AnyVal accept any value and aways creates a node with a provided one.
func AnyVal(val uast.Value) Op {
	return AnyNode(Is(val))
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

var _ ObjectOp = Obj{}

// Obj is a helper for defining a transformation on an object fields. See Object.
// Operations will be sorted by the field name before execution.
type Obj map[string]Op

// Object converts this helper to a full Object description.
func (o Obj) Object() Object {
	obj := Object{
		set:    make(map[string]struct{}, len(o)),
		fields: make(Fields, 0, len(o)),
	}
	for k, op := range o {
		obj.set[k] = struct{}{}
		obj.fields = append(obj.fields, Field{Name: k, Op: op})
	}
	sort.Sort(ByFieldName(obj.fields))
	return obj
}

// Check will make an Object operation from this helper and call Check on it.
func (o Obj) Check(st *State, n uast.Node) (bool, error) {
	return o.Object().Check(st, n)
}

// Construct will make an Object operation from this helper and call Construct on it.
func (o Obj) Construct(st *State, n uast.Node) (uast.Node, error) {
	return o.Object().Construct(st, n)
}

// ObjectOp is an operation that is executed on an object. See Object.
type ObjectOp interface {
	Op
	Object() Object
}

// Part defines a partial transformation of an object.
// All unused fields will be stored into variable with a specified name.
func Part(vr string, o ObjectOp) ObjectOp {
	obj := o.Object()
	obj.other = vr
	return obj
}

// Pre will execute provided field operation before executing the rest of operations for an object.
func Pre(fields Fields, o ObjectOp) ObjectOp {
	obj := o.Object()
	if err := obj.setFields(fields...); err != nil {
		panic(err)
	}
	arr := make([]Field, len(fields)+len(obj.fields))

	i := copy(arr, fields)
	copy(arr[i:], obj.fields)

	obj.fields = arr
	return obj
}

// Post will execute provided field operation after executing the rest of operations for an object.
func Post(o ObjectOp, fields Fields) ObjectOp {
	obj := o.Object()
	if err := obj.setFields(fields...); err != nil {
		panic(err)
	}
	arr := make([]Field, len(fields)+len(obj.fields))

	i := copy(arr, obj.fields)
	copy(arr[i:], fields)

	obj.fields = arr
	return obj
}

var _ ObjectOp = Fields{}

// Fields is a helper for multiple operations on object fields with a specific execution order. See Object.
type Fields []Field

// Object converts this helper to a full Object description.
func (o Fields) Object() Object {
	obj := Object{fields: o, set: make(map[string]struct{})}
	err := obj.setFields(obj.fields...)
	if err != nil {
		panic(err)
	}
	return obj
}

// Check will make an Object operation from this helper and call Check on it.
func (o Fields) Check(st *State, n uast.Node) (bool, error) {
	return o.Object().Check(st, n)
}

// Construct will make an Object operation from this helper and call Construct on it.
func (o Fields) Construct(st *State, n uast.Node) (uast.Node, error) {
	return o.Object().Construct(st, n)
}

// ByFieldName will sort fields descriptions by their names.
type ByFieldName []Field

func (arr ByFieldName) Len() int {
	return len(arr)
}

func (arr ByFieldName) Less(i, j int) bool {
	return arr[i].Name < arr[j].Name
}

func (arr ByFieldName) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

// Field is an operation on a specific field of an object.
type Field struct {
	Name string // name of the field
	// Optional can be set to make a field optional. Provided string is used as a variable name to the state of the field.
	// Note that "optional" means that the field may not exists in the object, and it does not mean that the field can be nil.
	// To handle nil fields, see Opt operation.
	Optional string
	Op       Op // operation used to check/construct the field value
}

// Object verifies that current node is an object and checks its fields with a
// defined operations. If field does not exist, object will be skipped.
// Reversal changes node type to object and creates all fields with a specified
// operations.
// Implementation will track a list of unprocessed object keys and will return an
// error in case the field was not used. To preserve all unprocessed keys use Part.
type Object struct {
	fields []Field
	set    map[string]struct{}
	other  string // preserve other fields
}

// Object returns the same object description.
func (o Object) Object() Object {
	return o
}

// GetField returns a field description by it's name.
func (o Object) GetField(k string) (Field, bool) {
	if _, ok := o.set[k]; !ok {
		return Field{}, false
	}
	for _, f := range o.fields {
		if f.Name == k {
			return f, true
		}
	}
	return Field{}, false
}

// SetField sets an operation for a named field. For more options, see SetFieldObj.
func (o *Object) SetField(k string, v Op) {
	o.SetFieldObj(Field{Name: k, Op: v})
}

// SetFieldObj sets a field to a specified description.
// It will override an existing field with the same name if it exists.
func (o *Object) SetFieldObj(f2 Field) {
	if _, ok := o.set[f2.Name]; ok {
		for i, f := range o.fields {
			if f.Name == f2.Name {
				o.fields[i] = f2
				return
			}
		}
	}
	if o.set == nil {
		o.set = make(map[string]struct{})
	}
	o.set[f2.Name] = struct{}{}
	o.fields = append(o.fields, f2)
}
func (o *Object) setFields(fields ...Field) error {
	for _, f := range fields {
		if _, ok := o.set[f.Name]; ok {
			return ErrDuplicateField.New(f.Name)
		}
		o.set[f.Name] = struct{}{}
	}
	return nil
}

// Check will verify that a node is an object and that fields matches a defined set of rules.
//
// If Part transform was not used, this operation will also ensure that all fields in the object are covered by field
// descriptions. If Pre was used, all unknown fields will be saved and restored to a new object on Construct.
//
// For information on optional fields see Field documentation.
func (o Object) Check(st *State, n uast.Node) (bool, error) {
	cur, ok := n.(uast.Object)
	if !ok {
		if errorOnFilterCheck {
			return filtered("%+v is not an object\n%+v", n, o)
		}
		return false, nil
	}
	for _, f := range o.fields {
		n, ok := cur[f.Name]
		if f.Optional != "" {
			if err := st.SetVar(f.Optional, uast.Bool(ok)); err != nil {
				return false, errKey.Wrap(err, f.Name)
			}
		}
		if !ok {
			if f.Optional != "" {
				continue
			}
			if errorOnFilterCheck {
				return filtered("field %+v is missing in %+v\n%+v", f, n, o)
			}
			return false, nil
		}
		ok, err := f.Op.Check(st, n)
		if err != nil {
			return false, errKey.Wrap(err, f.Name)
		} else if !ok {
			return false, nil
		}
	}
	if o.other == "" { // do not save unused fields
		if !allowUnusedFields {
			for k := range cur {
				if _, ok := o.set[k]; !ok {
					return false, ErrUnusedField.New(k)
				}
			}
		}
		return true, nil
	}
	// TODO: consider throwing an error if a transform is defined as partial, but in fact it's not
	left := make(uast.Object)
	for k, v := range cur {
		if _, ok := o.set[k]; !ok {
			left[k] = v
		}
	}
	err := st.SetVar(o.other, left)
	return err == nil, err
}

// Construct will create a new object and will populate it's fields according to field descriptions.
// If Part was used, it will also restore all unhandled fields.
func (o Object) Construct(st *State, old uast.Node) (uast.Node, error) {
	if err := noNode(old); err != nil {
		return nil, err
	}
	obj := make(uast.Object, len(o.fields))
	for _, f := range o.fields {
		if f.Optional != "" {
			on, err := st.MustGetVar(f.Optional)
			if err != nil {
				return obj, errKey.Wrap(err, f.Name)
			}
			exists, ok := on.(uast.Bool)
			if !ok {
				return obj, errKey.Wrap(ErrUnexpectedType.New(uast.Bool(false), on), f.Name)
			}
			if !exists {
				continue
			}
		}
		v, err := f.Op.Construct(st, nil)
		if err != nil {
			return obj, errKey.Wrap(err, f.Name)
		}
		obj[f.Name] = v
	}
	if o.other == "" {
		return obj, nil
	}
	v, err := st.MustGetVar(o.other)
	if err != nil {
		return obj, err
	}
	left, ok := v.(uast.Object)
	if !ok {
		return obj, ErrExpectedObject.New(v)
	}
	for k, v := range left {
		if v2, ok := obj[k]; ok {
			return nil, fmt.Errorf("trying to overwrite already set field with partial pbject data: %q: %v = %v",
				k, v2, v)
		}
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

// Bool asserts that value equals a specific boolean value.
func Bool(val bool) Op {
	return Is(uast.Bool(val))
}

// TypedObj is a shorthand for an object with a specific type
// and multiples operations on it.
func TypedObj(typ string, ops map[string]Op) Op {
	obj := Obj(ops)
	obj[uast.KeyType] = String(typ)
	return obj
}

// ArrayOp is a subset of operations that operates on an arrays with a pre-defined size. See Arr.
type ArrayOp interface {
	Op
	arr(st *State) (opArr, error)
}

// Arr checks if the current object is a list with a number of elements
// matching a number of ops, and applies ops to corresponding elements.
// Reversal creates a list of the size that matches the number of ops
// and creates each element with the corresponding op.
func Arr(ops ...Op) ArrayOp {
	return opArr(ops)
}

type opArr []Op

func (op opArr) arr(_ *State) (opArr, error) {
	return op, nil
}
func (op opArr) Check(st *State, n uast.Node) (bool, error) {
	arr, ok := n.(uast.Array)
	if !ok {
		return filtered("%+v is not a list, %+v", n, op)
	} else if len(arr) != len(op) {
		return filtered("%+v has wrong len for %+v", n, op)
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
	arr := make(uast.Array, 0, len(op))
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
		return false, nil
	}
	vn, ok := op.fwd[v]
	if !ok {
		return false, ErrUnhandledValueIn.New(v, op.fwd)
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
		return nil, ErrUnhandledValueIn.New(v, op.rev)
	}
	return vn, nil
}

// LookupVar is a shorthand to lookup value stored in variable.
func LookupVar(vr string, m map[uast.Value]uast.Value) Op {
	return Lookup(Var(vr), m)
}

// LookupOpVar is a conditional branch that takes a value of a variable and
// checks the map to find an appropriate operation to apply to current node.
// Note that the variable must be defined prior to this transformation, thus
// You might need to use Pre to define a variable used in this condition.
func LookupOpVar(vr string, cases map[uast.Value]Op) Op {
	def := cases[nil]
	delete(cases, nil)
	return opLookupOp{vr: vr, cases: cases, def: def}
}

type opLookupOp struct {
	vr    string
	def   Op
	cases map[uast.Value]Op
}

func (op opLookupOp) eval(st *State) (Op, error) {
	vn, err := st.MustGetVar(op.vr)
	if err != nil {
		return nil, err
	}
	v, ok := vn.(uast.Value)
	if !ok {
		return nil, ErrExpectedValue.New(vn)
	}
	sub, ok := op.cases[v]
	if !ok {
		if op.def == nil {
			return nil, ErrUnhandledValueIn.New(v, op.cases)
		}
		sub = op.def
	}
	return sub, nil
}

func (op opLookupOp) Check(st *State, n uast.Node) (bool, error) {
	sub, err := op.eval(st)
	if err != nil {
		return false, err
	}
	return sub.Check(st, n)
}

func (op opLookupOp) Construct(st *State, n uast.Node) (uast.Node, error) {
	sub, err := op.eval(st)
	if err != nil {
		return nil, err
	}
	return sub.Construct(st, n)
}

// LookupArrOpVar is like LookupOpVar but returns an array operation.
// Default value can be specified by setting the nil key.
func LookupArrOpVar(vr string, cases map[uast.Value]ArrayOp) ArrayOp {
	def := cases[nil]
	delete(cases, nil)
	return opLookupArrOp{vr: vr, cases: cases, def: def}
}

type opLookupArrOp struct {
	vr    string
	def   ArrayOp
	cases map[uast.Value]ArrayOp
}

func (op opLookupArrOp) arr(st *State) (opArr, error) {
	vn, err := st.MustGetVar(op.vr)
	if err != nil {
		return nil, err
	}
	v, ok := vn.(uast.Value)
	if !ok {
		return nil, ErrExpectedValue.New(vn)
	}
	sub, ok := op.cases[v]
	if !ok {
		if op.def == nil {
			return nil, ErrUnhandledValueIn.New(v, op.cases)
		}
		sub = op.def
	}
	return sub.arr(st)
}

func (op opLookupArrOp) Check(st *State, n uast.Node) (bool, error) {
	sub, err := op.arr(st)
	if err != nil {
		return false, err
	}
	return sub.Check(st, n)
}

func (op opLookupArrOp) Construct(st *State, n uast.Node) (uast.Node, error) {
	sub, err := op.arr(st)
	if err != nil {
		return nil, err
	}
	return sub.Construct(st, n)
}

// Append is like AppendArr but allows to set more complex first operation.
// Result of this operation should still be an array.
func Append(to Op, items ...ArrayOp) Op {
	if len(items) == 0 {
		return to
	}
	return opAppend{op: to, arrs: opAppendArr{arrs: items}}
}

type opAppend struct {
	op   Op
	arrs opAppendArr
}

func (op opAppend) Check(st *State, n uast.Node) (bool, error) {
	arr, ok := n.(uast.Array)
	if !ok {
		return filtered("%+v is not a list, %+v", n, op)
	}
	sarr, err := op.arrs.arr(st)
	if err != nil {
		return false, err
	}
	if len(sarr) > len(arr) {
		return filtered("array %+v is too small for %+v", n, op)
	}
	// split into array part that will go to sub op,
	// and the part we will use for sub-array checks
	tail := len(arr) - len(sarr)
	sub, arrs := arr[:tail], arr[tail:]
	if len(sub) == 0 {
		sub = nil
	}
	if ok, err := op.op.Check(st, sub); err != nil {
		return false, errAppend.Wrap(err)
	} else if !ok {
		return false, nil
	}
	return sarr.Check(st, arrs)
}

func (op opAppend) Construct(st *State, n uast.Node) (uast.Node, error) {
	n, err := op.op.Construct(st, n)
	if err != nil {
		return nil, err
	}
	arr, ok := n.(uast.Array)
	if !ok {
		return nil, ErrExpectedList.New(n)
	}
	sarr, err := op.arrs.arr(st)
	if err != nil {
		return nil, err
	}
	nn, err := sarr.Construct(st, nil)
	if err != nil {
		return nil, err
	}
	arr2, ok := nn.(uast.Array)
	if !ok {
		return nil, ErrExpectedList.New(n)
	}
	arr = append(arr, arr2...)
	return arr, nil
}

// AppendArr asserts that a node is a Array and checks that it contains a defined set of nodes at the end.
// Reversal uses sub-operation to create a Array and appends provided element lists at the end of it.
func AppendArr(items ...ArrayOp) ArrayOp {
	if len(items) == 1 {
		return items[0]
	}
	return opAppendArr{arrs: items}
}

type opAppendArr struct {
	arrs []ArrayOp
}

func (op opAppendArr) arr(st *State) (opArr, error) {
	var arr opArr
	for _, sub := range op.arrs {
		a, err := sub.arr(st)
		if err != nil {
			return nil, err
		}
		arr = append(arr, a...)
	}
	return arr, nil
}

func (op opAppendArr) Check(st *State, n uast.Node) (bool, error) {
	sarr, err := op.arr(st)
	if err != nil {
		return false, err
	}
	return sarr.Check(st, n)
}

func (op opAppendArr) Construct(st *State, n uast.Node) (uast.Node, error) {
	sarr, err := op.arr(st)
	if err != nil {
		return nil, err
	}
	return sarr.Construct(st, n)
}

// ValueFunc is a function that transforms values.
type ValueFunc func(uast.Value) (uast.Value, error)

// ValueConv converts a value with a provided function and passes it to sub-operation.
func ValueConv(on Op, conv, rev ValueFunc) Op {
	return opValueConv{op: on, conv: conv, rev: rev}
}

// StringFunc is a function that transforms string values.
type StringFunc func(string) (string, error)

// StringConv is like ValueConv, but only processes string arguments.
func StringConv(on Op, conv, rev StringFunc) Op {
	apply := func(fnc StringFunc) ValueFunc {
		return func(v uast.Value) (uast.Value, error) {
			sv, ok := v.(uast.String)
			if !ok {
				return nil, ErrUnexpectedType.New(uast.String(""), v)
			}
			s, err := fnc(string(sv))
			if err != nil {
				return nil, err
			}
			return uast.String(s), nil
		}
	}
	return ValueConv(on, apply(conv), apply(rev))
}

type opValueConv struct {
	op        Op
	conv, rev ValueFunc
}

func (op opValueConv) Check(st *State, n uast.Node) (bool, error) {
	v, ok := n.(uast.Value)
	if !ok {
		return false, nil
	}
	nv, err := op.conv(v)
	if ErrUnexpectedType.Is(err) {
		return false, nil // skip type mismatch errors on check
	} else if err != nil {
		return false, err
	}
	return op.op.Check(st, nv)
}

func (op opValueConv) Construct(st *State, n uast.Node) (uast.Node, error) {
	n, err := op.op.Construct(st, n)
	if err != nil {
		return nil, err
	}
	v, ok := n.(uast.Value)
	if !ok {
		return nil, ErrExpectedValue.New(n)
	}
	nv, err := op.rev(v)
	if err != nil {
		return nil, err
	}
	return nv, nil
}

// If checks if a named variable value is true and executes one of sub-operations.
func If(cond string, then, els Op) Op {
	return opIf{cond: cond, then: then, els: els}
}

type opIf struct {
	cond      string
	then, els Op
}

func (op opIf) Check(st *State, n uast.Node) (bool, error) {
	st1 := st.Clone()
	ok1, err1 := op.then.Check(st1, n)
	if ok1 && err1 == nil {
		st.ApplyFrom(st1)
		st.SetVar(op.cond, uast.Bool(true))
		return true, nil
	}
	st2 := st.Clone()
	ok2, err2 := op.els.Check(st2, n)
	if ok2 && err2 == nil {
		st.ApplyFrom(st2)
		st.SetVar(op.cond, uast.Bool(false))
		return true, nil
	}
	err := err1
	if err == nil {
		err = err2
	}
	return false, err
}

func (op opIf) Construct(st *State, n uast.Node) (uast.Node, error) {
	vn, err := st.MustGetVar(op.cond)
	if err != nil {
		return nil, err
	}
	cond, ok := vn.(uast.Bool)
	if !ok {
		return nil, ErrUnexpectedType.New(uast.Bool(false), vn)
	}
	if cond {
		return op.then.Construct(st, n)
	}
	return op.els.Construct(st, n)
}

// Each checks that current node is an array and applies sub-operation to each element.
// It uses a variable to store state of each element.
func Each(vr string, op Op) Op {
	return opEach{vr: vr, op: op}
}

type opEach struct {
	vr string
	op Op
}

func (op opEach) Check(st *State, n uast.Node) (bool, error) {
	arr, ok := n.(uast.Array)
	if !ok && n != nil {
		return filtered("%+v is not a list, %+v", n, op)
	}
	var subs []*State
	if arr != nil {
		subs = make([]*State, 0, len(arr))
	}
	for i, sub := range arr {
		sst := NewState()
		ok, err := op.op.Check(sst, sub)
		if err != nil {
			return false, errElem.Wrap(err, i, sub)
		} else if !ok {
			return false, nil
		}
		subs = append(subs, sst)
	}
	if err := st.SetStateVar(op.vr, subs); err != nil {
		return false, err
	}
	return true, nil
}

func (op opEach) Construct(st *State, n uast.Node) (uast.Node, error) {
	if err := noNode(n); err != nil {
		return nil, err
	}
	subs, ok := st.GetStateVar(op.vr)
	if !ok {
		return nil, ErrVariableNotDefined.New(op.vr)
	}
	if subs == nil {
		return nil, nil
	}
	arr := make(uast.Array, 0, len(subs))
	for i, stt := range subs {
		sub, err := op.op.Construct(stt, nil)
		if err != nil {
			return nil, errElem.Wrap(err, i, nil)
		}
		arr = append(arr, sub)
	}
	return arr, nil
}

// NotEmpty checks that node is not nil and contains one or more fields or elements.
func NotEmpty(op Op) Op {
	return opNotEmpty{op: op}
}

type opNotEmpty struct {
	op Op
}

func (op opNotEmpty) Check(st *State, n uast.Node) (bool, error) {
	switch n := n.(type) {
	case nil:
		return filtered("empty value %T for %v", n, op)
	case uast.Array:
		if len(n) == 0 {
			return filtered("empty value %T for %v", n, op)
		}
	case uast.Object:
		if len(n) == 0 {
			return filtered("empty value %T for %v", n, op)
		}
	}
	return op.op.Check(st, n)
}

func (op opNotEmpty) Construct(st *State, n uast.Node) (uast.Node, error) {
	n, err := op.op.Construct(st, n)
	if err != nil {
		return nil, err
	}
	switch n := n.(type) {
	case nil:
		return nil, ErrUnexpectedValue.New(n)
	case uast.Array:
		if len(n) == 0 {
			return nil, ErrUnexpectedValue.New(n)
		}
	case uast.Object:
		if len(n) == 0 {
			return nil, ErrUnexpectedValue.New(n)
		}
	}
	return n, nil
}

// Opt is an optional operation that uses a named variable to store the state.
func Opt(exists string, op Op) Op {
	return opOptional{vr: exists, op: op}
}

type opOptional struct {
	vr string
	op Op
}

func (op opOptional) Check(st *State, n uast.Node) (bool, error) {
	if err := st.SetVar(op.vr, uast.Bool(n != nil)); err != nil {
		return false, err
	}
	if n == nil {
		return true, nil
	}
	return op.op.Check(st, n)
}

func (op opOptional) Construct(st *State, n uast.Node) (uast.Node, error) {
	vn, err := st.MustGetVar(op.vr)
	if err != nil {
		return nil, err
	}
	exists, ok := vn.(uast.Bool)
	if !ok {
		return nil, ErrUnexpectedType.New(uast.Bool(false), vn)
	}
	if !exists {
		return nil, nil
	}
	return op.op.Construct(st, n)
}

// Check tests first check-only operation before applying the main op. It won't use the check-only argument for Construct.
// The check-only operation will not be able to set any variables or change state by other means.
func Check(s Sel, op Op) Op {
	return opCheck{sel: s, op: op}
}

type opCheck struct {
	sel Sel
	op  Op
}

func (op opCheck) Check(st *State, n uast.Node) (bool, error) {
	if ok, err := op.sel.Check(st.Clone(), n); err != nil || !ok {
		return ok, err
	}
	return op.op.Check(st, n)
}

func (op opCheck) Construct(st *State, n uast.Node) (uast.Node, error) {
	return op.op.Construct(st, n)
}

// Not negates the check.
func Not(s Sel) Sel {
	return opNot{sel: s}
}

type opNot struct {
	sel Sel
}

func (op opNot) Check(st *State, n uast.Node) (bool, error) {
	ok, err := op.sel.Check(st.Clone(), n)
	if err != nil {
		return false, err
	}
	return !ok, nil
}

// Any check matches if any of list elements matches sub-check.
func Any(s Sel) Sel {
	if s == nil {
		s = Is(nil)
	}
	return opAny{sel: s}
}

type opAny struct {
	sel Sel
}

func (op opAny) Check(st *State, n uast.Node) (bool, error) {
	l, ok := n.(uast.Array)
	if !ok {
		return false, nil
	}
	for _, o := range l {
		if ok, err := op.sel.Check(st.Clone(), o); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

// All check matches if all list elements matches sub-check.
func All(s Sel) Sel {
	return opAll{sel: s}
}

type opAll struct {
	sel Sel
}

func (op opAll) Check(st *State, n uast.Node) (bool, error) {
	l, ok := n.(uast.Array)
	if !ok {
		return false, nil
	}
	for _, o := range l {
		if ok, err := op.sel.Check(st.Clone(), o); err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}

var _ Sel = Has{}

// Has is a check-only operation that verifies that object has specific fields and they match given checks.
type Has map[string]Sel

// Check verifies that specified fields exists and matches the provided sub-operations.
func (m Has) Check(st *State, n uast.Node) (bool, error) {
	o, ok := n.(uast.Object)
	if !ok {
		return false, nil
	}
	for k, sel := range m {
		v, ok := o[k]
		if !ok {
			return false, nil
		}
		if ok, err := sel.Check(st.Clone(), v); err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}

// In check that the node is a value from a given list.
func In(vals ...uast.Value) Sel {
	m := make(map[uast.Value]struct{}, len(vals))
	for _, v := range vals {
		m[v] = struct{}{}
	}
	return opIn{m: m}
}

type opIn struct {
	m map[uast.Value]struct{}
}

func (op opIn) Check(st *State, n uast.Node) (bool, error) {
	v, ok := n.(uast.Value)
	if !ok && n != nil {
		return false, nil
	}
	_, ok = op.m[v]
	return ok, nil
}
