package uast

import (
	"errors"
)

var (
	ErrStop    = errors.New("stop iteration")
	ErrNoVisit = errors.New("do not visit")
)

// FindAll finds all nodes that satisfy a predicate.
func FindAll(n *Node, f func(Path) bool) []*Node {
	var nodes []*Node
	err := PreOrderVisit(n, func(ns Path) error {
		if f(ns) {
			nodes = append(nodes, ns.Node())
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	return nodes
}

// PreOrderVisit visits a tree in pre-order and applies the given
// function to every node path. If the function returns an error, iteration will
// stop and PreOrderVisit will return that error.
// If the function returns ErrStop, iteration will stop and PreOrderVisit will
// not return an error.
// If the function returns ErrNoVisit, children of the current node will not be
// visited.
func PreOrderVisit(n *Node, f func(Path) error) error {
	err := preOrderVisit(f, NewPath(n))
	if err == ErrStop {
		return nil
	}

	return err
}

func preOrderVisit(f func(Path) error, ns Path) error {
	if err := f(ns); err != nil {
		if err == ErrNoVisit {
			return nil
		}

		return err
	}

	n := ns[len(ns)-1]
	for _, c := range n.Children {
		if err := preOrderVisit(f, append(ns, c)); err != nil {
			return err
		}
	}

	return nil
}
