package buffer

import (
	"fmt"

	"go.lsp.dev/protocol"
)

func ParsePoint(p protocol.Position) Point {
	return Point{
		Row: int(p.Line),
		Col: int(p.Character),
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

func (p Point) Offset(row, col int) Point {
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

func (p Point) ToProtocol() protocol.Position {
	return protocol.Position{
		Line:      uint32(p.Row),
		Character: uint32(p.Col),
	}
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

func ParseRange(r protocol.Range) Range {
	return Range{
		Start: ParsePoint(r.Start),
		End:   ParsePoint(r.End),
	}
}

type Range struct {
	Start Point
	End   Point
}

func (r Range) Contains(p Point) bool {
	return p.GreaterThanOrEqual(r.Start) && p.LessThanOrEqual(r.End)
}

func (r Range) ContainsRow(row int) bool {
	return row >= r.Start.Row && row <= r.End.Row
}

func (r Range) ContainsRange(other Range) bool {
	return r.Start.LessThanOrEqual(other.Start) && r.End.GreaterThanOrEqual(other.End)
}

func (r Range) Overlaps(other Range) bool {
	return r.Contains(other.Start) || r.Contains(other.End)
}

func (r Range) Equal(other Range) bool {
	return r.Start.Equal(other.Start) && r.End.Equal(other.End)
}

func (r Range) String() string {
	return r.Start.String() + "-" + r.End.String()
}

func (r Range) IsEmpty() bool {
	return r.Start.Equal(r.End)
}

func (r Range) ToProtocol() protocol.Range {
	return protocol.Range{
		Start: r.Start.ToProtocol(),
		End:   r.End.ToProtocol(),
	}
}

func (r Range) Lines() int {
	return r.End.Row - r.Start.Row + 1
}

func (r Range) Compare(start Range) int {
	if c := r.Start.Compare(start.Start); c != 0 {
		return c
	}
	return r.End.Compare(start.End)
}

func (r Range) Zero() bool {
	return r.Start.Equal(r.End)
}
