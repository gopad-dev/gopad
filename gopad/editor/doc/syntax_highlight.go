package doc

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/tree-sitter/go-tree-sitter"
)

const (
	captureInjectionLanguage        = "injection.language"
	captureInjectionContent         = "injection.content"
	captureInjectionCombined        = "injection.combined"
	captureInjectionSelf            = "injection.self"
	captureInjectionIncludeChildren = "injection.include-children"

	captureLocal                = "local"
	captureLocalDefinition      = "local.definition"
	captureLocalDefinitionValue = "local.definition-value"
	captureLocalReference       = "local.reference"
	captureLocalScope           = "local.scope"
	captureLocalScopeInherits   = "local.scope-inherits"
)

type Highlight uint

type HighlightEvent interface {
	highlightEvent()
}

type HighlightEventLayerStart struct {
	LanguageName string
}

func (HighlightEventLayerStart) highlightEvent() {}

type HighlightEventLayerEnd struct{}

func (HighlightEventLayerEnd) highlightEvent() {}

type HighlightEventCaptureStart struct {
	Highlight Highlight
}

func (HighlightEventCaptureStart) highlightEvent() {}

type HighlightEventCaptureEnd struct{}

func (HighlightEventCaptureEnd) highlightEvent() {}

type HighlightEventSource struct {
	StartByte uint
	EndByte   uint
}

func (HighlightEventSource) highlightEvent() {}

func NewHighlightConfig(language *tree_sitter.Language, languageName string, highlightsQuery []byte, injectionQuery []byte, localsQuery []byte) (HighlightConfiguration, error) {
	var querySource []byte
	querySource = append(querySource, localsQuery...)
	highlightsQueryOffset := uint(len(querySource))
	querySource = append(querySource, highlightsQuery...)

	query, err := tree_sitter.NewQuery(language, string(querySource))
	if err != nil {
		return HighlightConfiguration{}, fmt.Errorf("error creating query: %w", err)
	}

	highlightsPatternIndex := uint(0)
	for i := range query.PatternCount() {
		patternOffset := query.StartByteForPattern(i)
		if patternOffset < highlightsQueryOffset {
			highlightsPatternIndex++
		}
	}

	injectionsQuery, err := tree_sitter.NewQuery(language, string(injectionQuery))
	if err != nil {
		return HighlightConfiguration{}, fmt.Errorf("error creating combined injections query: %w", err)
	}
	var combinedInjectionsPatterns []uint
	for i := range injectionsQuery.PatternCount() {
		settings := injectionsQuery.PropertySettings(i)
		if slices.ContainsFunc(settings, func(setting tree_sitter.QueryProperty) bool {
			return setting.Key == captureInjectionCombined
		}) {
			combinedInjectionsPatterns = append(combinedInjectionsPatterns, i)
		}
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
		localDefCaptureIndex      *uint
		localDefValueCaptureIndex *uint
		localRefCaptureIndex      *uint
		localScopeCaptureIndex    *uint
	)
	for i, captureName := range query.CaptureNames() {
		ui := uint(i)
		switch captureName {
		case captureLocalDefinition:
			localDefCaptureIndex = &ui
		case captureLocalDefinitionValue:
			localDefValueCaptureIndex = &ui
		case captureLocalReference:
			localRefCaptureIndex = &ui
		case captureLocalScope:
			localScopeCaptureIndex = &ui
		}
	}

	var (
		injectionContentCaptureIndex  *uint
		injectionLanguageCaptureIndex *uint
	)
	for i, captureName := range injectionsQuery.CaptureNames() {
		ui := uint(i)
		switch captureName {
		case captureInjectionContent:
			injectionContentCaptureIndex = &ui
		case captureInjectionLanguage:
			injectionLanguageCaptureIndex = &ui
		}
	}

	highlightIndices := make([]*Highlight, len(query.CaptureNames()))
	return HighlightConfiguration{
		Language:                      language,
		LanguageName:                  languageName,
		Query:                         query,
		InjectionsQuery:               injectionsQuery,
		CombinedInjectionsPatterns:    combinedInjectionsPatterns,
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

type HighlightConfiguration struct {
	Language                      *tree_sitter.Language
	LanguageName                  string
	Query                         *tree_sitter.Query
	InjectionsQuery               *tree_sitter.Query
	CombinedInjectionsPatterns    []uint
	HighlightsPatternIndex        uint
	HighlightIndices              []*Highlight
	NonLocalVariablePatterns      []bool
	InjectionContentCaptureIndex  *uint
	InjectionLanguageCaptureIndex *uint
	LocalScopeCaptureIndex        *uint
	LocalDefCaptureIndex          *uint
	LocalDefValueCaptureIndex     *uint
	LocalRefCaptureIndex          *uint
}

func (c *HighlightConfiguration) Configure(recognizedNames []string) {
	highlightIndices := make([]*Highlight, len(c.Query.CaptureNames()))
	for i, captureName := range c.Query.CaptureNames() {
		for {
			j := slices.Index(recognizedNames, captureName)
			if j != -1 {
				index := Highlight(j)
				highlightIndices[i] = &index
				break
			}

			lastDot := strings.LastIndex(captureName, ".")
			if lastDot == -1 {
				break
			}
			captureName = captureName[:lastDot]
		}
	}
	c.HighlightIndices = highlightIndices
}

func (c *HighlightConfiguration) NonconformantCaptureNames(captureNames []string) []string {
	var nonconformantNames []string
	for _, name := range c.Query.CaptureNames() {
		if !(strings.HasPrefix(name, "_") || strings.HasPrefix(name, "local") || slices.Contains(captureNames, name)) {
			nonconformantNames = append(nonconformantNames, name)
		}
	}

	return nonconformantNames
}

type LocalDef struct {
	Name      string
	Range     tree_sitter.Range
	Highlight *Highlight
}

type LocalScope struct {
	Inherits  bool
	Range     tree_sitter.Range
	LocalDefs []LocalDef
}

type InjectionCallback func(name string) *HighlightConfiguration

type iterRange struct {
	Start uint
	End   uint
	Depth int
}

type highlightIter struct {
	Ctx                context.Context
	Source             []byte
	ByteOffset         uint
	Layers             []*highlightIterLayer
	NextEvents         []HighlightEvent
	LastHighlightRange *iterRange
	Syntax             *SyntaxLayers
	LastLayer          *highlightIterLayer
}

func (h *highlightIter) emitEvents(offset uint, events ...HighlightEvent) (HighlightEvent, error) {
	var result HighlightEvent
	if h.ByteOffset < offset {
		result = HighlightEventSource{
			StartByte: h.ByteOffset,
			EndByte:   offset,
		}
		h.ByteOffset = offset
		h.NextEvents = append(events, events...)
	} else {
		if len(events) > 1 {
			h.NextEvents = append(h.NextEvents, events[1:]...)
		}
		result = events[0]
	}
	h.sortLayers()
	return result, nil
}

func (h *highlightIter) sortLayers() {
	for len(h.Layers) > 0 {
		key := h.Layers[0].sortKey()
		if key != nil {
			var i int
			for i+1 < len(h.Layers) {
				nextOffsetKey := h.Layers[i+1].sortKey()
				if nextOffsetKey != nil {
					if nextOffsetKey.GreaterThan(*key) {
						i += 1
						continue
					}
				}
				break
			}
			if i > 0 {
				h.Layers = append(rotateLeft(h.Layers[:i+1], 1), h.Layers[i+1:]...)
			}
			break
		}
		layer := h.Layers[0]
		h.Layers = h.Layers[1:]
		h.Syntax.parser.pushCursor(layer.Cursor)
	}
}

func rotateLeft[T any](s []T, i int) []T {
	return append(s[i:], s[:i]...)
}

func (h *highlightIter) next() (HighlightEvent, error) {
main:
	for {
		if len(h.NextEvents) > 0 {
			event := h.NextEvents[0]
			h.NextEvents = h.NextEvents[1:]
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
			sourceLen := uint(len(h.Source))
			if h.ByteOffset < sourceLen {
				result := HighlightEventSource{
					StartByte: h.ByteOffset,
					EndByte:   sourceLen,
				}
				h.ByteOffset = sourceLen
				return result, nil
			}
			return nil, nil
		}

		// Get the next capture from whichever layer has the earliest highlight boundary.
		layer := h.Layers[0]
		if layer != h.LastLayer {
			var events []HighlightEvent
			if h.LastLayer != nil {
				events = append(events, HighlightEventLayerEnd{})
			}
			h.LastLayer = layer

			return h.emitEvents(h.ByteOffset, append(events, HighlightEventLayerStart{
				LanguageName: layer.Config.LanguageName,
			})...)
		}

		var nextCaptureRange tree_sitter.Range
		if nextMatch, captureIndex, ok := layer.Captures.Peek(); ok {
			nextCapture := nextMatch.Captures[captureIndex]
			nextCaptureRange = nextCapture.Node.Range()

			// If any previous highlight ends before this node starts, then before
			// processing this capture, emit the source code up until the end of the
			// previous highlight, and an end event for that highlight.
			if len(layer.HighlightEndStack) > 0 {
				endByte := layer.HighlightEndStack[len(layer.HighlightEndStack)-1]
				if endByte <= nextCaptureRange.StartByte {
					layer.HighlightEndStack = layer.HighlightEndStack[:len(layer.HighlightEndStack)-1]
					return h.emitEvents(endByte, HighlightEventCaptureEnd{})
				}
			}
		} else {
			// If there are no more captures, then emit any remaining highlight end events.
			// And if there are none of those, then just advance to the end of the document.
			if len(layer.HighlightEndStack) > 0 {
				endByte := layer.HighlightEndStack[len(layer.HighlightEndStack)-1]
				layer.HighlightEndStack = layer.HighlightEndStack[:len(layer.HighlightEndStack)-1]
				return h.emitEvents(endByte, HighlightEventCaptureEnd{})
			}
			return h.emitEvents(uint(len(h.Source)), nil)
		}

		match, captureIndex, _ := layer.Captures.Next()
		capture := match.Captures[captureIndex]

		// Remove from the local scope stack any local scopes that have already ended.
		for nextCaptureRange.StartByte > layer.ScopeStack[len(layer.ScopeStack)-1].Range.EndByte {
			layer.ScopeStack = layer.ScopeStack[:len(layer.ScopeStack)-1]
		}

		// If this capture is for tracking local variables, then process the
		// local variable info.
		var referenceHighlight *Highlight
		var definitionHighlight *Highlight
		for match.PatternIndex < layer.Config.HighlightsPatternIndex {
			// If the node represents a local scope, push a new local scope onto
			// the scope stack.
			if layer.Config.LocalScopeCaptureIndex != nil && uint(capture.Index) == *layer.Config.LocalScopeCaptureIndex {
				definitionHighlight = nil
				scope := LocalScope{
					Inherits:  true,
					Range:     nextCaptureRange,
					LocalDefs: nil,
				}
				for _, prop := range layer.Config.Query.PropertySettings(match.PatternIndex) {
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
				for _, matchCapture := range match.Captures {
					if layer.Config.LocalDefValueCaptureIndex != nil && uint(matchCapture.Index) == *layer.Config.LocalDefValueCaptureIndex {
						valueRange = matchCapture.Node.Range()
					}
				}

				if len(h.Source) > int(nextCaptureRange.StartByte) && len(h.Source) > int(valueRange.EndByte) {
					name := string(h.Source[nextCaptureRange.StartByte:nextCaptureRange.EndByte])

					scope.LocalDefs = append(scope.LocalDefs, LocalDef{
						Name:      name,
						Range:     nextCaptureRange,
						Highlight: nil,
					})
					definitionHighlight = scope.LocalDefs[len(scope.LocalDefs)-1].Highlight
				}
			} else if layer.Config.LocalRefCaptureIndex != nil && uint(capture.Index) == *layer.Config.LocalRefCaptureIndex && definitionHighlight == nil {
				// If the node represents a reference, then try to find the corresponding
				// definition in the scope stack.
				definitionHighlight = nil
				if len(h.Source) > int(nextCaptureRange.StartByte) && len(h.Source) > int(nextCaptureRange.EndByte) {
					name := string(h.Source[nextCaptureRange.StartByte:nextCaptureRange.EndByte])
					for _, scope := range slices.Backward(layer.ScopeStack) {
						var highlight *Highlight
						for _, def := range slices.Backward(scope.LocalDefs) {
							if def.Name == name && nextCaptureRange.StartByte >= def.Range.EndByte {
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
			if nextMatch, nextCaptureIndex, ok := layer.Captures.Peek(); ok {
				nextCapture := nextMatch.Captures[nextCaptureIndex]
				if nextCapture.Node.Equals(capture.Node) {
					capture = nextCapture
					match, _, _ = layer.Captures.Next()
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
			if nextCaptureRange.StartByte == lastRange.Start && nextCaptureRange.EndByte == lastRange.End && layer.Depth < lastRange.Depth {
				h.sortLayers()
				continue main
			}
		}

		// Once a highlighting pattern is found for the current node, keep iterating over
		// any later highlighting patterns that also queryCapture this node and set the queryCapture to it.
		// Captures for a given node are ordered by pattern index, so these subsequent
		// captures are guaranteed to be for highlighting, not injections or
		// local variables.
		for {
			nextMatch, nextCaptureIndex, ok := layer.Captures.Peek()
			if !ok {
				break
			}

			nextCapture := nextMatch.Captures[nextCaptureIndex]
			if nextCapture.Node.Equals(capture.Node) {
				followingMatch, _, _ := layer.Captures.Next()
				// If the current node was found to be a local variable, then ignore
				// the following queryCapture if it's a highlighting pattern that is disabled
				// for local variables.
				if definitionHighlight != nil || referenceHighlight != nil && layer.Config.NonLocalVariablePatterns[followingMatch.PatternIndex] {
					continue
				}

				match.Remove()
				capture = nextCapture
				match = followingMatch
			} else {
				break
			}
		}

		currentHighlight := layer.Config.HighlightIndices[capture.Index]

		// If this node represents a local definition, then store the current
		// highlight value on the local scope entry representing this node.
		if definitionHighlight != nil {
			definitionHighlight = currentHighlight
		}

		// Emit a scope start event and push the node's end position to the stack.
		highlight := referenceHighlight
		if highlight == nil {
			highlight = currentHighlight
		}
		if highlight != nil {
			h.LastHighlightRange = &iterRange{
				Start: nextCaptureRange.StartByte,
				End:   nextCaptureRange.EndByte,
				Depth: layer.Depth,
			}
			layer.HighlightEndStack = append(layer.HighlightEndStack, nextCaptureRange.EndByte)
			return h.emitEvents(nextCaptureRange.StartByte, HighlightEventCaptureStart{
				Highlight: *highlight,
			})
		}

		h.sortLayers()
	}
}

func intersectRanges(parentRanges []tree_sitter.Range, nodes []tree_sitter.Node, includesChildren bool) []tree_sitter.Range {
	return []tree_sitter.Range{
		nodes[0].Range(),
	}

	// cursor := nodes[0].Walk()
	// results := make([]tree_sitter.Range, 0)
	// if len(parentRanges) == 0 {
	//	panic("parentRanges must not be empty")
	// }
	// parentRange := parentRanges[0]
	// parentRanges = parentRanges[1:]
	//
	// for _, node := range nodes {
	//	precedingRange := tree_sitter.Range{
	//		StartByte: 0,
	//		StartPoint: tree_sitter.Point{
	//			Row:    0,
	//			Column: 0,
	//		},
	//		EndByte:  node.StartByte(),
	//		EndPoint: node.StartPosition(),
	//	}
	//	followingRange := tree_sitter.Range{
	//		StartByte:  node.EndByte(),
	//		StartPoint: node.EndPosition(),
	//		EndByte:    ^uint(0),
	//		EndPoint: tree_sitter.Point{
	//			Row:    ^uint(0),
	//			Column: ^uint(0),
	//		},
	//	}
	//
	//	excludedRanges := make([]tree_sitter.Range, 0)
	//	cursor.Reset(node)
	//	cursor.GotoFirstChild()
	//	for range node.ChildCount() {
	//		child := cursor.Node()
	//		cursor.GotoNextSibling()
	//		if !includesChildren {
	//			excludedRanges = append(excludedRanges, tree_sitter.Range{
	//				StartByte:  child.StartByte(),
	//				StartPoint: child.StartPosition(),
	//				EndByte:    child.EndByte(),
	//				EndPoint:   child.EndPosition(),
	//			})
	//		}
	//	}
	//	excludedRanges = append(excludedRanges, followingRange)
	//
	//	for _, excludedRange := range excludedRanges {
	//		r := tree_sitter.Range{
	//			StartByte:  precedingRange.EndByte,
	//			StartPoint: precedingRange.EndPoint,
	//			EndByte:    excludedRange.StartByte,
	//			EndPoint:   excludedRange.StartPoint,
	//		}
	//		precedingRange = excludedRange
	//
	//		if r.EndByte < parentRange.StartByte {
	//			continue
	//		}
	//
	//		for parentRange.StartByte <= r.EndByte {
	//			if parentRange.EndByte > r.StartByte {
	//				if r.StartByte < parentRange.StartByte {
	//					r.StartByte = parentRange.StartByte
	//					r.StartPoint = parentRange.StartPoint
	//				}
	//
	//				if parentRange.EndByte < r.EndByte {
	//					if r.StartByte < parentRange.EndByte {
	//						results = append(results, tree_sitter.Range{
	//							StartByte:  r.StartByte,
	//							StartPoint: r.StartPoint,
	//							EndByte:    parentRange.EndByte,
	//							EndPoint:   parentRange.EndPoint,
	//						})
	//					}
	//					r.StartByte = parentRange.EndByte
	//					r.StartPoint = parentRange.EndPoint
	//				} else {
	//					if r.StartByte < r.EndByte {
	//						results = append(results, r)
	//					}
	//					break
	//				}
	//			}
	//
	//			if len(parentRanges) > 0 {
	//				parentRange = parentRanges[0]
	//				parentRanges = parentRanges[1:]
	//			} else {
	//				return results
	//			}
	//		}
	//	}
	// }
	//
	// return results
}

func (c HighlightConfiguration) injectionForMatch(query *tree_sitter.Query, queryMatch tree_sitter.QueryMatch, source []byte) (string, *tree_sitter.Node, bool) {
	if c.InjectionContentCaptureIndex == nil || c.InjectionLanguageCaptureIndex == nil {
		return "", nil, false
	}
	contentCaptureIndex := *c.InjectionContentCaptureIndex
	languageCaptureIndex := *c.InjectionLanguageCaptureIndex

	var languageName string
	var contentNode *tree_sitter.Node

	for _, capture := range queryMatch.Captures {
		index := uint(capture.Index)
		if index == languageCaptureIndex {
			languageName = capture.Node.Utf8Text(source)
		} else if index == contentCaptureIndex {
			contentNode = &capture.Node
		}
	}

	var includeChildren bool
	for _, property := range query.PropertySettings(queryMatch.PatternIndex) {
		switch property.Key {
		case captureInjectionLanguage:
			if languageName == "" {
				languageName = *property.Value
			}
		case captureInjectionSelf:
			if languageName == "" {
				languageName = c.LanguageName
			}
		case captureInjectionIncludeChildren:
			includeChildren = true
		}
	}

	return languageName, contentNode, includeChildren
}

type highlightIterLayer struct {
	Tree              *tree_sitter.Tree
	Cursor            *tree_sitter.QueryCursor
	Config            HighlightConfiguration
	HighlightEndStack []uint
	ScopeStack        []LocalScope
	Captures          *queryCapturesIter
	Depth             int
}

type sortKey struct {
	offset uint
	start  bool
	depth  int
}

// Compare compares the current sortKey (k) with another sortKey (other) lexicographically.
// Returns:
//
// -1 if other is greater
//
//	1 if k is greater
//
// 0 if both are equal
func (k sortKey) Compare(other sortKey) int {
	if k.offset < other.offset {
		return -1
	}
	if k.offset > other.offset {
		return 1
	}

	if !k.start && other.start {
		return -1
	}
	if k.start && !other.start {
		return 1
	}

	if k.depth < other.depth {
		return -1
	}
	if k.depth > other.depth {
		return 1
	}

	return 0
}

func (k sortKey) GreaterThan(other sortKey) bool {
	return k.Compare(other) == 1
}

func (k sortKey) LessThan(other sortKey) bool {
	return k.Compare(other) == -1
}

// First, sort scope boundaries by their byte offset in the document. At a
// given position, emit scope endings before scope beginnings. Finally, emit
// scope boundaries from deeper layers first.
func (h *highlightIterLayer) sortKey() *sortKey {
	depth := -int(h.Depth)

	var nextStart *uint
	if match, index, ok := h.Captures.Peek(); ok {
		startByte := match.Captures[index].Node.StartByte()
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
			return &sortKey{
				offset: *nextStart,
				start:  true,
				depth:  depth,
			}
		} else {
			return &sortKey{
				offset: *nextEnd,
				start:  false,
				depth:  depth,
			}
		}
	case nextStart != nil:
		return &sortKey{
			offset: *nextStart,
			start:  true,
			depth:  depth,
		}
	case nextEnd != nil:
		return &sortKey{
			offset: *nextEnd,
			start:  false,
			depth:  depth,
		}
	default:
		return nil
	}
}

type peekedCapture struct {
	match tree_sitter.QueryMatch
	index uint
	ok    bool
}

func newQueryCapturesIter(iter tree_sitter.QueryCaptures) *queryCapturesIter {
	return &queryCapturesIter{captures: iter}
}

type queryCapturesIter struct {
	captures tree_sitter.QueryCaptures
	peeked   *peekedCapture
}

func (q *queryCapturesIter) next() (tree_sitter.QueryMatch, uint, bool) {
	match, index := q.captures.Next()
	if match == nil {
		return tree_sitter.QueryMatch{}, index, false
	}

	match.Captures = slices.Clone(match.Captures)
	return *match, index, true
}

func (q *queryCapturesIter) Next() (tree_sitter.QueryMatch, uint, bool) {
	if q.peeked != nil {
		peeked := q.peeked
		q.peeked = nil
		return peeked.match, peeked.index, peeked.ok
	}
	return q.next()
}

func (q *queryCapturesIter) Peek() (tree_sitter.QueryMatch, uint, bool) {
	if q.peeked == nil {
		match, index, ok := q.next()
		q.peeked = &peekedCapture{
			match: match,
			index: index,
			ok:    ok,
		}
	}

	return q.peeked.match, q.peeked.index, q.peeked.ok
}
