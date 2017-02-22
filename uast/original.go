package uast

import (
	"fmt"
	"strconv"

	"sort"
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
	OriginalToNode(src map[string]interface{}) (*Node, error)
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
	// OffsetKey is a key that indicates the position offset.
	OffsetKey string
	// LineKey is a key that indicates the line number.
	LineKey string
	// TokenKeys is a slice of keys used to extract token content.
	TokenKeys map[string]bool
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
			switch {
			case c.isTokenKey(k):
				s, err := toString(o)
				if err != nil {
					return nil, err
				}

				if n.Token != nil {
					return nil, fmt.Errorf("two token keys for same node: %s", key)
				}

				n.Token = &s
			case c.InternalTypeKey == k:
				s, err := toString(o)
				if err != nil {
					return nil, err
				}

				n.InternalType = s
			case c.OffsetKey == k:
				i, err := toUint32(o)
				if err != nil {
					return nil, err
				}

				if n.StartPosition == nil {
					n.StartPosition = &Position{}
				}

				n.StartPosition.Offset = &i
			case c.LineKey == k:
				i, err := toUint32(o)
				if err != nil {
					return nil, err
				}

				if n.StartPosition == nil {
					n.StartPosition = &Position{}
				}

				n.StartPosition.Line = &i
			default:
				s, err := toString(o)
				if err != nil {
					return nil, err
				}

				n.Properties[k] = s
			}
		}
	}

	sort.Sort(byOffset(n.Children))
	return n, nil
}

func (c *BaseOriginalToNoder) isTokenKey(key string) bool {
	return c.TokenKeys != nil && c.TokenKeys[key]
}

func toString(v interface{}) (string, error) {
	switch o := v.(type) {
	case string:
		return o, nil
	case fmt.Stringer:
		return o.String(), nil
	case int:
		return strconv.Itoa(o), nil
	default:
		return "", fmt.Errorf("toString error: %#v", v)
	}
}

func toUint32(v interface{}) (uint32, error) {
	switch o := v.(type) {
	case string:
		i, err := strconv.ParseUint(o, 10, 32)
		if err != nil {
			return 0, err
		}

		return uint32(i), nil
	case uint32:
		return o, nil
	case int:
		return uint32(o), nil
	case int32:
		return uint32(o), nil
	case int64:
		return uint32(o), nil
	default:
		return 0, fmt.Errorf("toUint32 error: %#v", v)
	}
}

type byOffset []*Node

func (s byOffset) Len() int      { return len(s) }
func (s byOffset) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byOffset) Less(i, j int) bool {
	a := s[i]
	b := s[j]
	ao := a.offset()
	bo := b.offset()
	if ao == nil {
		return false
	}

	if bo == nil {
		return true
	}

	return *ao < *bo
}
