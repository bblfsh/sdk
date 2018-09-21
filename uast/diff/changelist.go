package diff

import (
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

type Changelist []Change

type Change interface {
	isChange()
	TransactionID() uint64
}

type changeBase struct {
	txID uint64
}

func (changeBase) isChange()                {}
func (ch changeBase) TransactionID() uint64 { return ch.txID }

// TODO: proper ID of a node somehow
type ID nodes.Comparable

// key in a node, string for nodes.Object and int for nodes.Array
type Key interface{ isKey() }

type String string
type Int int

func (Int) isKey()    {}
func (String) isKey() {}

// four change types

// Create a node. Each array and object is created separately.
type Create struct {
	changeBase
	Node nodes.Node
}

// delete a node by ID
type Delete struct {
	changeBase
	NodeID ID
}

// attach a node as a child of another node with a given key
type Attach struct {
	changeBase
	Parent ID
	Key    Key
	Child  ID
}

// deattach a child from a node
type Deattach struct {
	changeBase
	Parent ID
	Key    Key //or string, how to do alternative?
}
