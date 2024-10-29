package buffer

import (
	"fmt"
	"unicode/utf8"

	"go.gopad.dev/gopad/internal/xbytes"
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

func NewLine(data []byte) Line {
	return byteLine{
		data: data,
	}
}

func NewEmptyLine() Line {
	return byteLine{
		data: nil,
	}
}

type byteLine struct {
	data []byte
}

func (l byteLine) Len() int {
	return utf8.RuneCount(l.data)
}

func (l byteLine) LenBytes() int {
	return len(l.data)
}

func (l byteLine) Bytes() []byte {
	return l.data
}

func (l byteLine) String() string {
	return string(l.data)
}

func (l byteLine) Runes() []rune {
	return xbytes.Runes(l.data)
}

func (l byteLine) RunesRange(start int, end int) []rune {
	return xbytes.RunesRange(l.data, start, end)
}

func (l byteLine) Rune(index int) rune {
	return xbytes.Rune(l.data, index)
}

func (l byteLine) RuneBytes(index int) []byte {
	return []byte(string(xbytes.Rune(l.data, index)))
}

func (l byteLine) RunesBytesRange(start int, end int) []byte {
	return []byte(string(xbytes.RunesRange(l.data, start, end)))
}

func (l byteLine) RuneString(index int) string {
	return string(xbytes.Rune(l.data, index))
}

func (l byteLine) StringRange(start int, end int) string {
	return string(xbytes.CutRange(l.data, start, end))
}

func (l byteLine) RuneStrings() []string {
	runes := xbytes.Runes(l.data)
	strs := make([]string, len(runes))
	for i, r := range runes {
		strs[i] = string(r)
	}
	return strs
}

func (l byteLine) RuneIndex(index int) int {
	return xbytes.RuneIndex(l.data, index)
}

func (l byteLine) CutStart(index int) Line {
	l.data = xbytes.CutStart(l.data, index)
	return l
}

func (l byteLine) CutEnd(index int) Line {
	l.data = xbytes.CutEnd(l.data, index)
	return l
}

func (l byteLine) CutRange(start int, end int) Line {
	l.data = xbytes.CutRange(l.data, start, end)
	return l
}

func (l byteLine) Append(line Line) Line {
	l.data = xbytes.Append(l.data, line.Bytes()...)
	return l
}

func (l byteLine) AppendLines(lines ...Line) Line {
	for _, line := range lines {
		l.data = xbytes.Append(l.data, line.Bytes()...)
	}
	return l
}

func (l byteLine) Insert(index int, b []byte) Line {
	l.data = xbytes.Insert(l.data, index, b...)
	return l
}

func (l byteLine) Replace(index int, b []byte) Line {
	l.data = xbytes.Replace(l.data, index, b...)
	return l
}

func (l byteLine) ReplaceRange(start int, end int, b []byte) Line {
	l.data = xbytes.ReplaceRange(l.data, start, end, b...)
	return l
}

func (l byteLine) Copy() Line {
	data := make([]byte, len(l.data))
	copy(data, l.data)
	return byteLine{
		data: data,
	}
}
