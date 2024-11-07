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

const (
	CaptureInjectionCombined        = "injection.combined"
	CaptureInjectionLanguage        = "injection.language"
	CaptureInjectionSelf            = "injection.self"
	CaptureInjectionParent          = "injection.parent"
	CaptureInjectionIncludeChildren = "injection.include-children"
	CaptureLocal                    = "local"
	CaptureLocalScope               = "local.scope"
)

type Highlight = lipgloss.Style

type HighlightEvent interface {
	highlightEvent()
}

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

	querySource := rawInjectionQuery
	localsQueryOffset := uint32(len(querySource))
	querySource = append(querySource, rawLocalsQuery...)
	highlightsQueryOffset := uint32(len(querySource))
	querySource = append(querySource, rawHighlightsQuery...)

	query, err := sitter.NewQuery(querySource, language)
	if err != nil {
		return nil, fmt.Errorf("error creating query: %w", err)
	}

	localsPatternIndex := uint32(0)
	highlightsPatternIndex := uint32(0)
	for i := range query.PatternCount() {
		patternOffset := query.PatternStartByte(i)
		if patternOffset < highlightsQueryOffset {
			if patternOffset < highlightsQueryOffset {
				highlightsPatternIndex++
			}
			if patternOffset < localsQueryOffset {
				localsPatternIndex++
			}
		}
	}

	combinedInjectionsQuery, err := sitter.NewQuery(rawInjectionQuery, language)
	if err != nil {
		return nil, fmt.Errorf("error creating combined injections query: %w", err)
	}
	var hasCombinedQueries bool
	for i := range localsPatternIndex {
		settings := combinedInjectionsQuery.PropertySettings(i)
		if slices.ContainsFunc(settings, func(setting sitter.QueryProperty) bool {
			return setting.Key == CaptureInjectionCombined
		}) {
			hasCombinedQueries = true
			query.DisablePattern(i)
		} else {
			combinedInjectionsQuery.DisablePattern(i)
		}
	}
	if !hasCombinedQueries {
		combinedInjectionsQuery = nil
	}

	nonLocalVariablePatterns := make([]bool, 0)
	for i := range query.PatternCount() {
		predicates := query.PropertyPredicates(i)
		if slices.ContainsFunc(predicates, func(predicate sitter.CaptureQueryProperty) bool {
			return !predicate.IsCapture && predicate.Property.Key == CaptureLocal
		}) {
			nonLocalVariablePatterns = append(nonLocalVariablePatterns, true)
		}
	}

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
		CombinedInjectionsQuery:       combinedInjectionsQuery,
		LocalsPatternIndex:            localsPatternIndex,
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
	CombinedInjectionsQuery       *sitter.Query
	LocalsPatternIndex            uint32
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

func (h *HighlightIter) next() (HighlightEvent, error) {
	for {
		if h.NextEvent != nil {
			event := h.NextEvent
			h.NextEvent = nil
			return event, nil
		}

		// check for cancellation
		select {
		case <-h.Ctx.Done():
			return nil, h.Ctx.Err()
		default:
		}

		// If none of the layers have any more highlight boundaries, terminate.
		if len(h.Layers) == 0 {
			if h.ByteOffset < len(h.Source) {
				event := HighlightSource{
					Start: h.ByteOffset,
					End:   len(h.Source),
				}
				h.ByteOffset = len(h.Source)
				return event, nil
			}
			return nil, nil
		}

		// Get the next capture from whichever layer has the earliest highlight boundary.
		var r *sitter.Range
		layer := h.Layers[0]
	}

	return nil, nil
}

func (h *HighlightIter) sortLayers() {
	for len(h.Layers) > 1 {
		sortKey := h.Layers[0].sortKey()
		if sortKey != nil {
			var i int
			for i+1 < len(h.Layers) {
				nextOffset := h.Layers[i+1].sortKey()
				if nextOffset != nil {
					if nextOffset.position < sortKey.position {
						i++
						continue
					}
				}
				break
			}
			if i > 0 {
				h.Layers = append(h.Layers[:i], append([]HighlightIterLayer{h.Layers[0]}, h.Layers[i:]...)...)
			}
			break
		}
		layer := h.Layers[0]
		h.Layers = h.Layers[1:]
		h.Highlighter.cursors = append(h.Highlighter.cursors, layer.Cursor)
	}
}

func (h *HighlightIter) insertLayer(layer HighlightIterLayer) {
	sortKey := layer.sortKey()
	if sortKey != nil {
		i := 1
		for i < len(h.Layers) {
			sortKeyI := h.Layers[i].sortKey()
			if sortKeyI != nil {
				if sortKeyI.position > sortKey.position {
					h.Layers = slices.Insert(h.Layers, i, layer)
					return
				}
				i++
			} else {
				h.Layers = slices.Delete(h.Layers, i, i+1)
			}
		}
		h.Layers = append(h.Layers, layer)
	}
}

type highlightQueueItem struct {
	config HighlightConfig
	depth  int
	ranges []sitter.Range
}

type injectionItem struct {
	languageName    string
	nodes           []*sitter.Node
	includeChildren bool
}

func NewHighlightIterLayer(
	ctx context.Context,
	source []byte,
	parentName string,
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

		if config.CombinedInjectionsQuery != nil {
			injectionsByPatternIndex := make([]injectionItem, config.CombinedInjectionsQuery.PatternCount())
			cursor.Exec(config.CombinedInjectionsQuery, tree.RootNode())
			for {

				match, ok := cursor.NextMatch()
				if !ok {
					break
				}

				languageName, contentNode, includeChildren := injectionForMatch(config, parentName, match)
				if languageName == "" {
					injectionsByPatternIndex[match.PatternIndex].languageName = languageName
				}
				if contentNode != nil {
					injectionsByPatternIndex[match.PatternIndex].nodes = append(injectionsByPatternIndex[match.PatternIndex].nodes, contentNode)
				}
				injectionsByPatternIndex[match.PatternIndex].includeChildren = includeChildren
			}

			for _, injection := range injectionsByPatternIndex {
				if injection.languageName != "" && len(injection.nodes) > 0 {
					nextConfig := injectionCallback(injection.languageName)
					if nextConfig != nil {
						nextRanges := intersectRanges(ranges, injection.nodes, injection.includeChildren)
						if len(nextRanges) > 0 {
							queue = append(queue, highlightQueueItem{
								config: *nextConfig,
								depth:  depth + 1,
								ranges: nextRanges,
							})
						}
					}

				}
			}

		}

		captures := make([]sitter.QueryCapture, 0)
		cursor.Exec(config.Query, tree.RootNode())
		for {
			match, i, ok := cursor.NextCapture()
			if !ok {
				break
			}
			captures = append(captures, match.Captures[i])
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

func intersectRanges(parentRanges []sitter.Range, nodes []*sitter.Node, includesChildren bool) []sitter.Range {
	cursor := sitter.NewTreeCursor(nodes[0])
	results := make([]sitter.Range, 0)
	if len(parentRanges) == 0 {
		panic("parentRanges must not be empty")
	}
	parentRange := parentRanges[0]
	parentRanges = parentRanges[1:]

	for _, node := range nodes {
		precedingRange := sitter.Range{
			StartByte: 0,
			StartPoint: sitter.Point{
				Row:    0,
				Column: 0,
			},
			EndByte:  node.StartByte(),
			EndPoint: node.StartPoint(),
		}
		followingRange := sitter.Range{
			StartByte:  node.EndByte(),
			StartPoint: node.EndPoint(),
			EndByte:    ^uint32(0),
			EndPoint: sitter.Point{
				Row:    ^uint32(0),
				Column: ^uint32(0),
			},
		}

		excludedRanges := make([]sitter.Range, 0)
		cursor.Reset(node)
		cursor.GoToFirstChild()
		for range node.ChildCount() {
			child := cursor.CurrentNode()
			cursor.GoToNextSibling()
			if !includesChildren {
				excludedRanges = append(excludedRanges, sitter.Range{
					StartByte:  child.StartByte(),
					StartPoint: child.StartPoint(),
					EndByte:    child.EndByte(),
					EndPoint:   child.EndPoint(),
				})
			}
		}
		excludedRanges = append(excludedRanges, followingRange)

		for _, excludedRange := range excludedRanges {
			r := sitter.Range{
				StartByte:  precedingRange.EndByte,
				StartPoint: precedingRange.EndPoint,
				EndByte:    excludedRange.StartByte,
				EndPoint:   excludedRange.StartPoint,
			}
			precedingRange = excludedRange

			if r.EndByte < parentRange.StartByte {
				continue
			}

			for parentRange.StartByte <= r.EndByte {
				if parentRange.EndByte > r.StartByte {
					if r.StartByte < parentRange.StartByte {
						r.StartByte = parentRange.StartByte
						r.StartPoint = parentRange.StartPoint
					}

					if parentRange.EndByte < r.EndByte {
						if r.StartByte < parentRange.EndByte {
							results = append(results, sitter.Range{
								StartByte:  r.StartByte,
								StartPoint: r.StartPoint,
								EndByte:    parentRange.EndByte,
								EndPoint:   parentRange.EndPoint,
							})
						}
						r.StartByte = parentRange.EndByte
						r.StartPoint = parentRange.EndPoint
					} else {
						if r.StartByte < r.EndByte {
							results = append(results, r)
						}
						break
					}
				}

				if len(parentRanges) > 0 {
					parentRange = parentRanges[0]
					parentRanges = parentRanges[1:]
				} else {
					return results
				}
			}
		}
	}

	return results
}

func injectionForMatch(config HighlightConfig, parentName string, match *sitter.QueryMatch) (string, *sitter.Node, bool) {
	contentCaptureIndex := *config.InjectionContentCaptureIndex
	languageCaptureIndex := *config.InjectionLanguageCaptureIndex

	var languageName string
	var contentNode *sitter.Node

	for _, capture := range match.Captures {
		index := capture.Index
		if index == languageCaptureIndex {
			languageName = capture.Node.Content()
		} else if index == contentCaptureIndex {
			contentNode = capture.Node
		}
	}

	var includeChildren bool
	for _, property := range config.Query.PropertySettings(uint32(match.PatternIndex)) {
		switch property.Key {
		case CaptureInjectionLanguage:
			if languageName == "" {
				languageName = *property.Value
			}
		case CaptureInjectionSelf:
			if languageName == "" {
				languageName = config.LanguageName
			}
		case CaptureInjectionParent:
			if languageName == "" {
				languageName = parentName
			}
		case CaptureInjectionIncludeChildren:
			includeChildren = true
		}
	}

	return languageName, contentNode, includeChildren
}

type HighlightIterLayer struct {
	Tree              *sitter.Tree
	Cursor            *sitter.QueryCursor
	Config            HighlightConfig
	HighlightEndStack []uint32
	ScopeStack        []LocalScope
	Captures          []sitter.QueryCapture
	Ranges            []sitter.Range
	Depth             int
}

type sortKeyResult struct {
	position uint32
	start    bool
	depth    int
}

func (h *HighlightIterLayer) sortKey() *sortKeyResult {
	depth := -h.Depth

	var nextStart *uint32
	if len(h.Captures) > 0 {
		startByte := h.Captures[0].StartByte()
		nextStart = &startByte
	}

	var nextEnd *uint32
	if len(h.HighlightEndStack) > 0 {
		endByte := h.HighlightEndStack[len(h.HighlightEndStack)-1]
		nextEnd = &endByte
	}

	switch {
	case nextStart != nil && nextEnd != nil:
		if *nextStart < *nextEnd {
			return &sortKeyResult{
				position: *nextStart,
				start:    true,
				depth:    depth,
			}
		} else {
			return &sortKeyResult{
				position: *nextEnd,
				start:    false,
				depth:    depth,
			}
		}
	case nextStart != nil && nextEnd == nil:
		return &sortKeyResult{
			position: *nextStart,
			start:    true,
			depth:    depth,
		}
	case nextStart == nil && nextEnd != nil:
		return &sortKeyResult{
			position: *nextEnd,
			start:    false,
			depth:    depth,
		}
	default:
		return nil
	}
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
) iter.Seq2[HighlightEvent, error] {
	layers, err := NewHighlightIterLayer(ctx, source, "", h, injectionCallback, cfg, 0, nil)

	if err != nil {
		return func(yield func(HighlightEvent, error) bool) {
			yield(nil, err)
		}
	}

	return func(yield func(HighlightEvent, error) bool) {
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
		highlightIter.sortLayers()

		for {
			event, err := highlightIter.next()
			// error we are done
			if err != nil {
				yield(nil, err)
				return
			}

			// we're done if there are no more events
			if event == nil {
				break
			}

			// yield the event
			if yield(event, nil) {
				return
			}
		}
	}
}
