package native

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bblfsh/sdk/uast"
	"github.com/stretchr/testify/require"
)

var (
	fixtureDir = "../../uast/fixtures"
)

func TestToNodeErrUnsupported(t *testing.T) {
	require := require.New(t)
	p := &ObjectToNoder{}
	n, err := p.ToNode(struct{}{})
	require.Error(err)
	require.Nil(n)
	require.True(ErrUnsupported.Is(err))
}

func TestToNodeErrEmptyAST(t *testing.T) {
	topLevelIsRootNode := false
	testToNodeErrEmptyAST(t, topLevelIsRootNode)
	topLevelIsRootNode = true
	testToNodeErrEmptyAST(t, topLevelIsRootNode)
}

func testToNodeErrEmptyAST(t *testing.T, topIsRoot bool) {
	require := require.New(t)
	empty := make(map[string]interface{})
	p := &ObjectToNoder{TopLevelIsRootNode: topIsRoot}
	n, err := p.ToNode(empty)
	require.Error(err)
	require.Nil(n)
	require.True(ErrEmptyAST.Is(err))
}

func TestToNodeErrUnexpectedObjectSize(t *testing.T) {
	require := require.New(t)
	multiRoot := make(map[string]interface{})
	multiRoot["a"] = 0
	multiRoot["b"] = 0
	p := &ObjectToNoder{}
	n, err := p.ToNode(multiRoot)
	require.Error(err)
	require.Nil(n)
	require.True(ErrUnexpectedObjectSize.Is(err))
}

func TestToNodeWithTopLevelRoot(t *testing.T) {
	require := require.New(t)

	root, err := getRootAtTopLevelFromFixture()
	require.Nil(err)

	p := &ObjectToNoder{
		TopLevelIsRootNode: true,
		InternalTypeKey:    "internalClass",
		LineKey:            "line",
	}

	n, err := p.ToNode(root)
	require.NoError(err)
	require.NotNil(n)
}

// Returns a fixture of an AST with its root at the top level, by
// reusing the already existing fixture at java_example_1; it strips a
// few object from the top levels of the fixture (the CompilationUnit,
// then the types, then picks the first type) util we are left with a
// AST node at its top level.
func getRootAtTopLevelFromFixture() (map[string]interface{}, error) {
	ast, err := getFixture("java_example_1.json")
	if err != nil {
		return nil, err
	}

	// strip the CompilationUnit object
	compilationUnit, ok := ast["CompilationUnit"]
	if !ok {
		return nil, fmt.Errorf("key not found: CompilationUnit")
	}
	ast, ok = compilationUnit.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid cast: compilationUnit to map[string]interface{}")
	}

	// get the list of types
	types, ok := ast["types"]
	if !ok {
		return nil, fmt.Errorf("key not found: types")
	}
	list, ok := types.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid cast: types to []interface{}")
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("empty list of types")
	}

	first, ok := list[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid cast: first to map[string]interface{}")
	}

	return first, nil
}

func TestToNoderJava(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("java_example_1.json")
	require.NoError(err)

	var p ToNoder = &ObjectToNoder{
		InternalTypeKey: "internalClass",
		LineKey:         "line",
	}
	n, err := p.ToNode(f)
	require.NoError(err)
	require.NotNil(n)
}

// Test for promoting a specific property-list elements to its own node
func TestPropertyListPromotionSpecific(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("java_example_1.json")
	require.NoError(err)

	p := &ObjectToNoder{
		InternalTypeKey: "internalClass",
		LineKey:         "line",
	}
	n, err := p.ToNode(f)
	require.NoError(err)
	require.NotNil(n)

	child := findChildWithInternalType(n, "CompilationUnit.types")
	require.Nil(child)

	p = &ObjectToNoder{
		InternalTypeKey: "internalClass",
		LineKey:         "line",
		PromotedPropertyLists: map[string]map[string]bool{
			"CompilationUnit": {"types": true},
		},
		PromoteAllPropertyLists: false,
	}

	n, err = p.ToNode(f)
	require.NoError(err)
	require.NotNil(n)
}

// Test promoting all property-list elements to its own node
func TestPropertyListPromotionAll(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("java_example_1.json")
	require.NoError(err)

	p := &ObjectToNoder{
		InternalTypeKey: "internalClass",
		LineKey:         "line",
	}
	n, err := p.ToNode(f)
	require.NoError(err)
	require.NotNil(n)
	child := findChildWithInternalType(n, "CompilationUnit.types")
	require.Nil(child)

	p = &ObjectToNoder{
		InternalTypeKey:         "internalClass",
		LineKey:                 "line",
		PromoteAllPropertyLists: true,
	}

	n, err = p.ToNode(f)
	require.NoError(err)
	require.NotNil(n)

	child = findChildWithInternalType(n, "CompilationUnit.types")
	require.NotNil(child)
}

func TestSyntheticTokens(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("java_example_1.json")
	require.NoError(err)

	var c ToNoder = &ObjectToNoder{
		InternalTypeKey: "internalClass",
		LineKey:         "line",
		SyntheticTokens: map[string]string{
			"CompilationUnit": "TestToken",
		},
	}
	n, err := c.ToNode(f)
	require.NoError(err)
	require.NotNil(n)
	child := findChildWithInternalType(n, "CompilationUnit")

	require.Nil(child)
	n, err = c.ToNode(f)
	require.NoError(err)
	require.NotNil(n)
	require.True(n.Token == "TestToken")

}

func findChildWithInternalType(n *uast.Node, internalType string) *uast.Node {
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
