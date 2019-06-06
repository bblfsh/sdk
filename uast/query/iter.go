package query

import (
	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
)

type Empty = nodes.Empty

type IterOrder = nodes.IterOrder

const (
	IterAny       = nodes.IterAny
	PreOrder      = nodes.PreOrder
	PostOrder     = nodes.PostOrder
	LevelOrder    = nodes.LevelOrder
	ChildrenOrder = nodes.ChildrenOrder
	PositionOrder = ChildrenOrder + iota + 1
)

// NewIterator creates a new iterator with a given order.
// It's an extension of nodes.NewIterator that additionally supports PositionOrder.
func NewIterator(root nodes.External, order IterOrder) Iterator {
	if root == nil {
		return Empty{}
	}
	switch order {
	case PositionOrder:
		return uast.NewPositionalIterator(root)
	}
	return nodes.NewIterator(root, order)
}
