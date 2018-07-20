package xpath

import (
	"bytes"
	"fmt"
	"sort"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

// A nodeType is the type of a node.
type nodeType uint

const (
	// documentNode is a document object that, as the root of the document tree,
	// provides access to the entire XML document.
	documentNode nodeType = iota
	// elementNode is an element.
	elementNode
	// textNode is the text content of a node.
	textNode
)

// A node consists of a nodeType and some Data (tag name for
// element nodes, content for text) and are part of a tree of Nodes.
type node struct {
	Parent, PrevSibling, NextSibling, FirstChild, LastChild *node

	Type nodeType
	Data string
	Node nodes.Node

	level int
}

// ChildNodes gets all child nodes of the node.
func (n *node) ChildNodes() []*node {
	var a []*node
	for nn := n.FirstChild; nn != nil; nn = nn.NextSibling {
		a = append(a, nn)
	}
	return a
}

// InnerText gets the value of the node and all its child nodes.
func (n *node) InnerText() string {
	var output func(*bytes.Buffer, *node)
	output = func(buf *bytes.Buffer, n *node) {
		if n.Type == textNode {
			buf.WriteString(n.Data)
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			output(buf, child)
		}
	}
	var buf bytes.Buffer
	output(&buf, n)
	return buf.String()
}

// SelectElement finds the first of child elements with the
// specified name.
func (n *node) SelectElement(name string) *node {
	for nn := n.FirstChild; nn != nil; nn = nn.NextSibling {
		if nn.Data == name {
			return nn
		}
	}
	return nil
}

func parseValue(x nodes.Node, top *node, level int) {
	top.Node = x
	addNode := func(n *node) {
		if n.level == top.level {
			top.NextSibling = n
			n.PrevSibling = top
			n.Parent = top.Parent
			if top.Parent != nil {
				top.Parent.LastChild = n
			}
		} else if n.level > top.level {
			n.Parent = top
			if top.FirstChild == nil {
				top.FirstChild = n
				top.LastChild = n
			} else {
				t := top.LastChild
				t.NextSibling = n
				n.PrevSibling = t
				top.LastChild = n
			}
		}
	}
	switch v := x.(type) {
	case nodes.Array:
		for _, vv := range v {
			n := &node{Type: elementNode, level: level, Node: vv}
			addNode(n)
			parseValue(vv, n, level+1)
		}
	case nodes.Object:
		// The Go’s map iteration order is random.
		// (https://blog.golang.org/go-maps-in-action#Iteration-order)
		var keys []string
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			vv := v[key]
			n := &node{Data: key, Type: elementNode, level: level, Node: vv}
			addNode(n)
			parseValue(vv, n, level+1)
		}
	case nodes.String:
		n := &node{Data: string(v), Type: textNode, level: level, Node: v}
		addNode(n)
	case nodes.Value:
		s := fmt.Sprint(v)
		n := &node{Data: s, Type: textNode, level: level, Node: v}
		addNode(n)
	}
}

func conv(v nodes.Node) *node {
	doc := &node{Type: documentNode, Node: v}
	parseValue(v, doc, 1)
	return doc
}
