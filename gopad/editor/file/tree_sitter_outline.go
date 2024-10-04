package file

import (
	"slices"

	"go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/buffer"
)

type OutlineItem struct {
	Range buffer.Range
	Text  []OutlineItemChar
}

type OutlineItemChar struct {
	Char string
	Pos  *buffer.Position
}

type outlineBufferRange struct {
	r      byteRange
	isName bool
}

type byteRange struct {
	start int
	end   int
}

func (f *File) OutlineTree() []OutlineItem {
	if f.tree == nil || f.tree.Tree == nil || f.tree.Language.Grammar == nil || f.tree.Language.Grammar.OutlineQuery == nil {
		return nil
	}

	queryConfig := f.tree.Language.Grammar.OutlineQuery
	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(queryConfig.Query, f.tree.Tree.RootNode())

	var items []OutlineItem
	for {
		match, ok := queryCursor.NextMatch()
		if !ok {
			break
		}

		itemNodeIndex := slices.IndexFunc(match.Captures, func(capture sitter.QueryCapture) bool {
			return capture.Index == queryConfig.ItemCaptureID
		})
		if itemNodeIndex < 0 {
			continue
		}

		itemCapture := match.Captures[itemNodeIndex]
		itemRange := buffer.Range{
			Start: buffer.Position{
				Row: int(itemCapture.Node.StartPoint().Row),
				Col: int(itemCapture.Node.StartPoint().Column),
			},
			End: buffer.Position{
				Row: int(itemCapture.Node.EndPoint().Row),
				Col: int(itemCapture.Node.EndPoint().Column),
			},
		}

		var bufferRanges []outlineBufferRange
		for _, capture := range match.Captures {
			var isName bool
			if capture.Index == queryConfig.NameCaptureID {
				isName = true
			} else if (queryConfig.ContextCaptureID != nil && capture.Index == *queryConfig.ContextCaptureID) || (queryConfig.ExtraContextCaptureID != nil && capture.Index == *queryConfig.ExtraContextCaptureID) {
				isName = false
			} else {
				continue
			}

			r := byteRange{
				start: int(capture.Node.StartByte()),
				end:   int(capture.Node.EndByte()),
			}
			start := capture.Node.StartPoint()

			if capture.Node.EndPoint().Row > start.Row {
				r.end = r.start + f.buffer.LineLen(int(start.Row)) - int(start.Column)
			}

			bufferRanges = append(bufferRanges, outlineBufferRange{
				r:      r,
				isName: isName,
			})
		}

		if len(bufferRanges) == 0 {
			continue
		}

		var chars []OutlineItemChar
		var nameRanges []byteRange
		var lastBufferRangeEnd int
		for _, bufferRange := range bufferRanges {
			if len(chars) != 0 && bufferRange.r.start > lastBufferRangeEnd {
				chars = append(chars, OutlineItemChar{
					Char: " ",
				})
			}

			lastBufferRangeEnd = bufferRange.r.end
			if bufferRange.isName {
				start := len(chars)
				end := start + bufferRange.r.end - bufferRange.r.start

				if len(nameRanges) != 0 {
					start -= 1
				}

				nameRanges = append(nameRanges, byteRange{
					start: start,
					end:   end,
				})
			}

			start := f.buffer.Position(bufferRange.r.start)
			end := f.buffer.Position(bufferRange.r.end)

			for i := start.Row; i <= end.Row; i++ {
				line := f.buffer.Line(i)
				var colOffset int
				if i == start.Row && i == end.Row {
					line = line.CutRange(start.Col, end.Col)
					colOffset = start.Col
				} else if i == start.Row {
					line = line.CutStart(start.Col)
					colOffset = start.Col
				} else if i == end.Row {
					line = line.CutEnd(end.Col)
				}

				for j, char := range line.RuneStrings() {
					chars = append(chars, OutlineItemChar{
						Char: char,
						Pos: &buffer.Position{
							Row: i,
							Col: j + colOffset,
						},
					})
				}
			}
		}

		items = append(items, OutlineItem{
			Range: itemRange,
			Text:  chars,
		})
	}

	return items
}
