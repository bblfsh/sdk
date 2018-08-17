package xpath

import (
	"fmt"
	"strings"

	"github.com/antchfx/xpath"
	"gopkg.in/bblfsh/sdk.v2/uast"
)

var _ xpath.NodeNavigator = &nodeNavigator{}

// newNavigator creates a new xpath.nodeNavigator for the specified html.node.
func newNavigator(top *node) *nodeNavigator {
	return &nodeNavigator{cur: top, root: top}
}

// nodeNavigator is for navigating JSON document.
type nodeNavigator struct {
	root, cur *node
}

func (a *nodeNavigator) Current() *node {
	return a.cur
}

func (a *nodeNavigator) NodeType() xpath.NodeType {
	switch a.cur.Type {
	case textNode:
		return xpath.TextNode
	case documentNode:
		return xpath.RootNode
	case elementNode:
		return xpath.ElementNode
	default:
		panic(fmt.Sprintf("unknown node type %v", a.cur.Type))
	}
}

func (a *nodeNavigator) getType() [2]string {
	typ := uast.TypeOf(a.cur.Node)
	if typ == "" {
		return [2]string{"", a.cur.Data}
	}
	i := strings.Index(typ, ":")
	if i < 0 {
		return [2]string{"", typ}
	}
	return [2]string{typ[:i], typ[i+1:]}
}
func (a *nodeNavigator) LocalName() string {
	return a.getType()[1]
}

func (a *nodeNavigator) Prefix() string {
	return a.getType()[0]
}

func (a *nodeNavigator) Value() string {
	switch a.cur.Type {
	case elementNode:
		return a.cur.InnerText()
	case textNode:
		return a.cur.Data
	}
	return ""
}

func (a *nodeNavigator) Copy() xpath.NodeNavigator {
	n := *a
	return &n
}

func (a *nodeNavigator) MoveToRoot() {
	a.cur = a.root
}

func (a *nodeNavigator) MoveToParent() bool {
	if n := a.cur.Parent; n != nil {
		a.cur = n
		return true
	}
	return false
}

func (x *nodeNavigator) MoveToNextAttribute() bool {
	return false
}

func (a *nodeNavigator) MoveToChild() bool {
	if n := a.cur.FirstChild; n != nil {
		a.cur = n
		return true
	}
	return false
}

func (a *nodeNavigator) MoveToFirst() bool {
	for n := a.cur.PrevSibling; n != nil; n = n.PrevSibling {
		a.cur = n
	}
	return true
}

func (a *nodeNavigator) String() string {
	return a.Value()
}

func (a *nodeNavigator) MoveToNext() bool {
	if n := a.cur.NextSibling; n != nil {
		a.cur = n
		return true
	}
	return false
}

func (a *nodeNavigator) MoveToPrevious() bool {
	if n := a.cur.PrevSibling; n != nil {
		a.cur = n
		return true
	}
	return false
}

func (a *nodeNavigator) MoveTo(other xpath.NodeNavigator) bool {
	node, ok := other.(*nodeNavigator)
	if !ok || node.root != a.root {
		return false
	}
	a.cur = node.cur
	return true
}
