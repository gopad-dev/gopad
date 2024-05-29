package xio

import (
	"io"
)

// NopCloser returns a [io.WriteCloser] with a no-op Close method wrapping
// the provided [io.Writer] w.
func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
