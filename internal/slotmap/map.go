package slotmap

import (
	"iter"
	"math/rand"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().Unix()))

type LayerID uint64

func New[V any]() *Map[V] {
	return &Map[V]{m: make(map[LayerID]V)}
}

type Map[V any] struct {
	m map[LayerID]V
}

func (m *Map[V]) Iter() iter.Seq2[LayerID, V] {
	return func(yield func(LayerID, V) bool) {
		for k, v := range m.m {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (m *Map[V]) Len() int {
	return len(m.m)
}

func (m *Map[V]) Get(id LayerID) V {
	v, ok := m.m[id]
	if !ok {
		var zero V
		return zero
	}
	return v
}

func (m *Map[V]) Insert(v V) LayerID {
	id := LayerID(r.Uint64())
	m.m[id] = v
	return id
}

func (m *Map[V]) Remove(id LayerID) {
	delete(m.m, id)
}

func (m *Map[V]) Retain(f func(LayerID, V) bool) {
	for k, v := range m.m {
		if !f(k, v) {
			delete(m.m, k)
		}
	}
}

func (m *Map[V]) Clear() {
	clear(m.m)
}
