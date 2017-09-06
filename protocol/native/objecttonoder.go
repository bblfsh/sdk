package native

import (
	"fmt"
	"sort"
	"strconv"

	"gopkg.in/bblfsh/sdk.v0/uast"

	"gopkg.in/src-d/go-errors.v0"
)

var (
	ErrEmptyAST             = errors.NewKind("input AST was empty")
	ErrTwoTokensSameNode    = errors.NewKind("token was already set (%s != %s)")
	ErrTwoTypesSameNode     = errors.NewKind("internal type was already set (%s != %s)")
	ErrUnexpectedObject     = errors.NewKind("expected object of type %s, got: %#v")
	ErrUnexpectedObjectSize = errors.NewKind("expected object of size %d, got %d")
	ErrUnsupported          = errors.NewKind("unsupported: %s")
)

type FillType int

const (
	None FillType = iota
	OffsetFromLineCol
	LineColFromOffset
)

// ObjectToNoder is a ToNoder for trees that are represented as nested JSON objects.
// That is, an interface{} containing maps, slices, strings and integers. It
// then converts from that structure to *Node.
type ObjectToNoder struct {
	// InternalTypeKey is the name of the key that the native AST uses
	// to differentiate the type of the AST nodes. This internal key will then be
	// checkable in the AnnotationRules with the `HasInternalType` predicate. This
	// field is mandatory.
	InternalTypeKey string
	// OffsetKey is the key used in the native AST to indicate the absolute offset,
	// from the file start position, where the code mapped to the AST node starts.
	OffsetKey string
	// EndOffsetKey is the key used in the native AST to indicate the absolute offset,
	// from the file start position, where the code mapped to the AST node ends.
	EndOffsetKey string
	// LineKey is the key used in the native AST to indicate
	// the line number where the code mapped to the AST node starts.
	LineKey string
	// EndLineKey is the key used in the native AST to indicate
	// the line number where the code mapped to the AST node ends.
	EndLineKey string
	// ColumnKey is a key that indicates the column inside the line
	ColumnKey string
	// EndColumnKey is a key that indicates the column inside the line where the node ends.
	EndColumnKey string
	// TokenKeys establishes what properties (as in JSON
	// keys) in the native AST nodes can be mapped to Tokens in the UAST. If the
	// InternalTypeKey is the "type" of a node, the Token could be tough of as the
	// "value" representation; this could be a specific value for string/numeric
	// literals or the symbol name for others.  E.g.: if a native AST represents a
	// numeric literal as: `{"ast_type": NumLiteral, "value": 2}` then you should have
	// to add `"value": true` to the TokenKeys map.  Some native ASTs will use several
	// different fields as tokens depending on the node type; in that case, all should
	// be added to this map to ensure a correct UAST generation.
	TokenKeys map[string]bool
	// SyntheticTokens is a map of InternalType to string used to add
	// synthetic tokens to nodes depending on its InternalType; sometimes native ASTs just use an
	// InternalTypeKey for some node but we need to add a Token to the UAST node to
	// improve the representation. In this case we can add both the InternalKey and
	// what token it should generate. E.g.: an InternalTypeKey called "NullLiteral" in
	// Java should be mapped using this map to "null" adding ```"NullLiteral":
	// "null"``` to this map.
	SyntheticTokens map[string]string
	// PromotedPropertyLists allows to convert some properties in the native AST with a list value
	// to its own node with the list elements as children. 	By default the UAST
	// generation will set as children of a node any object that hangs from any of the
	// original native AST node properties. In this process, object key serving as
	// the parent is lost and its name is added as the "internalRole" key of the children.
	// This is usually fine since the InternalTypeKey of the parent AST node will
	// usually provide enough context and the node won't any other children. This map
	// allows you to change this default behavior for specific nodes so the properties
	// are "promoted" to a new node (with an InternalTypeKey named "Parent.KeyName")
	// and the objects in its list will be shown in the UAST as children. E.g.: if you
	// have a native AST where an "If" node has the JSON keys "body", "else" and
	// "condition" each with its own list of children, you could add an entry to
	// PromotedPropertyLists like
	//
	// "If": {"body": true, "orelse": true, "condition": true},
	//
	// In this case, the new nodes will have the InternalTypeKey "If.body", "If.orelse"
	// and "If.condition" and with these names you should be able to write specific
	// matching rules in the annotation.go file.
	PromotedPropertyLists map[string]map[string]bool
	// If this option is set, all properties mapped to a list will be promoted to its own node. Setting
	// this option to true will ignore the PromotedPropertyLists settings.
	PromoteAllPropertyLists bool
	// PromotedPropertyStrings allows to convert some properties which value is a string
	// in the native AST as a full node with the string value as Token like:
	//
	// "SomeKey": "SomeValue"
	//
	// that would be converted to a child node like:
	//
	// {"internalType": "SomeKey", "Token": "SomeValue"}
	PromotedPropertyStrings map[string]map[string]bool
	// TopLevelIsRootNode tells ToNode where to find the root node of
	// the AST.  If true, the root will be its input argument. If false,
	// the root will be the value of the only key present in its input
	// argument.
	TopLevelIsRootNode bool
	// PositionFill specifies if the noder has to fill missing positions (col, line, offset)
	// from ones that the native AST fills. The possible values are "None" (don't fill
	// anything), "OffsetFromLineCol" (fill the offset from the line and column values) and
	// "LineColFromOffset" (fill line and col from the offset).
	PositionFill FillType
}

func (c *ObjectToNoder) ToNode(v interface{}) (*uast.Node, error) {
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

func (c *ObjectToNoder) toNode(obj interface{}) (*uast.Node, error) {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return nil, ErrUnexpectedObject.New("map[string]interface{}", obj)
	}

	n := uast.NewNode()

	// We need to have the internalkey before iterating others
	internalKey, err := c.getInternalKeyFromObject(obj)
	if err != nil {
		return nil, err
	}

	var promotedListKeys map[string]bool
	if !c.PromoteAllPropertyLists && c.PromotedPropertyLists != nil {
		promotedListKeys = c.PromotedPropertyLists[internalKey]
	}
	var promotedStrKeys map[string]bool
	if c.PromotedPropertyStrings != nil {
		promotedStrKeys = c.PromotedPropertyStrings[internalKey]
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
			if c.PromoteAllPropertyLists || (promotedListKeys != nil && promotedListKeys[k]) {
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
			newKey := k
			if s, ok := o.(string); ok {
				if len(s) > 0 && promotedStrKeys != nil && promotedStrKeys[k] {
					newKey = internalKey + "." + k
					child := c.stringToNode(k, s, internalKey)
					if child != nil {
						n.Children = append(n.Children, child)
					}
				}
			}

			if err := c.addProperty(n, newKey, o); err != nil {
				return nil, err
			}
		}
	}

	sort.Stable(byOffset(n.Children))

	return n, nil
}

func (c *ObjectToNoder) mapToNode(k string, obj map[string]interface{}) (*uast.Node, error) {
	n, err := c.toNode(obj)
	if err != nil {
		return nil, err
	}

	n.Properties[uast.InternalRoleKey] = k

	return n, nil
}

func (c *ObjectToNoder) sliceToNodeWithChildren(k string, s []interface{}, parentKey string) (*uast.Node, error) {
	kn := uast.NewNode()

	var ns []*uast.Node
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

func (c *ObjectToNoder) stringToNode(k, v, parentKey string) *uast.Node {
	kn := uast.NewNode()

	c.setInternalKey(kn, parentKey+"."+k)
	kn.Properties["promotedPropertyString"] = "true"
	kn.Token = v

	return kn
}

func (c *ObjectToNoder) sliceToNodeSlice(k string, s []interface{}) ([]*uast.Node, error) {
	var ns []*uast.Node
	for _, v := range s {
		n, err := c.toNode(v)
		if err != nil {
			return nil, err
		}

		n.Properties[uast.InternalRoleKey] = k
		ns = append(ns, n)
	}

	return ns, nil
}

func (c *ObjectToNoder) addProperty(n *uast.Node, k string, o interface{}) error {
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

		if n.StartPosition == nil {
			n.StartPosition = &uast.Position{}
		}

		n.StartPosition.Offset = i
	case c.EndOffsetKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		if n.EndPosition == nil {
			n.EndPosition = &uast.Position{}
		}

		n.EndPosition.Offset = i
	case c.LineKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		if n.StartPosition == nil {
			n.StartPosition = &uast.Position{}
		}

		n.StartPosition.Line = i
	case c.EndLineKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		if n.EndPosition == nil {
			n.EndPosition = &uast.Position{}
		}

		n.EndPosition.Line = i
	case c.ColumnKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		if n.StartPosition == nil {
			n.StartPosition = &uast.Position{}
		}

		n.StartPosition.Col = i
	case c.EndColumnKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		if n.EndPosition == nil {
			n.EndPosition = &uast.Position{}
		}

		n.EndPosition.Col = i
	default:
		n.Properties[k] = fmt.Sprint(o)
	}

	return nil
}

func (c *ObjectToNoder) isTokenKey(key string) bool {
	return c.TokenKeys != nil && c.TokenKeys[key]
}

func (c *ObjectToNoder) syntheticToken(key string) string {
	if c.SyntheticTokens == nil {
		return ""
	}

	return c.SyntheticTokens[key]
}

func (c *ObjectToNoder) setInternalKey(n *uast.Node, k string) error {
	if n.InternalType != "" && n.InternalType != k {
		return ErrTwoTypesSameNode.New(n.InternalType, k)
	}

	n.InternalType = k
	return nil
}

func (c *ObjectToNoder) getInternalKeyFromObject(obj interface{}) (string, error) {
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

type byOffset []*uast.Node

func (s byOffset) Len() int      { return len(s) }
func (s byOffset) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byOffset) Less(i, j int) bool {
	a := s[i]
	b := s[j]
	apos := startPosition(a)
	bpos := startPosition(b)
	if apos == nil {
		return false
	}

	if bpos == nil {
		return true
	}

	return apos.Offset < bpos.Offset
}

func startPosition(n *uast.Node) *uast.Position {
	if n.StartPosition != nil {
		return n.StartPosition
	}

	var min *uast.Position
	for _, c := range n.Children {
		other := startPosition(c)
		if other == nil {
			continue
		}

		if min == nil || other.Offset < min.Offset {
			min = other
		}
	}

	return min
}
