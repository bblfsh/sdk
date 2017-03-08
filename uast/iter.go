package uast

type Iter interface {
	// Next returns the next node or nil if the are no more nodes.
	Next() *Node
}

func newSliceIter(elements ...*Node) Iter {
	return &sliceIter{elements: elements}
}

type sliceIter struct {
	idx      int
	elements []*Node
}

func (i *sliceIter) Next() *Node {
	if i.idx >= len(i.elements) {
		return nil
	}

	n := i.elements[i.idx]
	i.idx++
	return n
}

// NewPreOrderIter creates an iterator that iterates all tree nodes in pre-order.
func NewPreOrderIter(n *Node) Iter {
	return &preOrderIter{stack: []Iter{newSliceIter(n)}}
}

type preOrderIter struct {
	stack []Iter
}

func (i *preOrderIter) Next() *Node {
	for {
		if len(i.stack) == 0 {
			break
		}

		cur := i.stack[len(i.stack)-1]
		n := cur.Next()
		if n == nil {
			i.stack = i.stack[:len(i.stack)-1]
			continue
		}

		if len(n.Children) > 0 {
			i.stack = append(i.stack, newSliceIter(n.Children...))
		}

		return n
	}

	return nil
}
