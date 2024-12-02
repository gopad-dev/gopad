package buffer

import (
	"github.com/tree-sitter/go-tree-sitter"
)

func NewByteRange(startByte uint, endByte uint) ByteRange {
	return ByteRange{
		StartByte: startByte,
		EndByte:   endByte,
	}
}

func ParseByteRange(r tree_sitter.Range) ByteRange {
	return ByteRange{
		StartByte: r.StartByte,
		EndByte:   r.EndByte,
	}
}

type ByteRange struct {
	StartByte uint
	EndByte   uint
}

func (r ByteRange) Contains(i uint) bool {
	return i >= r.StartByte && i <= r.EndByte
}

func (r ByteRange) ContainsRange(other ByteRange) bool {
	return r.StartByte <= other.StartByte && r.EndByte >= other.EndByte
}

func (r ByteRange) IsEmpty() bool {
	return r.StartByte == r.EndByte
}

func (r ByteRange) Bytes() uint {
	return r.EndByte - r.StartByte
}

func (r ByteRange) ToTreeSitter() tree_sitter.Range {
	return tree_sitter.Range{
		StartByte: r.StartByte,
		EndByte:   r.EndByte,
	}
}
