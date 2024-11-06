package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"slices"

	"github.com/charmbracelet/lipgloss"
	"go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/cmd/grammar"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/buffer"
)

const (
	configDir  = "config"
	queriesDir = "queries"

	queryHighlightsFileName = "highlights.scm"
	queryInjectionsFileName = "injections.scm"
	queryLocalsFileName     = "locals.scm"
	queryOutlineFileName    = "outline.scm"
)

type Highlight = lipgloss.Style

type HighlightEvent interface {
	highlightEvent()
}

type HighlightError struct {
	Err error
}

func (HighlightError) highlightEvent() {}

type HighlightSource struct {
	Start int
	End   int
}

func (HighlightSource) highlightEvent() {}

type HighlightStart struct {
	Highlight
}

func (HighlightStart) highlightEvent() {}

type HighlightEnd struct{}

func (HighlightEnd) highlightEvent() {}

func NewHighlightConfig(languageName string, cfg config.GrammarConfig, defaultConfigs embed.FS) (*HighlightConfig, error) {
	name := cfg.Name
	if name == "" {
		name = languageName
	}

	symbolName := cfg.SymbolName
	if symbolName == "" {
		symbolName = cfg.Name
	}

	libPath := cfg.Path
	if libPath == "" {
		libPath = filepath.Join(config.Path, "grammars", grammar.LibName(name))
	}

	// compiled tree-sitter grammars are stored on disk only
	_, err := os.Stat(libPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("error checking lib %q: %w", libPath, err)
	}

	language, err := sitter.LoadLanguage(symbolName, libPath)
	if err != nil {
		return nil, fmt.Errorf("error loading lib %q: %w", libPath, err)
	}

	queriesConfigDir := cfg.QueriesDir
	if queriesConfigDir == "" {
		queriesConfigDir = filepath.Join(config.Path, queriesDir, name)
	}

	rawHighlightsQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryHighlightsFileName)
	if err != nil {
		return nil, fmt.Errorf("error reading highlights query: %w", err)
	}

	rawInjectionQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryInjectionsFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error reading injections query: %w", err)
	}

	rawLocalsQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryLocalsFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error reading locals query: %w", err)
	}

	querySource := rawLocalsQuery
	highlightsQueryOffset := uint32(len(querySource))
	querySource = append(querySource, rawHighlightsQuery...)

	query, err := sitter.NewQuery(querySource, language)
	if err != nil {
		return nil, fmt.Errorf("error creating query: %w", err)
	}

	highlightsPatternIndex := uint32(0)
	for i := range query.PatternCount() {
		patternOffset := query.PatternStartByte(i)
		if patternOffset < highlightsQueryOffset {
			highlightsPatternIndex += 1
		}
	}

	injectionsQuery, err := sitter.NewQuery(rawInjectionQuery, language)
	if err != nil {
		return nil, fmt.Errorf("error creating injections query: %w", err)
	}
	nonLocalVariablePatterns := make([]bool, 0) // TODO: needs changes in go-tree-sitter

	var (
		injectionContentCaptureIndex  *uint32
		injectionLanguageCaptureIndex *uint32
		localDefCaptureIndex          *uint32
		localDefValueCaptureIndex     *uint32
		localRefCaptureIndex          *uint32
		localScopeCaptureIndex        *uint32
	)

	for i := range query.CaptureCount() {
		captureName := query.CaptureNameForID(i)
		switch captureName {
		case "injection.content":
			injectionContentCaptureIndex = &i
		case "injection.language":
			injectionLanguageCaptureIndex = &i
		case "local.definition":
			localDefCaptureIndex = &i
		case "local.definition-value":
			localDefValueCaptureIndex = &i
		case "local.reference":
			localRefCaptureIndex = &i
		case "local.scope":
			localScopeCaptureIndex = &i
		}
	}

	highlightIndices := make([]*Highlight, query.CaptureCount())

	return &HighlightConfig{
		Language:                      language,
		LanguageName:                  languageName,
		Query:                         query,
		InjectionsQuery:               injectionsQuery,
		CombinedInjectionsQuery:       nil, // TODO: needs changes in go-tree-sitter
		HighlightsPatternIndex:        highlightsPatternIndex,
		HighlightIndices:              highlightIndices,
		NonLocalVariablePatterns:      nonLocalVariablePatterns,
		InjectionContentCaptureIndex:  injectionContentCaptureIndex,
		InjectionLanguageCaptureIndex: injectionLanguageCaptureIndex,
		LocalScopeCaptureIndex:        localScopeCaptureIndex,
		LocalDefCaptureIndex:          localDefCaptureIndex,
		LocalDefValueCaptureIndex:     localDefValueCaptureIndex,
		LocalRefCaptureIndex:          localRefCaptureIndex,
	}, nil
}

func readQuery(config string, defaultConfigs embed.FS, name string, query string) ([]byte, error) {
	_, err := os.Stat(filepath.Join(config, query))

	var f fs.File
	if errors.Is(err, os.ErrNotExist) {
		f, err = defaultConfigs.Open(filepath.Join(configDir, queriesDir, name, query))
	} else if err == nil {
		f, err = os.Open(filepath.Join(config, query))
	}

	if err != nil {
		return nil, fmt.Errorf("error opening query %q: %w", query, err)
	}

	defer func() {
		_ = f.Close()
	}()

	return io.ReadAll(f)
}

type HighlightConfig struct {
	Language                      *sitter.Language
	LanguageName                  string
	Query                         *sitter.Query
	InjectionsQuery               *sitter.Query
	CombinedInjectionsQuery       *sitter.Query
	HighlightsPatternIndex        uint32
	HighlightIndices              []*Highlight
	NonLocalVariablePatterns      []bool
	InjectionContentCaptureIndex  *uint32
	InjectionLanguageCaptureIndex *uint32
	LocalScopeCaptureIndex        *uint32
	LocalDefCaptureIndex          *uint32
	LocalDefValueCaptureIndex     *uint32
	LocalRefCaptureIndex          *uint32
}

type LocalDef struct {
	Name      string
	Range     buffer.Range
	Highlight *Highlight
}

type LocalScope struct {
	Inherits  bool
	Range     buffer.Range
	LocalDefs []LocalDef
}

type InjectionCallback func(name string) *HighlightConfig

type HighlightRange struct {
	Start int
	End   int
	Depth int
}

type HighlightIter struct {
	Ctx                context.Context
	Source             []byte
	LanguageName       string
	ByteOffset         int
	Highlighter        *Highlighter
	InjectionCallback  InjectionCallback
	Layers             []HighlightIterLayer
	IterCount          int
	NextEvent          HighlightEvent
	LastHighlightRange *HighlightRange
}

func (h *HighlightIter) next(emitEvent func(HighlightEvent)) bool {
	return false
}

func (h *HighlightIter) SortLayers() {
	for len(h.Layers) > 1 {
		l
	}
}

type highlightQueueItem struct {
	config   HighlightConfig
	depth    int
	ranges   []sitter.Range
	combined bool
}

func NewHighlightIterLayer(
	ctx context.Context,
	source []byte,
	parentName *string,
	highlighter *Highlighter,
	injectionCallback InjectionCallback,
	config HighlightConfig,
	depth int,
	ranges []sitter.Range,
) ([]HighlightIterLayer, error) {
	var result []HighlightIterLayer
	var queue []highlightQueueItem
	for {
		if len(ranges) > 0 {
			highlighter.Parser.SetIncludedRanges(ranges)
		}
		highlighter.Parser.SetLanguage(config.Language)
		tree, err := highlighter.Parser.ParseCtx(ctx, nil, source)
		if err != nil {
			return nil, err
		}

		cursor := highlighter.popCursor()
		if cursor == nil {
			cursor = sitter.NewQueryCursor()
		}

		cursor.Exec(config.InjectionsQuery, tree.RootNode())

		var captures []sitter.QueryCapture
		for {
			match, index, ok := cursor.NextCapture()
			if !ok {
				break
			}

			captures = append(captures, match.Captures[index])

			capture := match.Captures[index]
			if capture.Index != *config.InjectionContentCaptureIndex {
				continue
			}

			language, combined := getLanguageByMatch(parentName, match)
			if language == nil || language.Config.Grammar == nil {
				continue
			}

			start := capture.StartPoint()
			end := capture.EndPoint()

			if injectionCallback == nil {
				continue
			}

			nextConfig := injectionCallback(language.Name)
			if nextConfig == nil {
				continue
			}

			if combined {
				queueIndex := slices.IndexFunc(queue, func(item highlightQueueItem) bool {
					if !item.combined {
						return false
					}

					return item.config.LanguageName == nextConfig.LanguageName && item.depth == depth
				})

				if queueIndex > -1 {
					queue[queueIndex].ranges = append(queue[queueIndex].ranges, sitter.Range{
						StartPoint: start,
						EndPoint:   end,
					})
					continue
				}
			}

			queue = append(queue, highlightQueueItem{
				config: *nextConfig,
				depth:  depth + 1,
				ranges: []sitter.Range{
					{
						StartPoint: start,
						EndPoint:   end,
					},
				},
				combined: combined,
			})
		}

		result = append(result, HighlightIterLayer{
			Tree:              tree,
			Cursor:            cursor,
			Config:            config,
			HighlightEndStack: nil,
			ScopeStack: []LocalScope{
				{
					Inherits:  false,
					Range:     buffer.Range{},
					LocalDefs: nil,
				},
			},
			Captures: captures,
			Ranges:   ranges,
			Depth:    depth,
		})

		if len(queue) == 0 {
			break
		}

		var next highlightQueueItem
		next, queue = queue[0], append(queue, queue[1:]...)

		config = next.config
		depth = next.depth
		ranges = next.ranges
	}

	return result, nil
}

func getLanguageByMatch(parentName *string, match *sitter.QueryMatch) (*file.Language, bool) {
	var combined bool
	if combinedStr, ok := match.Properties["injection.combined"]; ok && combinedStr == "true" {
		combined = true
	}

	if subLanguageName, ok := match.Properties["injection.language"]; ok {
		if language := file.GetLanguage(subLanguageName); language != nil {
			return language, combined
		}
	}

	if subFileName, ok := match.Properties["injection.filename"]; ok {
		if language := file.GetLanguageByFilename(subFileName); language != nil {
			return language, combined
		}
	}

	if subMIMEType, ok := match.Properties["injection.mimetype"]; ok {
		if language := file.GetLanguageByMIMEType(subMIMEType); language != nil {
			return language, combined
		}
	}

	if _, ok := match.Properties["injection.parent"]; ok {
		if parentName != nil {
			return file.GetLanguage(*parentName), combined
		}
	}

	return nil, false
}

type HighlightIterLayer struct {
	Tree              *sitter.Tree
	Cursor            *sitter.QueryCursor
	Config            HighlightConfig
	HighlightEndStack []int
	ScopeStack        []LocalScope
	Captures          []sitter.QueryCapture
	Ranges            []sitter.Range
	Depth             int
}

func (h *HighlightIterLayer) SortKey() {
	depth := -h.Depth
	nextStart := h.Captures[0].StartPoint()

}

func NewHighlighter() *Highlighter {
	return &Highlighter{
		Parser: sitter.NewParser(),
	}
}

type Highlighter struct {
	Parser  *sitter.Parser
	cursors []*sitter.QueryCursor
}

func (h *Highlighter) popCursor() *sitter.QueryCursor {
	if len(h.cursors) == 0 {
		return nil
	}

	cursor := h.cursors[len(h.cursors)-1]
	h.cursors = h.cursors[:len(h.cursors)-1]
	return cursor
}

func (h *Highlighter) Highlight(
	ctx context.Context,
	cfg HighlightConfig,
	source []byte,
	injectionCallback InjectionCallback,
) iter.Seq[HighlightEvent] {
	layers, err := NewHighlightIterLayer(ctx, source, nil, h, injectionCallback, cfg, 0, nil)

	if err != nil {
		return func(yield func(HighlightEvent) bool) {
			yield(HighlightError{
				Err: err,
			})
		}
	}

	return func(yield func(HighlightEvent) bool) {
		highlightIter := &HighlightIter{
			Ctx:                ctx,
			Source:             source,
			LanguageName:       cfg.LanguageName,
			ByteOffset:         0,
			Highlighter:        h,
			InjectionCallback:  injectionCallback,
			Layers:             layers,
			IterCount:          0,
			NextEvent:          nil,
			LastHighlightRange: nil,
		}

		highlightIter.SortLayers()

		var done bool
		emitEvent := func(event HighlightEvent) {
			done = yield(event)
		}

		for highlightIter.next(emitEvent) {
			if done {
				return
			}

		}
	}
}
