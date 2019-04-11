package diff

import (
	"errors"
	"fmt"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

// Apply is a method that takes a tree (nodes.Node) and applies the current changelist to
// that tree.
func (cl Changelist) Apply(root nodes.Node) (nodes.Node, error) {
	nodeDict := make(map[ID]nodes.Node)
	nodes.WalkPreOrder(root, func(node nodes.Node) bool {
		nodeDict[nodes.UniqueKey(node)] = node
		return true
	})

	for _, change := range cl {
		switch ch := change.(type) {
		case Create:
			// create a node and add to the dictionary
			nodeDict[nodes.UniqueKey(ch.Node)] = ch.Node

		case Attach:
			// get src and chld from the dictionary, attach (modify src)
			parent, ok := nodeDict[ch.Parent]
			if !ok {
				return nil, errors.New("diff: invalid attachment point")
			}
			child, ok := nodeDict[ch.Child]
			if !ok {
				child, ok = ch.Child.(nodes.Value)
				if !ok {
					return nil, fmt.Errorf("diff: unknown type of a child: %T", ch.Child)
				}
			}

			switch key := ch.Key.(type) {
			case String:
				parent := parent.(nodes.Object)
				parent[string(key)] = child
			case Int:
				parent := parent.(nodes.Array)
				parent[int(key)] = child
			default:
				return nil, fmt.Errorf("diff: unknown type of a key: %T", ch.Key)
			}
		case Detach:
			// get the src from the dictionary, deatach (modify src)
			parent := nodeDict[ch.Parent]

			switch key := ch.Key.(type) {
			case String:
				parent := parent.(nodes.Object)
				delete(parent, string(key))
			case Int:
				return nil, errors.New("diff: cannot detach from an Array")
			default:
				return nil, fmt.Errorf("diff: unknown type of a key: %T", ch.Key)
			}

		case Delete:
			return nil, errors.New("diff: delete is not supported in a Changelist")
		default:
			return nil, fmt.Errorf("diff: unknown change of type %T", change)
		}
	}
	return root, nil
}
