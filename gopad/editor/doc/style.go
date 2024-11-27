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
	Start        int
	End          int
}

type highlightStyle struct {
	highlight    Highlight
	languageName string
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
	activeHighlights []highlightStyle
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
			case HighlightEventStart:
				log.Println("HighlightEventStart", event)
				i.activeHighlights = append(i.activeHighlights, highlightStyle{
					highlight:    event.Highlight,
					languageName: event.LanguageName,
				})
			case HighlightEventEnd:
				log.Println("HighlightEventEnd", event)
				_, i.activeHighlights = xslices.Pop(i.activeHighlights)
			case HighlightEventSource:
				ch := CharStyle{
					Style:     i.textStyle,
					StyleName: "text",
					Start:     i.buf.RuneIndex(int(event.StartByte)),
					End:       i.buf.RuneIndex(int(event.EndByte)),
				}

				if len(i.activeHighlights) > 0 {
					highlight := xslices.Last(i.activeHighlights)

					ch.Style = i.theme.Highlight(int(highlight.highlight), highlight.languageName)
					ch.StyleName = i.theme.Scope(int(highlight.highlight))
					ch.LanguageName = highlight.languageName
				}

				if ok := yield(ch); !ok {
					return
				}
			}
		}
	}
}
