package transformer

import (
	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/src-d/go-errors.v1"
)

type Transformer interface {
	Do(n uast.Node) (uast.Node, error)
}

var _ Transformer = (TransformFunc)(nil)

type TransformFunc func(n uast.Node) (uast.Node, bool, error)

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

type CodeTransformer interface {
	OnCode(code string) Transformer
}

var (
	ErrVariableRedeclared = errors.NewKind("variable %q redeclared (%v vs %v)")
	ErrVariableNotDefined = errors.NewKind("variable %q is not defined")
	ErrExpectedObject     = errors.NewKind("expected object, got %T")
	ErrExpectedValue      = errors.NewKind("expected value, got %T")
	ErrUnhandledValue     = errors.NewKind("unhandled value: %v")
	ErrUnexpectedNode     = errors.NewKind("expected node to be nil, got: %v")
	ErrAmbiguousValue     = errors.NewKind("map has ambiguous value %v")

	errAnd  = errors.NewKind("op %d (%T)")
	errKey  = errors.NewKind("key %q")
	errElem = errors.NewKind("elem %d (%T)")
)

func Map(src, dst Op) Mapping {
	return Mapping{src: src, dst: dst}
}

var _ Transformer = Mapping{}

type Mapping struct {
	src, dst Op
}

func (m Mapping) Reverse() Mapping {
	m.src, m.dst = m.dst, m.src
	return m
}
func (m Mapping) Do(n uast.Node) (uast.Node, error) {
	var errs []error
	nn, ok := uast.Apply(n, func(n uast.Node) (uast.Node, bool) {
		st := NewState()
		if ok, err := m.src.Check(st, n); err != nil {
			errs = append(errs, err)
			return n, false
		} else if !ok {
			return n, false
		}
		nn, err := m.dst.Construct(st, nil)
		if err != nil {
			errs = append(errs, err)
			return n, false
		}
		return nn, true
	})
	var first error
	if len(errs) != 0 {
		first = errs[0] // TODO: return multi-error
	}
	if ok {
		return nn, first
	}
	return n, first
}

// NewState creates a new state for Ops to work on.
// It stores variables, flags and anything that necessary
// for transformation steps to persist data.
func NewState() *State {
	return &State{
		vars: make(map[string]uast.Node),
	}
}

type State struct {
	vars map[string]uast.Node
}

func (st *State) GetVar(name string) (uast.Node, bool) {
	n, ok := st.vars[name]
	return n, ok
}

func (st *State) SetVar(name string, val uast.Node) error {
	cur, ok := st.vars[name]
	if !ok {
		// not declared
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

type Sel interface {
	Check(st *State, n uast.Node) (bool, error)
}

type Mod interface {
	Construct(st *State, n uast.Node) (uast.Node, error)
}

type Op interface {
	Sel
	Mod
}
