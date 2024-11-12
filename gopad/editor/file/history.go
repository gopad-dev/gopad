package file

import (
	"time"

	"go.gopad.dev/gopad/gopad/editor/buffer"
)

type Revision struct {
	Parent      int
	LastChild   *int
	Transaction Transaction
	Inversion   Transaction
	Time        time.Time
}

func NewHistory() *History {
	return &History{
		revisions: []Revision{
			{
				Parent:      0,
				LastChild:   nil,
				Transaction: NewTransactionFrom(NewChangeSet()),
				Inversion:   NewTransactionFrom(NewChangeSet()),
				Time:        time.Now(),
			},
		},
		current: 0,
	}
}

type History struct {
	revisions []Revision
	current   int
}

func (h *History) CurrentRevision() int {
	return h.current
}

func (h *History) AtRoot() bool {
	return h.current == 0
}

func (h *History) CommitRevision(transaction Transaction, originalBuf buffer.Buffer) {
	inversion := transaction.Invert(originalBuf)

	newCurrent := len(h.revisions)
	h.revisions[h.current].LastChild = &newCurrent
	h.revisions = append(h.revisions, Revision{
		Parent:      h.current,
		LastChild:   nil,
		Transaction: transaction,
		Inversion:   inversion,
		Time:        time.Now(),
	})
	h.current = newCurrent
}

func (h *History) Undo() *Transaction {
	if h.AtRoot() {
		return nil
	}

	currentRevision := h.revisions[h.current]
	h.current = currentRevision.Parent
	return &currentRevision.Inversion
}

func (h *History) Redo() *Transaction {
	currentRevision := h.revisions[h.current]
	lastChild := currentRevision.LastChild
	if lastChild == nil {
		return nil
	}

	h.current = *lastChild

	return &h.revisions[*lastChild].Transaction
}
