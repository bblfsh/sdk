package diff

import (
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

type Changelist []Change

type Change interface {
	isChange()
}

type changeBase struct {
	txID uint64
}

func (_ *changeBase) isChange() {}

// TODO: proper ID of a node somehow
type ID interface{}

// key in a node, string for nodes.Object and int for nodes.Array
type Key interface{ isKey() }

type String string
type Int int

func (_ Int) isKey()    {}
func (_ String) isKey() {}

// four change types

// create a node
type Create struct {
	changeBase
	node nodes.Node
}

// delete a node by ID
type Delete struct {
	changeBase
	nodeID ID
}

// attach a node as a child of another node with a given key
type Attach struct {
	changeBase
	parent ID
	key    Key
	child  ID
}

// deattach a child from a node
type Deattach struct {
	changeBase
	parent ID
	key    Key //or string, how to do alternative?
}
