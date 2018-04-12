package transformer

import (
	"fmt"

	"gopkg.in/bblfsh/sdk.v1/uast"
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
	return Mapping{Name: name, Steps: []Step{
		{Name: "src", Op: src},
		{Name: "dst", Op: dst},
	}}
}

var _ Transformer = Mapping{}

// Step is a single transformation step. See Mapping.
type Step struct {
	Name string
	Op   Op
}

// Mapping is a set of transformation steps executed in order.
type Mapping struct {
	Name  string
	Steps []Step
}

// Reverse changes a transformation direction, allowing to construct the source tree.
func (m Mapping) Reverse() Mapping {
	n := len(m.Steps)
	steps := make([]Step, n)
	for i, s := range m.Steps {
		steps[n-1-i] = s
	}
	m.Steps = steps
	return m
}

func applyMap(src, dst Op, root uast.Node) (uast.Node, error) {
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
	if len(m.Steps) <= 1 {
		return n, ErrFewSteps.New()
	}
	steps := m.Steps
	var err error
	for len(steps) >= 2 {
		src, dst := steps[0], steps[1]
		n, err = applyMap(src.Op, dst.Op, n)
		if err != nil {
			return n, errMapping.Wrap(err, m.Name)
		}
		steps = steps[1:]
	}
	return n, err
}

// Mappings takes multiple mappings and optimizes the process of applying them as a single transformation.
func Mappings(maps ...Mapping) Transformer {
	if len(maps) == 0 {
		return mappings{}
	}
	names := make([]string, 0, len(maps))
	steps := make([]multiStep, 0, len(maps[0].Steps))
	typedAll := make([][]Op, len(steps))
	for _, st := range maps[0].Steps {
		ops := make([]Op, 0, len(maps))
		steps = append(steps, multiStep{
			name:     st.Name,
			all:      ops,
			typedObj: make(map[string][]Op),
		})
	}
	for _, m := range maps {
		names = append(names, m.Name)
		if len(m.Steps) != len(steps) {
			panic(fmt.Errorf("wrong number of steps for %q", m.Name))
		}
		for j, st := range m.Steps {
			if steps[j].name != st.Name {
				panic(fmt.Errorf("unexpected step %q.%q", m.Name, st.Name))
			}
			op := st.Op
			// pre-compile object operations (sort fields for unordered ops, etc)
			// TODO: recurse somehow
			if oop, ok := op.(ObjectOp); ok {
				if _, ok := op.(Object); !ok {
					op = oop.Object()
				}
			}
			// all operations
			steps[j].all = append(steps[j].all, op)
			// switch by operation type and make a separate list
			// next time we will see a node with matching type, we will apply only specific ops
			switch op := op.(type) {
			case ObjectOp:
				steps[j].objs = append(steps[j].objs, op)
				if f, _ := op.Object().GetField(uast.KeyType); f.Optional == "" {
					if is, ok := f.Op.(opIs); ok {
						if typ, ok := is.v.(uast.String); ok {
							m := steps[j].typedObj
							s := string(typ)
							m[s] = append(m[s], op)
						}
					}
				}
			case ArrayOp:
				steps[j].arrs = append(steps[j].arrs, op)
			default:
				steps[j].others = append(steps[j].others, op)
				// the type is unknown, thus we should try to apply it to objects and array as well
				typedAll[j] = append(typedAll[j], op)
				steps[j].objs = append(steps[j].objs, op)
				steps[j].arrs = append(steps[j].arrs, op)
			}
		}
	}
	for j, arr := range typedAll {
		if len(arr) == 0 {
			continue
		}
		m := steps[j].typedObj
		for typ, ops := range m {
			m[typ] = append(ops, arr...)
		}
	}
	return mappings{names: names, steps: steps}
}

type multiStep struct {
	name     string
	objs     []Op
	typedObj map[string][]Op
	typedAll []Op
	arrs     []Op
	others   []Op
	all      []Op
}

type mappings struct {
	names []string
	steps []multiStep
}

func (m mappings) apply(msrc, mdst multiStep, root uast.Node) (uast.Node, error) {
	var errs []error
	st := NewState()
	nn, ok := uast.Apply(root, func(old uast.Node) (uast.Node, bool) {
		src, dst := msrc.all, mdst.all
		switch old := old.(type) {
		case nil:
			// apply all
		case uast.Object:
			src, dst = msrc.objs, mdst.objs
			if typ, ok := old[uast.KeyType].(uast.String); ok {
				if ops, ok := msrc.typedObj[string(typ)]; ok {
					src = ops
				}
				if ops, ok := mdst.typedObj[string(typ)]; ok {
					dst = ops
				}
			}
		case uast.Array:
			src, dst = msrc.arrs, mdst.arrs
		default:
			src, dst = msrc.others, mdst.others
		}

		n := old
		applied := false
		for i, src := range src {
			dst := dst[i]
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
func (m mappings) Do(n uast.Node) (uast.Node, error) {
	if len(m.steps) <= 1 {
		return n, ErrFewSteps.New()
	}
	steps := m.steps
	var err error
	for len(steps) >= 2 {
		src, dst := steps[0], steps[1]
		n, err = m.apply(src, dst, n)
		if err != nil {
			return n, errMapping.Wrap(err, dst.name)
		}
		steps = steps[1:]
	}
	return n, err
}

// NewState creates a new state for Ops to work on.
// It stores variables, flags and anything that necessary
// for transformation steps to persist data.
func NewState() *State {
	return &State{}
}

// State stores all variables (placeholder values, flags and wny other state) between Check and Construct steps.
type State struct {
	vars   map[string]uast.Node
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
		st2.vars = make(map[string]uast.Node)
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
		st.vars = make(map[string]uast.Node)
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

// SetVar sets a named variable. It will return ErrVariableRedeclared if a variable with the same name is already set.
// It will ignore the operation if variable already exists and has the same value (uast.Value).
func (st *State) SetVar(name string, val uast.Node) error {
	cur, ok := st.vars[name]
	if !ok {
		// not declared
		if st.vars == nil {
			st.vars = make(map[string]uast.Node)
		}
		st.vars[name] = val
		return nil
	}
	v1, ok1 := cur.(uast.Value)
	v2, ok2 := val.(uast.Value)
	// the only exception is two primitive values that are equal
	if ok1 && ok2 && v1 == v2 {
		// already declared, and value is alredy in the map
		return nil
	}
	return ErrVariableRedeclared.New(name, cur, val)
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
