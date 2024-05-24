package overlay

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Render whitespaces.
func render(width int) string {
	r := []rune(" ")
	j := 0
	b := strings.Builder{}

	// Cycle through runes and print them into the whitespace.
	for i := 0; i < width; {
		b.WriteRune(r[j])
		j++
		if j >= len(r) {
			j = 0
		}
		i += ansi.StringWidth(string(r[j]))
	}

	// Fill any extra gaps white spaces. This might be necessary if any runes
	// are more than one cell wide, which could leave a one-rune gap.
	short := width - ansi.StringWidth(b.String())
	if short > 0 {
		b.WriteString(strings.Repeat(" ", short))
	}

	return b.String()
}

type config struct {
	marginX int
	marginY int
}

type Option func(*config)

func WithMarginX(x int) Option {
	return func(c *config) {
		c.marginX = x
	}
}

func WithMarginY(y int) Option {
	return func(c *config) {
		c.marginY = y
	}
}
