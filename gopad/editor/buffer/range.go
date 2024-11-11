package buffer

import (
	"go.lsp.dev/protocol"
)

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
