package uast

import (
	"sort"

	"gopkg.in/bblfsh/sdk.v1/uast/role"
	"gopkg.in/src-d/go-errors.v1"
)

var (
	ErrEmptyAST             = errors.NewKind("empty AST given")
	ErrTwoTokensSameNode    = errors.NewKind("token was already set (%s != %s)")
	ErrTwoTypesSameNode     = errors.NewKind("internal type was already set (%s != %s)")
	ErrUnexpectedObject     = errors.NewKind("expected object of type %s, got: %#v")
	ErrUnexpectedObjectSize = errors.NewKind("expected object of size %d, got %d")
	ErrUnsupported          = errors.NewKind("unsupported: %s")
)

// Special field keys for Object
const (
	KeyType  = "@type"  // InternalType
	KeyToken = "@token" // Token
	KeyRoles = "@role"  // Roles, represented as List(Int(role1), Int(role2))
	// TODO: a single @pos field with "start" and "end" fields?
	KeyStart = "@start" // StartPosition
	KeyEnd   = "@end"   // EndPosition
)

// NewNode creates a default AST node with Unannotated role.
func NewNode() Object {
	return Object{KeyRoles: RoleList(role.Unannotated)}
}

// EmptyNode creates a new empty node with no fields.
func EmptyNode() Object {
	return Object{}
}

// Node is a generic interface for structures used in AST.
//
// Can be one of:
//	* Object
//	* List
//	* Value
type Node interface {
	// Clone creates a deep copy of the node.
	Clone() Node
	isNode() // to limit possible type
}

// Value is a generic interface for values of AST node fields.
//
// Can be one of:
//	* String
//	* Int
//	* Bool
type Value interface {
	Node
	isValue() // to limit possible type
}

// Properties are written directly to object map: obj[k] = String(m[k]).
// Children are not flatten to single array, but written as fields: obj[k] = Object(m[k]) or obj[k] = List(m[k]).

// Object is a representation of generic AST node with fields.
type Object map[string]Node

func (Object) isNode() {}

// Keys returns a sorted list of node keys.
func (m Object) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (m Object) Clone() Node {
	out := make(Object, len(m))
	for k, v := range m {
		out[k] = v.Clone()
	}
	return out
}

// CloneObject clones this AST node only, without deep copy of field values.
func (m Object) CloneObject() Object {
	out := make(Object, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// CloneProperties returns an object containing all field that are values.
func (m Object) CloneProperties() Object {
	out := make(Object)
	for k, v := range m {
		if v, ok := v.(Value); ok {
			out[k] = v
		}
	}
	return out
}

// Children returns a list of all internal nodes of type Object and List.
func (m Object) Children() []Node {
	out := make([]Node, 0, len(m))
	// order should be predictable
	for _, k := range m.Keys() {
		v := m[k]
		if _, ok := v.(Value); !ok {
			out = append(out, v)
		}
	}
	return out
}

// Properties returns a map containing all field of object that are values.
func (m Object) Properties() map[string]Value {
	out := make(map[string]Value)
	for k, v := range m {
		if v, ok := v.(Value); ok {
			out[k] = v
		}
	}
	return out
}

// SetProperty is a helper for setting node properties.
func (m Object) SetProperty(k, v string) Object {
	m[k] = String(v)
	return m
}

// Type is a helper for getting node type (see KeyType).
func (m Object) Type() string {
	s, _ := m[KeyType].(String)
	return string(s)
}

// SetType is a helper for setting node type (see KeyType).
func (m Object) SetType(typ string) Object {
	return m.SetProperty(KeyType, typ)
}

// Token is a helper for getting node token (see KeyToken).
func (m Object) Token() string {
	s, _ := m[KeyToken].(String)
	return string(s)
}

// SetToken is a helper for setting node type (see KeyToken).
func (m Object) SetToken(tok string) Object {
	return m.SetProperty(KeyToken, tok)
}

// Roles is a helper for getting node UAST roles (see KeyRoles).
func (m Object) Roles() []role.Role {
	arr, _ := m[KeyRoles].(List)
	out := make([]role.Role, 0, len(arr))
	for _, v := range arr {
		if r, ok := v.(Int); ok {
			// TODO: use String, and define string lookup on Role
			out = append(out, role.Role(r))
		}
	}
	return out
}

// SetRoles is a helper for setting node UAST roles (see KeyRoles).
func (m Object) SetRoles(roles ...role.Role) Object {
	m[KeyRoles] = RoleList(roles...)
	return m
}

// StartPosition returns start position of the node in source file.
func (m Object) StartPosition() *Position {
	o, _ := m[KeyStart].(Object)
	return AsPosition(o)
}

// EndPosition returns start position of the node in source file.
func (m Object) EndPosition() *Position {
	o, _ := m[KeyEnd].(Object)
	return AsPosition(o)
}

// List is an ordered list of AST nodes.
type List []Node

func (List) isNode() {}

func (m List) Clone() Node {
	out := make(List, 0, len(m))
	for _, v := range m {
		out = append(out, v.Clone())
	}
	return out
}

func (m List) CloneList() List {
	out := make(List, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

// String is a string value used in AST fields.
type String string

func (String) isNode()  {}
func (String) isValue() {}
func (v String) Clone() Node {
	return v
}

// Int is a integer value used in AST fields.
type Int int64

func (Int) isNode()  {}
func (Int) isValue() {}
func (v Int) Clone() Node {
	return v
}

// Bool is a boolean value used in AST fields.
type Bool bool

func (Bool) isNode()  {}
func (Bool) isValue() {}
func (v Bool) Clone() Node {
	return v
}

// Apply takes a root node and applies callback to each node of the tree recursively.
// Apply returns an old or a new node and a flag that indicates if node was changed or not.
// If callback returns true and a new node, Apply will make a copy of parent node and
// will replace an old value with a new one. It will make a copy of all parent
// nodes recursively in this case.
func Apply(root Node, apply func(n Node) (Node, bool)) (Node, bool) {
	if root == nil {
		return nil, false
	}
	var changed bool
	switch n := root.(type) {
	case Object:
		var nn Object
		for k, v := range n {
			if nv, ok := Apply(v, apply); ok {
				if nn == nil {
					nn = n.CloneObject()
				}
				nn[k] = nv
			}
		}
		if nn != nil {
			changed = true
			root = nn
		}
	case List:
		var nn List
		for i, v := range n {
			if nv, ok := Apply(v, apply); ok {
				if nn == nil {
					nn = n.CloneList()
				}
				nn[i] = nv
			}
		}
		if nn != nil {
			changed = true
			root = nn
		}
	}
	nn, changed2 := apply(root)
	return nn, changed || changed2
}

/*
const (
	// InternalRoleKey is a key string uses in properties to use the internal
	// role of a node in the AST, if any.
	InternalRoleKey = "internalRole"
)

// ObjectToNode transform trees that are represented as nested JSON objects.
// That is, an interface{} containing maps, slices, strings and integers. It
// then converts from that structure to *Node.
type ObjectToNode struct {
	// IsNode is used to identify witch map[string]interface{} are nodes, if
	// nil, any map[string]interface{} is considered a node.
	IsNode func(map[string]interface{}) bool
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
	// SpecificTokenKeys allow to map specific nodes, by their internal type, to a
	// concrete field of the node. This can solve conflicts on some nodes that the token
	// represented by a very unique field or have more than one of the fields specified in
	// TokenKeys.
	SpecificTokenKeys map[string]string
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
	// generation will set as children of a node any uast. that hangs from any of the
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
	// OnToNode is called, if defined, just before the method ToNode is called,
	// allowing any modification or alteration of the AST before being
	// processed.
	OnToNode func(interface{}) (interface{}, error)
	//Modifier function is called, if defined, to modify a
	// map[string]interface{} (which normally would be converted to a Node)
	// before it's processed.
	Modifier func(map[string]interface{}) error
}

func (c *ObjectToNode) ToNode(v interface{}) (*Node, error) {
	if c.OnToNode != nil {
		var err error
		v, err = c.OnToNode(v)
		if err != nil {
			return nil, err
		}
	}

	src, ok := v.(map[string]interface{})
	if !ok {
		return nil, ErrUnsupported.New("non-object root node")
	}

	root, err := findRoot(src, c.TopLevelIsRootNode)
	if err != nil {
		return nil, err
	}

	nodes, err := c.toNodes(root)
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, ErrEmptyAST.New()
	}

	if len(nodes) != 1 {
		return nil, ErrUnsupported.New("multiple root nodes found")
	}

	return nodes[0], err
}

func findRoot(m map[string]interface{}, topLevelIsRootNode bool) (interface{}, error) {
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

func (c *ObjectToNode) toNodes(obj interface{}) ([]*Node, error) {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return nil, ErrUnexpectedObject.New("map[string]interface{}", obj)
	}

	if err := c.applyModifier(m); err != nil {
		return nil, err
	}

	internalKey := c.getInternalKeyFromObject(m)

	var promotedListKeys map[string]bool
	if !c.PromoteAllPropertyLists && c.PromotedPropertyLists != nil {
		promotedListKeys = c.PromotedPropertyLists[internalKey]
	}
	var promotedStrKeys map[string]bool
	if c.PromotedPropertyStrings != nil {
		promotedStrKeys = c.PromotedPropertyStrings[internalKey]
	}

	n := NewNode()
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
			if ov == nil {
				continue
			}
			c.maybeAddComposedPositionProperties(n, ov)
			children, err := c.mapToNodes(k, ov)
			if err != nil {
				return nil, err
			}

			n.Children = append(n.Children, children...)
		case []interface{}:
			if c.PromoteAllPropertyLists || (promotedListKeys != nil && promotedListKeys[k]) {
				// This property->List  must be promoted to its own node
				children, err := c.sliceToNodeWithChildren(k, ov, internalKey)
				if err != nil {
					return nil, err
				}

				n.Children = append(n.Children, children...)
				continue
			}

			// This property -> List elements will be added as the current node Children
			children, err := c.sliceToNodeSlice(k, ov)
			// List of non-nodes
			if ErrUnexpectedObject.Is(err) {
				err = c.addProperty(n, k, ov)
			}
			if err != nil {
				return nil, err
			}

			n.Children = append(n.Children, children...)
		case nil:
			// ignoring key with nil values
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

	if c.IsNode != nil && !c.IsNode(m) {
		return n.Children, nil
	}

	return []*Node{n}, nil
}
func (c *ObjectToNode) applyModifier(m map[string]interface{}) error {
	if c.Modifier == nil {
		return nil
	}

	return c.Modifier(m)
}
func (c *ObjectToNode) mapToNodes(k string, obj map[string]interface{}) (Node, error) {
	nodes, err := c.toNodes(obj)
	if err != nil {
		return nil, err
	}

	for _, n := range nodes {
		n.Properties[InternalRoleKey] = k
	}

	return nodes, nil
}

func (c *ObjectToNode) sliceToNodeWithChildren(k string, s []interface{}, parentKey string) (Node, error) {
	kn := NewNode()

	var ns []*Node
	for _, v := range s {
		n, err := c.toNodes(v)
		if err != nil {
			return nil, err
		}

		ns = append(ns, n...)
	}

	if len(ns) == 0 {
		// should be still create new nodes for empty slices or add it as an option?
		return nil, nil
	}
	c.setInternalKey(kn, parentKey+"."+k)
	kn.Properties["promotedPropertyList"] = "true"
	kn.Children = append(kn.Children, ns...)

	return []*Node{kn}, nil
}

func (c *ObjectToNode) stringToNode(k, v, parentKey string) Node {
	kn := make(Object)

	c.setInternalKey(kn, parentKey+"."+k)
	kn.Properties["promotedPropertyString"] = "true"
	kn.Token = v

	return kn
}

func (c *ObjectToNode) sliceToNodeSlice(k string, s []interface{}) ([]*Node, error) {
	var ns []*Node
	for _, v := range s {
		nodes, err := c.toNodes(v)
		if err != nil {
			return nil, err
		}

		for _, n := range nodes {
			n.Properties[InternalRoleKey] = k
		}

		ns = append(ns, nodes...)
	}

	return ns, nil
}

func (c *ObjectToNode) maybeAddComposedPositionProperties(n *Node, o map[string]interface{}) {
	keys := []string{c.OffsetKey, c.LineKey, c.ColumnKey, c.EndOffsetKey, c.EndLineKey, c.EndColumnKey}
	for _, k := range keys {
		if !strings.Contains(k, ".") {
			continue
		}
		xs := strings.SplitAfterN(k, ".", 2)
		v, err := lookup.LookupString(o, xs[1])
		if err != nil {
			continue
		}

		c.addProperty(n, k, v.Interface())
	}
}

func (c *ObjectToNode) addProperty(n *Node, k string, o interface{}) error {
	switch {
	case c.isTokenKey(n, k):
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
			n.StartPosition = &Position{}
		}

		n.StartPosition.Offset = i
	case c.EndOffsetKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		if n.EndPosition == nil {
			n.EndPosition = &Position{}
		}

		n.EndPosition.Offset = i
	case c.LineKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		if n.StartPosition == nil {
			n.StartPosition = &Position{}
		}

		n.StartPosition.Line = i
	case c.EndLineKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		if n.EndPosition == nil {
			n.EndPosition = &Position{}
		}

		n.EndPosition.Line = i
	case c.ColumnKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		if n.StartPosition == nil {
			n.StartPosition = &Position{}
		}

		n.StartPosition.Col = i
	case c.EndColumnKey == k:
		i, err := toUint32(o)
		if err != nil {
			return err
		}

		if n.EndPosition == nil {
			n.EndPosition = &Position{}
		}

		n.EndPosition.Col = i
	default:
		v, err := toPropValue(o)
		if err != nil {
			return err
		}
		n.Properties[k] = v
	}

	return nil
}

func (c *ObjectToNode) isTokenKey(n *Node, key string) bool {

	if c.SpecificTokenKeys != nil && n.InternalType != "" {
		if tokenKey, ok := c.SpecificTokenKeys[n.InternalType]; ok {
			// Nodes of this internalType use a specific property as token
			return tokenKey == key
		}
	}

	return c.TokenKeys != nil && c.TokenKeys[key]
}

func (c *ObjectToNode) syntheticToken(key string) string {

	if c.SyntheticTokens == nil {
		return ""
	}

	return c.SyntheticTokens[key]
}

func (c *ObjectToNode) setInternalKey(n *Node, k string) error {
	if n.InternalType != "" && n.InternalType != k {
		return ErrTwoTypesSameNode.New(n.InternalType, k)
	}

	n.InternalType = k
	return nil
}

func (c *ObjectToNode) getInternalKeyFromObject(m map[string]interface{}) string {
	if val, ok := m[c.InternalTypeKey].(string); ok {
		return val
	}

	// should this be an error?
	return ""
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
	apos := startPosition(a)
	bpos := startPosition(b)
	if apos == nil {
		return false
	}

	if bpos == nil {
		return false
	}

	return apos.Offset < bpos.Offset
}

func startPosition(n *Node) *Position {
	if n.StartPosition != nil {
		return n.StartPosition
	}

	var min *Position
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

func toPropValue(o interface{}) (string, error) {
	if o == nil {
		return "null", nil
	}

	t := reflect.TypeOf(o)
	switch t.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		b, err := json.Marshal(o)
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		return fmt.Sprint(o), nil
	}
}
*/
