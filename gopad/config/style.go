package config

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Style struct {
	Foreground       string `toml:"foreground"`
	Background       string `toml:"background"`
	BorderForeground string `toml:"border_foreground"`
	BorderBackground string `toml:"border_background"`
	UnderlineColor   string `toml:"underline_color"`

	Bold            bool `toml:"bold"`
	Italic          bool `toml:"italic"`
	Underline       bool `toml:"underline"`
	DoubleUnderline bool `toml:"double_underline"`
	CurlyUnderline  bool `toml:"curly_underline"`
	DottedUnderline bool `toml:"dotted_underline"`
	DashedUnderline bool `toml:"dashed_underline"`
	Strikethrough   bool `toml:"strikethrough"`
	Reverse         bool `toml:"reverse"`
	Blink           bool `toml:"blink"`
	Faint           bool `toml:"faint"`
}

func (s Style) Style(colors ColorStyles) lipgloss.Style {
	style := lipgloss.NewStyle()

	if s.Foreground != "" {
		style = style.Foreground(getColor(colors, s.Foreground))
	}
	if s.Background != "" {
		style = style.Background(getColor(colors, s.Background))
	}
	if s.BorderForeground != "" {
		style = style.BorderForeground(getColor(colors, s.BorderForeground))
	}
	if s.BorderBackground != "" {
		style = style.BorderBackground(getColor(colors, s.BorderBackground))
	}
	if s.UnderlineColor != "" {
		style = style.UnderlineColor(getColor(colors, s.UnderlineColor))
	}

	style = style.Bold(s.Bold)
	style = style.Italic(s.Italic)
	style = style.Underline(s.Underline)
	style = style.DoubleUnderline(s.DoubleUnderline)
	style = style.CurlyUnderline(s.CurlyUnderline)
	style = style.DottedUnderline(s.DottedUnderline)
	style = style.Strikethrough(s.Strikethrough)
	style = style.Reverse(s.Reverse)
	style = style.Blink(s.Blink)
	style = style.Faint(s.Faint)

	return style
}

func getColor(colors ColorStyles, color string) color.Color {
	if color == "" {
		return lipgloss.NoColor{}
	}

	colorRef, ok := strings.CutPrefix(color, "$")
	if ok {
		if c, ok := colors[colorRef]; ok {
			return c
		}
	}

	return parseColor(color)
}

func parseColor(color string) color.Color {
	if ok := strings.HasPrefix(color, "#"); ok {
		return lipgloss.Color(color)
	}

	i, err := strconv.ParseUint(color, 10, 8)
	if err != nil {
		return lipgloss.Color(color)
	}

	return lipgloss.ANSIColor(i)
}
