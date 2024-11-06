package main

import (
	"context"
	"embed"
	"log"
	"time"

	"go.gopad.dev/gopad/gopad/config"
)

var source = []byte(`package main

func main() {
	println("Hello, World!")
}
`)

func main() {
	highlighter := NewHighlighter()

	cfg, err := NewHighlightConfig("go", config.GrammarConfig{}, embed.FS{})
	if err != nil {
		log.Fatalf("failed to create highlight config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	highlights := highlighter.Highlight(ctx, *cfg, source, nil)

	for event := range highlights {
		switch e := event.(type) {
		case HighlightSource:
			log.Printf("highlighting source from %d to %d", e.Start, e.End)
		case HighlightStart:
			log.Printf("starting highlight with style %v", e.Highlight)
		case HighlightEnd:
			log.Printf("ending highlight")
		}
	}
}
