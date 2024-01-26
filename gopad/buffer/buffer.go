package buffer

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"slices"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"go.gopad.dev/gopad/internal/bubbles/searchbar"
	"go.gopad.dev/gopad/internal/xbytes"
)

// New creates a new buffer from an io.Reader.
func New(name string, r io.Reader, encoding string, lineEnding LineEnding, onDisk bool) (*Buffer, error) {
	fileEncoding, err := htmlindex.Get(encoding)
	if err != nil {
		fileEncoding = unicode.UTF8
	}

	hasher := sha256.New()

	br := bufio.NewReader(io.TeeReader(transform.NewReader(r, fileEncoding.NewDecoder()), hasher))
	var lines []Line
	for {
		data, err := br.ReadBytes('\n')
		if len(data) > 0 {
			if data[len(data)-1] == '\n' {
				data = data[:len(data)-1]
			}

			if len(data) > 0 && data[len(data)-1] == '\r' {
				data = data[:len(data)-1]
				if lineEnding == LineEndingAuto {
					lineEnding = LineEndingCRLF
				}
			} else if lineEnding == LineEndingAuto {
				lineEnding = LineEndingLF
			}

			lines = append(lines, NewLine(data))
		} else {
			lines = append(lines, NewEmptyLine())
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}
	}

	b := &Buffer{
		name:       name,
		encoding:   encoding,
		lineEnding: lineEnding,
		lines:      lines,
		checksum:   hasher.Sum(nil),
		onDisk:     onDisk,
	}

	return b, nil
}

// NewFromFile creates a new buffer from a file on disk.
func NewFromFile(name string, encoding string, lineEnding LineEnding) (*Buffer, error) {
	var err error
	name, err = filepath.Abs(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute file path: %w", err)
	}
	file, err := readFile(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return New(name, file, encoding, lineEnding, true)
}

// Buffer represents a file in memory.
type Buffer struct {
	name       string
	encoding   string
	lineEnding LineEnding
	version    int32
	lines      []Line
	checksum   []byte
	onDisk     bool
	dirty      bool
}

// Name returns the full path of the buffer.
func (b *Buffer) Name() string {
	return b.name
}

// FileName returns the file name of the buffer.
func (b *Buffer) FileName() string {
	return filepath.Base(b.name)
}

// Encoding returns the encoding of the buffer. If the encoding is not recognized, UTF-8 is returned.
func (b *Buffer) Encoding() encoding.Encoding {
	fileEncoding, err := htmlindex.Get(b.encoding)
	if err != nil {
		return unicode.UTF8
	}
	return fileEncoding
}

func (b *Buffer) EncodingName() string {
	return b.encoding
}

// SetEncoding sets the encoding of the buffer.
func (b *Buffer) SetEncoding(encoding string) {
	b.encoding = encoding
}

// LineEnding returns the line ending of the buffer.
func (b *Buffer) LineEnding() LineEnding {
	return b.lineEnding
}

// SetLineEnding sets the line ending of the buffer.
func (b *Buffer) SetLineEnding(lineEnding LineEnding) {
	b.lineEnding = lineEnding
}

// Version returns the version of the buffer.
func (b *Buffer) Version() int32 {
	return b.version
}

// Checksum returns the sha256 checksum of the buffer.
func (b *Buffer) Checksum() []byte {
	return b.checksum
}

// Dirty returns whether the buffer has unsaved changes.
func (b *Buffer) Dirty() bool {
	if !b.onDisk {
		return true
	}
	return b.dirty
}

// Save saves the buffer to the file it represents.
func (b *Buffer) Save() error {
	file, err := writeFile(b.name)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	hasher := sha256.New()
	fileEncoding := b.Encoding()
	eol := b.lineEnding.Bytes()

	w := bufio.NewWriter(transform.NewWriter(io.MultiWriter(hasher, file), fileEncoding.NewEncoder()))
	defer func() {
		w.Flush()
		file.Sync()
	}()

	for i, line := range b.lines {
		if line.Len() > 0 {
			if _, err = w.Write(line.data); err != nil {
				return fmt.Errorf("error writing line: %w", err)
			}
		}

		if i < len(b.lines)-1 {
			if _, err = w.Write(eol); err != nil {
				return fmt.Errorf("error writing line ending: %w", err)
			}
		}
	}

	b.dirty = false
	b.onDisk = true
	b.checksum = hasher.Sum(nil)

	return nil
}

// Rename renames the buffer and the file it represents.
func (b *Buffer) Rename(name string) error {
	if err := b.Save(); err != nil {
		return fmt.Errorf("failed to save new file: %w", err)
	}

	if err := renameFile(b.name, name); err != nil {
		return fmt.Errorf("failed to delete old file: %w", err)
	}

	b.name = name
	return nil
}

// Delete deletes the buffer and the file it represents.
func (b *Buffer) Delete() error {
	if err := deleteFile(b.name); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Search searches for the given term in the buffer.
func (b *Buffer) Search(term string) []searchbar.Result {
	var results []searchbar.Result

	bytesTerm := []byte(term)

	buf := b.Bytes()
	var lastIndex int
	for {
		index := bytes.Index(buf[lastIndex:], bytesTerm)
		if index == -1 {
			break
		}

		lastIndex += index
		endIndex := index + len(term) - 1

		row, col := b.Index(index)
		rowEnd, colEnd := b.Index(endIndex)
		results = append(results, searchbar.Result{
			RowStart: row,
			ColStart: col,
			RowEnd:   rowEnd,
			ColEnd:   colEnd + 1,
		})
	}

	return results
}

// ByteIndex returns the byte index in the buffer for the given row and col.
func (b *Buffer) ByteIndex(row int, col int) int {
	var n int
	for i, line := range b.lines {
		if i == row {
			return n + len(line.CutEnd(col).Bytes())
		}
		n += len(line.Bytes()) + 1
	}
	return n
}

// Bytes returns the buffer as a byte slice. This uses \n as the line ending.
func (b *Buffer) Bytes() []byte {
	var bs []byte
	for i, line := range b.lines {
		bs = append(bs, line.Bytes()...)
		if i < len(b.lines)-1 {
			bs = append(bs, '\n')
		}
	}
	return bs
}

func (b *Buffer) BytesRange(start Position, end Position) []byte {
	var bs []byte
	for i := start.Row; i <= end.Row; i++ {
		line := b.lines[i]
		if i == start.Row {
			line = line.CutStart(start.Col)
		}
		if i == end.Row {
			line = line.CutEnd(end.Col)
		}
		bs = append(bs, line.Bytes()...)
		if i < end.Row {
			bs = append(bs, '\n')
		}
	}

	return bs
}

// String returns the buffer as a string. This uses \n as the line ending.
func (b *Buffer) String() string {
	return string(b.Bytes())
}

func (b *Buffer) refreshDirty() {
	hasher := sha256.New()
	fileEncoding := b.Encoding()
	eol := b.lineEnding.Bytes()

	w := transform.NewWriter(hasher, fileEncoding.NewEncoder())

	for i, line := range b.lines {
		_, _ = w.Write(line.data)

		if i < len(b.lines)-1 {
			_, _ = w.Write(eol)
		}
	}

	checksum := hasher.Sum(nil)

	b.dirty = !bytes.Equal(b.checksum, checksum)
}

// Index returns the row and column for the given byte index.
func (b *Buffer) Index(index int) (int, int) {
	var n int
	for i, line := range b.lines {
		if n+len(line.Bytes()) >= index {
			return i, index - n
		}
		n += len(line.Bytes()) + 1
	}
	return len(b.lines) - 1, b.lines[len(b.lines)-1].Len()
}

// LinesLen returns the number of lines in the buffer.
func (b *Buffer) LinesLen() int {
	return len(b.lines)
}

// Lines returns the lines in the buffer.
func (b *Buffer) Lines() []Line {
	return b.lines
}

// Len returns the rune length of the buffer. The actual byte length may be different due to line endings & encoding.
func (b *Buffer) Len() int {
	var n int
	for _, line := range b.lines {
		n += line.Len() + 1
	}
	return n
}

// Line returns the line at the given row.
func (b *Buffer) Line(row int) Line {
	return b.lines[row]
}

// LineLen returns the rune length of the line at the given row.
func (b *Buffer) LineLen(row int) int {
	return b.lines[row].Len()
}

// InsertNewLine inserts a new line at the given position.
func (b *Buffer) InsertNewLine(row int, col int) (int, int) {
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	line := b.lines[row]
	b.lines[row] = line.CutEnd(col)
	b.lines = slices.Insert(b.lines, row+1, NewEmptyLine())
	b.lines[row+1] = line.CutStart(col)

	return row + 1, 0
}

// Insert inserts text at the given position.
func (b *Buffer) Insert(row int, col int, text []byte) (int, int) {
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	for _, r := range xbytes.Runes(text) {
		if r == '\n' {
			line := b.lines[row]
			b.lines[row] = line.CutEnd(col)
			b.lines = slices.Insert(b.lines, row+1, line.CutStart(col))
			row++
			col = 0
			continue
		}
		b.lines[row] = b.lines[row].Insert(col, []byte(string(r))...)
		col++
	}

	return row, col
}

func (b *Buffer) Replace(fromRow int, fromCol int, toRow int, toCol int, text []byte) (int, int) {
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	row, col := b.DeleteRange(fromRow, fromCol, toRow, toCol)
	return b.Insert(row, col, text)
}

// DuplicateLine duplicates the line at the given row.
func (b *Buffer) DuplicateLine(row int) int {
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	line := b.lines[row]
	b.lines = slices.Insert(b.lines, row+1, line.Copy())

	return row + 1
}

// DeleteLine deletes the line at the given row.
func (b *Buffer) DeleteLine(row int) int {
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	if row == 0 && len(b.lines) == 1 {
		b.lines[0] = NewEmptyLine()
		return row
	}
	if row == len(b.lines)-1 {
		b.lines = b.lines[:row]
		return max(row-1, 0)
	}
	b.lines = append(b.lines[:row], b.lines[row+1:]...)

	return row
}

// DeleteBefore deletes count characters before the current position.
func (b *Buffer) DeleteBefore(row int, col int, count int) (int, int) {
	if row == 0 && col == 0 {
		return row, col
	}
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	for i := 0; i < count; i++ {
		if col == 0 {
			line := b.lines[row]
			col = b.lines[row-1].Len()
			if line.Len() > 0 {
				b.lines[row-1] = b.lines[row-1].Append(line)
			}
			b.lines = append(b.lines[:row], b.lines[row+1:]...)
			row = max(row-1, 0)
		} else if col > 0 {
			b.lines[row] = b.lines[row].CutEnd(col - 1).Append(b.lines[row].CutStart(col))
			col = max(col-1, 0)
		} else if row > 0 {
			row = max(col-1, 0)
			col = b.lines[row].Len()
		}
	}

	return row, col
}

// DeleteAfter deletes count characters after the current position.
func (b *Buffer) DeleteAfter(row int, col int, count int) (int, int) {
	if row == len(b.lines)-1 && col == b.lines[row].Len() {
		return row, col
	}
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	for i := 0; i < count; i++ {
		if col == b.LineLen(row) {
			b.lines[row] = b.lines[row].Append(b.lines[row+1])
			b.lines = append(b.lines[:row+1], b.lines[row+2:]...)
		} else if col < b.lines[row].Len() {
			b.lines[row] = b.lines[row].CutEnd(col).Append(b.lines[row].CutStart(col + 1))
		}
	}

	return row, col
}

// DeleteRange deletes the range of text between the two positions.
func (b *Buffer) DeleteRange(startRow int, startCol int, endRow int, endCol int) (int, int) {
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	if startRow == endRow {
		b.lines[startRow] = b.lines[startRow].CutRange(startCol, endCol)
	} else {
		b.lines[startRow] = b.lines[startRow].CutEnd(startCol).Append(b.lines[endRow].CutStart(endCol))
		b.lines = append(b.lines[:startRow+1], b.lines[endRow+1:]...)
	}

	return startRow, startCol
}

// RemoveTab removes a tab character at the front of the current line.
func (b *Buffer) RemoveTab(row int, col int) (int, int) {
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	if b.lines[row].Rune(0) == '\t' {
		b.lines[row] = b.lines[row].CutStart(1)
		return row, col - 1
	}

	return row, col
}
