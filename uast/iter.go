package uast

// PathIter iterates node paths.
type PathIter interface {
	// Next returns the next node path or nil if the are no more nodes.
	Next() Path
}

// PathIter iterates node paths, optionally stepping to avoid visiting children
// of some nodes.
type PathStepIter interface {
	PathIter
	// If Step is called, children of the last node returned by Next() will
	// not be visited.
	Step()
}

func newSliceIter(elements ...Path) PathIter {
	return &sliceIter{elements: elements}
}

func newNodeSliceIter(prefix Path, nodes ...*Node) PathIter {
	paths := make([]Path, 0, len(nodes))
	for _, n := range nodes {
		paths = append(paths, append(prefix, n))
	}

	return newSliceIter(paths...)
}

type sliceIter struct {
	idx      int
	elements []Path
}

func (i *sliceIter) Next() Path {
	if i.idx >= len(i.elements) {
		return nil
	}

	n := i.elements[i.idx]
	i.idx++
	return n
}

// NewPreOrderPathIter creates an iterator that iterates all tree nodes in pre-order.
func NewPreOrderPathIter(p Path) PathStepIter {
	return &preOrderPathIter{
		stack: []PathIter{newSliceIter(p)},
	}
}

type preOrderPathIter struct {
	stack []PathIter
	last  Path
}

func (i *preOrderPathIter) Next() Path {
	for {
		if !i.last.IsEmpty() {
			n := i.last.Node()
			if len(n.Children) > 0 {
				i.stack = append(i.stack, newNodeSliceIter(i.last, n.Children...))
			}
		}

		i.last = nil

		if len(i.stack) == 0 {
			break
		}

		cur := i.stack[len(i.stack)-1]
		p := cur.Next()
		if p.IsEmpty() {
			i.stack = i.stack[:len(i.stack)-1]
			continue
		}

		i.last = p
		return p
	}

	return NewPath()
}

func (i *preOrderPathIter) Step() {
	i.last = nil
}
