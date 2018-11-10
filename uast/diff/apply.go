package diff

import (
	"fmt"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

func Apply(root nodes.Node, changelist Changelist) nodes.Node {
	nodeDict := make(map[ID]nodes.Node)
	nodes.WalkPreOrder(root, func(node nodes.Node) bool {
		nodeDict[nodes.UniqueKey(node)] = node
		return true
	})

	for _, change := range changelist {
		switch ch := change.(type) {
		case Create:
			//create a node and add to the dictionary
			nodeDict[nodes.UniqueKey(ch.Node)] = ch.Node

		case Attach:
			//get src and chld from the dictionary, attach (modify src)
			parent, ok := nodeDict[ch.Parent]
			if !ok {
				panic("invalid attachment point")
			}
			child, ok := nodeDict[ch.Child]
			if !ok {
				child, ok = ch.Child.(nodes.Value)
				if !ok {
					panic(fmt.Errorf("unknown type of a child: %v (type %T)", ch.Child, ch.Child))
				}
			}

			switch key := ch.Key.(type) {

			case String:
				parent := parent.(nodes.Object)
				parent[string(key)] = child

			case Int:
				parent := parent.(nodes.Array)
				parent[int(key)] = child
			}

		case Deatach:
			//get the src from the dictionary, deatach (modify src)
			parent := nodeDict[ch.Parent]

			switch key := ch.Key.(type) {

			case String:
				parent := parent.(nodes.Object)
				delete(parent, string(key))

			case Int:
				panic(fmt.Errorf("cannot deatach from an Array"))
			}

		case Delete:
			panic(fmt.Errorf("delete is not supported in a Changelist"))

		default:
			panic(fmt.Sprintf("unknown change %v of type %T", change, change))
		}
	}
	return root
}
