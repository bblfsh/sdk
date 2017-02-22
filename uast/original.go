package uast

import (
	"fmt"
	"strconv"

	"srcd.works/go-errors.v0"
)

var (
	ErrEmptyAST             = errors.NewKind("input AST was empty")
	ErrUnexpectedObject     = errors.NewKind("expected object of type %s, got: %#v")
	ErrUnexpectedObjectSize = errors.NewKind("expected object of size %d, got %d")
	ErrUnsupported          = errors.NewKind("unsupported: %s")
)

// OriginalToNoder is a converter of source ASTs to *Node.
type OriginalToNoder interface {
	// OriginalToNode converts the source AST to a *Node.
	OriginalToNode(src map[interface{}]interface{}) (*Node, error)
}

// topLevelIsRootNode is true if the top level object is the root node
// of the AST. If false, top level object should have a single key, that
// being the root node.
const (
	topLevelIsRootNode = false
)

// BaseOriginalToNoder is an implementation of OriginalToNoder that aims to work
// for the most common source ASTs.
type BaseOriginalToNoder struct {
	// InternalTypeKey is a key in the source AST that can be used to get the
	// InternalType of a node.
	InternalTypeKey string
}

func (c *BaseOriginalToNoder) OriginalToNode(src map[string]interface{}) (*Node, error) {
	if len(src) == 0 {
		return nil, ErrEmptyAST.New()
	}

	if topLevelIsRootNode {
		return nil, ErrUnsupported.New("top level object as root node")
	}

	if len(src) > 1 {
		return nil, ErrUnexpectedObjectSize.New(1, len(src))
	}

	for key, obj := range src {
		return c.toNode(key, obj)
	}

	panic("not reachable")
}

func (c *BaseOriginalToNoder) toNode(key interface{}, obj interface{}) (*Node, error) {
	skey, ok := key.(string)
	if !ok {
		return nil, ErrUnexpectedObject.New("string", key)
	}

	m, ok := obj.(map[string]interface{})
	if !ok {
		return nil, ErrUnexpectedObject.New("map[string]interface{}", obj)
	}

	n := NewNode()
	//TODO: More flexibility here
	n.InternalType = skey
	for k, o := range m {

		switch ov := o.(type) {
		case map[string]interface{}:
			child, err := c.toNode(k, o)
			if err != nil {
				return nil, err
			}

			n.Children = append(n.Children, child)
		case []interface{}:
			for _, v := range ov {
				child, err := c.toNode("", v)
				if err != nil {
					return nil, err
				}

				n.Children = append(n.Children, child)
			}
		default:
			s := c.toString(o)
			switch k {
			case c.InternalTypeKey:
				n.InternalType = s
			default:
				n.Properties[k] = s
			}
		}
	}

	return n, nil
}

func (c *BaseOriginalToNoder) toString(v interface{}) string {
	switch o := v.(type) {
	case string:
		return o
	case fmt.Stringer:
		return o.String()
	case int:
		return strconv.Itoa(o)
	default:
		panic(fmt.Errorf("toString error: %#v", v))
	}
}
