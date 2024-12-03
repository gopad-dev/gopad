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
	BytesRange(r ByteRange) []byte
	// Rune returns the rune at the given index.
	Rune(i uint) []byte

	// ByteIndex converts the rune index to a byte index.
	ByteIndex(i uint) uint
	// RuneIndex converts the byte index to a rune index.
	RuneIndex(i uint) uint

	// ByteIndexByPoint returns the byte index for the given point.
	ByteIndexByPoint(p Point) uint
	// Point returns the point for the given byte index.
	Point(i uint) Point

	// Len returns the rune length of the buffer. The actual byte length may be different due to line endings & encoding.
	Len() uint
	// BytesLen returns the byte length of the buffer.
	BytesLen() uint
	// LinesLen returns the number of lines in the buffer.
	LinesLen() uint
	// LineLen returns the rune length of the line at the given row.
	LineLen(l uint) uint

	// Line returns the line at the given row.
	Line(l int) Line
	// Lines returns the lines in the buffer.
	Lines() []Line

	// Insert inserts text at the given position.
	Insert(i uint, text []byte)
	// Replace replaces the text in the given range with the given text.
	Replace(r ByteRange, text []byte)
	// Delete deletes the range of text between the two positions.
	Delete(r ByteRange)
}

// Line represents a line in a buffer.
type Line interface {
	// Clone returns a copy of the line.
	Clone() Line
	// Len returns the rune length of the line.
	Len() uint
	// BytesLen returns the byte length of the line.
	BytesLen() uint

	// ByteIndex converts the rune index to a byte index.
	ByteIndex(i uint) uint
	// Column converts the byte index to a rune index.
	Column(i uint) uint

	// Bytes returns the line as a byte slice.
	Bytes() []byte
	// BytesRange returns the line as a byte slice from the given range.
	BytesRange(r ByteRange) []byte

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
	Insert(i uint, text []byte) Line
	// Replace replaces the text between the given start and end indexes with the given text.
	Replace(r ByteRange, text []byte) Line
	// Delete deletes the text between the given start and end indexes.
	Delete(r ByteRange) Line
}
