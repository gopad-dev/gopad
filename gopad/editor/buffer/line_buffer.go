package buffer

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"slices"

	"go.gopad.dev/gopad/internal/xbytes"
)

// New creates a new buffer from an io.Reader.
func New(r io.Reader, lineEnding LineEnding) (Buffer, error) {
	hasher := sha256.New()
	defer hasher.Reset()

	br := bufio.NewReader(io.TeeReader(r, hasher))
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
	checksum := hasher.Sum(nil)

	b := &lineBuffer{
		lineEnding: lineEnding,
		lines:      lines,
		checksum:   checksum,
		hasher:     hasher,
	}

	return b, nil
}

// NewFromFile creates a new buffer from a file on disk.
func NewFromFile(name string, lineEnding LineEnding) (Buffer, error) {
	var err error
	name, err = filepath.Abs(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute file path: %w", err)
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	return New(file, lineEnding)
}

type lineBuffer struct {
	lineEnding LineEnding
	lines      []Line
	checksum   []byte
	hasher     hash.Hash
}

func (b *lineBuffer) WriteTo(w io.Writer) (int64, error) {
	var n int64
	for _, line := range b.lines {
		n1, err := w.Write(line.Bytes())
		if err != nil {
			return n, err
		}
		n += int64(n1)

		n2, err := w.Write([]byte{byte(b.lineEnding)})
		if err != nil {
			return n, err
		}
		n += int64(n2)
	}
	return n, nil
}

func (b *lineBuffer) LineEnding() LineEnding {
	return b.lineEnding
}

func (b *lineBuffer) SetLineEnding(lineEnding LineEnding) {
	b.lineEnding = lineEnding
}

func (b *lineBuffer) Checksum() []byte {
	return b.checksum
}

func (b *lineBuffer) UpdateChecksum() error {
	defer b.hasher.Reset()

	for _, line := range b.lines {
		if _, err := b.hasher.Write(line.Bytes()); err != nil {
			return fmt.Errorf("failed to update checksum: %w", err)
		}
		if _, err := b.hasher.Write([]byte{byte(b.lineEnding)}); err != nil {
			return fmt.Errorf("failed to update checksum: %w", err)
		}
	}
	b.checksum = b.hasher.Sum(nil)

	return nil
}

func (b *lineBuffer) Dirty() bool {
	return !bytes.Equal(b.checksum, b.Checksum())
}

func (b *lineBuffer) Clone() Buffer {
	lines := make([]Line, len(b.lines))
	for i, line := range b.lines {
		lines[i] = line.Clone()
	}
	checksum := make([]byte, len(b.checksum))
	copy(checksum, b.checksum)

	return &lineBuffer{
		lineEnding: b.lineEnding,
		lines:      lines,
		checksum:   checksum,
	}
}

func (b *lineBuffer) Bytes() []byte {
	var bs []byte
	for i, line := range b.lines {
		bs = append(bs, line.Bytes()...)
		if i < len(b.lines)-1 {
			bs = append(bs, '\n')
		}
	}
	return bs
}

func (b *lineBuffer) BytesRange(r Range) []byte {
	if r.Start.Row == r.End.Row {
		return b.lines[r.Start.Row].CutRange(r.Start.Col, r.End.Col).Bytes()
	}

	var bs []byte
	for i := r.Start.Row; i <= r.End.Row; i++ {
		line := b.lines[i]
		if i == r.Start.Row {
			line = line.CutStart(r.Start.Col)
		}
		if i == r.End.Row {
			line = line.CutEnd(r.End.Col)
		}
		bs = append(bs, line.Bytes()...)
		if i < r.End.Row {
			bs = append(bs, '\n')
		}
	}

	return bs
}

func (b *lineBuffer) Rune(i int) rune {
	for _, line := range b.lines {
		if i < line.Len() {
			return line.Rune(i)
		}
		if i == line.Len() {
			return '\n'
		}
		i -= line.Len() + 1
	}
	return 0
}

func (b *lineBuffer) ByteIndex(i int) int {
	var n int
	for _, line := range b.lines {
		runeLen := line.Len() + 1
		if n+runeLen >= i {
			return n + line.ByteIndex(i-n)
		}
		n += runeLen
	}
	return n
}

func (b *lineBuffer) RuneIndex(i int) int {
	var n int
	for _, line := range b.lines {
		byteLen := len(line.Bytes())
		if n+byteLen >= i {
			return n + line.RuneIndex(i-n)
		}
		n += byteLen + 1
	}
	return n
}

func (b *lineBuffer) ByteIndexByPoint(p Point) int {
	var n int
	for i, line := range b.lines {
		if i == p.Row {
			return n + len(line.CutEnd(p.Col).Bytes())
		}
		n += len(line.Bytes()) + 1
	}
	return n
}

func (b *lineBuffer) Position(index int) Point {
	var n int
	for i, line := range b.lines {
		if n+len(line.Bytes()) >= index {
			return Point{Row: i, Col: index - n}
		}
		n += len(line.Bytes()) + 1
	}
	if n == 0 {
		return Point{
			Row: 0,
			Col: 0,
		}
	}
	return Point{Row: len(b.lines) - 1, Col: b.lines[len(b.lines)-1].Len()}
}

func (b *lineBuffer) Len() int {
	var n int
	for _, line := range b.lines {
		n += line.Len() + 1
	}
	return n
}

func (b *lineBuffer) BytesLen() int {
	var n int
	for _, line := range b.lines {
		n += line.BytesLen() + 1
	}
	return n
}

func (b *lineBuffer) LinesLen() int {
	return len(b.lines)
}

func (b *lineBuffer) Lines() []Line {
	return b.lines
}

func (b *lineBuffer) Line(row int) Line {
	return b.lines[row]
}

func (b *lineBuffer) LineLen(row int) int {
	if row < 0 || row >= len(b.lines) {
		return 0
	}
	return b.lines[row].Len()
}

func (b *lineBuffer) Insert(p Point, text []byte) {
	if len(text) == 0 {
		return
	}

	if len(text) == 1 && text[0] == '\n' {
		b.insertNewLine(p)
		return
	}

	for _, r := range xbytes.Runes(text) {
		if r == '\n' {
			line := b.lines[p.Row]
			b.lines[p.Row] = line.CutEnd(p.Col)
			b.lines = slices.Insert(b.lines, p.Row+1, line.CutStart(p.Col))
			p.Row++
			p.Col = 0
			continue
		}
		b.lines[p.Row] = b.lines[p.Row].Insert(p.Col, []byte(string(r)))
		p.Col++
	}
}

func (b *lineBuffer) Replace(r Range, text []byte) {
	b.Delete(r)
	if len(text) == 0 {
		return
	}
	b.Insert(r.Start, text)
}

func (b *lineBuffer) Delete(r Range) {
	if r.Start.Row == r.End.Row {
		b.lines[r.Start.Row] = b.lines[r.Start.Row].CutEnd(r.Start.Col).Append(b.lines[r.Start.Row].CutStart(r.End.Col))
	} else {
		b.lines[r.Start.Row] = b.lines[r.Start.Row].CutEnd(r.Start.Col).Append(b.lines[r.End.Row].CutStart(r.End.Col))
		b.lines = append(b.lines[:r.Start.Row+1], b.lines[r.End.Row+1:]...)
	}
}

func (b *lineBuffer) insertNewLine(p Point) {
	line := b.lines[p.Row]
	b.lines[p.Row] = line.CutEnd(p.Col)
	b.lines = slices.Insert(b.lines, p.Row+1, NewEmptyLine())
	b.lines[p.Row+1] = line.CutStart(p.Col)
}
