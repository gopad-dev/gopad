package editor

import (
	"bytes"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"go.gopad.dev/gopad/gopad/editor/buffer"
)

type ModelType int

const (
	ModelTypeNone ModelType = iota
	ModelTypeFile
	ModelTypeSearchBar
	ModelTypeFileTree
)

func Focus(model ModelType) tea.Cmd {
	return func() tea.Msg {
		return FocusMsg{
			Model: model,
		}
	}
}

type FocusMsg struct {
	Model ModelType
}

func Search(term string) tea.Cmd {
	return func() tea.Msg {
		return SearchMsg{
			Term: term,
		}
	}
}

type SearchMsg struct {
	Term string
}

func ExecSearch(term string, buff buffer.Buffer) tea.Cmd {
	return func() tea.Msg {
		bytesTerm := []byte(term)
		termWidth := ansi.StringWidth(term)
		buf := buff.Bytes()

		var offset int
		var results []Result
		for {
			index := bytes.Index(buf[offset:], bytesTerm)
			if index == -1 {
				break
			}

			row, col := buff.Index(index + offset)
			rowEnd, colEnd := buff.Index(index + offset + termWidth)

			results = append(results, Result(buffer.Range{
				Start: buffer.Point{
					Row: row,
					Col: col,
				},
				End: buffer.Point{
					Row: rowEnd,
					Col: colEnd + 1,
				},
			}))

			offset += index + termWidth
		}

		return searchResultMsg{
			Results: results,
		}
	}
}

type searchResultMsg struct {
	Results []Result
}
