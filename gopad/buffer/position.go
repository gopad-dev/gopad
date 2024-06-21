package buffer

import (
	"fmt"

	"go.lsp.dev/protocol"
)

func ParsePosition(p protocol.Position) Position {
	return Position{
		Row: int(p.Line),
		Col: int(p.Character),
	}
}

type Position struct {
	Row int
	Col int
}

func (p Position) LessThan(other Position) bool {
	if p.Row == other.Row {
		return p.Col < other.Col
	}
	return p.Row < other.Row
}

func (p Position) LessThanOrEqual(other Position) bool {
	if p.Row == other.Row {
		return p.Col <= other.Col
	}
	return p.Row < other.Row
}

func (p Position) GreaterThan(other Position) bool {
	if p.Row == other.Row {
		return p.Col > other.Col
	}
	return p.Row > other.Row
}

func (p Position) GreaterThanOrEqual(other Position) bool {
	if p.Row == other.Row {
		return p.Col >= other.Col
	}
	return p.Row > other.Row
}

func (p Position) Equal(other Position) bool {
	return p.Row == other.Row && p.Col == other.Col
}

func (p Position) String() string {
	return fmt.Sprintf("[%d:%d]", p.Row+1, p.Col+1)
}

func (p Position) ToProtocol() protocol.Position {
	return protocol.Position{
		Line:      uint32(p.Row),
		Character: uint32(p.Col),
	}
}

func (p Position) Compare(start Position) int {
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
		Start: ParsePosition(r.Start),
		End:   ParsePosition(r.End),
	}
}

type Range struct {
	Start Position
	End   Position
}

func (r Range) Contains(p Position) bool {
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
