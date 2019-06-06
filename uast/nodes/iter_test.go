package nodes_test

import (
	"testing"

	"github.com/bblfsh/sdk/v3/uast/nodes"
)

func TestIterPreOrder(t *testing.T) {
	a := nodes.Object{
		"k1": nodes.Int(1),
		"k2": nodes.Int(2),
	}
	b := nodes.String("v")
	root := nodes.Array{a, b}

	it := nodes.NewIterator(root, nodes.PreOrder)
	got := allNodes(it)
	exp := []nodes.Node{
		root,
		a, nodes.Int(1), nodes.Int(2),
		b,
	}
	if !nodes.Equal(nodes.Array(exp), nodes.Array(got)) {
		t.Fatalf("wrong order: %v", got)
	}
}

func TestIterPostOrder(t *testing.T) {
	a := nodes.Object{
		"k1": nodes.Int(1),
		"k2": nodes.Int(2),
	}
	b := nodes.String("v")
	root := nodes.Array{a, b}

	it := nodes.NewIterator(root, nodes.PostOrder)
	got := allNodes(it)
	exp := []nodes.Node{
		nodes.Int(1), nodes.Int(2), a,
		b,
		root,
	}
	if !nodes.Equal(nodes.Array(exp), nodes.Array(got)) {
		t.Fatalf("wrong order: %v", got)
	}
}

func TestIterLevelOrder(t *testing.T) {
	a := nodes.Object{
		"k1": nodes.Int(1),
		"k2": nodes.Int(2),
	}
	b := nodes.String("v")
	root := nodes.Array{a, b}

	it := nodes.NewIterator(root, nodes.LevelOrder)
	got := allNodes(it)
	exp := []nodes.Node{
		root,
		a, b,
		nodes.Int(1), nodes.Int(2),
	}
	if !nodes.Equal(nodes.Array(exp), nodes.Array(got)) {
		t.Fatalf("wrong order: %v", got)
	}
}

func TestIterChildren(t *testing.T) {
	a := nodes.Object{
		"k1": nodes.Int(1),
		"k2": nodes.Array{nodes.Int(2)},
	}
	b := nodes.String("v")
	c := nodes.Array{b}
	root := nodes.Array{a, b, c}

	it := nodes.NewIterator(root, nodes.ChildrenOrder)
	got := allNodes(it)
	exp := []nodes.Node{a, c}
	if !nodes.Equal(nodes.Array(exp), nodes.Array(got)) {
		t.Fatalf("wrong order: %v", got)
	}
}

func allNodes(it nodes.Iterator) []nodes.Node {
	var out []nodes.Node
	for it.Next() {
		n, _ := it.Node().(nodes.Node)
		out = append(out, n)
	}
	return out
}
