package diff

import (
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

// Changelist is a list of changes, a result of tree difference. Applying all changes from a
// changelist on a source tree will result in it being transformed into the destination tree.
type Changelist []Change

// Change is a single operation performed
type Change interface {
	isChange()
	TransactionID() uint64
}

type changeBase struct {
	txID uint64
}

func (changeBase) isChange()                {}
func (ch changeBase) TransactionID() uint64 { return ch.txID }

// ID is a type representing node unique ID that can be compared in O(1)
type ID nodes.Comparable

// Key in a node, string for nodes.Object and int for nodes.Array
type Key interface{ isKey() }

// String is a wrapped string type for the Key interface.
type String string

// Int is a wrapped int type for the Key interface.
type Int int

func (Int) isKey()    {}
func (String) isKey() {}

// four change types

// Create a node. Each array and object is created separately.
type Create struct {
	changeBase
	Node nodes.Node
}

// Delete a node by ID
type Delete struct {
	changeBase
	NodeID ID
}

// Attach a node as a child of another node with a given key
type Attach struct {
	changeBase
	Parent ID
	Key    Key
	Child  ID
}

// Deatach a child from a node
type Deatach struct {
	changeBase
	Parent ID
	Key    Key // Currently deatach semantics are only defined for nodes.Object so the Key is
	// practically always a string
}
