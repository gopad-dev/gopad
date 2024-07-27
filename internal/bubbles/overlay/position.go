package overlay

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/muesli/reflow/ansi"
)

func PlacePosition(xPos lipgloss.Position, yPos lipgloss.Position, fg string, bg string, opts ...Option) string {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	x := calculateXOffset(xPos, lipgloss.Width(fg), lipgloss.Width(bg), cfg.marginX)
	y := calculateYOffset(yPos, lipgloss.Height(fg), lipgloss.Height(bg), cfg.marginY)

	return Place(x, y, fg, bg)
}

// Place places fg on top of bg.
func Place(x int, y int, fg string, bg string) string {
	fgLines, fgWidth := getLines(fg)
	bgLines, bgWidth := getLines(bg)
	bgHeight := len(bgLines)
	fgHeight := len(fgLines)

	if fgWidth >= bgWidth && fgHeight >= bgHeight {
		return fg
	}

	var b strings.Builder
	for i, bgLine := range bgLines {
		if i > 0 {
			b.WriteByte('\n')
		}
		if i < y || i >= y+fgHeight {
			b.WriteString(bgLine)
			continue
		}

		pos := 0
		if x > 0 {
			left := xansi.Truncate(bgLine, x, "")
			pos = lipgloss.Width(left)
			b.WriteString(left)
			if pos < x {
				b.WriteString(render(x - pos))
				pos = x
			}
		}

		fgLine := fgLines[i-y]
		b.WriteString(fgLine)
		pos += lipgloss.Width(fgLine)

		right := cutLeft(bgLine, pos)
		bgWidth := lipgloss.Width(bgLine)
		rightWidth := lipgloss.Width(right)
		if rightWidth <= bgWidth-pos {
			b.WriteString(render(bgWidth - rightWidth - pos))
		}

		b.WriteString(right)
	}

	return b.String()
}

func getLines(s string) (lines []string, widest int) {
	lines = strings.Split(s, "\n")

	for _, l := range lines {
		w := lipgloss.Width(l)
		if widest < w {
			widest = w
		}
	}

	return lines, widest
}

func calculateXOffset(xPos lipgloss.Position, fgWidth int, bgWidth int, margin int) int {
	switch xPos {
	case lipgloss.Left:
		return 0 + margin
	case lipgloss.Center:
		return (bgWidth - fgWidth) / 2
	case lipgloss.Right:
		return bgWidth - fgWidth - margin
	default:
		x := int(float64(bgWidth-fgWidth) * float64(xPos))
		if xPos < 0.5 {
			return x + margin
		}
		return x - margin
	}
}

func calculateYOffset(yPos lipgloss.Position, fgHeight int, bgHeight int, margin int) int {
	switch yPos {
	case lipgloss.Top:
		return 0 + margin
	case lipgloss.Center:
		return (bgHeight - fgHeight) / 2
	case lipgloss.Bottom:
		return bgHeight - fgHeight - margin
	default:
		y := int(float64(bgHeight-fgHeight) * float64(yPos))
		if yPos < 0.5 {
			return y + margin
		}
		return y - margin
	}
}

// cutLeft cuts printable characters from the left.
// This function is heavily based on muesli's ansi and truncate packages.
func cutLeft(s string, cutWidth int) string {
	var (
		pos    int
		isAnsi bool
		ab     bytes.Buffer
		b      bytes.Buffer
	)
	for _, c := range s {
		var w int
		if c == ansi.Marker || isAnsi {
			isAnsi = true
			ab.WriteRune(c)
			if ansi.IsTerminator(c) {
				isAnsi = false
				if bytes.HasSuffix(ab.Bytes(), []byte("[0m")) {
					ab.Reset()
				}
			}
		} else {
			w = xansi.StringWidth(string(c))
		}

		if pos >= cutWidth {
			if b.Len() == 0 {
				if ab.Len() > 0 {
					b.Write(ab.Bytes())
				}
				if pos-cutWidth > 1 {
					b.WriteByte(' ')
					continue
				}
			}
			b.WriteRune(c)
		}
		pos += w
	}
	return b.String()
}
