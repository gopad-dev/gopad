package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
)

var source = []byte(`package main

func main() {
	println("Hello, World!")
}
`)

//go:embed config/*
var defaultConfigs embed.FS

func main() {
	if err := config.Load(configDir, defaultConfigs); err != nil {
		log.Panicln("failed to load config:", err)
	}

	cfg, err := NewHighlightConfig("go", config.GrammarConfig{
		Name:       "go",
		SymbolName: "go",
		QueriesDir: "config/queries/go",
		Path:       "config/grammars/libtree-sitter-go.so",
	}, embed.FS{})
	if err != nil {
		log.Fatalf("failed to create highlight config: %v", err)
	}
	if cfg == nil {
		log.Fatalf("tree-sitter grammar not found")
	}

	highlighter := NewHighlighter()
	highlights := highlighter.Highlight(context.Background(), *cfg, source, func(name string) *HighlightConfig {
		log.Println("loading highlight config for", name)
		return nil
	})

	var highlightName *string
	for event, err := range highlights {
		if err != nil {
			log.Panicf("failed to highlight source: %v", err)
		}
		switch e := event.(type) {
		case HighlightSource:
			var style lipgloss.Style
			if highlightName != nil {
				style = getMatchingStyle(*highlightName, cfg.LanguageName)
			}

			print(style.Render(string(source[e.Start:e.End])))

		case HighlightStart:
			highlightName = (*string)(&e.Highlight)
		case HighlightEnd:
			highlightName = nil
		}
	}
}

func getMatchingStyle(matchType string, languageName string) lipgloss.Style {
	var currentStyle lipgloss.Style

	for {
		codeStyle, ok := config.Theme.CodeStyles[fmt.Sprintf("%s.%s", matchType, languageName)]
		if ok {
			currentStyle = codeStyle
			break
		}
		codeStyle, ok = config.Theme.CodeStyles[matchType]
		if ok {
			currentStyle = codeStyle
			break
		}
		lastDot := strings.LastIndex(matchType, ".")
		if lastDot == -1 {
			break
		}
		matchType = matchType[:lastDot]
	}

	return currentStyle
}
