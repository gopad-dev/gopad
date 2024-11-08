package buffer

import (
	"golang.org/x/text/encoding"
)

// BlockCommentToken represents how block comments are represented in a file.
type BlockCommentToken struct {
	// Start is the token that starts a block comment.
	Start string
	// End is the token that ends a block comment.
	End string
}

// Buffer keeps track of the contents of a file.
type Buffer interface {
	// Name returns the full path of the buffer.
	Name() string
	// FileName returns the file name of the buffer.
	FileName() string
	// Encoding returns the encoding of the buffer. If the encoding is not recognized, UTF-8 is returned.
	Encoding() encoding.Encoding
	// EncodingName returns the name of the encoding of the buffer.
	EncodingName() string
	// SetEncoding sets the encoding of the buffer.
	SetEncoding(encoding string)
	// LineEnding returns the line ending of the buffer.
	LineEnding() LineEnding
	// SetLineEnding sets the line ending of the buffer.
	SetLineEnding(lineEnding LineEnding)
	// Version returns the version of the buffer.
	Version() uint64
	// Checksum returns the sha256 checksum of the buffer.
	Checksum() []byte
	// Dirty returns whether the buffer has unsaved changes.
	Dirty() bool

	// Copy returns a copy of the buffer.
	Copy() Buffer
	// Save saves the buffer to the file it represents.
	Save() error
	// Rename renames the buffer and the file it represents.
	Rename(name string) error
	// Delete deletes the buffer and the file it represents.
	Delete() error

	// Bytes returns the buffer as a byte slice. This uses \n as the line ending.
	Bytes() []byte
	// BytesRange returns the buffer as a byte slice from the given range. This uses \n as the line ending.
	BytesRange(r Range) []byte
	// String returns the buffer as a string. This uses \n as the line ending.
	String() string

	// ByteIndex returns the byte index in the buffer for the given row and col.
	ByteIndex(p Point) int
	// Position returns the row and column for the given byte index.
	Position(i int) Point
	// Index returns the row and column for the given byte index.
	Index(i int) (int, int)
	// LinesLen returns the number of lines in the buffer.
	LinesLen() int
	// Lines returns the lines in the buffer.
	Lines() []Line
	// Len returns the rune length of the buffer. The actual byte length may be different due to line endings & encoding.
	Len() int
	// Line returns the line at the given row.
	Line(row int) Line
	// LineLen returns the rune length of the line at the given row.
	LineLen(row int) int

	// Insert inserts text at the given position.
	Insert(p Point, text []byte)
	// InsertNewLine inserts a new line at the given position.
	InsertNewLine(p Point)
	// Replace replaces the text in the given range with the given text.
	Replace(r Range, text []byte)
	// DuplicateLine duplicates the line at the given row.
	DuplicateLine(row int)
	// DeleteLine deletes the line at the given row.
	DeleteLine(row int)
	// DeleteBefore deletes count characters before the current position.
	DeleteBefore(p Point)
	// DeleteAfter deletes count characters after the current position.
	DeleteAfter(p Point)
	// DeleteRange deletes the range of text between the two positions.
	DeleteRange(r Range)
	// AddTab adds a tab character at the front of the current line.
	AddTab(row int)
	// RemoveTab removes a tab character at the front of the current line.
	RemoveTab(row int)
}

// Line represents a line in a buffer.
type Line interface {
	// Copy returns a copy of the line.
	Copy() Line
	// Len returns the rune length of the line.
	Len() int
	// LenBytes returns the byte length of the line.
	LenBytes() int
	// RuneIndex returns the byte index in the line for the given rune index.
	RuneIndex(index int) int

	// Rune returns the rune at the given index.
	Rune(index int) rune
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
	CutStart(index int) Line
	// CutEnd returns a new line with the text after the given index.
	CutEnd(index int) Line
	// CutRange returns a new line with the text between the given start and end indexes.
	CutRange(start int, end int) Line
	// Append appends the given line to the current line.
	Append(line Line) Line
	// AppendLines appends the given lines to the current line.
	AppendLines(lines ...Line) Line
	// Insert inserts text at the given index.
	Insert(index int, text []byte) Line
	// Replace replaces the text at the given index with the given text.
	Replace(index int, text []byte) Line
	// ReplaceRange replaces the text between the given start and end indexes with the given text.
	ReplaceRange(start int, end int, text []byte) Line
}
