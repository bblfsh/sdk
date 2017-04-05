package uast

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	fixtureDir = "fixtures"
)

func TestToNoderJava(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("java_example_1.json")
	require.NoError(err)

	var c ToNoder = &BaseToNoder{
		InternalTypeKey: "internalClass",
		LineKey:         "line",
	}
	n, err := c.ToNode(f)
	require.NoError(err)
	require.NotNil(n)
}

// Test for promoting a specific property-list elements to its own node
func TestPropertyListPromotionSpecific(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("java_example_1.json")
	require.NoError(err)

	var c ToNoder = &BaseToNoder{
		InternalTypeKey: "internalClass",
		LineKey:         "line",
	}
	n, err := c.ToNode(f)
	require.NoError(err)
	require.NotNil(n)

	child := findChildWithInternalType(n, "CompilationUnit.types")
	require.Nil(child)

	c = &BaseToNoder{
		InternalTypeKey: "internalClass",
		LineKey:         "line",
		PromotedPropertyLists: map[string]map[string]bool {
			"CompilationUnit" : { "types" : true },
		},
		PromoteAllPropertyLists: false,
	}

	n, err = c.ToNode(f)
	require.NoError(err)
	require.NotNil(n)
}

// Test promoting all property-list elements to its own node
func TestPropertyListPromotionAll(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("java_example_1.json")
	require.NoError(err)

	var c ToNoder = &BaseToNoder{
		InternalTypeKey: "internalClass",
		LineKey:         "line",
	}
	n, err := c.ToNode(f)
	require.NoError(err)
	require.NotNil(n)
	child := findChildWithInternalType(n, "CompilationUnit.types")
	require.Nil(child)

	c = &BaseToNoder{
		InternalTypeKey: "internalClass",
		LineKey:         "line",
		PromoteAllPropertyLists: true,
	}

	n, err = c.ToNode(f)
	require.NoError(err)
	require.NotNil(n)

	child = findChildWithInternalType(n, "CompilationUnit.types")
	require.NotNil(child)
}

func findChildWithInternalType(n *Node, internalType string) (*Node) {
	for _, child := range n.Children {
		if child.InternalType == internalType {
			return child
		}
	}
	return nil
}

func getFixture(name string) (map[string]interface{}, error) {
	path := filepath.Join(fixtureDir, name)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(f)
	data := map[string]interface{}{}
	if err := d.Decode(&data); err != nil {
		_ = f.Close()
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	return data, nil
}
