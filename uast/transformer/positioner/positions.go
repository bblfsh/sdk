package positioner

import (
	"fmt"
	"sort"
	"unicode/utf16"
	"unicode/utf8"

	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/transformer"
)

var _ transformer.CodeTransformer = Positioner{}

const cloneObj = false

// FromLineCol fills the Offset field of all Position nodes by using their Line and Col.
func FromLineCol() Positioner {
	return Positioner{method: fromLineCol}
}

// NewFillOffsetFromLineCol fills the Offset field of all Position nodes by using
// their Line and Col.
//
// Deprecated: see FromLineCol
func NewFillOffsetFromLineCol() Positioner {
	return FromLineCol()
}

// FromOffset fills the Line and Col fields of all Position nodes by using their Offset.
func FromOffset() Positioner {
	return Positioner{method: fromOffset}
}

// NewFillLineColFromOffset fills the Line and Col fields of all Position nodes by using
// their Offset.
//
// Deprecated: see FromOffset
func NewFillLineColFromOffset() Positioner {
	return FromOffset()
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
	method  func(*positionIndex, *uast.Position) error
}

// OnCode uses the source code to update positional information of UAST nodes.
func (t Positioner) OnCode(code string) transformer.Transformer {
	var idx *positionIndex
	if t.unicode {
		idx = newPositionIndexUnicode([]byte(code))
	} else {
		idx = newPositionIndex([]byte(code))
	}
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

func fromLineCol(idx *positionIndex, pos *uast.Position) error {
	offset, err := idx.Offset(int(pos.Line), int(pos.Col))
	if err != nil {
		return err
	}
	pos.Offset = uint32(offset)
	return nil
}

func fromOffset(idx *positionIndex, pos *uast.Position) error {
	line, col, err := idx.LineCol(int(pos.Offset))
	if err != nil {
		return err
	}
	pos.Line = uint32(line)
	pos.Col = uint32(col)
	return nil
}

func fromUnicodeOffset(idx *positionIndex, pos *uast.Position) error {
	off, err := idx.RuneOffset(int(pos.Offset))
	if err != nil {
		return err
	}
	pos.Offset = uint32(off)
	return fromOffset(idx, pos)
}

func fromUTF16Offset(idx *positionIndex, pos *uast.Position) error {
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

type positionIndex struct {
	offsetByLine []int
	spans        []runeSpan
	size         int
}

func newPositionIndex(data []byte) *positionIndex {
	idx := &positionIndex{
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

func newPositionIndexUnicode(data []byte) *positionIndex {
	idx := &positionIndex{
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

func (idx *positionIndex) addLineOffset(offset int) {
	idx.offsetByLine = append(idx.offsetByLine, offset)
}

func (idx *positionIndex) lineOffset(line int) int {
	return idx.offsetByLine[line-1]
}

func (idx *positionIndex) lineEnd(line int) int {
	if line == len(idx.offsetByLine) {
		return idx.size - 1
	}
	return idx.offsetByLine[line] - 1
}

// LineCol returns a one-based line and col given a zero-based byte offset.
// It returns an error if the given offset is out of bounds.
func (idx *positionIndex) LineCol(offset int) (int, int, error) {
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

func (idx *positionIndex) checkLine(line int) error {
	const minLine = 1
	maxLine := len(idx.offsetByLine)
	if line < minLine || line > maxLine {
		return fmt.Errorf("line out of bounds: %d [%d, %d]", line, minLine, maxLine)
	}
	return nil
}

// Offset returns a zero-based byte offset given a one-based line and column.
// It returns an error if the given line and column are out of bounds.
func (idx *positionIndex) Offset(line, col int) (int, error) {
	if err := idx.checkLine(line); err != nil {
		return -1, err
	}
	const minCol = 1

	lineOffset := idx.lineOffset(line)
	maxCol := idx.lineEnd(line) - lineOffset + 1

	// For empty files with 1-indexed drivers, set maxCol to 1
	if maxCol == 0 && col == 1 {
		maxCol = 1
	}

	if col < minCol || (maxCol > 0 && col-1 > maxCol) {
		return 0, fmt.Errorf("column out of bounds: %d [%d, %d]", col, minCol, maxCol)
	}
	return lineOffset + col - 1, nil
}

// RuneOffset returns a zero-based byte offset given a zero-based Unicode character offset.
func (idx *positionIndex) RuneOffset(offset int) (int, error) {
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
func (idx *positionIndex) UTF16Offset(offset int) (int, error) {
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
