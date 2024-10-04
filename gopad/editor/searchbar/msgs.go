package searchbar

import (
	"bytes"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"go.gopad.dev/gopad/gopad/buffer"
)

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

func ExecSearch(term string, buffer *buffer.Buffer) tea.Cmd {
	return func() tea.Msg {
		bytesTerm := []byte(term)
		termWidth := ansi.StringWidth(term)
		buf := buffer.Bytes()

		var offset int
		var results []Result
		for {
			index := bytes.Index(buf[offset:], bytesTerm)
			if index == -1 {
				break
			}

			row, col := buffer.Index(index + offset)
			rowEnd, colEnd := buffer.Index(index + offset + termWidth)

			results = append(results, Result{
				RowStart: row,
				ColStart: col,
				RowEnd:   rowEnd,
				ColEnd:   colEnd + 1,
			})

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
