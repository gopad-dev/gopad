package file

import (
	"fmt"
	"strings"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/ls"
)

func NewAutocompleter() *Autocompleter {
	return &Autocompleter{}
}

type Autocompleter struct {
	completions []ls.CompletionItem
	completion  int
	offset      int

	show bool
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

func (s *Autocompleter) Selected() ls.CompletionItem {
	if len(s.completions) == 0 {
		return ls.CompletionItem{}
	}

	return s.completions[s.completion]
}

func (s *Autocompleter) calculateOffset(height int) {
	if height == 0 {
		s.offset = 0
		return
	}

	if s.completion > s.offset+height {
		s.offset = s.completion - height + 1
	} else if s.completion < s.offset {
		s.offset = s.completion
	}
}

func (s *Autocompleter) View(width int, height int) string {
	width = min(width, 80)
	height = min(height, 10)

	s.calculateOffset(height)

	autocompleteStyle := config.Theme.Editor.Autocomplete.Style

	labelWidth := width - autocompleteStyle.GetHorizontalFrameSize()

	var view string
	for i := range height {
		ii := i + s.offset
		if ii >= len(s.completions) {
			break
		}

		completion := s.completions[ii]
		style := config.Theme.Editor.Autocomplete.ItemStyle
		details := completion.Kind.String()

		if i == s.completion {
			style = config.Theme.Editor.Autocomplete.SelectedItemStyle
			if completion.Detail != "" {
				details = completion.Detail
			}
		}

		view += style.Render(fmt.Sprintf("%s%s%s", completion.Label, strings.Repeat(" ", labelWidth-len(completion.Label)-len(details)), details)) + "\n"
	}

	view = strings.TrimRight(view, "\n")

	return autocompleteStyle.Width(width).MaxHeight(height).Render(view)
}
