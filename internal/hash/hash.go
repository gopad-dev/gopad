package hash

import (
	"hash"
	"hash/fnv"
)

type Hash interface {
	Hash(h *Hasher)
}

func New() *Hasher {
	return &Hasher{
		h: fnv.New64a(),
	}
}

type Hasher struct {
	h hash.Hash64
}

func (h *Hasher) Write(b []byte) (int, error) {
	return h.h.Write(b)
}

func (h *Hasher) Hash(a Hash) {
	a.Hash(h)
}

func (h *Hasher) Finish() uint64 {
	return h.h.Sum64()
}

func (h *Hasher) HashOne(a Hash) uint64 {
	h.Hash(a)
	defer h.h.Reset()
	return h.Finish()
}
