package uast

import (
	"errors"
)

var (
	ErrStop = errors.New("stop iteration")
)

// FindAll finds all nodes that satisfy a predicate.
func FindAll(n *Node, f func(...*Node) bool) []*Node {
	var nodes []*Node
	err := PreOrderVisit(n, func(ns ...*Node) error {
		if f(ns...) {
			nodes = append(nodes, ns[len(ns) - 1])
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
// If the fu nction returns ErrStop, iteration will stop and PreOrderVisit will
// not return an error.
func PreOrderVisit(n *Node, f func(...*Node) error) error {
	err := preOrderVisit(f, n)
	if err == ErrStop {
		return nil
	}

	return err
}

func preOrderVisit(f func(...*Node) error, ns ...*Node) error {
	if err := f(ns...); err != nil {
		return err
	}

	n := ns[len(ns)-1]
	for _, c := range n.Children {
		if err := preOrderVisit(f, append(ns, c)...); err != nil {
			return err
		}
	}

	return nil
}
