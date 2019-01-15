package positioner

import (
	"fmt"
	"sort"
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

// runeSpan represents a sequence of UTF8 characters of the same size in bytes.
type runeSpan struct {
	firstRuneInd int // index of the first rune
	byteOff      int // bytes offset of the first rune
	numRunes     int // number of runes
	runeSize     int // in bytes
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
		runeSize: 1,
	}
	runes := 0
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
		if n == cur.runeSize {
			// continue this span
			cur.numRunes++
			runes++
			i += n - 1
			continue
		}
		if cur.numRunes != 0 {
			idx.spans = append(idx.spans, cur)
		}
		// make a new span
		cur = runeSpan{
			byteOff:      i,
			firstRuneInd: runes,
			numRunes:     1,
			runeSize:     n,
		}
		runes++
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

// Offset returns a zero-based byte offset given a one-based line and column.
// It returns an error if the given line and column are out of bounds.
func (idx *positionIndex) Offset(line, col int) (int, error) {
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
func (idx *positionIndex) RuneOffset(offset int) (int, error) {
	var last int
	if len(idx.spans) != 0 {
		s := idx.spans[len(idx.spans)-1]
		last = s.firstRuneInd + s.numRunes
	}
	if offset == last {
		// special case — EOF position
		return idx.size, nil
	}
	if offset < 0 || offset >= last {
		return -1, fmt.Errorf("rune out of bounds: %d [%d, %d)", offset, 0, last)
	}
	i := sort.Search(len(idx.spans), func(i int) bool {
		s := idx.spans[i]
		return offset < s.firstRuneInd
	})
	s := idx.spans[i-1]
	return s.byteOff + s.runeSize*(offset-s.firstRuneInd), nil
}
