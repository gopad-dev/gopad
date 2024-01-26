package xbytes

import (
	"slices"
	"unicode/utf8"
)

func RuneIndex(s []byte, i int) int {
	if i == 0 {
		return 0
	}

	byteIndex := 0
	runeIndex := 0
	for len(s) > 0 {
		_, l := utf8.DecodeRune(s)
		s = s[l:]

		if runeIndex == i {
			return byteIndex
		}

		byteIndex += l
		runeIndex++
	}
	return -1
}

func RuneLen(s []byte, i int) int {
	runeIndex := 0
	for len(s) > 0 {
		_, l := utf8.DecodeRune(s)
		s = s[l:]

		if runeIndex == i {
			return l
		}

		runeIndex++
	}
	return -1
}

func Runes(s []byte) []rune {
	t := make([]rune, utf8.RuneCount(s))
	i := 0
	for len(s) > 0 {
		r, l := utf8.DecodeRune(s)
		t[i] = r
		i++
		s = s[l:]
	}
	return t
}

func RunesRange(s []byte, start int, end int) []rune {
	t := make([]rune, end-start)
	i := 0
	runeIndex := 0
	for len(s) > 0 {
		r, l := utf8.DecodeRune(s)
		s = s[l:]

		if runeIndex >= start && runeIndex < end {
			t[i] = r
			i++
		}

		runeIndex++
	}
	return t
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

func Append(s []byte, b ...byte) []byte {
	return append(s, b...)
}

func Insert(s []byte, i int, b ...byte) []byte {
	if i == 0 {
		return append(b, s...)
	}

	ri := RuneIndex(s, i)
	if ri == -1 || ri == len(s) {
		return append(s, b...)
	}

	return slices.Insert(s, ri, b...)
}

func Replace(s []byte, i int, b ...byte) []byte {
	ri := RuneIndex(s, i)
	if ri == -1 || ri == len(s) {
		return append(s, b...)
	}

	rl := RuneLen(s, i)

	return slices.Replace(s, ri, ri+rl, b...)
}

func ReplaceRange(s []byte, start int, end int, b ...byte) []byte {
	startIndex := RuneIndex(s, start)
	endIndex := RuneIndex(s, end)
	if startIndex == -1 || endIndex == -1 {
		return append(s, b...)
	}

	return slices.Replace(s, startIndex, endIndex, b...)
}
