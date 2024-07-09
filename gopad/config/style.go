package config

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Style struct {
	Foreground       string `toml:"foreground"`
	Background       string `toml:"background"`
	Bold             bool   `toml:"bold"`
	Italic           bool   `toml:"italic"`
	Underline        bool   `toml:"underline"`
	UnderlineColor   string `toml:"underline_color"`
	UnderlineStyle   string `toml:"underline_style"`
	Strikethrough    bool   `toml:"strikethrough"`
	Reverse          bool   `toml:"reverse"`
	Blink            bool   `toml:"blink"`
	Faint            bool   `toml:"faint"`
	BorderForeground string `toml:"border_foreground"`
	BorderBackground string `toml:"border_background"`
}

func (s Style) Style(colors Colors) lipgloss.Style {
	style := lipgloss.NewStyle()

	if s.Foreground != "" {
		style = style.Foreground(color(colors, s.Foreground))
	}

	if s.Background != "" {
		style = style.Background(color(colors, s.Background))
	}

	if s.BorderForeground != "" {
		style = style.BorderForeground(color(colors, s.BorderForeground))
	}

	if s.BorderBackground != "" {
		style = style.BorderBackground(color(colors, s.BorderBackground))
	}

	style = style.Bold(s.Bold)
	style = style.Italic(s.Italic)
	style = style.Underline(s.Underline)
	if s.UnderlineColor != "" {
		style = style.UnderlineColor(color(colors, s.UnderlineColor))
	}
	style = style.UnderlineStyle(underlineStyle(s.UnderlineStyle))

	style = style.Strikethrough(s.Strikethrough)
	style = style.Reverse(s.Reverse)
	style = style.Blink(s.Blink)
	style = style.Faint(s.Faint)

	return style
}

func color(colors Colors, color string) lipgloss.Color {
	colorRef, ok := strings.CutPrefix(color, "$")
	if ok {
		if c, ok := colors[colorRef]; ok {
			return c
		}

		panic("color ref not found: " + color)
	}

	return lipgloss.Color(color)
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
