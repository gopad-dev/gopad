package buffer

import (
	"bytes"
	"log"
	"unicode"
)

type BlockCommentToken struct {
	Start string
	End   string
}

func (b *Buffer) ToggleBlockComment(from Position, to Position, row int, col int, tokens []BlockCommentToken) (int, int) {
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	if from.Row == to.Row {
		line := b.lines[from.Row]

		if line.Len() == 0 {
			return row, col
		}

		var hasStartComment bool
		var hasEndComment bool
		var blockToken *BlockCommentToken
		for _, token := range tokens {
			if string(line.RunesRange(from.Col, from.Col+len(token.Start))) == token.Start {
				hasStartComment = true
			}
			if string(line.RunesRange(to.Col-len(token.End), to.Col)) == token.End {
				hasEndComment = true
			}
			if hasStartComment && hasEndComment {
				blockToken = &token
				break
			}
		}

		if blockToken == nil {
			blockToken = &tokens[0]
		}

		if hasStartComment && hasEndComment {
			log.Println("removing inline block comment")
			b.lines[from.Row] = line.ReplaceRange(from.Col, from.Col+len(blockToken.Start)).ReplaceRange(to.Col-len(blockToken.Start), to.Col-len(blockToken.Start)+len(blockToken.End))
			if col >= from.Col {
				col -= len(blockToken.Start)
			}
			if col >= to.Col {
				col -= len(blockToken.End)
			}
		} else {
			log.Println("adding inline block comment")
			b.lines[from.Row] = line.Insert(from.Col, []byte(blockToken.Start)...).Insert(to.Col+len(blockToken.Start), []byte(blockToken.End)...)
			if col >= from.Col {
				col += len(blockToken.Start)
			}
			if col >= to.Col {
				col += len(blockToken.End)
			}
		}

		return row, col
	}

	startLine := b.lines[from.Row]
	endLine := b.lines[to.Row]

	var hasStartComment bool
	var hasEndComment bool
	var blockToken *BlockCommentToken
	for _, token := range tokens {
		if startLine.Len() > 0 {
			if string(startLine.RunesRange(from.Col, from.Col+len(token.Start))) == token.Start {
				hasStartComment = true
			}
		}

		if endLine.Len() > 0 {
			if string(endLine.RunesRange(max(0, to.Col-len(token.End)), to.Col)) == token.End {
				hasEndComment = true
			}
		}

		if hasStartComment && hasEndComment {
			blockToken = &token
			break
		}

		hasStartComment = false
		hasEndComment = false
	}

	if blockToken == nil {
		blockToken = &tokens[0]
	}

	if hasStartComment && hasEndComment {
		log.Println("removing line block comment")
		b.lines[from.Row] = startLine.ReplaceRange(from.Col, from.Col+len(blockToken.Start))
		b.lines[to.Row] = endLine.ReplaceRange(to.Col-len(blockToken.End), to.Col)
		if row == from.Row && col >= from.Col {
			col -= len(blockToken.Start)
		} else if row == to.Row && col >= to.Col {
			col -= len(blockToken.End)
		}
	} else {
		log.Println("adding line block comment")
		b.lines[from.Row] = startLine.Insert(from.Col, []byte(blockToken.Start)...)
		b.lines[to.Row] = endLine.Insert(to.Col, []byte(blockToken.End)...)
		if row == from.Row && col >= from.Col {
			col += len(blockToken.Start)
		} else if row == to.Row && col >= to.Col {
			col += len(blockToken.End)
		}
	}

	return row, col
}

func (b *Buffer) ToggleLineComment(row int, col int, tokens []string) (int, int) {
	defer func() {
		b.version++
		b.refreshDirty()
	}()

	line := b.lines[row]

	if line.Len() == 0 {
		return row, col
	}

	for i, r := range line.Runes() {
		if !unicode.IsSpace(r) {
			lineData := line.CutStart(i).Bytes()

			if ok, token := hasPrefixes(lineData, tokens); ok {
				b.lines[row] = line.CutEnd(i).Append(line.CutStart(i + len(token)))
				if col >= i+len(token) {
					col -= len(token)
				}
				return row, col
			}

			token := tokens[0]
			b.lines[row] = line.Insert(i, []byte(token)...)
			if col >= i {
				col += len(token)
			}

			return row, col
		}
	}

	return row, col
}

func hasPrefixes(b []byte, prefixes []string) (bool, string) {
	for _, prefix := range prefixes {
		if bytes.HasPrefix(b, []byte(prefix)) {
			return true, prefix
		}
	}
	return false, ""
}
