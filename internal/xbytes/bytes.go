package xbytes

import (
	"slices"
	"unicode/utf8"
)

// ByteIndex converts the rune index to a byte index.
func ByteIndex(s []byte, i int) int {
	if i == 0 {
		return 0
	}

	var (
		runeIndex int
		byteIndex int
	)
	for len(s) > 0 {
		_, l := utf8.DecodeRune(s)

		if runeIndex == i {
			return byteIndex
		}

		s = s[l:]
		runeIndex++
		byteIndex += l
	}

	return -1
}

// ByteIndexEnd converts the rune index to a byte index at the end of the rune.
func ByteIndexEnd(s []byte, i int) int {
	var (
		runeIndex int
		byteIndex int
	)
	for len(s) > 0 {
		_, l := utf8.DecodeRune(s)

		if runeIndex == i {
			return byteIndex + l
		}

		s = s[l:]
		runeIndex++
		byteIndex += l
	}

	return -1
}

// RuneIndex converts the byte index to a rune index.
func RuneIndex(s []byte, i int) int {
	if i == 0 {
		return 0
	}

	var (
		runeIndex int
		byteIndex int
	)
	for len(s) > 0 {
		_, l := utf8.DecodeRune(s)

		if byteIndex == i {
			return runeIndex
		}

		s = s[l:]
		runeIndex++
		byteIndex += l
	}
	return -1
}

func RuneLen(s []byte, i int) int {
	var runeIndex int
	for len(s) > 0 {
		_, l := utf8.DecodeRune(s)

		if runeIndex == i {
			return l
		}

		s = s[l:]
		runeIndex++
	}

	return -1
}

func RuneCount(s []byte) int {
	return utf8.RuneCount(s)
}

func Runes(s []byte) []rune {
	runes := make([]rune, 0)
	for len(s) > 0 {
		r, l := utf8.DecodeRune(s)
		runes = append(runes, r)
		s = s[l:]
	}
	return runes
}

func RunesRange(s []byte, start int, end int) []rune {
	runes := make([]rune, end-start)

	var (
		runeIndex int
		i         int
	)
	for len(s) > 0 {
		r, l := utf8.DecodeRune(s)
		s = s[l:]

		if runeIndex >= start && runeIndex < end {
			runes[i] = r
			i++
		}

		runeIndex++
	}
	return runes
}

func Rune(s []byte, i int) rune {
	runeIndex := 0
	for len(s) > 0 {
		r, l := utf8.DecodeRune(s)
		s = s[l:]

		if runeIndex == i {
			return r
		}

		runeIndex++
	}

	return utf8.RuneError
}

func CutStart(s []byte, index int) []byte {
	slen := len(s)
	i := 0
	totalLen := 0
	for totalLen <= slen {
		_, l := utf8.DecodeRune(s[totalLen:])

		if i >= index {
			return s[totalLen:]
		}

		totalLen += l
		i++
	}

	return s
}

func CutEnd(s []byte, index int) []byte {
	slen := len(s)
	i := 0
	totalLen := 0
	for totalLen <= slen {
		_, l := utf8.DecodeRune(s[totalLen:])

		if i >= index {
			return s[:totalLen]
		}

		totalLen += l
		i++
	}

	return s
}

func CutRange(s []byte, start int, end int) []byte {
	slen := len(s)
	i := 0
	startLen := 0
	endLen := 0
	for endLen <= slen {
		_, l := utf8.DecodeRune(s[endLen:])

		if i >= end {
			return s[startLen:endLen]
		}

		if i < start {
			startLen += l
		}
		endLen += l
		i++
	}

	return s
}

// Append appends the given bytes to the given bytes.
func Append(s []byte, b ...byte) []byte {
	return append(s, b...)
}

// Prepend prepends the given bytes to the given bytes.
func Prepend(s []byte, b ...byte) []byte {
	return Append(b, s...)
}

// Insert inserts the given bytes at the given index.
func Insert(s []byte, i int, b ...byte) []byte {
	if i == 0 {
		return Append(b, s...)
	}

	index := ByteIndex(s, i)
	if index == -1 {
		return append(s, b...)
	}

	return slices.Insert(s, index, b...)
}

// Replace replaces the range of bytes from start (inclusive) to end (inclusive) with the given bytes.
func Replace(s []byte, start int, end int, b ...byte) []byte {
	startIndex := ByteIndex(s, start)
	endIndex := ByteIndexEnd(s, end)
	if startIndex == -1 {
		panic("start index out of range")
	}
	if endIndex == -1 {
		panic("end index out of range")
	}

	return slices.Replace(s, startIndex, endIndex, b...)
}

// Delete deletes the range of bytes from start (inclusive) to end (inclusive).
func Delete(s []byte, start int, end int) []byte {
	startIndex := ByteIndex(s, start)
	endIndex := ByteIndexEnd(s, end)
	if startIndex == -1 {
		panic("start index out of range")
	}
	if endIndex == -1 {
		panic("end index out of range")
	}

	return slices.Delete(s, startIndex, endIndex)
}
