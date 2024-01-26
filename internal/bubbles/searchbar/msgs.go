package searchbar

import (
	"github.com/charmbracelet/bubbletea"
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

func SearchResult(results []Result) tea.Cmd {
	return func() tea.Msg {
		return searchResultMsg{
			Results: results,
		}
	}
}

type searchResultMsg struct {
	Results []Result
}
