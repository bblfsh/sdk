package uast

import (
	"fmt"
	"sort"
	"strconv"

	"srcd.works/go-errors.v0"
)

var (
	ErrEmptyAST             = errors.NewKind("input AST was empty")
	ErrTwoTokensSameNode    = errors.NewKind("token was already set (%s != %s)")
	ErrTwoTypesSameNode     = errors.NewKind("internal type was already set (%s != %s)")
	ErrUnexpectedObject     = errors.NewKind("expected object of type %s, got: %#v")
	ErrUnexpectedObjectSize = errors.NewKind("expected object of size %d, got %d")
	ErrUnsupported          = errors.NewKind("unsupported: %s")
)

// ToNoder is a converter of source ASTs to *Node.
type ToNoder interface {
	// ToNode converts the source AST to a *Node.
	ToNode(src interface{}) (*Node, error)
}

const (
	// topLevelIsRootNode is true if the top level object is the root node
	// of the AST. If false, top level object should have a single key, that
	// being the root node.
	topLevelIsRootNode = false
	// InternalRoleKey is a key string uses in properties to use the internal
	// role of a node in the AST, if any.
	InternalRoleKey = "internalRole"
)

// BaseOriginalToNoder is an implementation of OriginalToNoder that aims to work
// for the most common source ASTs.
type BaseToNoder struct {
	// InternalTypeKey is a key in the source AST that can be used to get the
	// InternalType of a node.
	InternalTypeKey string
	// OffsetKey is a key that indicates the position offset.
	OffsetKey string
	// LineKey is a key that indicates the line number.
	LineKey string
	// TokenKeys is a slice of keys used to extract token content.
	TokenKeys map[string]bool
	// SyntheticTokens is a map of InternalType to string used to add
	// synthetic tokens to nodes depending on its InternalType.
	SyntheticTokens map[string]string
}

func (c *BaseToNoder) ToNode(v interface{}) (*Node, error) {
	src, ok := v.(map[string]interface{})
	if !ok {
		return nil, ErrUnsupported.New("non-object root node")
	}

	if topLevelIsRootNode {
		return nil, ErrUnsupported.New("top level object as root node")
	}

	if len(src) > 1 {
		return nil, ErrUnexpectedObjectSize.New(1, len(src))
	}

	if len(src) == 0 {
		return nil, ErrEmptyAST.New()
	}

	var vobj interface{}
	for _, obj := range src {
		vobj = obj
		break
	}

	return c.toNode(vobj)
}

func (c *BaseToNoder) toNode(obj interface{}) (*Node, error) {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return nil, ErrUnexpectedObject.New("map[string]interface{}", obj)
	}

	n := NewNode()
	for k, o := range m {
		switch ov := o.(type) {
		case map[string]interface{}:
			child, err := c.mapToNode(k, ov)
			if err != nil {
				return nil, err
			}

			n.Children = append(n.Children, child)
		case []interface{}:
			children, err := c.sliceToNodes(k, ov)
			if err != nil {
				return nil, err
			}

			n.Children = append(n.Children, children...)
		default:
			if err := c.addProperty(n, k, o); err != nil {
				return nil, err
			}
		}
	}

	sort.Stable(byOffset(n.Children))

	return n, nil
}

func (c *BaseToNoder) mapToNode(k string, obj map[string]interface{}) (*Node, error) {
	n, err := c.toNode(obj)
	if err != nil {
		return nil, err
	}

	n.Properties[InternalRoleKey] = k

	return n, nil
}

func (c *BaseToNoder) sliceToNodes(k string, s []interface{}) ([]*Node, error) {
	var ns []*Node
	for _, v := range s {
		n, err := c.toNode(v)
		if err != nil {
			return nil, err
		}

		n.Properties[InternalRoleKey] = k
		ns = append(ns, n)
	}

	return ns, nil
}

func (c *BaseToNoder) addProperty(n *Node, k string, o interface{}) error {
	switch {
	case c.isTokenKey(k):
		s := fmt.Sprint(o)
		if n.Token != "" && n.Token != s {
			return ErrTwoTokensSameNode.New(n.Token, s)
		}

		n.Token = s
	case c.InternalTypeKey == k:
		s := fmt.Sprint(o)
		if err := c.setInternalKey(n, s); err != nil {
			return err
		}

		tk := c.syntheticToken(s)
		if tk != "" {
			if n.Token != "" && n.Token != tk {
				return ErrTwoTokensSameNode.New(n.Token, tk)
			}

			n.Token = tk
		}
	case c.OffsetKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		n.StartPosition.Offset = i
	case c.LineKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		n.StartPosition.Line = i
	default:
		n.Properties[k] = fmt.Sprint(0)
	}

	return nil
}

func (c *BaseToNoder) isTokenKey(key string) bool {
	return c.TokenKeys != nil && c.TokenKeys[key]
}

func (c *BaseToNoder) syntheticToken(key string) string {
	if c.SyntheticTokens == nil {
		return ""
	}

	return c.SyntheticTokens[key]
}

func (c *BaseToNoder) setInternalKey(n *Node, k string) error {
	if n.InternalType != "" && n.InternalType != k {
		return ErrTwoTypesSameNode.New(n.InternalType, k)
	}

	n.InternalType = k
	return nil
}

// toUint32 converts a JSON value to a uint32.
// The only expected values are string or int64.
func toUint32(v interface{}) (uint32, error) {
	switch o := v.(type) {
	case string:
		i, err := strconv.ParseUint(o, 10, 32)
		if err != nil {
			return 0, err
		}

		return uint32(i), nil
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
	apos := a.startPosition()
	bpos := b.startPosition()
	if apos.IsEmpty() {
		return false
	}

	if bpos.IsEmpty() {
		return true
	}

	return apos.Offset < bpos.Offset
}
