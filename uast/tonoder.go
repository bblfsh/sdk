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
	// ColumnKey is a key that indicates the column inside the line
	ColumnKey string
	// TokenKeys is a slice of keys used to extract token content.
	TokenKeys map[string]bool
	// SyntheticTokens is a map of InternalType to string used to add
	// synthetic tokens to nodes depending on its InternalType.
	SyntheticTokens map[string]string
	// PromotedPropertyLists allows to convert some properties in the native AST with a list value
	// to its own node with the list elements as children. The key of the first map is the name
	// of the InternalType which can have promotions and the value is a map where keys must be the names
	// of the properties to be promoted if the value is true. For example to promote a "body" property
	// inside an "If" InternalKey the map should contain: ["If"]["body"] = true.
	PromotedPropertyLists map[string]map[string]bool
	// If this option is set, all properties mapped to a list will be promoted to its own node. Setting
	// this option to true will ignore the PromotedPropertyLists settings.
	PromoteAllPropertyLists bool
	// TopLevelIsRootNode tells ToNode where to find the root node of
	// the AST.  If true, the root will be its input argument. If false,
	// the root will be the value of the only key present in its input
	// argument.
	TopLevelIsRootNode bool
}

func (c *BaseToNoder) ToNode(v interface{}) (*Node, error) {
	src, ok := v.(map[string]interface{})
	if !ok {
		return nil, ErrUnsupported.New("non-object root node")
	}

	root, err := findRoot(src, c.TopLevelIsRootNode)
	if err != nil {
		return nil, err
	}

	return c.toNode(root)
}

func findRoot(m map[string]interface{}, topLevelIsRootNode bool) (
	interface{}, error) {

	if len(m) == 0 {
		return nil, ErrEmptyAST.New()
	}

	if topLevelIsRootNode {
		return m, nil
	}

	if len(m) > 1 {
		return nil, ErrUnexpectedObjectSize.New(1, len(m))
	}

	for _, root := range m {
		return root, nil
	}

	panic("unreachable")
}

func (c *BaseToNoder) toNode(obj interface{}) (*Node, error) {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return nil, ErrUnexpectedObject.New("map[string]interface{}", obj)
	}

	n := NewNode()

	// We need to have the internalkey before iterating others
	internalKey, err := c.getInternalKeyFromObject(obj)
	if err != nil {
		return nil, err
	}

	var promotedKeys map[string]bool
	if !c.PromoteAllPropertyLists && c.PromotedPropertyLists != nil {
		promotedKeys = c.PromotedPropertyLists[internalKey]
	}

	if err := c.setInternalKey(n, internalKey); err != nil {
		return nil, err
	}

	// Sort the keys of the map so the integration tests that currently do a
	// textual diff doesn't fail because of sort order
	var keys []string
	for listkey := range m {
		keys = append(keys, listkey)
	}
	sort.Strings(keys)

	for _, k := range keys {
		o := m[k]
		switch ov := o.(type) {
		case map[string]interface{}:
			child, err := c.mapToNode(k, ov)
			if err != nil {
				return nil, err
			}

			n.Children = append(n.Children, child)
		case []interface{}:
			if c.PromoteAllPropertyLists || (promotedKeys != nil && promotedKeys[k]) {
				// This property->List  must be promoted to its own node
				child, err := c.sliceToNodeWithChildren(k, ov, internalKey)
				if err != nil {
					return nil, err
				}
				if child != nil {
					n.Children = append(n.Children, child)
				}
				continue
			}

			// This property -> List elements will be added as the current node Children
			children, err := c.sliceToNodeSlice(k, ov)
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

func (c *BaseToNoder) sliceToNodeWithChildren(k string, s []interface{}, parentKey string) (*Node, error) {
	kn := NewNode()

	var ns []*Node
	for _, v := range s {
		n, err := c.toNode(v)
		if err != nil {
			return nil, err
		}

		ns = append(ns, n)
	}

	if len(ns) == 0 {
		// should be still create new nodes for empty slices or add it as an option?
		return nil, nil
	}
	c.setInternalKey(kn, parentKey+"."+k)
	kn.Properties["promotedPropertyList"] = "true"
	kn.Children = append(kn.Children, ns...)

	return kn, nil
}

func (c *BaseToNoder) sliceToNodeSlice(k string, s []interface{}) ([]*Node, error) {
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
		// InternalType should be already set by toNode, but check if
		// they InternalKey is one of the ones in SyntheticTokens.
		s := fmt.Sprint(o)
		tk := c.syntheticToken(s)
		if tk != "" {
			if n.Token != "" && n.Token != tk {
				return ErrTwoTokensSameNode.New(n.Token, tk)
			}

			n.Token = tk
 		}
		return nil
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
	case c.ColumnKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		n.StartPosition.Col = i
	default:
		n.Properties[k] = fmt.Sprint(o)
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

func (c *BaseToNoder) getInternalKeyFromObject(obj interface{}) (string, error) {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return "", ErrUnexpectedObject.New("map[string]interface{}", obj)
	}

	if val, ok := m[c.InternalTypeKey].(string); ok {
		return val, nil
	}
	// should this be an error?
	return "", nil
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
	case float64:
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
