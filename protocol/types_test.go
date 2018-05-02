package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/role"
)

func TestNodeMarshal(t *testing.T) {
	exp := &Node{
		InternalType:  "node",
		Properties:    map[string]string{"key": "val"},
		Children:      []*Node{NewNode()},
		Token:         "tok",
		StartPosition: &uast.Position{Offset: 7, Line: 2, Col: 1},
		EndPosition:   &uast.Position{Offset: 9, Line: 3, Col: 1},
		Roles:         []role.Role{role.File, role.Incomplete},
	}
	data, err := exp.Marshal()
	require.NoError(t, err)

	got := new(Node)
	err = got.Unmarshal(data)
	require.NoError(t, err)

	require.Equal(t, exp.String(), got.String())
}
