package xstrings

import (
	"strings"
)

func CutNewlines(s string) string {
	return strings.SplitN(s, "\n", 2)[0]
}

func Clamp(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}
