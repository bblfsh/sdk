package positioner

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/transformer"
)

var _ transformer.CodeTransformer = Positioner{}

const cloneObj = false

var (
	errNoUnicodeIndex = errors.New("unicode index is disabled")
)

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
// interpreting their Offset as a 0-based UTF-16 code unit index.
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
	idx := NewIndex([]byte(code), &IndexOptions{Unicode: t.unicode})
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
	off, err := idx.FromRuneOffset(int(pos.Offset))
	if err != nil {
		return err
	}
	pos.Offset = uint32(off)
	return fromOffset(idx, pos)
}

func fromUTF16Offset(idx *Index, pos *uast.Position) error {
	off, err := idx.FromUTF16Offset(int(pos.Offset))
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
	firstUTF16Ind int // index of the first UTF-16 code unit
	byteOff       int // bytes offset of the first rune

	// size invariant:
	// runeSize8 >= runeSize16

	runeSize8  int // in bytes
	runeSize16 int // in utf16 code units (2 = surrogate pair)

	numRunes int // number of runes in this span
}

// Index is a positional index.
type Index struct {
	offsetByLine []int
	spans        []runeSpan // if nil, multi-byte rune indexing is disabled
	size         int
}

// IndexOptions is a set of options for positional index.
type IndexOptions struct {
	// Unicode flag controls if an index is build to accept UTF-8/UTF-16 rune offsets in
	// addition to byte offsets.
	// If the flag is not set, RuneOffset and UTF16Offset will always fail with an error.
	Unicode bool
}

// NewIndex creates a new positional index.
// If opt == nil, all options default to their zero values.
func NewIndex(data []byte, opt *IndexOptions) *Index {
	if opt == nil {
		opt = &IndexOptions{}
	}
	if opt.Unicode {
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
		size:  len(data),
		spans: []runeSpan{}, // indicates that unicode index is enabled
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

		if n != cur.runeSize8 {
			if cur.numRunes != 0 {
				// save previous span
				idx.spans = append(idx.spans, cur)
			}
			// start a new span
			cur = runeSpan{
				byteOff:       i,
				firstRuneInd:  runes,
				firstUTF16Ind: codePoints,
				runeSize8:     n,
				runeSize16:    1,
			}
			if r1, r2 := utf16.EncodeRune(r); r1 != utf8.RuneError || r2 != utf8.RuneError {
				// surrogate pair: needs two UTF-16 code units
				cur.runeSize16 = 2
			}
		}

		cur.numRunes++
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

// lineOffset returns a byte offset for a given line.
// The line number must be valid, or the function will panic.
func (idx *Index) lineOffset(line int) int {
	return idx.offsetByLine[line-1]
}

// lineEnd returns an inclusive end byte offset for a given line.
// The line number must be valid, or the function will panic.
func (idx *Index) lineEnd(line int) int {
	if line == len(idx.offsetByLine) {
		return idx.size - 1
	}
	return idx.offsetByLine[line] - 1
}

// offsetToLine returns a line given a byte offset.
func (idx *Index) offsetToLine(offset int) (int, error) {
	line := sort.Search(len(idx.offsetByLine), func(i int) bool {
		return offset < idx.offsetByLine[i]
	})
	if line <= 0 || line > len(idx.offsetByLine) {
		return -1, fmt.Errorf("offset not found in index: %d", offset)
	}
	return line, nil
}

// checkByteOffset checks if the byte offset is in bounds.
// As a special case, it returns io.EOF if the offset equals to the size of the file.
func (idx *Index) checkByteOffset(offset int) error {
	last := idx.size
	if offset < 0 || offset > last {
		return fmt.Errorf("byte offset out of bounds: %d [%d, %d]", offset, 0, last)
	} else if offset == last {
		return io.EOF
	}
	return nil
}

// LineCol returns a one-based line and one-based column offset given a zero-based byte offset within a file.
// It returns an error if the given offset is out of bounds.
func (idx *Index) LineCol(offset int) (int, int, error) {
	err := idx.checkByteOffset(offset)
	if err != nil && err != io.EOF {
		return -1, -1, err
	}

	line, err := idx.offsetToLine(offset)
	if err != nil {
		return -1, -1, err
	}
	return line, offset - idx.lineOffset(line) + 1, nil
}

func (idx *Index) checkLine(line int) error {
	const minLine = 1
	maxLine := len(idx.offsetByLine)
	if line < minLine || line > maxLine {
		return fmt.Errorf("line out of bounds: %d [%d, %d]", line, minLine, maxLine)
	}
	return nil
}

// Offset returns a zero-based byte offset within a file given a one-based line and one-based column offset.
// It returns an error if the given line and column are out of bounds.
func (idx *Index) Offset(line, col int) (int, error) {
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
		return -1, fmt.Errorf("column out of bounds: %d [%d, %d]", col, minCol, maxCol)
	}
	return lineOffset + col - 1, nil
}

// unicodeOffset returns a zero-based byte offset within a file given a zero-based Unicode character offset or
// UTF-16 code unit offset within a file.
func (idx *Index) unicodeOffset(offset int, isUTF16 bool) (int, error) {
	if idx.spans == nil {
		return -1, errNoUnicodeIndex
	}
	var last int
	if len(idx.spans) != 0 {
		s := idx.spans[len(idx.spans)-1]
		if isUTF16 {
			last = s.firstUTF16Ind + s.numRunes*s.runeSize16
		} else {
			last = s.firstRuneInd + s.numRunes
		}
	}
	if offset < 0 || offset > last {
		str := "rune"
		if isUTF16 {
			str = "code unit"
		}
		return -1, fmt.Errorf("%s out of bounds: %d [%d, %d)", str, offset, 0, last)
	} else if offset == last {
		// special case — EOF position
		return idx.size, nil
	}
	cmp := func(i int) bool {
		return offset < idx.spans[i].firstRuneInd
	}
	if isUTF16 {
		cmp = func(i int) bool {
			return offset < idx.spans[i].firstUTF16Ind
		}
	}
	i := sort.Search(len(idx.spans), cmp)
	s := idx.spans[i-1]
	if isUTF16 {
		return s.byteOff + s.runeSize8*((offset-s.firstUTF16Ind)/s.runeSize16), nil
	}
	return s.byteOff + s.runeSize8*(offset-s.firstRuneInd), nil
}

// FromRuneOffset returns a zero-based byte offset given a zero-based Unicode character offset within a file.
func (idx *Index) FromRuneOffset(offset int) (int, error) {
	return idx.unicodeOffset(offset, false)
}

// RuneOffset returns a zero-based byte offset given a zero-based Unicode character offset within a file.
//
// Deprecated: use FromRuneOffset.
func (idx *Index) RuneOffset(offset int) (int, error) {
	return idx.unicodeOffset(offset, false)
}

// FromUTF16Offset returns a zero-based byte offset given a zero-based UTF-16 code unit offset within a file.
func (idx *Index) FromUTF16Offset(offset int) (int, error) {
	return idx.unicodeOffset(offset, true)
}

// UTF16Offset returns a zero-based byte offset given a zero-based UTF-16 code unit offset within a file.
//
// Deprecated: use FromUTF16Offset
func (idx *Index) UTF16Offset(offset int) (int, error) {
	return idx.unicodeOffset(offset, true)
}

// toUnicodeOffset returns a zero-based Unicode character offset or a UTF-16 code unit offset given a zero-based byte offset within a file.
func (idx *Index) toUnicodeOffset(offset int, isUTF16 bool) (int, error) {
	if idx.spans == nil {
		return -1, errNoUnicodeIndex
	}
	err := idx.checkByteOffset(offset)
	if err == io.EOF {
		// special case — EOF position
		if len(idx.spans) == 0 {
			return 0, nil
		}
		s := idx.spans[len(idx.spans)-1]
		if isUTF16 {
			return s.firstUTF16Ind + s.numRunes*s.runeSize16, nil
		}
		return s.firstRuneInd + s.numRunes, nil
	} else if err != nil {
		return -1, err
	}
	i := sort.Search(len(idx.spans), func(i int) bool {
		return offset < idx.spans[i].byteOff
	})
	s := idx.spans[i-1]
	if isUTF16 {
		return s.firstUTF16Ind + (offset-s.byteOff)/s.runeSize16, nil
	}
	return s.firstRuneInd + (offset-s.byteOff)/s.runeSize8, nil
}

// ToRuneOffset returns a zero-based Unicode character offset given a zero-based byte offset within a file.
func (idx *Index) ToRuneOffset(offset int) (int, error) {
	return idx.toUnicodeOffset(offset, false)
}

// ToUTF16Offset returns a zero-based UTF-16 code unit offset given a zero-based byte offset within a file.
func (idx *Index) ToUTF16Offset(offset int) (int, error) {
	return idx.toUnicodeOffset(offset, true)
}

// toUnicodeLineCol returns a one-based line and one-based column in Unicode characters or a UTF-16 code units given a
// zero-based byte offset within a file.
func (idx *Index) toUnicodeLineCol(offset int, isUTF16 bool) (int, int, error) {
	if idx.spans == nil {
		return -1, -1, errNoUnicodeIndex
	}
	err := idx.checkByteOffset(offset)
	if err != nil && err != io.EOF {
		return -1, -1, err
	}
	line, err := idx.offsetToLine(offset)
	if err != nil {
		return -1, -1, err
	}
	spans := idx.spans

	// find start span (line start)
	lineStart := idx.lineOffset(line)
	i := sort.Search(len(spans), func(i int) bool {
		return lineStart < spans[i].byteOff
	})
	spans = spans[i-1:]

	// find end span (the input offset)
	i = sort.Search(len(spans), func(i int) bool {
		return offset < spans[i].byteOff
	})
	spans = spans[:i]

	// scan spans and calculate the column
	offset -= lineStart
	col := 0
	for offset > 0 && len(spans) > 0 {
		s := spans[0]
		spans = spans[1:]
		n := s.numRunes
		if spanBytes := n * s.runeSize8; offset < spanBytes {
			n = offset / s.runeSize8
			offset = 0
		} else {
			offset -= spanBytes
		}
		if isUTF16 {
			col += n * s.runeSize16
		} else {
			col += n
		}
	}
	col++ // one-based
	return line, col, nil
}

// ToUnicodeLineCol returns a one-based line and one-based col in Unicode characters given a zero-based byte offset within a file.
// It returns an error if the given offset is out of bounds.
func (idx *Index) ToUnicodeLineCol(offset int) (int, int, error) {
	return idx.toUnicodeLineCol(offset, false)
}

// ToUTF16LineCol returns a one-based line and one-based col in UTF-16 code units given a zero-based byte offset within a file.
// It returns an error if the given offset is out of bounds.
func (idx *Index) ToUTF16LineCol(offset int) (int, int, error) {
	return idx.toUnicodeLineCol(offset, true)
}
