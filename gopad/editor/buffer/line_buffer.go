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
	unicode2 "unicode"

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
func NewFromFile(fileName string, lineEnding LineEnding) (Buffer, error) {
	var err error
	fileName, err = filepath.Abs(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute file path: %w", err)
	}
	file, err := os.Open(fileName)
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
	version    uint64
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

func (b *lineBuffer) Copy() Buffer {
	lines := make([]Line, len(b.lines))
	for i, line := range b.lines {
		lines[i] = line.Copy()
	}
	checksum := make([]byte, len(b.checksum))
	copy(checksum, b.checksum)

	return &lineBuffer{
		lineEnding: b.lineEnding,
		version:    b.version,
		lines:      lines,
		checksum:   checksum,
	}
}

func (b *lineBuffer) LineEnding() LineEnding {
	return b.lineEnding
}

func (b *lineBuffer) SetLineEnding(lineEnding LineEnding) {
	b.lineEnding = lineEnding
}

func (b *lineBuffer) Version() uint64 {
	return b.version
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

func (b *lineBuffer) ByteIndex(p Point) int {
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
	return Point{Row: len(b.lines) - 1, Col: b.lines[len(b.lines)-1].Len()}
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

func (b *lineBuffer) String() string {
	return string(b.Bytes())
}

func (b *lineBuffer) LinesLen() int {
	return len(b.lines)
}

func (b *lineBuffer) Lines() []Line {
	return b.lines
}

func (b *lineBuffer) Len() int {
	var n int
	for _, line := range b.lines {
		n += line.Len() + 1
	}
	return n
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

func (b *lineBuffer) insertNewLine(p Point) {

	line := b.lines[p.Row]
	b.lines[p.Row] = line.CutEnd(p.Col)
	b.lines = slices.Insert(b.lines, p.Row+1, NewEmptyLine())
	b.lines[p.Row+1] = line.CutStart(p.Col)
}

func (b *lineBuffer) Insert(p Point, text []byte) {
	if len(text) == 0 {
		return
	}

	defer func() {
		b.version++
	}()

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
	defer func() {
		b.version++

	}()

	b.DeleteRange(r)
	if len(text) == 0 {
		return
	}
	b.Insert(r.Start, text)
}

func (b *lineBuffer) DuplicateLine(row int) {
	defer func() {
		b.version++

	}()

	line := b.lines[row]
	b.lines = slices.Insert(b.lines, row+1, line.Copy())
}

func (b *lineBuffer) DeleteLine(row int) {
	defer func() {
		b.version++

	}()

	if row == 0 && len(b.lines) == 1 {
		b.lines[0] = NewEmptyLine()
		return
	}
	if row == len(b.lines)-1 {
		b.lines = b.lines[:row]
		return
	}
	b.lines = append(b.lines[:row], b.lines[row+1:]...)
}

func (b *lineBuffer) DeleteBefore(p Point) {
	if p.Row == 0 && p.Col == 0 {
		return
	}
	defer func() {
		b.version++

	}()

	if p.Col == 0 {
		line := b.lines[p.Row]
		p.Col = b.lines[p.Row-1].Len()
		if line.Len() > 0 {
			b.lines[p.Row-1] = b.lines[p.Row-1].Append(line)
		}
		b.lines = append(b.lines[:p.Row], b.lines[p.Row+1:]...)
		p.Row = max(p.Row-1, 0)
	} else if p.Col > 0 {
		b.lines[p.Row] = b.lines[p.Row].CutEnd(p.Col - 1).Append(b.lines[p.Row].CutStart(p.Col))
		p.Col = max(p.Col-1, 0)
	} else if p.Row > 0 {
		p.Row = max(p.Col-1, 0)
		p.Col = b.lines[p.Row].Len()
	}
}

func (b *lineBuffer) DeleteAfter(p Point) {
	if p.Row == len(b.lines)-1 && p.Col == b.lines[p.Row].Len() {
		return
	}
	defer func() {
		b.version++

	}()

	if p.Col == b.LineLen(p.Row) {
		b.lines[p.Row] = b.lines[p.Row].Append(b.lines[p.Row+1])
		b.lines = append(b.lines[:p.Row+1], b.lines[p.Row+2:]...)
	} else if p.Col < b.lines[p.Row].Len() {
		b.lines[p.Row] = b.lines[p.Row].CutEnd(p.Col).Append(b.lines[p.Row].CutStart(p.Col + 1))
	}
}

func (b *lineBuffer) DeleteRange(r Range) {
	defer func() {
		b.version++

	}()

	if r.Start.Row == r.End.Row {
		b.lines[r.Start.Row] = b.lines[r.Start.Row].CutEnd(r.Start.Col).Append(b.lines[r.Start.Row].CutStart(r.End.Col))
	} else {
		b.lines[r.Start.Row] = b.lines[r.Start.Row].CutEnd(r.Start.Col).Append(b.lines[r.End.Row].CutStart(r.End.Col))
		b.lines = append(b.lines[:r.Start.Row+1], b.lines[r.End.Row+1:]...)
	}
}

func (b *lineBuffer) AddTab(row int) {
	defer func() {
		b.version++

	}()

	b.lines[row] = b.lines[row].Insert(0, []byte("\t"))
}

func (b *lineBuffer) RemoveTab(row int) {
	defer func() {
		b.version++
	}()

	if b.lines[row].Rune(0) == '\t' {
		b.lines[row] = b.lines[row].CutStart(1)
	}
}

func (b *lineBuffer) ToggleBlockComment(r Range, tokens []BlockCommentToken) {
	defer func() {
		b.version++

	}()

	if r.Start.Row == r.End.Row {
		line := b.lines[r.Start.Row]

		if line.Len() == 0 {
			return
		}

		var hasStartComment bool
		var hasEndComment bool
		var blockToken *BlockCommentToken
		for _, token := range tokens {
			if string(line.RunesRange(r.Start.Col, r.Start.Col+len(token.Start))) == token.Start {
				hasStartComment = true
			}
			if string(line.RunesRange(r.End.Col-len(token.End), r.End.Col)) == token.End {
				hasEndComment = true
			}
			if hasStartComment && hasEndComment {
				blockToken = &token
				break
			}
		}

		if blockToken == nil {
			blockToken = &tokens[0]
		}

		if hasStartComment && hasEndComment {
			b.lines[r.Start.Row] = line.ReplaceRange(r.Start.Col, r.Start.Col+len(blockToken.Start), nil).ReplaceRange(r.End.Col-len(blockToken.Start), r.End.Col-len(blockToken.Start)+len(blockToken.End), nil)
		} else {
			b.lines[r.Start.Row] = line.Insert(r.Start.Col, []byte(blockToken.Start)).Insert(r.End.Col+len(blockToken.Start), []byte(blockToken.End))
		}

		return
	}

	startLine := b.lines[r.Start.Row]
	endLine := b.lines[r.End.Row]

	var hasStartComment bool
	var hasEndComment bool
	var blockToken *BlockCommentToken
	for _, token := range tokens {
		if startLine.Len() > 0 {
			if string(startLine.RunesRange(r.Start.Col, r.Start.Col+len(token.Start))) == token.Start {
				hasStartComment = true
			}
		}

		if endLine.Len() > 0 {
			if string(endLine.RunesRange(max(0, r.End.Col-len(token.End)), r.End.Col)) == token.End {
				hasEndComment = true
			}
		}

		if hasStartComment && hasEndComment {
			blockToken = &token
			break
		}

		hasStartComment = false
		hasEndComment = false
	}

	if blockToken == nil {
		blockToken = &tokens[0]
	}

	if hasStartComment && hasEndComment {
		b.lines[r.Start.Row] = startLine.ReplaceRange(r.Start.Col, r.Start.Col+len(blockToken.Start), nil)
		b.lines[r.End.Row] = endLine.ReplaceRange(r.End.Col-len(blockToken.End), r.End.Col, nil)
	} else {
		b.lines[r.Start.Row] = startLine.Insert(r.Start.Col, []byte(blockToken.Start))
		b.lines[r.End.Row] = endLine.Insert(r.End.Col, []byte(blockToken.End))
	}
}

func (b *lineBuffer) ToggleLineComment(row int, tokens []string) {
	defer func() {
		b.version++
	}()

	line := b.lines[row]

	if line.Len() == 0 {
		return
	}

	for i, r := range line.Runes() {
		if !unicode2.IsSpace(r) {
			lineData := line.CutStart(i).Bytes()

			if ok, token := hasPrefixes(lineData, tokens); ok {
				b.lines[row] = line.CutEnd(i).Append(line.CutStart(i + len(token)))
				return
			}

			token := tokens[0]
			b.lines[row] = line.Insert(i, []byte(token))
			return
		}
	}
}

func hasPrefixes(b []byte, prefixes []string) (bool, string) {
	for _, prefix := range prefixes {
		if bytes.HasPrefix(b, []byte(prefix)) {
			return true, prefix
		}
	}
	return false, ""
}
