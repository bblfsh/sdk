package uast

import (
	"fmt"
	"strings"

	"gopkg.in/bblfsh/sdk.v1/uast/role"
)

// PathIter iterates node paths.
type PathIter interface {
	// Next returns the next node path or nil if the are no more nodes.
	Next() Path
}

// PathStepIter iterates node paths, optionally stepping to avoid visiting children
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

func newNodeSliceIter(prefix Path, nodes ...Node) PathIter {
	paths := make([]Path, 0, len(nodes))
	for _, n := range nodes {
		paths = append(paths, prefix.Child(n))
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

type orderPathIter struct {
	stack []PathIter
	last  Path
}

// NewOrderPathIter creates an iterator that iterates all tree nodes (by default it
// will use preorder traversal but will switch to inorder or postorder if the Infix and
// Postfix roles are found).
func NewOrderPathIter(p Path) PathStepIter {
	return &orderPathIter{
		stack: []PathIter{newSliceIter(p)},
	}
}

const (
	preOrder = iota
	inOrder
	postOrder
)

func getNextIterType(n Node) int {
	var order int
	o, _ := n.(Object)
	for _, r := range o.Roles() {
		switch r {
		case role.Infix:
			order = inOrder
		case role.Postfix:
			order = postOrder
		default:
			order = preOrder
		}
	}

	return order
}

// Make a copy of the Node removing the children. Used to
// add nodes with the InOrder or PostOrder roles to the stack
// when their children have been already added
func noChildrenNodeCopy(n Object) Object {
	m := make(Object)
	for k, v := range n {
		// only clone attributes and special fields
		if _, ok := v.(Value); ok || strings.HasPrefix(k, "@") {
			m[k] = v
		}
	}
	return m
}

// Adds to the orderPathIter stack with the right order depending on
// the order Role with (if set) can be Infix, Postfix or Prefix. Defaults to Preorder
// if the order Role is not set. This also updates i.last.
func (i *orderPathIter) addToStackWithOrder(n Node) {
	var children []Node
	switch n := n.(type) {
	case Object:
		children = make([]Node, 0, len(n))
		for _, k := range n.Keys() {
			children = append(children, n[k])
		}
	case Array:
		children = n
	default:
		return
	}

	switch getNextIterType(n) {
	case inOrder:
		// Right
		if l := len(children); l != 2 {
			panic(fmt.Sprintf("unsupported iteration over node with %d children", l))
		}
		i.stack = append(i.stack, newNodeSliceIter(i.last, children[1]))
		if obj, ok := n.(Object); ok {
			// Relator
			i.stack = append(i.stack, newNodeSliceIter(i.last, noChildrenNodeCopy(obj)))
		}
		// left
		i.stack = append(i.stack, newNodeSliceIter(i.last, children[0]))
	case postOrder:
		if obj, ok := n.(Object); ok {
			// Children
			i.stack = append(i.stack, newNodeSliceIter(i.last, noChildrenNodeCopy(obj)))
		}
		// Relator
		i.stack = append(i.stack, newNodeSliceIter(i.last, children...))
	default:
		// no order role or (default) preOrder
		// (children not added to the stack):
		i.stack = append(i.stack, newNodeSliceIter(i.last, children...))
	}
}

func (i *orderPathIter) Next() Path {
	for {
		if !i.last.IsEmpty() {
			n := i.last.Node()
			i.addToStackWithOrder(n)
		}

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
		n := p.Node()
		if _, ok := n.(Value); ok {
			continue
		}

		// Check if the item has the role inOrder or postOrder and have children; in that
		// case skip it since the children and the (childless) copy of the node have already
		// been added in addToStackWithOrder in the correct order
		iterType := getNextIterType(n)
		obj, _ := n.(Object)
		if (iterType == inOrder || iterType == postOrder) && len(obj) == 0 {
			continue
		}

		return p
	}

	return NewPath()
}

func (i *orderPathIter) Step() {
	i.last = nil
}
