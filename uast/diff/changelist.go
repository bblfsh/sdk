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
	// GroupID is the same for multiple changes that are a part of one high-level operation.
	GroupID() uint64
}

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
	group uint64

	// Node to create.
	Node nodes.Node
}

func (Create) isChange() {}

func (ch Create) GroupID() uint64 { return ch.group }

// Delete a node by ID
type Delete struct {
	group uint64

	// Node to delete.
	Node ID
}

func (Delete) isChange() {}

func (ch Delete) GroupID() uint64 { return ch.group }

// Attach a node as a child of another node with a given key
type Attach struct {
	group uint64

	// Parent node ID to attach the key to.
	Parent ID
	// Key to attach.
	Key Key
	// Child to attach on a given key.
	Child ID
}

func (Attach) isChange() {}

func (ch Attach) GroupID() uint64 { return ch.group }

// Detach a child from a node
type Detach struct {
	group uint64

	// Parent node ID to detach the key from.
	Parent ID
	// Key to detach. Currently detach semantics are only defined for nodes.Object so the
	// Key is practically always a String.
	Key Key
}

func (Detach) isChange() {}

func (ch Detach) GroupID() uint64 { return ch.group }
