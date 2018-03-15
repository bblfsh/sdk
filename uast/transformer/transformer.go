package transformer

import (
	"fmt"

	//"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/uast"
)

//type Tranformer interface {
//	Do(code string, e protocol.Encoding, n *uast.Node) error
//}

func Map(src, dst Op) Mapping {
	return Mapping{src: src, dst: dst}
}

// TODO: rename to Transformer?
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
	return fmt.Errorf("variable %q is redeclared (%v vs %v)", name, cur, val)
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
