package xslices

// Pop removes the last element from the slice and returns it along with the new slice.
func Pop[T any](s []T) (T, []T) {
	if len(s) == 0 {
		var zero T
		return zero, s
	}

	return s[len(s)-1], s[:len(s)-1]
}

// Last returns the last element of the slice.
func Last[T any](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}

	return s[len(s)-1]
}
