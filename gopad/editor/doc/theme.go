package doc

import (
	"iter"
	"log"

	"github.com/charmbracelet/lipgloss/v2"

	"go.gopad.dev/gopad/gopad/editor/buffer"
	"go.gopad.dev/gopad/internal/xslices"
)

type Theme interface {
	Highlight(i int, languageName string) lipgloss.Style
	Scope(i int) string
	Scopes() []string
}

type CharStyle struct {
	Style        lipgloss.Style
	StyleName    string
	LanguageName string
	Start        uint
	End          uint
}

func newStyleIter(highlightIter iter.Seq2[HighlightEvent, error], buf buffer.Buffer, theme Theme) iter.Seq[CharStyle] {
	iterator := styleIterator{
		activeHighlights: nil,
		highlightIter:    highlightIter,
		buf:              buf,
		theme:            theme,
	}

	return iterator.iter()
}

type styleIterator struct {
	textStyle        lipgloss.Style
	activeHighlights []int
	activeLanguages  []string
	highlightIter    iter.Seq2[HighlightEvent, error]
	buf              buffer.Buffer
	theme            Theme
}

func (i *styleIterator) iter() iter.Seq[CharStyle] {
	return func(yield func(CharStyle) bool) {
		for event, err := range i.highlightIter {
			if err != nil {
				log.Printf("error getting highlight event: %v", err)
				continue
			}

			switch event := event.(type) {
			case HighlightEventLayerStart:
				i.activeHighlights = append(i.activeHighlights, -1)
				i.activeLanguages = append(i.activeLanguages, event.LanguageName)
			case HighlightEventLayerEnd:
				_, i.activeHighlights = xslices.Pop(i.activeHighlights)
				_, i.activeLanguages = xslices.Pop(i.activeLanguages)
			case HighlightEventCaptureStart:
				i.activeHighlights = append(i.activeHighlights, int(event.Highlight))
			case HighlightEventCaptureEnd:
				_, i.activeHighlights = xslices.Pop(i.activeHighlights)
			case HighlightEventSource:
				languageName := xslices.Last(i.activeLanguages)

				ch := CharStyle{
					Style:        i.textStyle,
					StyleName:    "text",
					LanguageName: languageName,
					Start:        i.buf.RuneIndex(event.StartByte),
					End:          i.buf.RuneIndex(event.EndByte),
				}

				if len(i.activeHighlights) > 0 {
					highlight := xslices.Last(i.activeHighlights)
					if highlight >= 0 {
						ch.Style = i.theme.Highlight(highlight, languageName)
						ch.StyleName = i.theme.Scope(highlight)
					}
				}

				if ok := yield(ch); !ok {
					return
				}
			}
		}
	}
}
