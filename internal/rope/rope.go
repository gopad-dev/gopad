package rope

import (
	"bytes"
	"strings"
)

const (
	maxDepth    = 64
	maxLeafSize = 4096
)

func New() Rope {
	return Rope{}
}

func NewString(s string) Rope {
	return Rope{content: s, len: len(s)}
}

func NewBytes(b []byte) Rope {
	return NewString(string(b))
}

type Rope struct {
	content string
	len     int
	depth   int
	left    *Rope
	right   *Rope
}

func (r Rope) concat(r2 Rope) Rope {
	switch {
	case r.len == 0:
		return r2
	case r2.len == 0:
		return r
	case r.len+r2.len <= maxLeafSize:
		return NewString(r.String() + r2.String())
	default:
		depth := r.depth
		if r2.depth > depth {
			depth = r2.depth
		}
		return Rope{
			len:   r.len + r2.len,
			depth: depth + 1,
			left:  &r,
			right: &r2,
		}
	}
}

func (r Rope) Append(r2 Rope) Rope {
	return r.concat(r2).rebalanceIfNeeded()
}

func (r Rope) AppendString(s string) Rope {
	return r.Append(NewString(s))
}

func (r Rope) AppendBytes(b []byte) Rope {
	return r.Append(NewBytes(b))
}

func (r Rope) Delete(offset int, length int) Rope {
	if length == 0 || offset == r.len {
		return r
	}

	left, right := r.Split(offset)
	_, newRight := right.Split(length)
	return left.Append(newRight)
}

func (r Rope) Equal(r2 Rope) bool {
	if r == r2 {
		return true
	}

	if r.len != r2.len {
		return false
	}

	for i := 0; i < r.len; i += maxLeafSize {
		if !bytes.Equal(r.Slice(i, i+maxLeafSize), r2.Slice(i, i+maxLeafSize)) {
			return false
		}
	}

	return true
}

func (r Rope) Insert(at int, other Rope) Rope {
	switch at {
	case 0:
		return other.Append(r)
	case r.len:
		return r.Append(other)
	default:
		left, right := r.Split(at)
		return left.concat(other).Append(right)
	}
}

func (r Rope) InsertString(at int, s string) Rope {
	return r.Insert(at, NewString(s))
}

func (r Rope) InsertBytes(at int, b []byte) Rope {
	return r.Insert(at, NewBytes(b))
}

func (r Rope) Len() int {
	return r.len
}

func (r Rope) Rebalance() Rope {
	if r.isBalanced() {
		return r
	}

	var leaves []Rope
	r.walk(func(node Rope) {
		leaves = append(leaves, node)
	})

	return merge(leaves, 0, len(leaves))
}

func (r Rope) Slice(a, b int) []byte {
	p := make([]byte, b-a)
	n, _ := r.ReadAt(p, int64(a))
	return p[:n]
}

func (r Rope) Split(at int) (Rope, Rope) {
	switch {
	case r.isLeaf():
		return NewString(r.content[0:at]), NewString(r.content[at:])

	case at == 0:
		return Rope{}, r

	case at == r.len:
		return r, Rope{}

	case at < r.left.len:
		left, right := r.left.Split(at)
		return left, right.Append(*r.right)

	case at > r.left.len:
		left, right := r.right.Split(at - r.left.len)
		return r.left.Append(left), right

	default:
		return *r.left, *r.right
	}
}

func (r Rope) String() string {
	if r.isLeaf() {
		return r.content
	}

	var builder strings.Builder
	r.walk(func(node Rope) {
		builder.WriteString(node.content)
	})

	return builder.String()
}

func (r Rope) Bytes() []byte {
	if r.isLeaf() {
		return []byte(r.content)
	}

	var b []byte
	r.walk(func(node Rope) {
		b = append(b, []byte(node.content)...)
	})

	return b
}

func (r Rope) isBalanced() bool {
	switch {
	case r.isLeaf():
		return true
	case r.depth >= len(fibonacci)-2:
		return false
	default:
		return fibonacci[r.depth+2] <= r.len
	}
}

func (r Rope) isLeaf() bool {
	return r.left == nil
}

func (r Rope) leafForOffset(at int) (Rope, int) {
	switch {
	case r.isLeaf():
		return r, at
	case at < r.left.len:
		return r.left.leafForOffset(at)
	default:
		return r.right.leafForOffset(at - r.left.len)
	}
}

func (r Rope) rebalanceIfNeeded() Rope {
	if r.isBalanced() || abs(r.left.depth-r.right.depth) < maxDepth {
		return r
	}

	return r.Rebalance()
}

func (r Rope) walk(callback func(Rope)) {
	if r.isLeaf() {
		callback(r)
	} else {
		r.left.walk(callback)
		r.right.walk(callback)
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func merge(leaves []Rope, start int, end int) Rope {
	length := end - start
	switch length {
	case 1:
		return leaves[start]
	case 2:
		return leaves[start].concat(leaves[start+1])
	default:
		mid := start + length/2
		return merge(leaves, start, mid).concat(merge(leaves, mid, end))
	}
}

var fibonacci []int

func init() {
	// The heurstic for whether a rope is balanced depends on the Fibonacci sequence;
	// we initialize the table of Fibonacci numbers here.
	first := 0
	second := 1

	for c := 0; c < maxDepth+3; c++ {
		next := 0
		if c <= 1 {
			next = c
		} else {
			next = first + second
			first = second
			second = next
		}
		fibonacci = append(fibonacci, next)
	}
}
