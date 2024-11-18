package file

import (
	"context"
	"fmt"
	"iter"
	"log"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/tree-sitter/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/buffer"
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

type Highlight uint

type HighlightEvent interface {
	highlightEvent()
}

type HighlightEventSource struct {
	StartByte uint
	EndByte   uint
}

func (HighlightEventSource) highlightEvent() {}

type HighlightEventStart struct {
	Highlight    Highlight
	LanguageName string
}

func (HighlightEventStart) highlightEvent() {}

type HighlightEventEnd struct{}

func (HighlightEventEnd) highlightEvent() {}

func NewHighlightConfig(language *tree_sitter.Language, languageName string, highlightsQuery []byte, injectionQuery []byte, localsQuery []byte) (*HighlightConfiguration, error) {
	var querySource []byte
	querySource = append(querySource, localsQuery...)
	highlightsQueryOffset := uint(len(querySource))
	querySource = append(querySource, highlightsQuery...)

	query, err := tree_sitter.NewQuery(language, string(querySource))
	if err != nil {
		return nil, fmt.Errorf("error creating query: %w", err)
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
		return nil, fmt.Errorf("error creating combined injections query: %w", err)
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
		injectionContentCaptureIndex  *uint
		injectionLanguageCaptureIndex *uint
		localDefCaptureIndex          *uint
		localDefValueCaptureIndex     *uint
		localRefCaptureIndex          *uint
		localScopeCaptureIndex        *uint
	)

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
		}
	}

	highlightIndices := make([]*Highlight, len(query.CaptureNames()))
	return &HighlightConfiguration{
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
		captureParts := strings.Split(captureName, ".")

		var bestIndex *Highlight
		var bestMatchLen int
		for j, recognizedName := range recognizedNames {
			var matchLen int
			matches := true
			for _, part := range strings.Split(recognizedName, ".") {
				matchLen++
				if !slices.Contains(captureParts, part) {
					matches = false
					break
				}
			}
			if matches && matchLen > bestMatchLen {
				index := Highlight(j)
				bestIndex = &index
				bestMatchLen = matchLen
			}
		}
		highlightIndices[i] = bestIndex
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
	Range     ByteRange
	Highlight *Highlight
}

type LocalScope struct {
	Inherits  bool
	Range     ByteRange
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
	NextEvent          HighlightEvent
	LastHighlightRange *iterRange
	Syntax             *SyntaxLayers
}

func (h *highlightIter) emitEvent(offset uint, event HighlightEvent) (HighlightEvent, error) {
	var result HighlightEvent
	if h.ByteOffset < offset {
		result = HighlightEventSource{
			StartByte: h.ByteOffset,
			EndByte:   offset,
		}
		h.ByteOffset = offset
		h.NextEvent = event
	} else {
		result = event
	}
	h.sortLayers()
	return result, nil
}

func (h *highlightIter) sortLayers() {
	for len(h.Layers) > 0 {
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
				} else {
					layer := h.Layers[i+1]
					h.Layers = append(h.Layers[:i], h.Layers[i+1:]...)
					h.Syntax.parser.pushCursor(layer.Cursor)
				}
				break
			}
			if i > 0 {
				h.Layers = append(h.Layers[:i], append([]*highlightIterLayer{h.Layers[0]}, h.Layers[i:]...)...)
			}
			break
		} else {
			layer := h.Layers[0]
			h.Layers = h.Layers[1:]
			h.Syntax.parser.pushCursor(layer.Cursor)
		}
	}
}

func (h *highlightIter) next() (HighlightEvent, error) {
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
					return h.emitEvent(endByte, HighlightEventEnd{})
				}
			}
		} else {
			// If there are no more captures, then emit any remaining highlight end events.
			// And if there are none of those, then just advance to the end of the document.
			if len(layer.HighlightEndStack) > 0 {
				endByte := layer.HighlightEndStack[len(layer.HighlightEndStack)-1]
				layer.HighlightEndStack = layer.HighlightEndStack[:len(layer.HighlightEndStack)-1]
				return h.emitEvent(endByte, HighlightEventEnd{})
			}
			return h.emitEvent(uint(len(h.Source)), nil)
		}

		match := layer.Captures[0]
		layer.Captures = layer.Captures[1:]
		capture := match.Match.Captures[match.Index]

		// Remove from the local scope stack any local scopes that have already ended.
		for r.StartByte > layer.ScopeStack[len(layer.ScopeStack)-1].Range.EndByte {
			layer.ScopeStack = layer.ScopeStack[:len(layer.ScopeStack)-1]
		}

		// If this capture is for tracking local variables, then process the
		// local variable info.
		var referenceHighlight *Highlight
		var definitionHighlight *Highlight
		for match.Match.PatternIndex < layer.Config.HighlightsPatternIndex {
			// If the node represents a local scope, push a new local scope onto
			// the scope stack.
			if layer.Config.LocalScopeCaptureIndex != nil && uint(capture.Index) == *layer.Config.LocalScopeCaptureIndex {
				definitionHighlight = nil
				scope := LocalScope{
					Inherits:  true,
					Range:     ByteRangeFromRange(r),
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

					scope.LocalDefs = append(scope.LocalDefs, LocalDef{
						Name:      name,
						Range:     ByteRangeFromRange(r),
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
						var highlight *Highlight
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
			definitionHighlight = currentHighlight
		}

		// Emit a scope start event and push the node's end position to the stack.
		highlight := referenceHighlight
		if highlight == nil {
			highlight = currentHighlight
		}
		if highlight != nil {
			h.LastHighlightRange = &iterRange{
				Start: r.StartByte,
				End:   r.EndByte,
				Depth: layer.Depth,
			}
			layer.HighlightEndStack = append(layer.HighlightEndStack, r.EndByte)
			return h.emitEvent(r.StartByte, HighlightEventStart{
				Highlight:    *highlight,
				LanguageName: layer.Config.LanguageName,
			})
		}

		h.sortLayers()
	}
}

func intersectRanges(parentRanges []tree_sitter.Range, nodes []tree_sitter.Node, includesChildren bool) []tree_sitter.Range {
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

func (c HighlightConfiguration) injectionForMatch(query *tree_sitter.Query, match *tree_sitter.QueryMatch, source []byte) (string, *tree_sitter.Node, bool) {
	if c.InjectionContentCaptureIndex == nil || c.InjectionLanguageCaptureIndex == nil {
		return "", nil, false
	}
	contentCaptureIndex := *c.InjectionContentCaptureIndex
	languageCaptureIndex := *c.InjectionLanguageCaptureIndex

	var languageName string
	var contentNode *tree_sitter.Node

	for _, capture := range match.Captures {
		index := capture.Index
		if uint(index) == languageCaptureIndex {
			languageName = capture.Node.Utf8Text(source)
		} else if uint(index) == contentCaptureIndex {
			contentNode = &capture.Node
		}
	}

	var includeChildren bool
	for _, property := range query.PropertySettings(match.PatternIndex) {
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

type queryCapture struct {
	Match *tree_sitter.QueryMatch
	Index uint
}

type highlightIterLayer struct {
	Tree              *tree_sitter.Tree
	Cursor            *tree_sitter.QueryCursor
	Config            HighlightConfiguration
	HighlightEndStack []uint
	ScopeStack        []LocalScope
	Captures          []queryCapture
	Depth             int
}

type sortKeyResult struct {
	position uint
	start    bool
	depth    int
}

func (h *highlightIterLayer) sortKey() *sortKeyResult {
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

type CharStyle struct {
	Style        lipgloss.Style
	StyleName    string
	LanguageName string
	End          int
}

func newStyleIter(highlightIter iter.Seq2[HighlightEvent, error], buf buffer.Buffer, styles *config.CodeStyles) iter.Seq[CharStyle] {
	iterator := styleIterator{
		activeHighlights: nil,
		highlightIter:    highlightIter,
		buf:              buf,
		styles:           styles,
	}

	return iterator.iter()
}

type highlightStyle struct {
	highlight    Highlight
	languageName string
}

type styleIterator struct {
	activeHighlights []highlightStyle
	highlightIter    iter.Seq2[HighlightEvent, error]
	buf              buffer.Buffer
	styles           *config.CodeStyles
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
				i.activeHighlights = append(i.activeHighlights, highlightStyle{
					highlight:    event.Highlight,
					languageName: event.LanguageName,
				})
			case HighlightEventEnd:
				i.activeHighlights = i.activeHighlights[:len(i.activeHighlights)-1]
			case HighlightEventSource:
				ch := CharStyle{
					// End:       i.buf.RuneIndex(int(event.EndByte)), TODO: RuneIndex seems to be broken, investigate
					End: int(event.EndByte),
				}

				if len(i.activeHighlights) > 0 {
					highlight := i.activeHighlights[len(i.activeHighlights)-1]

					ch.Style = i.styles.Highlight(int(highlight.highlight), highlight.languageName)
					ch.StyleName = i.styles.Scope(int(highlight.highlight))
					ch.LanguageName = highlight.languageName
				}

				if ok := yield(ch); !ok {
					return
				}
			}
		}
	}
}
