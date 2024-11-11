package buffer

import (
	"fmt"
	"io"
)

var (
	lineEndingCRLF = []byte("\r\n")
	lineEndingLF   = []byte("\n")
)

type LineEnding int

const (
	LineEndingAuto LineEnding = iota
	LineEndingLF
	LineEndingCRLF
)

func (l LineEnding) String() string {
	switch l {
	case LineEndingAuto:
		return "Auto"
	case LineEndingLF:
		return "LF"
	case LineEndingCRLF:
		return "CRLF"
	}
	return "Unknown"
}

func (l LineEnding) Bytes() []byte {
	switch l {
	case LineEndingCRLF:
		return lineEndingCRLF
	case LineEndingLF:
		return lineEndingLF
	default:
		panic(fmt.Sprintf("unknown line ending: %d", l))
	}
}

// Buffer keeps track of the contents of a file.
type Buffer interface {
	io.WriterTo

	// LineEnding returns the line ending of the buffer.
	LineEnding() LineEnding
	// SetLineEnding sets the line ending of the buffer.
	SetLineEnding(lineEnding LineEnding)
	// Version returns the version of the buffer.
	Version() uint64
	// Checksum returns the sha256 checksum of the buffer.
	Checksum() []byte
	// UpdateChecksum recalculates the sha256 checksum of the buffer.
	UpdateChecksum() error
	// Dirty returns whether the buffer has been modified.
	Dirty() bool

	// Clone returns a copy of the buffer.
	Clone() Buffer

	// Bytes returns the buffer as a byte slice. This uses \n as the line ending.
	Bytes() []byte
	// BytesRange returns the buffer as a byte slice from the given range. This uses \n as the line ending.
	BytesRange(r Range) []byte

	// ByteIndex returns the byte index in the buffer for the given point.
	ByteIndex(p Point) int
	// Position returns the point for the given byte index.
	Position(i int) Point

	// Len returns the rune length of the buffer. The actual byte length may be different due to line endings & encoding.
	Len() int
	// LinesLen returns the number of lines in the buffer.
	LinesLen() int
	// Lines returns the lines in the buffer.
	Lines() []Line

	// Line returns the line at the given row.
	Line(l int) Line
	// LineLen returns the rune length of the line at the given row.
	LineLen(l int) int

	// Insert inserts text at the given position.
	Insert(p Point, text []byte)
	// Replace replaces the text in the given range with the given text.
	Replace(r Range, text []byte)
	// Delete deletes the range of text between the two positions.
	Delete(r Range)
}

// Line represents a line in a buffer.
type Line interface {
	// Clone returns a copy of the line.
	Clone() Line
	// Len returns the rune length of the line.
	Len() int
	// BytesLen returns the byte length of the line.
	BytesLen() int
	// Index returns the byte index in the line for the given rune index.
	Index(i int) int

	// Rune returns the rune at the given index.
	Rune(i int) rune
	// Runes returns the runes in the line.
	Runes() []rune
	// RunesRange returns the runes in the line from the given range.
	RunesRange(start int, end int) []rune

	// RuneBytes returns the rune at the given index as a byte slice.
	RuneBytes(index int) []byte
	// Bytes returns the line as a byte slice.
	Bytes() []byte
	// RunesBytesRange returns the runes in the line from the given range as a byte slice.
	RunesBytesRange(start int, end int) []byte

	// RuneString returns the rune at the given index as a string.
	RuneString(index int) string
	// String returns the line as a string.
	String() string
	// RuneStrings returns the runes in the line as strings.
	RuneStrings() []string
	// StringRange returns the string in the line from the given range.
	StringRange(start int, end int) string

	// CutStart returns a new line with the text before the given index.
	CutStart(i int) Line
	// CutEnd returns a new line with the text after the given index.
	CutEnd(i int) Line
	// CutRange returns a new line with the text between the given start and end indexes.
	CutRange(start int, end int) Line

	// Append appends the given lines to the current line.
	Append(lines ...Line) Line
	// Prepend prepends the given lines to the current line.
	Prepend(lines ...Line) Line
	// Insert inserts text at the given index.
	Insert(i int, text []byte) Line
	// Replace replaces the text between the given start and end indexes with the given text.
	Replace(start int, end int, text []byte) Line
	// Delete deletes the text between the given start and end indexes.
	Delete(start int, end int) Line
}
