package diff

import (
	"fmt"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

func Apply(tree nodes.Node, changelist Changelist) nodes.Node {
	for _, change := range changelist {
		switch change.(type) {
		case Create:
		case Delete:
			panic("delete is not supported!")
		case Attach:
		case Deattach:
		default:
			panic(fmt.Sprintf("unknown change %T of type %v", change, change))
		}
	}
	return tree
}
