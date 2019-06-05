package positioner

import (
	"testing"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/stretchr/testify/require"
)

func offset(v int) nodes.Object {
	return uast.Position{Offset: uint32(v)}.ToObject()
}

func lineCol(line, col int) nodes.Object {
	return uast.Position{Line: uint32(line), Col: uint32(col)}.ToObject()
}

func fullPos(off, line, col int) nodes.Object {
	return uast.Position{Offset: uint32(off), Line: uint32(line), Col: uint32(col)}.ToObject()
}

func TestFillLineColNested(t *testing.T) {
	require := require.New(t)

	data := "hello\n\nworld"

	input := nodes.Object{
		"a": nodes.Object{
			uast.KeyStart: offset(0),
			uast.KeyEnd:   offset(4),
		},
		"b": nodes.Array{nodes.Object{
			uast.KeyStart: offset(7),
			uast.KeyEnd:   offset(12),
		}},
	}

	expected := nodes.Object{
		"a": nodes.Object{
			uast.KeyStart: fullPos(0, 1, 1),
			uast.KeyEnd:   fullPos(4, 1, 5),
		},
		"b": nodes.Array{nodes.Object{
			uast.KeyStart: fullPos(7, 3, 1),
			uast.KeyEnd:   fullPos(12, 3, 6),
		}},
	}

	p := FromOffset()
	out, err := p.OnCode(data).Do(input)
	require.NoError(err)
	require.Equal(expected, out)
}

func TestFillOffsetNested(t *testing.T) {
	require := require.New(t)

	data := "hello\n\nworld"

	input := nodes.Object{
		"a": nodes.Object{
			uast.KeyStart: lineCol(1, 1),
			uast.KeyEnd:   lineCol(1, 5),
		},
		"b": nodes.Array{nodes.Object{
			uast.KeyStart: lineCol(3, 1),
			uast.KeyEnd:   lineCol(3, 6),
		}},
	}

	expected := nodes.Object{
		"a": nodes.Object{
			uast.KeyStart: fullPos(0, 1, 1),
			uast.KeyEnd:   fullPos(4, 1, 5),
		},
		"b": nodes.Array{nodes.Object{
			uast.KeyStart: fullPos(7, 3, 1),
			uast.KeyEnd:   fullPos(12, 3, 6),
		}},
	}

	p := FromLineCol()
	out, err := p.OnCode(data).Do(input)
	require.NoError(err)
	require.Equal(expected, out)
}

func TestFillOffsetEmptyFile(t *testing.T) {
	require := require.New(t)

	data := ""

	input := nodes.Object{
		uast.KeyStart: lineCol(1, 1),
		uast.KeyEnd:   lineCol(1, 1),
	}

	expected := nodes.Object{
		uast.KeyStart: fullPos(0, 1, 1),
		uast.KeyEnd:   fullPos(0, 1, 1),
	}

	p := FromLineCol()
	out, err := p.OnCode(data).Do(input)
	require.NoError(err)
	require.Equal(expected, out)
}

func TestPosIndexSpans(t *testing.T) {
	const source = `line1
𝓏𝓏2
ё3
a4`
	ind := NewIndex([]byte(source), &IndexOptions{Unicode: true})
	require.Equal(t, []runeSpan{
		{firstRuneInd: 0, firstUTF16Ind: 0, byteOff: 0, runeSize8: 1, runeSize16: 1, numRunes: 6},
		{firstRuneInd: 6, firstUTF16Ind: 6, byteOff: 6, runeSize8: 4, runeSize16: 2, numRunes: 2},
		{firstRuneInd: 8, firstUTF16Ind: 10, byteOff: 14, runeSize8: 1, runeSize16: 1, numRunes: 2},
		{firstRuneInd: 10, firstUTF16Ind: 12, byteOff: 16, runeSize8: 2, runeSize16: 1, numRunes: 1},
		{firstRuneInd: 11, firstUTF16Ind: 13, byteOff: 18, runeSize8: 1, runeSize16: 1, numRunes: 4},
	}, ind.spans)
}

func TestPosIndex(t *testing.T) {
	// Verify that a multi-byte Unicode rune does not displace offsets after
	// its occurrence in the input. Test few other simple cases as well.
	const source = `line1
ё2
a3`
	var cases = []uast.Position{
		{Offset: 0, Line: 1, Col: 1},
		{Offset: 4, Line: 1, Col: 5},

		// multi-byte unicode rune
		{Offset: 6, Line: 2, Col: 1},
		{Offset: 8, Line: 2, Col: 3}, // col is a byte offset+1, not a rune index

		{Offset: 10, Line: 3, Col: 1},
		{Offset: 11, Line: 3, Col: 2},

		// special case — EOF position
		{Offset: 12, Line: 3, Col: 3},
	}

	ind := NewIndex([]byte(source), nil)
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			line, col, err := ind.LineCol(int(c.Offset))
			require.NoError(t, err)
			require.Equal(t, c.Line, uint32(line))
			require.Equal(t, c.Col, uint32(col))

			off, err := ind.Offset(int(c.Line), int(c.Col))
			require.NoError(t, err)
			require.Equal(t, c.Offset, uint32(off))
		})
	}
}

func TestPosIndexUnicode(t *testing.T) {
	// Verify that a rune offset -> byte offset conversion works.
	const source = `line1
𝓏𝓏2
ё3
a4`
	var cases = []struct {
		runeOff   int
		byteOff   int
		line, col int
		runeCol   int
	}{
		{runeOff: 0, byteOff: 0, line: 1, col: 1, runeCol: 1},

		// first 4-byte rune
		{runeOff: 6, byteOff: 6, line: 2, col: 1, runeCol: 1},
		// second 4-byte rune
		{runeOff: 7, byteOff: 10, line: 2, col: 5, runeCol: 2},
		// end of the second rune
		{runeOff: 8, byteOff: 14, line: 2, col: 9, runeCol: 3},
		// EOL
		{runeOff: 9, byteOff: 15, line: 2, col: 10, runeCol: 4},

		// 2-byte rune
		{runeOff: 10, byteOff: 16, line: 3, col: 1, runeCol: 1},
		// end of the rune
		{runeOff: 11, byteOff: 18, line: 3, col: 3, runeCol: 2},
		// EOL
		{runeOff: 12, byteOff: 19, line: 3, col: 4, runeCol: 3},

		// last line with 1-byte runes
		{runeOff: 13, byteOff: 20, line: 4, col: 1, runeCol: 1},

		// special case — EOF position
		{runeOff: 15, byteOff: 22, line: 4, col: 3, runeCol: 3},
	}

	ind := newIndexUnicode([]byte(source))
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			off, err := ind.RuneOffset(c.runeOff)
			require.NoError(t, err)
			require.Equal(t, c.byteOff, off)

			roff, err := ind.ToRuneOffset(c.byteOff)
			require.NoError(t, err)
			require.Equal(t, c.runeOff, roff)

			// verify that offset -> line/col conversion still works
			line, col, err := ind.LineCol(off)
			require.NoError(t, err)
			require.Equal(t, c.line, line)
			require.Equal(t, c.col, col)

			// check Unicode line/col as well
			line, col, err = ind.ToUnicodeLineCol(off)
			require.NoError(t, err)
			require.Equal(t, c.line, line)
			require.Equal(t, c.runeCol, col)
		})
	}
}

func TestPosIndexUTF16(t *testing.T) {
	// Verify that a UTF-16 code unit offset -> byte offset conversion works.
	// Also test UTF-16 surrogate pairs.
	const source = `line1
𝓏𝓏2
ё3
a4`
	var cases = []struct {
		cuOff     int
		byteOff   int
		line, col int
		cuCol     int
	}{
		{cuOff: 0, byteOff: 0, line: 1, col: 1, cuCol: 1},

		// first 4-byte rune (surrogate pair; 2 code units)
		{cuOff: 6, byteOff: 6, line: 2, col: 1, cuCol: 1},
		// second 4-byte rune (surrogate pair; 2 code units)
		{cuOff: 8, byteOff: 10, line: 2, col: 5, cuCol: 3},
		// end of the second rune
		{cuOff: 10, byteOff: 14, line: 2, col: 9, cuCol: 5},
		// EOL
		{cuOff: 11, byteOff: 15, line: 2, col: 10, cuCol: 6},

		// 2-byte rune (1 code point)
		{cuOff: 12, byteOff: 16, line: 3, col: 1, cuCol: 1},
		// end of the rune
		{cuOff: 13, byteOff: 18, line: 3, col: 3, cuCol: 2},
		// EOL
		{cuOff: 14, byteOff: 19, line: 3, col: 4, cuCol: 3},

		// last line with 1-byte runes
		{cuOff: 15, byteOff: 20, line: 4, col: 1, cuCol: 1},

		// special case — EOF position
		{cuOff: 17, byteOff: 22, line: 4, col: 3, cuCol: 3},
	}

	ind := newIndexUnicode([]byte(source))
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			off, err := ind.UTF16Offset(c.cuOff)
			require.NoError(t, err)
			require.Equal(t, c.byteOff, off)

			roff, err := ind.ToUTF16Offset(c.byteOff)
			require.NoError(t, err)
			require.Equal(t, c.cuOff, roff)

			// verify that offset -> line/col conversion still works
			line, col, err := ind.LineCol(off)
			require.NoError(t, err)
			require.Equal(t, c.line, line)
			require.Equal(t, c.col, col)

			// check Unicode line/col as well
			line, col, err = ind.ToUTF16LineCol(off)
			require.NoError(t, err)
			require.Equal(t, c.line, line)
			require.Equal(t, c.cuCol, col)
		})
	}
}
