package positioner

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v1/uast"
)

func TestFillLineColFromOffset(t *testing.T) {
	require := require.New(t)

	data := "hello\n\nworld"

	input := uast.Object{
		uast.KeyStart: uast.Position{Offset: 0}.ToObject(),
		uast.KeyEnd:   uast.Position{Offset: 4}.ToObject(),
		"a": uast.Array{uast.Object{
			uast.KeyStart: uast.Position{Offset: 7}.ToObject(),
			uast.KeyEnd:   uast.Position{Offset: 12}.ToObject(),
		}},
	}

	expected := uast.Object{
		uast.KeyStart: uast.Position{Offset: 0, Line: 1, Col: 1}.ToObject(),
		uast.KeyEnd:   uast.Position{Offset: 4, Line: 1, Col: 5}.ToObject(),
		"a": uast.Array{uast.Object{
			uast.KeyStart: uast.Position{Offset: 7, Line: 3, Col: 1}.ToObject(),
			uast.KeyEnd:   uast.Position{Offset: 12, Line: 3, Col: 6}.ToObject(),
		}},
	}

	p := NewFillLineColFromOffset()
	out, err := p.OnCode(data).Do(input)
	require.NoError(err)
	require.Equal(expected, out)
}

func TestFillOffsetFromLineCol(t *testing.T) {
	require := require.New(t)

	data := "hello\n\nworld"
	input := uast.Object{
		uast.KeyStart: uast.Position{Line: 1, Col: 1}.ToObject(),
		uast.KeyEnd:   uast.Position{Line: 1, Col: 5}.ToObject(),
		"a": uast.Array{uast.Object{
			uast.KeyStart: uast.Position{Line: 3, Col: 1}.ToObject(),
			uast.KeyEnd:   uast.Position{Line: 3, Col: 5}.ToObject(),
		}},
	}

	expected := uast.Object{
		uast.KeyStart: uast.Position{Offset: 0, Line: 1, Col: 1}.ToObject(),
		uast.KeyEnd:   uast.Position{Offset: 4, Line: 1, Col: 5}.ToObject(),
		"a": uast.Array{uast.Object{
			uast.KeyStart: uast.Position{Offset: 7, Line: 3, Col: 1}.ToObject(),
			uast.KeyEnd:   uast.Position{Offset: 11, Line: 3, Col: 5}.ToObject(),
		}},
	}

	p := NewFillOffsetFromLineCol()
	out, err := p.OnCode(data).Do(input)
	require.NoError(err)
	require.Equal(expected, out)
}

func TestEmptyFile(t *testing.T) {
	require := require.New(t)

	data := ""

	input := uast.Object{
		uast.KeyStart: uast.Position{Line: 1, Col: 1}.ToObject(),
		uast.KeyEnd:   uast.Position{Line: 1, Col: 1}.ToObject(),
	}

	expected := uast.Object{
		uast.KeyStart: uast.Position{Offset: 0, Line: 1, Col: 1}.ToObject(),
		uast.KeyEnd:   uast.Position{Offset: 0, Line: 1, Col: 1}.ToObject(),
	}

	p := NewFillOffsetFromLineCol()
	out, err := p.OnCode(data).Do(input)
	require.NoError(err)
	require.Equal(expected, out)
}
