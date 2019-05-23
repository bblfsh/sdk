package positioner

import (
	"fmt"
	"sort"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/transformer"
)

var _ transformer.CodeTransformer = Positioner{}

const cloneObj = false

// FromLineCol fills the Offset field of all Position nodes by using their Line and Col.
func FromLineCol() Positioner {
	return Positioner{method: fromLineCol}
}

// FromOffset fills the Line and Col fields of all Position nodes by using their Offset.
func FromOffset() Positioner {
	return Positioner{method: fromOffset}
}

// FromUnicodeOffset fills the Line, Col and Offset fields of all Position nodes by
// interpreting their Offset as a 0-based Unicode character index.
func FromUnicodeOffset() Positioner {
	return Positioner{unicode: true, method: fromUnicodeOffset}
}

// FromUTF16Offset fills the Line, Col and Offset fields of all Position nodes by
// interpreting their Offset as a 0-based UTF-16 code point index.
func FromUTF16Offset() Positioner {
	return Positioner{unicode: true, method: fromUTF16Offset}
}

// Positioner is a transformation that only changes positional information.
// The transformation should be initialized with the source code by calling OnCode.
type Positioner struct {
	unicode bool
	method  func(*Index, *uast.Position) error
}

// OnCode uses the source code to update positional information of UAST nodes.
func (t Positioner) OnCode(code string) transformer.Transformer {
	idx := NewIndex([]byte(code), t.unicode)
	return transformer.TransformObjFunc(func(o nodes.Object) (nodes.Object, bool, error) {
		pos := uast.AsPosition(o)
		if pos == nil {
			return o, false, nil
		}
		if err := t.method(idx, pos); err != nil {
			return o, false, err
		}
		if cloneObj {
			o = o.CloneObject()
		}
		for k, v := range pos.ToObject() {
			o[k] = v
		}
		return o, cloneObj, nil
	})
}

func fromLineCol(idx *Index, pos *uast.Position) error {
	offset, err := idx.Offset(int(pos.Line), int(pos.Col))
	if err != nil {
		return err
	}
	pos.Offset = uint32(offset)
	return nil
}

func fromOffset(idx *Index, pos *uast.Position) error {
	line, col, err := idx.LineCol(int(pos.Offset))
	if err != nil {
		return err
	}
	pos.Line = uint32(line)
	pos.Col = uint32(col)
	return nil
}

func fromUnicodeOffset(idx *Index, pos *uast.Position) error {
	off, err := idx.RuneOffset(int(pos.Offset))
	if err != nil {
		return err
	}
	pos.Offset = uint32(off)
	return fromOffset(idx, pos)
}

func fromUTF16Offset(idx *Index, pos *uast.Position) error {
	off, err := idx.UTF16Offset(int(pos.Offset))
	if err != nil {
		return err
	}
	pos.Offset = uint32(off)
	return fromOffset(idx, pos)
}

// runeSpan represents a sequence of UTF8 characters of the same size in bytes.
type runeSpan struct {
	// offset/index invariant:
	// byteOff >= firstUTF16Ind >= firstRuneInd

	firstRuneInd  int // index of the first rune
	firstUTF16Ind int // index of the first UTF-16 code point
	byteOff       int // bytes offset of the first rune

	// size invariant:
	// runeSize8 >= runeSize16

	runeSize8  int // in bytes
	runeSize16 int // in utf16 code points (2 = surrogate pair)

	numRunes int // number of runes in this span
}

// Index is a positional index.
type Index struct {
	offsetByLine []int
	spans        []runeSpan
	size         int
}

// NewIndex creates a new positional index.
// Unicode flag controls if an index is build to accept UTF-8/UTF-16 rune offsets in
// addition to byte offsets.
func NewIndex(data []byte, unicode bool) *Index {
	if unicode {
		return newIndexUnicode(data)
	}
	idx := &Index{
		size: len(data),
	}
	idx.addLineOffset(0)
	for i, b := range data {
		if b == '\n' {
			idx.addLineOffset(i + 1)
		}
	}
	return idx
}

func newIndexUnicode(data []byte) *Index {
	idx := &Index{
		size: len(data),
	}
	idx.addLineOffset(0)

	cur := runeSpan{
		runeSize8:  1,
		runeSize16: 1,
	}
	runes := 0
	codePoints := 0
	// decode UTF8 runes and collect a slice of UTF8 character spans
	// each span only contains characters with the same size in bytes
	for i := 0; i < len(data); i++ {
		r, n := utf8.DecodeRune(data[i:])
		if n == 0 {
			break // EOF, should not happen
		}
		if r == '\n' {
			idx.addLineOffset(i + 1)
		}
		if n == cur.runeSize8 {
			// continue this span
			cur.numRunes++
			runes++
			codePoints += cur.runeSize16
			i += n - 1
			continue
		}
		if cur.numRunes != 0 {
			idx.spans = append(idx.spans, cur)
		}
		// make a new span
		cur = runeSpan{
			byteOff:       i,
			firstRuneInd:  runes,
			firstUTF16Ind: codePoints,
			numRunes:      1,
			runeSize8:     n,
			runeSize16:    1,
		}
		if r1, r2 := utf16.EncodeRune(r); r1 != utf8.RuneError || r2 != utf8.RuneError {
			// surrogate pair: needs two UTF-16 code points
			cur.runeSize16 = 2
		}
		runes++
		codePoints += cur.runeSize16
		i += n - 1
	}
	if cur.numRunes != 0 {
		idx.spans = append(idx.spans, cur)
	}
	return idx
}

func (idx *Index) addLineOffset(offset int) {
	idx.offsetByLine = append(idx.offsetByLine, offset)
}

// LineCol returns a one-based line and col given a zero-based byte offset.
// It returns an error if the given offset is out of bounds.
func (idx *Index) LineCol(offset int) (int, int, error) {
	var (
		minOffset = 0
		maxOffset = idx.size
	)

	if offset < minOffset || offset > maxOffset {
		return 0, 0, fmt.Errorf("offset out of bounds: %d [%d, %d]", offset, minOffset, maxOffset)
	}

	line := sort.Search(len(idx.offsetByLine), func(i int) bool {
		return offset < idx.offsetByLine[i]
	})
	if line <= 0 || line > len(idx.offsetByLine) {
		return 0, 0, fmt.Errorf("offset not found in index: %d", offset)
	}

	lineOffset := idx.offsetByLine[line-1]
	col := offset - lineOffset + 1
	return line, col, nil
}

// Offset returns a zero-based byte offset given a one-based line and column.
// It returns an error if the given line and column are out of bounds.
func (idx *Index) Offset(line, col int) (int, error) {
	var (
		minLine = 1
		maxLine = len(idx.offsetByLine)
		minCol  = 1
	)

	maxOffset := idx.size - 1

	if line < minLine || line > maxLine {
		return -1, fmt.Errorf("line out of bounds: %d [%d, %d]", line, minLine, maxLine)
	}

	nextLine := line
	line = line - 1
	if nextLine < len(idx.offsetByLine) {
		maxOffset = idx.offsetByLine[nextLine] - 1
	}

	maxCol := maxOffset - idx.offsetByLine[line] + 1

	// For empty files with 1-indexed drivers, set maxCol to 1
	if maxCol == 0 && col == 1 {
		maxCol = 1
	}

	if col < minCol || (maxCol > 0 && col-1 > maxCol) {
		return 0, fmt.Errorf("column out of bounds: %d [%d, %d]", col, minCol, maxCol)
	}

	offset := idx.offsetByLine[line] + col - 1
	return offset, nil
}

// RuneOffset returns a zero-based byte offset given a zero-based Unicode character offset.
func (idx *Index) RuneOffset(offset int) (int, error) {
	var last int
	if len(idx.spans) != 0 {
		s := idx.spans[len(idx.spans)-1]
		last = s.firstRuneInd + s.numRunes
	}
	if offset < 0 || offset > last {
		return -1, fmt.Errorf("rune out of bounds: %d [%d, %d)", offset, 0, last)
	} else if offset == last {
		// special case — EOF position
		return idx.size, nil
	}
	i := sort.Search(len(idx.spans), func(i int) bool {
		s := idx.spans[i]
		return offset < s.firstRuneInd
	})
	s := idx.spans[i-1]
	return s.byteOff + s.runeSize8*(offset-s.firstRuneInd), nil
}

// UTF16Offset returns a zero-based byte offset given a zero-based UTF-16 code point offset.
func (idx *Index) UTF16Offset(offset int) (int, error) {
	var last int
	if len(idx.spans) != 0 {
		s := idx.spans[len(idx.spans)-1]
		last = s.firstUTF16Ind + s.numRunes*s.runeSize16
	}
	if offset < 0 || offset > last {
		return -1, fmt.Errorf("code point out of bounds: %d [%d, %d)", offset, 0, last)
	} else if offset == last {
		// special case — EOF position
		return idx.size, nil
	}
	i := sort.Search(len(idx.spans), func(i int) bool {
		s := idx.spans[i]
		return offset < s.firstUTF16Ind
	})
	s := idx.spans[i-1]
	return s.byteOff + s.runeSize8*((offset-s.firstUTF16Ind)/s.runeSize16), nil
}
