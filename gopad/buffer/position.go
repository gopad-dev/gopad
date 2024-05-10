package buffer

import (
	"fmt"
)

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
