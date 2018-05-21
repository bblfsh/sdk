package transformer

import (
	"sort"
	"strings"

	"gopkg.in/bblfsh/sdk.v2/uast"
)

// Transformer is an interface for transformations that operates on AST trees.
// An implementation is responsible for walking the tree and executing transformation on each AST node.
type Transformer interface {
	Do(root uast.Node) (uast.Node, error)
}

// CodeTransformer is a special case of Transformer that needs an original source code to operate.
type CodeTransformer interface {
	OnCode(code string) Transformer
}

// Sel is an operation that can verify if a specific node matches a set of constraints or not.
type Sel interface {
	// Check will verify constraints for a single node and returns true if an objects matches them.
	// It can also populate the State with variables that can be used later to Construct a different object from the State.
	Check(st *State, n uast.Node) (bool, error)
}

// Mod is an operation that can reconstruct an AST node from a given State.
type Mod interface {
	// Construct will use variables stored in State to reconstruct an AST node.
	// Node that is provided as an argument may be used as a base for reconstruction.
	Construct(st *State, n uast.Node) (uast.Node, error)
}

// Op is a generic AST transformation step that describes a shape of an AST tree.
// It can be used to either check the constraints for a specific node and populate state, or to reconstruct an AST shape
// from a the same state (probably produced by another Op).
type Op interface {
	Sel
	Mod
}

// Transformers appends all provided transformer slices into single one.
func Transformers(arr ...[]Transformer) []Transformer {
	var out []Transformer
	for _, a := range arr {
		out = append(out, a...)
	}
	return out
}

var _ Transformer = (TransformFunc)(nil)

// TransformFunc is a function that will be applied to each AST node to transform the tree.
// It returns a new AST and true if tree was changed, or an old node and false if no modifications were done.
// The the tree will be traversed automatically and the callback will be called for each node.
type TransformFunc func(n uast.Node) (uast.Node, bool, error)

// Do runs a transformation function for each AST node.
func (f TransformFunc) Do(n uast.Node) (uast.Node, error) {
	var last error
	nn, ok := uast.Apply(n, func(n uast.Node) (uast.Node, bool) {
		nn, ok, err := f(n)
		if err != nil {
			last = err
			return n, false
		} else if !ok {
			return n, false
		}
		return nn, ok
	})
	if ok {
		return nn, last
	}
	return n, last
}

var _ Transformer = (TransformObjFunc)(nil)

// TransformObjFunc is like TransformFunc, but only matches Object nodes.
type TransformObjFunc func(n uast.Object) (uast.Object, bool, error)

// Func converts this TransformObjFunc to a regular TransformFunc by skipping all non-object nodes.
func (f TransformObjFunc) Func() TransformFunc {
	return TransformFunc(func(n uast.Node) (uast.Node, bool, error) {
		obj, ok := n.(uast.Object)
		if !ok {
			return n, false, nil
		}
		nn, ok, err := f(obj)
		if err != nil {
			return n, false, err
		} else if !ok {
			return n, false, nil
		}
		return nn, ok, nil
	})
}

// Do runs a transformation function for each AST node.
func (f TransformObjFunc) Do(n uast.Node) (uast.Node, error) {
	return f.Func().Do(n)
}

// Map creates a two-way mapping between two transform operations.
// The first operation will be used to check constraints for each node and store state, while the second one will use
// the state to construct a new tree.
func Map(name string, src, dst Op) Mapping {
	return Mapping{name: name, src: src, dst: dst}
}

var _ Transformer = Mapping{}

// Mapping is a set of transformation steps executed in order.
type Mapping struct {
	name     string
	src, dst Op
}

// Reverse changes a transformation direction, allowing to construct the source tree.
func (m Mapping) Reverse() Mapping {
	m.src, m.dst = m.dst, m.src
	return m
}

func (m Mapping) apply(root uast.Node) (uast.Node, error) {
	src, dst := m.src, m.dst
	var errs []error
	_, objOp := src.(ObjectOp)
	_, arrOp := src.(ArrayOp)
	st := NewState()
	nn, ok := uast.Apply(root, func(n uast.Node) (uast.Node, bool) {
		if n != nil {
			if objOp {
				if _, ok := n.(uast.Object); !ok {
					return n, false
				}
			} else if arrOp {
				if _, ok := n.(uast.Array); !ok {
					return n, false
				}
			}
		}
		st.Reset()
		if ok, err := src.Check(st, n); err != nil {
			errs = append(errs, errCheck.Wrap(err))
			return n, false
		} else if !ok {
			return n, false
		}
		nn, err := dst.Construct(st, nil)
		if err != nil {
			errs = append(errs, errConstruct.Wrap(err))
			return n, false
		}
		return nn, true
	})
	err := NewMultiError(errs...)
	if ok {
		return nn, err
	}
	return root, err
}

// Do will traverse the whole tree and will apply transformation steps for each node.
func (m Mapping) Do(n uast.Node) (uast.Node, error) {
	nn, err := m.apply(n)
	if err != nil {
		return n, errMapping.Wrap(err, m.name)
	}
	return nn, nil
}

// Mappings takes multiple mappings and optimizes the process of applying them as a single transformation.
func Mappings(maps ...Mapping) Transformer {
	if len(maps) == 0 {
		return mappings{}
	} else if len(maps) == 1 {
		return maps[0]
	}
	mp := mappings{
		all: maps,
	}
	mp.index()
	return mp
}

type mappings struct {
	all []Mapping

	// indexed mappings

	objs   []Mapping // mappings applied to objects
	arrs   []Mapping // mappings applied to arrays
	others []Mapping // mappings to other types

	typedObj map[string][]Mapping // mappings for objects with specific type
	typedAny []Mapping            // mappings for any typed object (operations that does not mention the type)
}

func (m *mappings) index() {
	precompile := func(op Op) Op {
		// TODO: recurse somehow
		if oop, ok := op.(ObjectOp); ok {
			if _, ok := op.(Object); !ok {
				return oop.Object()
			}
		}
		return op
	}
	type ordered struct {
		ind int
		mp  Mapping
	}
	var typedAny []ordered
	typed := make(map[string][]ordered)
	for i, mp := range m.all {
		// pre-compile object operations (sort fields for unordered ops, etc)
		mp.src, mp.dst = precompile(mp.src), precompile(mp.dst)

		oop := mp.src
		if chk, ok := oop.(opCheck); ok {
			oop = chk.op
		}
		// switch by operation type and make a separate list
		// next time we will see a node with matching type, we will apply only specific ops
		switch op := oop.(type) {
		case ObjectOp:
			m.objs = append(m.objs, mp)
			specific := false
			if f, _ := op.Object().GetField(uast.KeyType); f.Optional == "" {
				if is, ok := f.Op.(opIs); ok {
					if typ, ok := is.v.(uast.String); ok {
						s := string(typ)
						typed[s] = append(typed[s], ordered{ind: i, mp: mp})
						specific = true
					}
				}
			}
			if !specific {
				typedAny = append(typedAny, ordered{ind: i, mp: mp})
			}
		case ArrayOp:
			m.arrs = append(m.arrs, mp)
		default:
			m.others = append(m.others, mp)
			// the type is unknown, thus we should try to apply it to objects and array as well
			typedAny = append(typedAny, ordered{ind: i, mp: mp})
			m.objs = append(m.objs, mp)
			m.arrs = append(m.arrs, mp)
		}
	}
	m.typedObj = make(map[string][]Mapping, len(typed))
	for typ, ord := range typed {
		ord = append(ord, typedAny...)
		sort.Slice(ord, func(i, j int) bool {
			return ord[i].ind < ord[j].ind
		})
		maps := make([]Mapping, 0, len(ord))
		for _, o := range ord {
			maps = append(maps, o.mp)
		}
		m.typedObj[typ] = maps
	}
}

func (m mappings) Do(root uast.Node) (uast.Node, error) {
	var errs []error
	st := NewState()
	nn, ok := uast.Apply(root, func(old uast.Node) (uast.Node, bool) {
		maps := m.all
		switch old := old.(type) {
		case nil:
			// apply all
		case uast.Object:
			maps = m.objs
			if typ, ok := old[uast.KeyType].(uast.String); ok {
				if mp, ok := m.typedObj[string(typ)]; ok {
					maps = mp
				}
			}
		case uast.Array:
			maps = m.arrs
		default:
			maps = m.others
		}

		n := old
		applied := false
		for _, mp := range maps {
			src, dst := mp.src, mp.dst
			st.Reset()
			if ok, err := src.Check(st, n); err != nil {
				errs = append(errs, errCheck.Wrap(err))
				continue
			} else if !ok {
				continue
			}
			applied = true
			nn, err := dst.Construct(st, nil)
			if err != nil {
				errs = append(errs, errConstruct.Wrap(err))
				continue
			}
			n = nn
		}

		if !applied {
			return old, false
		}
		return n, true
	})
	err := NewMultiError(errs...)
	if ok {
		return nn, err
	}
	return root, err
}

// NewState creates a new state for Ops to work on.
// It stores variables, flags and anything that necessary
// for transformation steps to persist data.
func NewState() *State {
	return &State{}
}

// Vars is a set of variables with their values.
type Vars map[string]uast.Node

// State stores all variables (placeholder values, flags and wny other state) between Check and Construct steps.
type State struct {
	vars   Vars
	states map[string][]*State
}

// Reset clears the state and allows to reuse an object.
func (st *State) Reset() {
	st.vars = nil
	st.states = nil
}

// Clone will return a copy of the State. This can be used to apply Check and throw away any variables produced by it.
// To merge a cloned state back use ApplyFrom on a parent state.
func (st *State) Clone() *State {
	st2 := NewState()
	if len(st.vars) != 0 {
		st2.vars = make(Vars)
	}
	for k, v := range st.vars {
		st2.vars[k] = v
	}
	if len(st.states) != 0 {
		st2.states = make(map[string][]*State)
	}
	for k, v := range st.states {
		st2.states[k] = v
	}
	return st2
}

// ApplyFrom merges a provided state into this state object.
func (st *State) ApplyFrom(st2 *State) {
	if len(st2.vars) != 0 && st.vars == nil {
		st.vars = make(Vars)
	}
	for k, v := range st2.vars {
		if _, ok := st.vars[k]; !ok {
			st.vars[k] = v
		}
	}
	if len(st2.states) != 0 && st.states == nil {
		st.states = make(map[string][]*State)
	}
	for k, v := range st2.states {
		if _, ok := st.states[k]; !ok {
			st.states[k] = v
		}
	}
}

// GetVar looks up a named variable.
func (st *State) GetVar(name string) (uast.Node, bool) {
	n, ok := st.vars[name]
	return n, ok
}

// MustGetVar looks up a named variable and returns ErrVariableNotDefined in case it does not exists.
func (st *State) MustGetVar(name string) (uast.Node, error) {
	n, ok := st.GetVar(name)
	if !ok {
		return nil, ErrVariableNotDefined.New(name)
	}
	return n, nil
}

// VarsPtrs is a set of variable pointers.
type VarsPtrs map[string]uast.NodePtr

// MustGetVars is like MustGetVar but fetches multiple variables in one operation.
func (st *State) MustGetVars(vars VarsPtrs) error {
	for name, dst := range vars {
		n, ok := st.GetVar(name)
		if !ok {
			return ErrVariableNotDefined.New(name)
		}
		if err := dst.SetNode(n); err != nil {
			return err
		}
	}
	return nil
}

// SetVar sets a named variable. It will return ErrVariableRedeclared if a variable with the same name is already set.
// It will ignore the operation if variable already exists and has the same value (uast.Value).
func (st *State) SetVar(name string, val uast.Node) error {
	cur, ok := st.vars[name]
	if !ok {
		// not declared
		if st.vars == nil {
			st.vars = make(Vars)
		}
		st.vars[name] = val
		return nil
	}
	if uast.Equal(cur, val) {
		// already declared, and the same value is already in the map
		return nil
	}
	return ErrVariableRedeclared.New(name, cur, val)
}

// SetVars is like SetVar but sets multiple variables in one operation.
func (st *State) SetVars(vars Vars) error {
	for k, v := range vars {
		if err := st.SetVar(k, v); err != nil {
			return err
		}
	}
	return nil
}

// GetStateVar returns a stored sub-state from a named variable.
func (st *State) GetStateVar(name string) ([]*State, bool) {
	n, ok := st.states[name]
	return n, ok
}

// SetStateVar sets a sub-state variable. It returns ErrVariableRedeclared if the variable with this name already exists.
func (st *State) SetStateVar(name string, sub []*State) error {
	cur, ok := st.states[name]
	if ok {
		return ErrVariableRedeclared.New(name, cur, sub)
	}
	if st.states == nil {
		st.states = make(map[string][]*State)
	}
	st.states[name] = sub
	return nil
}

// DefaultNamespace is a transform that sets a specified namespace for predicates and values that doesn't have a namespace.
func DefaultNamespace(ns string) Transformer {
	return TransformFunc(func(n uast.Node) (uast.Node, bool, error) {
		obj, ok := n.(uast.Object)
		if !ok {
			return n, false, nil
		}
		tp, ok := obj[uast.KeyType].(uast.String)
		if !ok {
			return n, false, nil
		}
		if strings.Contains(string(tp), ":") {
			return n, false, nil
		}
		obj = obj.CloneObject()
		obj[uast.KeyType] = uast.String(ns + ":" + string(tp))
		return obj, true, nil
	})
}
