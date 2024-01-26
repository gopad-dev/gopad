package buffer

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

type Range struct {
	Start Position
	End   Position
}

func (r Range) Contains(p Position) bool {
	return p.GreaterThanOrEqual(r.Start) && p.LessThan(r.End)
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
