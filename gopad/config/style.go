package config

import (
	"github.com/charmbracelet/lipgloss"
)

type Style struct {
	Foreground     string `toml:"foreground"`
	Background     string `toml:"background"`
	Bold           bool   `toml:"bold"`
	Italic         bool   `toml:"italic"`
	Underline      bool   `toml:"underline"`
	UnderlineColor string `toml:"underline_color"`
	UnderlineStyle string `toml:"underline_style"`
	Strikethrough  bool   `toml:"strikethrough"`
	Reverse        bool   `toml:"reverse"`
	Blink          bool   `toml:"blink"`
	Faint          bool   `toml:"faint"`
}

func (s Style) Style() lipgloss.Style {
	style := lipgloss.NewStyle()

	if s.Foreground != "" {
		style = style.Foreground(lipgloss.Color(s.Foreground))
	}

	if s.Background != "" {
		style = style.Background(lipgloss.Color(s.Background))
	}

	style = style.Bold(s.Bold)
	style = style.Italic(s.Italic)
	style = style.Underline(s.Underline)
	if s.UnderlineColor != "" {
		style = style.UnderlineColor(lipgloss.Color(s.UnderlineColor))
	}
	style = style.UnderlineStyle(underlineStyle(s.UnderlineStyle))

	style = style.Strikethrough(s.Strikethrough)
	style = style.Reverse(s.Reverse)
	style = style.Blink(s.Blink)
	style = style.Faint(s.Faint)

	return style
}

func underlineStyle(s string) lipgloss.UnderlineStyle {
	switch s {
	case "single":
		return lipgloss.UnderlineStyleSingle
	case "double":
		return lipgloss.UnderlineStyleDouble
	case "curly":
		return lipgloss.UnderlineStyleCurly
	case "dotted":
		return lipgloss.UnderlineStyleDotted
	case "dashed":
		return lipgloss.UnderlineStyleDashed
	}

	return lipgloss.UnderlineStyleSingle
}
