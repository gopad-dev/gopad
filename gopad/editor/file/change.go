package file

import (
	"go.gopad.dev/gopad/gopad/editor/buffer"
)

type Insertion struct {
	At   int
	Text []byte
}

type Deletion struct {
	From int
	To   int
}

type Change struct {
	From int
	To   int
	Text []byte
}

func FromInsert(buf buffer.Buffer, i Insertion) Transaction {
	return Transaction{}
}

func FromDelete(buf buffer.Buffer, d Deletion) Transaction {
	return Transaction{}
}

func FromChange(buf buffer.Buffer, changes []Change) Transaction {
	return Transaction{}
}

type Transaction struct {
}
