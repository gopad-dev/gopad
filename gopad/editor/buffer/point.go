package buffer

import (
	"fmt"

	"github.com/tree-sitter/go-tree-sitter"
	"go.lsp.dev/protocol"
)

func ParsePoint(p protocol.Position) Point {
	return Point{
		Row: int(p.Line),
		Col: int(p.Character),
	}
}

func NewPoint(row int, col int) Point {
	return Point{
		Row: row,
		Col: col,
	}
}

type Point struct {
	Row int
	Col int
}

func (p Point) Point() (int, int) {
	return p.Row, p.Col
}

func (p Point) Add(p2 Point) Point {
	return Point{
		Row: p.Row + p2.Row,
		Col: p.Col + p2.Col,
	}
}

func (p Point) Sub(p2 Point) Point {
	return Point{
		Row: p.Row - p2.Row,
		Col: p.Col - p2.Col,
	}
}

func (p Point) Offset(row int, col int) Point {
	return Point{
		Row: p.Row + row,
		Col: p.Col + col,
	}
}

func (p Point) LessThan(other Point) bool {
	if p.Row == other.Row {
		return p.Col < other.Col
	}
	return p.Row < other.Row
}

func (p Point) LessThanOrEqual(other Point) bool {
	if p.Row == other.Row {
		return p.Col <= other.Col
	}
	return p.Row < other.Row
}

func (p Point) GreaterThan(other Point) bool {
	if p.Row == other.Row {
		return p.Col > other.Col
	}
	return p.Row > other.Row
}

func (p Point) GreaterThanOrEqual(other Point) bool {
	if p.Row == other.Row {
		return p.Col >= other.Col
	}
	return p.Row > other.Row
}

func (p Point) Equal(other Point) bool {
	return p.Row == other.Row && p.Col == other.Col
}

func (p Point) String() string {
	return fmt.Sprintf("[%d:%d]", p.Row+1, p.Col+1)
}

func (p Point) Compare(start Point) int {
	if p.Row < start.Row {
		return -1
	}
	if p.Row > start.Row {
		return 1
	}
	if p.Col < start.Col {
		return -1
	}
	if p.Col > start.Col {
		return 1
	}
	return 0
}

func (p Point) ToProtocol() protocol.Position {
	return protocol.Position{
		Line:      uint32(p.Row),
		Character: uint32(p.Col),
	}
}

func (p Point) ToTreeSitter() tree_sitter.Point {
	return tree_sitter.Point{
		Row:    uint(p.Row),
		Column: uint(p.Col),
	}
}
