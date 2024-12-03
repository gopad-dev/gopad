package doc

import (
	"context"
	"fmt"

	"go.gopad.dev/gopad/gopad/editor/buffer"
)

func injectionCallback(languageName string) *HighlightConfiguration {
	language := GetLanguage(languageName)
	if language == nil || language.Grammar == nil {
		return nil
	}

	return &language.Grammar.Highlight
}

func NewSyntax(ctx context.Context, language *Language, source []byte) (*Syntax, error) {
	if language == nil || language.Grammar == nil {
		return nil, nil
	}

	layers, err := NewSyntaxLayers(ctx, source, language.Grammar.Highlight)
	if err != nil {
		return nil, fmt.Errorf("error creating syntax layers: %w", err)
	}

	syntax := &Syntax{
		Language: language,
		Layers:   layers,
	}

	if err = syntax.parse(ctx, 0, source, nil); err != nil {
		return nil, fmt.Errorf("error parsing syntax: %w", err)
	}

	return syntax, nil
}

type Syntax struct {
	Rev      uint64
	Language *Language
	Layers   *SyntaxLayers
}

func (s *Syntax) parse(ctx context.Context, newRev uint64, newSource []byte, edits []SyntaxEdit) error {
	if s.Layers.layers.Len() == 0 {
		return nil
	}

	filteredEdits := make([]SyntaxEdit, 0)
	for _, edit := range edits {
		if newRev == s.Rev+uint64(len(edits)) {
			filteredEdits = append(filteredEdits, edit)
		}
	}

	if err := s.Layers.Update(ctx, s.Rev, newRev, newSource, filteredEdits); err != nil {
		return err
	}

	s.Rev = newRev

	return nil
}

func (s *Syntax) Update(ctx context.Context, newRev uint64, newBuf buffer.Buffer, oldBuf buffer.Buffer, changeSet ChangeSet) error {
	edits := generateEdits(oldBuf, changeSet)
	return s.parse(ctx, newRev, newBuf.Bytes(), []SyntaxEdit{edits})
}
