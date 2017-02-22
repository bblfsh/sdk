package uast

import (
	"errors"
)

var (
	ErrStop = errors.New("stop iteration")
)

// PreOrderVisit visits a tree in pre-order and applies the given
// function to every node. If the function returns an error, iteration will
// stop and PreOrderVisit will return that error.
// If the fu nction returns ErrStop, iteration will stop and PreOrderVisit will
// not return an error.
func PreOrderVisit(n *Node, f func(*Node) error) error {
	err := preOrderVisit(n, f)
	if err == ErrStop {
		return nil
	}

	return err
}

func preOrderVisit(n *Node, f func(*Node) error) error {
	if err := f(n); err != nil {
		return err
	}

	for _, c := range n.Children {
		if err := preOrderVisit(c, f); err != nil {
			return err
		}
	}

	return nil
}
