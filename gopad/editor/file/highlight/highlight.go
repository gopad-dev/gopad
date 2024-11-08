package highlight

import (
	"context"
	"fmt"
	"iter"
	"log"
	"slices"

	"github.com/tree-sitter/go-tree-sitter"
)

const (
	captureInjectionCombined        = "injection.combined"
	captureInjectionLanguage        = "injection.language"
	captureInjectionSelf            = "injection.self"
	captureInjectionParent          = "injection.parent"
	captureInjectionIncludeChildren = "injection.include-children"
	captureLocal                    = "local"
	captureLocalScopeInherits       = "local.scope-inherits"
)

type Event interface {
	highlightEvent()
}

type EventSource struct {
	Start uint
	End   uint
}

func (EventSource) highlightEvent() {}

type EventStart struct {
	CaptureName  string
	LanguageName string
}

func (EventStart) highlightEvent() {}

type EventEnd struct{}

func (EventEnd) highlightEvent() {}

func NewConfig(language *tree_sitter.Language, languageName string, highlightsQuery []byte, injectionQuery []byte, localsQuery []byte) (*Config, error) {
	querySource := injectionQuery
	localsQueryOffset := uint(len(querySource))
	querySource = append(querySource, localsQuery...)
	highlightsQueryOffset := uint(len(querySource))
	querySource = append(querySource, highlightsQuery...)

	query, err := tree_sitter.NewQuery(language, string(querySource))
	if err != nil {
		return nil, fmt.Errorf("error creating query: %w", err)
	}

	localsPatternIndex := uint(0)
	highlightsPatternIndex := uint(0)
	for i := range query.PatternCount() {
		patternOffset := query.StartByteForPattern(i)
		if patternOffset < highlightsQueryOffset {
			if patternOffset < highlightsQueryOffset {
				highlightsPatternIndex++
			}
			if patternOffset < localsQueryOffset {
				localsPatternIndex++
			}
		}
	}

	combinedInjectionsQuery, err := tree_sitter.NewQuery(language, string(injectionQuery))
	if err != nil {
		return nil, fmt.Errorf("error creating combined injections query: %w", err)
	}
	var hasCombinedQueries bool
	for i := range localsPatternIndex {
		settings := combinedInjectionsQuery.PropertySettings(i)
		if slices.ContainsFunc(settings, func(setting tree_sitter.QueryProperty) bool {
			return setting.Key == captureInjectionCombined
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
		if slices.ContainsFunc(predicates, func(predicate tree_sitter.PropertyPredicate) bool {
			return !predicate.Positive && predicate.Property.Key == captureLocal
		}) {
			nonLocalVariablePatterns = append(nonLocalVariablePatterns, true)
		}
	}

	var (
		injectionContentCaptureIndex  *uint
		injectionLanguageCaptureIndex *uint
		localDefCaptureIndex          *uint
		localDefValueCaptureIndex     *uint
		localRefCaptureIndex          *uint
		localScopeCaptureIndex        *uint
	)

	highlightIndices := make([]string, 0)
	for i, captureName := range query.CaptureNames() {
		ui := uint(i)
		switch captureName {
		case "injection.content":
			injectionContentCaptureIndex = &ui
		case "injection.language":
			injectionLanguageCaptureIndex = &ui
		case "local.definition":
			localDefCaptureIndex = &ui
		case "local.definition-value":
			localDefValueCaptureIndex = &ui
		case "local.reference":
			localRefCaptureIndex = &ui
		case "local.scope":
			localScopeCaptureIndex = &ui
		default:
			highlightIndices = append(highlightIndices, captureName)
		}
	}

	return &Config{
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

type Config struct {
	Language                      *tree_sitter.Language
	LanguageName                  string
	Query                         *tree_sitter.Query
	CombinedInjectionsQuery       *tree_sitter.Query
	LocalsPatternIndex            uint
	HighlightsPatternIndex        uint
	HighlightIndices              []string
	NonLocalVariablePatterns      []bool
	InjectionContentCaptureIndex  *uint
	InjectionLanguageCaptureIndex *uint
	LocalScopeCaptureIndex        *uint
	LocalDefCaptureIndex          *uint
	LocalDefValueCaptureIndex     *uint
	LocalRefCaptureIndex          *uint
}

type localDef struct {
	Name      string
	Range     tree_sitter.Range
	Highlight *string
}

type localScope struct {
	Inherits  bool
	Range     tree_sitter.Range
	LocalDefs []localDef
}

type InjectionCallback func(name string) *Config

type iterRange struct {
	Start uint
	End   uint
	Depth int
}

type highlightIter struct {
	Ctx                context.Context
	Source             []byte
	LanguageName       string
	ByteOffset         uint
	Highlighter        *Highlighter
	InjectionCallback  InjectionCallback
	Layers             []*iterLayer
	NextEvent          Event
	LastHighlightRange *iterRange
}

func (h *highlightIter) emitEvent(offset uint, event Event) (Event, error) {
	var result Event
	if h.ByteOffset < offset {
		result = EventSource{
			Start: h.ByteOffset,
			End:   offset,
		}
		h.ByteOffset = offset
		h.NextEvent = event
	} else {
		result = event
	}
	h.sortLayers()
	return result, nil
}

func (h *highlightIter) next() (Event, error) {
main:
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
			if h.ByteOffset < uint(len(h.Source)) {
				event := EventSource{
					Start: h.ByteOffset,
					End:   uint(len(h.Source)),
				}
				h.ByteOffset = uint(len(h.Source))
				return event, nil
			}
			log.Println("we are done")
			return nil, nil
		}

		// Get the next capture from whichever layer has the earliest highlight boundary.
		var r tree_sitter.Range
		layer := h.Layers[0]
		if len(layer.Captures) > 0 {
			nextMatch := layer.Captures[0]
			nextCapture := nextMatch.Match.Captures[nextMatch.Index]
			r = nextCapture.Node.Range()

			// If any previous highlight ends before this node starts, then before
			// processing this capture, emit the source code up until the end of the
			// previous highlight, and an end event for that highlight.
			if len(layer.HighlightEndStack) > 0 {
				endByte := layer.HighlightEndStack[len(layer.HighlightEndStack)-1]
				if endByte <= r.StartByte {
					layer.HighlightEndStack = layer.HighlightEndStack[:len(layer.HighlightEndStack)-1]
					return h.emitEvent(endByte, EventEnd{})
				}
			}
		} else {
			// If there are no more captures, then emit any remaining highlight end events.
			// And if there are none of those, then just advance to the end of the document.
			if len(layer.HighlightEndStack) > 0 {
				endByte := layer.HighlightEndStack[len(layer.HighlightEndStack)-1]
				layer.HighlightEndStack = layer.HighlightEndStack[:len(layer.HighlightEndStack)-1]
				return h.emitEvent(endByte, EventEnd{})
			}
			return h.emitEvent(uint(len(h.Source)), nil)
		}

		match := layer.Captures[0]
		layer.Captures = layer.Captures[1:]
		capture := match.Match.Captures[match.Index]

		if match.Match.PatternIndex < layer.Config.LocalsPatternIndex {
			languageName, contentNode, includeChildren := injectionForMatch(layer.Config, h.LanguageName, match.Match, h.Source)

			match.Match.Remove()

			if languageName != "" && contentNode != nil {
				newConfig := h.InjectionCallback(languageName)
				if newConfig != nil {
					ranges := intersectRanges(h.Layers[0].Ranges, []*tree_sitter.Node{contentNode}, includeChildren)
					if len(ranges) > 0 {
						newLayers, err := newIterLayers(h.Ctx, h.Source, h.LanguageName, h.Highlighter, h.InjectionCallback, *newConfig, h.Layers[0].Depth+1, ranges)
						if err != nil {
							return nil, err
						}
						for _, newLayer := range newLayers {
							h.insertLayer(newLayer)
						}
					}
				}
			}

			h.sortLayers()
			continue main
		}

		// Remove from the local scope stack any local scopes that have already ended.
		for r.StartByte > layer.ScopeStack[len(layer.ScopeStack)-1].Range.EndByte {
			layer.ScopeStack = layer.ScopeStack[:len(layer.ScopeStack)-1]
		}

		// If this capture is for tracking local variables, then process the
		// local variable info.
		var referenceHighlight *string
		var definitionHighlight *string
		for match.Match.PatternIndex < layer.Config.HighlightsPatternIndex {
			// If the node represents a local scope, push a new local scope onto
			// the scope stack.
			if layer.Config.LocalScopeCaptureIndex != nil && uint(capture.Index) == *layer.Config.LocalScopeCaptureIndex {
				definitionHighlight = nil
				scope := localScope{
					Inherits:  true,
					Range:     r,
					LocalDefs: nil,
				}
				for _, prop := range layer.Config.Query.PropertySettings(match.Match.PatternIndex) {
					if prop.Key == captureLocalScopeInherits {
						scope.Inherits = *prop.Value == "true"
					}
				}
				layer.ScopeStack = append(layer.ScopeStack, scope)
			} else if layer.Config.LocalDefCaptureIndex != nil && uint(capture.Index) == *layer.Config.LocalDefCaptureIndex {
				// If the node represents a definition, add a new definition to the
				// local scope at the top of the scope stack.
				referenceHighlight = nil
				definitionHighlight = nil
				scope := layer.ScopeStack[len(layer.ScopeStack)-1]

				var valueRange tree_sitter.Range
				for _, matchCapture := range match.Match.Captures {
					if layer.Config.LocalDefValueCaptureIndex != nil && uint(matchCapture.Index) == *layer.Config.LocalDefValueCaptureIndex {
						valueRange = matchCapture.Node.Range()
					}
				}

				if len(h.Source) > int(r.StartByte) && len(h.Source) > int(valueRange.EndByte) {
					name := string(h.Source[r.StartByte:r.EndByte])

					scope.LocalDefs = append(scope.LocalDefs, localDef{
						Name:      name,
						Range:     r,
						Highlight: nil,
					})
					definitionHighlight = scope.LocalDefs[len(scope.LocalDefs)-1].Highlight
				}
			} else if layer.Config.LocalRefCaptureIndex != nil && uint(capture.Index) == *layer.Config.LocalRefCaptureIndex && definitionHighlight == nil {
				// If the node represents a reference, then try to find the corresponding
				// definition in the scope stack.
				definitionHighlight = nil
				if len(h.Source) > int(r.StartByte) && len(h.Source) > int(r.EndByte) {
					name := string(h.Source[r.StartByte:r.EndByte])
					for _, scope := range slices.Backward(layer.ScopeStack) {
						var highlight *string
						for _, def := range slices.Backward(scope.LocalDefs) {
							if def.Name == name && r.StartByte >= def.Range.EndByte {
								highlight = def.Highlight
							}
						}
						if highlight != nil {
							referenceHighlight = highlight
							break
						}
						if !scope.Inherits {
							break
						}
					}
				}
			}

			// Continue processing any additional matches for the same node.
			if len(layer.Captures) > 0 {
				nextMatch := layer.Captures[0]
				nextCapture := nextMatch.Match.Captures[nextMatch.Index]
				if nextCapture.Node.Equals(capture.Node) {
					capture = nextCapture
					match = nextMatch
					layer.Captures = layer.Captures[1:]
					continue
				}
			}

			h.sortLayers()
			continue main
		}

		// Otherwise, this capture must represent a highlight.
		// If this exact range has already been highlighted by an earlier pattern, or by
		// a different layer, then skip over this one.
		if h.LastHighlightRange != nil {
			lastRange := *h.LastHighlightRange
			if r.StartByte == lastRange.Start && r.EndByte == lastRange.End && layer.Depth < lastRange.Depth {
				h.sortLayers()
				continue main
			}
		}

		// Once a highlighting pattern is found for the current node, keep iterating over
		// any later highlighting patterns that also match this node and set the match to it.
		// Captures for a given node are ordered by pattern index, so these subsequent
		// captures are guaranteed to be for highlighting, not injections or
		// local variables.
		for len(layer.Captures) > 0 {
			nextMatch := layer.Captures[0]
			nextCapture := nextMatch.Match.Captures[nextMatch.Index]
			if nextCapture.Node.Equals(capture.Node) {
				followingMatch := nextMatch
				layer.Captures = layer.Captures[1:]
				// If the current node was found to be a local variable, then ignore
				// the following match if it's a highlighting pattern that is disabled
				// for local variables.
				if definitionHighlight != nil || referenceHighlight != nil && layer.Config.NonLocalVariablePatterns[followingMatch.Match.PatternIndex] {
					continue
				}

				match.Match.Remove()
				capture = nextCapture
				match = nextMatch
			} else {
				break
			}
		}

		currentHighlight := layer.Config.HighlightIndices[capture.Index]

		// If this node represents a local definition, then store the current
		// highlight value on the local scope entry representing this node.
		if definitionHighlight != nil {
			definitionHighlight = &currentHighlight
		}

		// Emit a scope start event and push the node's end position to the stack.
		highlight := referenceHighlight
		if highlight == nil {
			highlight = &currentHighlight
		}
		if highlight != nil {
			h.LastHighlightRange = &iterRange{
				Start: r.StartByte,
				End:   r.EndByte,
				Depth: layer.Depth,
			}
			layer.HighlightEndStack = append(layer.HighlightEndStack, r.EndByte)
			return h.emitEvent(r.StartByte, EventStart{
				CaptureName:  *highlight,
				LanguageName: layer.Config.LanguageName,
			})
		}

		h.sortLayers()
	}
}

func (h *highlightIter) sortLayers() {
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
				h.Layers = append(h.Layers[:i], append([]*iterLayer{h.Layers[0]}, h.Layers[i:]...)...)
			}
			break
		}
		layer := h.Layers[0]
		h.Layers = h.Layers[1:]
		h.Highlighter.cursors = append(h.Highlighter.cursors, layer.Cursor)
	}
}

func (h *highlightIter) insertLayer(layer *iterLayer) {
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
	config Config
	depth  int
	ranges []tree_sitter.Range
}

type injectionItem struct {
	languageName    string
	nodes           []*tree_sitter.Node
	includeChildren bool
}

func newIterLayers(
	ctx context.Context,
	source []byte,
	parentName string,
	highlighter *Highlighter,
	injectionCallback InjectionCallback,
	config Config,
	depth int,
	ranges []tree_sitter.Range,
) ([]*iterLayer, error) {
	var result []*iterLayer
	var queue []highlightQueueItem
	for {
		if err := highlighter.Parser.SetIncludedRanges(ranges); err == nil {
			if err = highlighter.Parser.SetLanguage(config.Language); err != nil {
				return nil, fmt.Errorf("error setting language: %w", err)
			}
			tree := highlighter.Parser.ParseCtx(ctx, source, nil)

			cursor := highlighter.popCursor()
			if cursor == nil {
				cursor = tree_sitter.NewQueryCursor()
			}

			if config.CombinedInjectionsQuery != nil {
				injectionsByPatternIndex := make([]injectionItem, config.CombinedInjectionsQuery.PatternCount())

				matches := cursor.Matches(config.CombinedInjectionsQuery, tree.RootNode(), source)
				for {
					match := matches.Next()
					if match == nil {
						break
					}

					languageName, contentNode, includeChildren := injectionForMatch(config, parentName, match, source)
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

			captures := make([]queryCapture, 0)
			queryCaptures := cursor.Captures(config.Query, tree.RootNode(), source)
			for {
				capture, i := queryCaptures.Next()
				if capture == nil {
					break
				}
				captures = append(captures, queryCapture{
					Match: capture,
					Index: i,
				})
			}

			result = append(result, &iterLayer{
				Tree:              tree,
				Cursor:            cursor,
				Config:            config,
				HighlightEndStack: nil,
				ScopeStack: []localScope{
					{
						Inherits: false,
						Range: tree_sitter.Range{
							StartByte: 0,
							StartPoint: tree_sitter.Point{
								Row:    0,
								Column: 0,
							},
							EndByte: ^uint(0),
							EndPoint: tree_sitter.Point{
								Row:    ^uint(0),
								Column: ^uint(0),
							},
						},
						LocalDefs: nil,
					},
				},
				Captures: captures,
				Ranges:   ranges,
				Depth:    depth,
			})
		}

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

func intersectRanges(parentRanges []tree_sitter.Range, nodes []*tree_sitter.Node, includesChildren bool) []tree_sitter.Range {
	cursor := nodes[0].Walk()
	results := make([]tree_sitter.Range, 0)
	if len(parentRanges) == 0 {
		panic("parentRanges must not be empty")
	}
	parentRange := parentRanges[0]
	parentRanges = parentRanges[1:]

	for _, node := range nodes {
		precedingRange := tree_sitter.Range{
			StartByte: 0,
			StartPoint: tree_sitter.Point{
				Row:    0,
				Column: 0,
			},
			EndByte:  node.StartByte(),
			EndPoint: node.StartPosition(),
		}
		followingRange := tree_sitter.Range{
			StartByte:  node.EndByte(),
			StartPoint: node.EndPosition(),
			EndByte:    ^uint(0),
			EndPoint: tree_sitter.Point{
				Row:    ^uint(0),
				Column: ^uint(0),
			},
		}

		excludedRanges := make([]tree_sitter.Range, 0)
		cursor.Reset(node)
		cursor.GotoFirstChild()
		for range node.ChildCount() {
			child := cursor.Node()
			cursor.GotoNextSibling()
			if !includesChildren {
				excludedRanges = append(excludedRanges, tree_sitter.Range{
					StartByte:  child.StartByte(),
					StartPoint: child.StartPosition(),
					EndByte:    child.EndByte(),
					EndPoint:   child.EndPosition(),
				})
			}
		}
		excludedRanges = append(excludedRanges, followingRange)

		for _, excludedRange := range excludedRanges {
			r := tree_sitter.Range{
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
							results = append(results, tree_sitter.Range{
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

func injectionForMatch(config Config, parentName string, match *tree_sitter.QueryMatch, source []byte) (string, *tree_sitter.Node, bool) {
	contentCaptureIndex := *config.InjectionContentCaptureIndex
	languageCaptureIndex := *config.InjectionLanguageCaptureIndex

	var languageName string
	var contentNode *tree_sitter.Node

	for _, capture := range match.Captures {
		index := uint(capture.Index)
		if index == languageCaptureIndex {
			languageName = capture.Node.Utf8Text(source)
		} else if index == contentCaptureIndex {
			contentNode = capture.Node
		}
	}

	var includeChildren bool
	for _, property := range config.Query.PropertySettings(match.PatternIndex) {
		switch property.Key {
		case captureInjectionLanguage:
			if languageName == "" {
				languageName = *property.Value
			}
		case captureInjectionSelf:
			if languageName == "" {
				languageName = config.LanguageName
			}
		case captureInjectionParent:
			if languageName == "" {
				languageName = parentName
			}
		case captureInjectionIncludeChildren:
			includeChildren = true
		}
	}

	return languageName, contentNode, includeChildren
}

type queryCapture struct {
	Match *tree_sitter.QueryMatch
	Index uint
}

type iterLayer struct {
	Tree              *tree_sitter.Tree
	Cursor            *tree_sitter.QueryCursor
	Config            Config
	HighlightEndStack []uint
	ScopeStack        []localScope
	Captures          []queryCapture
	Ranges            []tree_sitter.Range
	Depth             int
}

type sortKeyResult struct {
	position uint
	start    bool
	depth    int
}

func (h *iterLayer) sortKey() *sortKeyResult {
	depth := -h.Depth

	var nextStart *uint
	if len(h.Captures) > 0 {
		capture := h.Captures[0]
		startByte := capture.Match.Captures[capture.Index].Node.StartByte()
		nextStart = &startByte
	}

	var nextEnd *uint
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

func New() *Highlighter {
	return &Highlighter{
		Parser: tree_sitter.NewParser(),
	}
}

type Highlighter struct {
	Parser  *tree_sitter.Parser
	cursors []*tree_sitter.QueryCursor
}

func (h *Highlighter) popCursor() *tree_sitter.QueryCursor {
	if len(h.cursors) == 0 {
		return nil
	}

	cursor := h.cursors[len(h.cursors)-1]
	h.cursors = h.cursors[:len(h.cursors)-1]
	return cursor
}

func (h *Highlighter) Highlight(
	ctx context.Context,
	cfg Config,
	source []byte,
	injectionCallback InjectionCallback,
) iter.Seq2[Event, error] {
	layers, err := newIterLayers(ctx, source, "", h, injectionCallback, cfg, 0, nil)
	if err != nil {
		return func(yield func(Event, error) bool) {
			yield(nil, err)
		}
	}

	return func(yield func(Event, error) bool) {
		hIter := &highlightIter{
			Ctx:                ctx,
			Source:             source,
			LanguageName:       cfg.LanguageName,
			ByteOffset:         0,
			Highlighter:        h,
			InjectionCallback:  injectionCallback,
			Layers:             layers,
			NextEvent:          nil,
			LastHighlightRange: nil,
		}
		hIter.sortLayers()

		for {
			event, err := hIter.next()
			// error we are done
			if err != nil {
				yield(nil, err)
				return
			}

			// we're done if there are no more events
			if event == nil {
				return
			}

			// yield the event
			if !yield(event, nil) {
				return
			}
		}
	}
}
