package nodes

import "fmt"

// Iterator over nodes.
type Iterator interface {
	// Next advances an iterator.
	Next() bool
	// Node returns a current node.
	Node() External
}

var _ Iterator = Empty{}

// Empty is an empty iterator.
type Empty struct{}

// Next implements Iterator.
func (Empty) Next() bool { return false }

// Node implements Iterator.
func (Empty) Node() External { return nil }

// IterOrder is a tree iteration order.
type IterOrder int

const (
	// IterAny is a native iteration order of the tree. It's the fastest iteration order that lists all nodes in the tree.
	// The iteration order is not guaranteed to be the same for consecutive iterations over the same tree.
	// This order is more suitable for searching for nodes in the fastest way possible.
	IterAny = IterOrder(iota)
	// PreOrder is a pre-order depth-first search.
	PreOrder
	// PostOrder is a post-order depth-first search.
	PostOrder
	// LevelOrder is a breadth-first search.
	LevelOrder
	// ChildrenOrder is similar to LevelOrder, but list only the first level.
	ChildrenOrder
)

// NewIterator creates a new iterator with a given order.
func NewIterator(root External, order IterOrder) Iterator {
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
		return &levelOrderIter{level: []External{root}, i: -1}
	case ChildrenOrder:
		return newChildrenIterator(root)
	default:
		panic(fmt.Errorf("unsupported iterator order: %v", order))
	}
}

func eachChild(n External, fnc func(v External)) {
	switch KindOf(n) {
	case KindObject:
		if m, ok := n.(ExternalObject); ok {
			keys := m.Keys()
			for _, k := range keys {
				if v, _ := m.ValueAt(k); v != nil {
					fnc(v)
				}
			}
		}
	case KindArray:
		if m, ok := n.(ExternalArray); ok {
			sz := m.Size()
			for i := 0; i < sz; i++ {
				if v := m.ValueAt(i); v != nil {
					fnc(v)
				}
			}
		}
	}
}

func eachChildRev(n External, fnc func(v External)) {
	switch KindOf(n) {
	case KindObject:
		if m, ok := n.(ExternalObject); ok {
			keys := m.Keys()
			// reverse order
			for i := len(keys) - 1; i >= 0; i-- {
				if v, _ := m.ValueAt(keys[i]); v != nil {
					fnc(v)
				}
			}
		}
	case KindArray:
		if m, ok := n.(ExternalArray); ok {
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
	cur External
	q   []External
}

func (it *preOrderIter) push(n External) {
	if n == nil {
		return
	}
	it.q = append(it.q, n)
}
func (it *preOrderIter) pop() External {
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
	return KindOf(it.cur) != KindNil
}
func (it *preOrderIter) Node() External {
	return it.cur
}

type postOrderIter struct {
	cur External
	s   [][]External
}

func (it *postOrderIter) start(n External) {
	kind := KindOf(n)
	if kind == KindNil {
		return
	}
	si := len(it.s)
	q := []External{n}
	it.s = append(it.s, nil)
	eachChildRev(n, func(v External) {
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
func (it *postOrderIter) Node() External {
	return it.cur
}

type levelOrderIter struct {
	level []External
	i     int
}

func (it *levelOrderIter) Next() bool {
	if len(it.level) == 0 {
		return false
	} else if it.i+1 < len(it.level) {
		it.i++
		return true
	}
	var next []External
	for _, n := range it.level {
		eachChild(n, func(v External) {
			next = append(next, v)
		})
	}
	it.i = 0
	it.level = next
	return len(it.level) > 0
}

func (it *levelOrderIter) Node() External {
	if it.i >= len(it.level) {
		return nil
	}
	return it.level[it.i]
}

func addUnfoldingArrays(nodes []External, n External) []External {
	switch n.Kind() {
	case KindArray:
		if m, ok := n.(ExternalArray); ok {
			sz := m.Size()
			for i := 0; i < sz; i++ {
				if v := m.ValueAt(i); v != nil {
					nodes = addUnfoldingArrays(nodes, v)
				}
			}
		}
	case KindObject:
		nodes = append(nodes, n)
	}
	return nodes
}

func newChildrenIterator(n External) Iterator {
	var nodes []External
	eachChild(n, func(v External) {
		nodes = addUnfoldingArrays(nodes, v)
	})
	return newFixedIterator(nodes)
}

// newFixedIterator creates a node iterator that list nodes in the given slice. It won't recurse into those nodes.
func newFixedIterator(nodes []External) Iterator {
	return &fixedIter{nodes: nodes, first: true}
}

type fixedIter struct {
	nodes []External
	first bool
}

// Next implements Iterator.
func (it *fixedIter) Next() bool {
	if it.first {
		it.first = false
		return len(it.nodes) > 0
	} else if len(it.nodes) == 0 {
		return false
	}
	it.nodes = it.nodes[1:]
	return len(it.nodes) > 0
}

// Node implements Iterator.
func (it *fixedIter) Node() External {
	if len(it.nodes) == 0 {
		return nil
	}
	return it.nodes[0]
}
