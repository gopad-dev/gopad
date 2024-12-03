package doc

import (
	"iter"
	"slices"

	"go.gopad.dev/gopad/gopad/editor/buffer"
	"go.gopad.dev/gopad/internal/xbytes"
)

type Deletion struct {
	From int
	To   int
}

type Change struct {
	From int
	To   int
	Text []byte
}

type Operation interface {
	operation()
}

type Move struct {
	N int
}

func (Move) operation() {}

type Delete struct {
	N int
}

func (Delete) operation() {}

type Insert struct {
	Text []byte
}

func (Insert) operation() {}

func NewChangeSet() ChangeSet {
	return ChangeSet{
		Changes:  nil,
		Len:      0,
		LenAfter: 0,
	}
}

func NewChangeSetFromBuf(buf buffer.Buffer) ChangeSet {
	bufLen := buf.Len()
	return ChangeSet{
		Changes:  nil,
		Len:      bufLen,
		LenAfter: bufLen,
	}
}

type ChangeSet struct {
	Changes  []Operation
	Len      int
	LenAfter int
}

func (c *ChangeSet) Delete(n int) {
	if n == 0 {
		return
	}

	c.Len += n

	if len(c.Changes) > 0 {
		d, ok := c.Changes[len(c.Changes)-1].(Delete)
		if ok {
			c.Changes[len(c.Changes)-1] = Delete{
				N: d.N + n,
			}
			return
		}
	}

	c.Changes = append(c.Changes, Delete{
		N: n,
	})
}

func (c *ChangeSet) Insert(text []byte) {
	if len(text) == 0 {
		return
	}

	c.LenAfter += xbytes.RuneCount(text)

	if len(c.Changes) > 0 {
		i, ok := c.Changes[len(c.Changes)-1].(Insert)
		if ok {
			c.Changes[len(c.Changes)-1] = Insert{
				Text: append(i.Text, text...),
			}
			return
		}
	}

	c.Changes = append(c.Changes, Insert{
		Text: text,
	})
}

func (c *ChangeSet) Move(n int) {
	if n == 0 {
		return
	}

	c.Len += n
	c.LenAfter += n

	if len(c.Changes) > 0 {
		m, ok := c.Changes[len(c.Changes)-1].(Move)
		if ok {
			c.Changes[len(c.Changes)-1] = Move{
				N: m.N + n,
			}
			return
		}
	}

	c.Changes = append(c.Changes, Move{
		N: n,
	})
}

func (c ChangeSet) Invert(originalBuf buffer.Buffer) ChangeSet {
	if originalBuf.Len() != c.Len {
		panic("invalid change set")
	}

	changes := ChangeSet{
		Changes:  make([]Operation, 0, len(c.Changes)),
		Len:      0,
		LenAfter: 0,
	}

	var pos uint
	for _, change := range c.Changes {
		switch change := change.(type) {
		case Move:
			changes.Move(change.N)
			pos = uint(int(pos) + change.N)
		case Delete:
			text := originalBuf.BytesRange(buffer.ByteRange{StartByte: pos, EndByte: uint(int(pos) + change.N)})
			changes.Insert(text)
			pos = uint(int(pos) + change.N)
		case Insert:
			changes.Delete(xbytes.RuneCount(change.Text))
		}
	}

	return changes
}

func (c ChangeSet) Apply(buf buffer.Buffer) bool {
	if buf.Len() != c.Len {
		return false
	}

	var pos uint
	for _, change := range c.Changes {
		switch change := change.(type) {
		case Move:
			pos = uint(int(pos) + change.N)
		case Delete:
			buf.Delete(buffer.ByteRange{StartByte: pos, EndByte: uint(int(pos) + change.N)})
		case Insert:
			buf.Insert(pos, change.Text)
			pos += uint(xbytes.RuneCount(change.Text))
		}
	}

	return true
}

func (c ChangeSet) IsEmpty() bool {
	return len(c.Changes) == 0
}

func (c ChangeSet) Iter() iter.Seq[Change] {
	i := newChangeIterator(c)

	return func(yield func(Change) bool) {
		for {
			change, ok := i.next()
			if !ok {
				return
			}
			if !yield(change) {
				return
			}
		}
	}
}

// Merge merges the changes from the given change set into the current change set.
func (c ChangeSet) Merge(other ChangeSet) ChangeSet {
	if c.Len != other.Len {
		panic("invalid change set")
	}

	if c.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return c
	}

	changesA := c.Changes
	nextA := func() Operation {
		if len(changesA) == 0 {
			return nil
		}
		next := changesA[0]
		changesA = changesA[1:]
		return next
	}
	headA := nextA()

	changesB := other.Changes
	nextB := func() Operation {
		if len(changesB) == 0 {
			return nil
		}
		next := changesB[0]
		changesB = changesB[1:]
		return next
	}
	headB := nextB()

	changes := NewChangeSet()
	for {
		if headA == nil && headB == nil {
			return changes
		}

		deleteA, deleteAOk := headA.(Delete)
		deleteB, deleteBOk := headB.(Delete)
		insertA, insertAOk := headA.(Insert)
		insertB, insertBOk := headB.(Insert)
		moveA, moveAOk := headA.(Move)
		moveB, moveBOk := headB.(Move)

		switch {
		case deleteAOk:
			changes.Delete(deleteA.N)
			headA = nextA()
		case insertBOk:
			changes.Insert(insertB.Text)
			headB = nextB()
		case moveAOk && moveBOk:
			if moveA.N < moveB.N {
				changes.Move(moveA.N)
				headB = Move{N: moveB.N - moveA.N}
			} else if moveA.N == moveB.N {
				changes.Move(moveA.N)
				headA = nextA()
				headB = nextB()
			} else {
				changes.Move(moveB.N)
				headA = Move{N: moveA.N - moveB.N}
				headB = nextB()
			}
		case insertAOk && deleteBOk:
			charLen := xbytes.RuneCount(insertA.Text)
			if charLen < deleteB.N {
				headA = nextA()
				headB = Delete{N: deleteB.N - charLen}
			} else if charLen == deleteB.N {
				headA = nextA()
				headB = nextB()
			} else {
				headA = Insert{Text: insertA.Text[deleteB.N:]} // TODO: check if this is correct
				headB = nextB()
			}
		case insertAOk && moveBOk:
			charLen := xbytes.RuneCount(insertA.Text)
			if charLen < moveB.N {
				changes.Insert(insertA.Text)
				headA = nextA()
				headB = Move{N: moveB.N - charLen}
			} else if charLen == moveB.N {
				changes.Insert(insertA.Text)
				headA = nextA()
				headB = nextB()
			} else {
				changes.Insert(insertA.Text[:moveB.N])
				headA = Insert{Text: insertA.Text[moveB.N:]} // TODO: check if this is correct
				headB = nextB()
			}
		case moveAOk && deleteBOk:
			if moveA.N < deleteB.N {
				changes.Delete(moveA.N)
				headA = nextA()
				headB = Delete{N: deleteB.N - moveA.N}
			} else if moveA.N == deleteB.N {
				changes.Delete(deleteB.N)
				headA = nextA()
				headB = nextB()
			} else {
				changes.Delete(deleteB.N)
				headA = Move{N: moveA.N - deleteB.N}
				headB = nextB()
			}
		}
	}
}

func (c ChangeSet) UpdatePosition(selections []Selection) []Selection {
	slices.SortFunc(selections, func(a, b Selection) int {
		return int(a.Anchor) - int(b.Anchor)
	})

	for i, selection := range selections {
		anchor := selection.Anchor
		head := selection.Head
		for _, change := range c.Changes {
			switch change := change.(type) {
			case Move:
				anchor = uint(int(anchor) + change.N)
				head = uint(int(head) + change.N)
			case Delete:
				anchor = uint(int(anchor) + change.N)
				head += uint(int(head) + change.N)
			case Insert:
				anchor += uint(xbytes.RuneCount(change.Text))
				head += uint(xbytes.RuneCount(change.Text))
			}
		}
		selections[i] = Selection{
			Anchor: anchor,
			Head:   head,
		}
	}

	return selections
}

func newChangeIterator(changeSet ChangeSet) *changeIterator {
	return &changeIterator{
		changes: changeSet.Changes,
		pos:     0,
	}
}

type changeIterator struct {
	changes []Operation
	pos     int
}

func (c *changeIterator) next() (Change, bool) {
	for {
		if len(c.changes) == 0 {
			return Change{}, false
		}

		var operation Operation
		operation, c.changes = c.changes[0], c.changes[1:]

		switch operation := operation.(type) {
		case Move:
			c.pos += operation.N
		case Delete:
			start := c.pos
			c.pos += operation.N
			return Change{
				From: start,
				To:   c.pos,
				Text: nil,
			}, true
		case Insert:
			start := c.pos

			// Check if the next operation is a delete operation, if so, we can merge the two operations into a replace operation.
			if len(c.changes) > 0 {
				next, ok := c.changes[0].(Delete)
				if ok {
					c.changes = c.changes[1:]
					c.pos += next.N
					return Change{
						From: start,
						To:   c.pos,
						Text: operation.Text,
					}, true
				}
			}
			return Change{
				From: start,
				To:   start,
				Text: operation.Text,
			}, true
		}
	}
}

func NewTransactionFromChange(buf buffer.Buffer, changes []Change) Transaction {
	docLen := buf.Len()

	changeSet := NewChangeSet()

	var last int
	for _, change := range changes {
		changeSet.Move(change.From - last)
		span := change.To - change.From
		if len(change.Text) > 0 {
			changeSet.Insert(change.Text)
			changeSet.Delete(span)
		} else {
			changeSet.Delete(span)
		}
		last = change.To
	}

	changeSet.Move(docLen - last)

	return Transaction{
		Changes: changeSet,
	}
}

func NewTransactionFromDelete(buf buffer.Buffer, deletions []Deletion) Transaction {
	docLen := buf.Len()

	changeSet := NewChangeSet()

	var last int
	for _, deletion := range deletions {
		if last > deletion.To {
			continue
		}
		if last > deletion.From {
			deletion.From = last
		}

		changeSet.Move(deletion.From - last)
		changeSet.Delete(deletion.To - deletion.From)
		last = deletion.To
	}

	changeSet.Move(docLen - last)

	return Transaction{
		Changes: changeSet,
	}
}

func NewTransactionFromInsert(buf buffer.Buffer, index int, text []byte) Transaction {
	return NewTransactionFromChange(buf, []Change{{
		From: index,
		To:   index,
		Text: text,
	}})
}

func NewTransaction(buf buffer.Buffer) Transaction {
	return NewTransactionFrom(NewChangeSetFromBuf(buf))
}

func NewTransactionFrom(changes ChangeSet) Transaction {
	return Transaction{
		Changes: changes,
	}
}

type Transaction struct {
	Changes ChangeSet
	Cursor  *Cursor
}

func (t *Transaction) Apply(buf buffer.Buffer) bool {
	if t.Changes.IsEmpty() {
		return true
	}

	return t.Changes.Apply(buf)
}

func (t *Transaction) Invert(originalBuf buffer.Buffer) Transaction {
	return Transaction{
		Changes: t.Changes.Invert(originalBuf),
		Cursor:  nil,
	}
}

func (t *Transaction) Insert(text []byte) {
	t.Changes.Insert(text)
}

func (t *Transaction) WithCursor(cursor Cursor) {
	t.Cursor = &cursor
}

func (t *Transaction) Iter() iter.Seq[Change] {
	return t.Changes.Iter()
}

func (t *Transaction) IsEmpty() bool {
	return t.Changes.IsEmpty()
}
