package driver

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v0/uast"
)

func TestFillLineColFromOffset(t *testing.T) {
	require := require.New(t)

	var (
		data     []byte
		input    *uast.Node
		expected *uast.Node
		err      error
	)

	data = []byte("hello\n\nworld")
	input = &uast.Node{StartPosition: &uast.Position{Offset: 0}, EndPosition: &uast.Position{Offset: 4}, Children: []*uast.Node{
		{StartPosition: &uast.Position{Offset: 7}, EndPosition: &uast.Position{Offset: 11}},
	}}
	expected = &uast.Node{StartPosition: &uast.Position{Offset: 0, Line: 1, Col: 1}, EndPosition: &uast.Position{Offset: 4, Line: 1, Col: 5}, Children: []*uast.Node{
		{StartPosition: &uast.Position{Offset: 7, Line: 3, Col: 1}, EndPosition: &uast.Position{Offset: 11, Line: 3, Col: 5}},
	}}
	err = FillLineColFromOffset(data, input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestFillOffsetFromLineCol(t *testing.T) {
	require := require.New(t)

	var (
		data     []byte
		input    *uast.Node
		expected *uast.Node
		err      error
	)

	data = []byte("hello\n\nworld")
	input = &uast.Node{StartPosition: &uast.Position{Line: 1, Col: 1}, EndPosition: &uast.Position{Line: 1, Col: 5}, Children: []*uast.Node{
		{StartPosition: &uast.Position{Line: 3, Col: 1}, EndPosition: &uast.Position{Line: 3, Col: 5}},
	}}
	expected = &uast.Node{StartPosition: &uast.Position{Offset: 0, Line: 1, Col: 1}, EndPosition: &uast.Position{Offset: 4, Line: 1, Col: 5}, Children: []*uast.Node{
		{StartPosition: &uast.Position{Offset: 7, Line: 3, Col: 1}, EndPosition: &uast.Position{Offset: 11, Line: 3, Col: 5}},
	}}
	err = FillOffsetFromLineCol(data, input)
	require.NoError(err)
	require.Equal(expected, input)
}
