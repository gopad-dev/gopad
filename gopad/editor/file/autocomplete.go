package file

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/ls"
)

func NewAutocompleter(f *File) *Autocompleter {
	return &Autocompleter{
		file: f,
	}
}

type Autocompleter struct {
	file        *File
	completions []ls.CompletionItem
	completion  int
	offset      int

	show bool
}

func (s *Autocompleter) Update() tea.Cmd {
	if s.Visible() {
		row, col := s.file.Cursor()
		return ls.GetAutocompletion(s.file.Name(), row, col)
	}

	return nil
}

func (s *Autocompleter) Visible() bool {
	return s.show
}

func (s *Autocompleter) SetCompletions(completions []ls.CompletionItem) {
	s.show = true
	s.completions = completions
	if len(s.completions) > 0 && len(s.completions)-1 < s.completion {
		s.completion = len(s.completions) - 1
	}
}

func (s *Autocompleter) ClearCompletions() {
	s.show = false
	s.completions = nil
	s.completion = 0
}

func (s *Autocompleter) Next() {
	if s.completion < len(s.completions)-1 {
		s.completion++
	}
}

func (s *Autocompleter) Previous() {
	if s.completion > 0 {
		s.completion--
	}
}

func (s *Autocompleter) Selected() *ls.CompletionItem {
	if len(s.completions) == 0 || s.completion >= len(s.completions) {
		return nil
	}

	item := s.completions[s.completion]
	return &item
}

func (s *Autocompleter) calculateOffset(height int) {
	if height == 0 {
		s.offset = 0
		return
	}

	if s.completion >= s.offset+height {
		s.offset = s.completion - height + 1
	} else if s.completion < s.offset {
		s.offset = s.completion
	}
}

func (s *Autocompleter) View(width int, height int) string {
	width = min(width, 80)
	height = min(height, 10)

	s.calculateOffset(height)

	autocompleteStyle := config.Theme.UI.Autocomplete.Style

	labelWidth := width - autocompleteStyle.GetHorizontalFrameSize()

	if len(s.completions) == 0 {
		view := config.Theme.UI.Autocomplete.ItemStyle.Render("No completions")
		return autocompleteStyle.Width(width).MaxHeight(height).Render(view)
	}

	var view string
	for i := range height {
		ii := i + s.offset
		if ii >= len(s.completions) {
			break
		}

		completion := s.completions[ii]
		style := config.Theme.UI.Autocomplete.ItemStyle

		icon := completion.Kind.Icon()
		details := completion.Kind.String()
		if i == s.completion {
			style = config.Theme.UI.Autocomplete.SelectedItemStyle
			if completion.Detail != "" {
				details = completion.Detail
			}
		}

		viewLabel := " " + completion.Label + strings.Repeat(" ", labelWidth-2-len(completion.Label)-len(details)) + details
		view += fmt.Sprintf("%s%s", style.Render(icon), style.Render(viewLabel)) + "\n"
	}

	view = strings.TrimRight(view, "\n")

	return autocompleteStyle.Width(width).MaxHeight(height).Render(view)
}
