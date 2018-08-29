package query

import (
	"fmt"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

var _ Iterator = Empty{}

type Empty struct{}

func (Empty) Next() bool           { return false }
func (Empty) Node() nodes.External { return nil }

type IterOrder int

const (
	IterAny = IterOrder(iota)
	PreOrder
	PostOrder
	LevelOrder
	PositionOrder
)

func NewIterator(root nodes.External, order IterOrder) Iterator {
	if root == nil {
		return Empty{}
	}
	if order == IterAny {
		order = PreOrder
	}
	switch order {
	case PreOrder:
		it := &preOrderIter{}
		it.push(root)
		return it
	case PostOrder:
		it := &postOrderIter{}
		it.start(root)
		return it
	case LevelOrder:
		return &levelOrderIter{level: []nodes.External{root}, i: -1}
	default:
		panic(fmt.Errorf("unsupported iterator order: %v", order))
	}
}

func eachChild(n nodes.External, fnc func(v nodes.External)) {
	switch nodes.KindOf(n) {
	case nodes.KindObject:
		if m, ok := n.(nodes.ExternalObject); ok {
			keys := m.Keys()
			for _, k := range keys {
				if v, _ := m.ValueAt(k); v != nil {
					fnc(v)
				}
			}
		}
	case nodes.KindArray:
		if m, ok := n.(nodes.ExternalArray); ok {
			sz := m.Size()
			for i := 0; i < sz; i++ {
				if v := m.ValueAt(i); v != nil {
					fnc(v)
				}
			}
		}
	}
}

func eachChildRev(n nodes.External, fnc func(v nodes.External)) {
	switch nodes.KindOf(n) {
	case nodes.KindObject:
		if m, ok := n.(nodes.ExternalObject); ok {
			keys := m.Keys()
			// reverse order
			for i := len(keys) - 1; i >= 0; i-- {
				if v, _ := m.ValueAt(keys[i]); v != nil {
					fnc(v)
				}
			}
		}
	case nodes.KindArray:
		if m, ok := n.(nodes.ExternalArray); ok {
			sz := m.Size()
			// reverse order
			for i := sz - 1; i >= 0; i-- {
				if v := m.ValueAt(i); v != nil {
					fnc(v)
				}
			}
		}
	}
}

type preOrderIter struct {
	cur nodes.External
	q   []nodes.External
}

func (it *preOrderIter) push(n nodes.External) {
	if n == nil {
		return
	}
	it.q = append(it.q, n)
}
func (it *preOrderIter) pop() nodes.External {
	l := len(it.q)
	if l == 0 {
		return nil
	}
	n := it.q[l-1]
	it.q = it.q[:l-1]
	return n
}

func (it *preOrderIter) Next() bool {
	cur := it.cur
	it.cur = nil
	eachChildRev(cur, it.push)
	it.cur = it.pop()
	return nodes.KindOf(it.cur) != nodes.KindNil
}
func (it *preOrderIter) Node() nodes.External {
	return it.cur
}

type postOrderIter struct {
	cur nodes.External
	s   [][]nodes.External
}

func (it *postOrderIter) start(n nodes.External) {
	kind := nodes.KindOf(n)
	if kind == nodes.KindNil {
		return
	}
	si := len(it.s)
	q := []nodes.External{n}
	it.s = append(it.s, nil)
	eachChildRev(n, func(v nodes.External) {
		q = append(q, v)
	})
	if l := len(q); l > 1 {
		it.start(q[l-1])
		q = q[:l-1]
	}
	it.s[si] = q
}

func (it *postOrderIter) Next() bool {
	down := false
	for {
		l := len(it.s)
		if l == 0 {
			return false
		}
		l--
		top := it.s[l]
		if len(top) == 0 {
			it.s = it.s[:l]
			down = true
			continue
		}
		i := len(top) - 1
		if down && i > 0 {
			down = false
			n := top[i]
			it.s[l] = top[:i]
			it.start(n)
			continue
		}
		down = false
		it.cur = top[i]
		it.s[l] = top[:i]
		return true
	}
}
func (it *postOrderIter) Node() nodes.External {
	return it.cur
}

type levelOrderIter struct {
	level []nodes.External
	i     int
}

func (it *levelOrderIter) Next() bool {
	if len(it.level) == 0 {
		return false
	} else if it.i+1 < len(it.level) {
		it.i++
		return true
	}
	var next []nodes.External
	for _, n := range it.level {
		eachChild(n, func(v nodes.External) {
			next = append(next, v)
		})
	}
	it.i = 0
	it.level = next
	return len(it.level) > 0
}
func (it *levelOrderIter) Node() nodes.External {
	if it.i >= len(it.level) {
		return nil
	}
	return it.level[it.i]
}
