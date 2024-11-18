package file

import (
	"context"
	"encoding/binary"
	"fmt"
	"iter"
	"log"
	"slices"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/tree-sitter/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/editor/buffer"
	"go.gopad.dev/gopad/internal/hash"
	"go.gopad.dev/gopad/internal/slotmap"
	"go.gopad.dev/gopad/internal/xbytes"
)

const (
	TreeSitterMatchLimit = 256
)

var injectionCallback = func(languageName string) *HighlightConfiguration {
	language := GetLanguage(languageName)
	if language == nil || language.Grammar == nil {
		return nil
	}

	return &language.Grammar.Highlight
}

type SyntaxEdit []*tree_sitter.InputEdit

func NewByteRange(startByte uint, endByte uint) ByteRange {
	return ByteRange{
		StartByte: startByte,
		EndByte:   endByte,
	}
}

func ByteRangeFromRange(r tree_sitter.Range) ByteRange {
	return ByteRange{
		StartByte: r.StartByte,
		EndByte:   r.EndByte,
	}
}

type ByteRange struct {
	StartByte uint
	EndByte   uint
}

func NewSyntax(language *Language, source []byte) (*Syntax, error) {
	layers, err := NewSyntaxLayers(source, language.Grammar.Highlight)
	if err != nil {
		return nil, fmt.Errorf("error creating syntax layers: %w", err)
	}

	syntax := &Syntax{
		Language: language,
		Layers:   layers,
	}

	if err = syntax.Parse(context.Background(), 0, source, nil); err != nil {
		return nil, fmt.Errorf("error parsing syntax: %w", err)
	}

	return syntax, nil
}

type Syntax struct {
	Rev      uint64
	Language *Language
	Layers   *SyntaxLayers
}

func (s *Syntax) Parse(ctx context.Context, newRev uint64, newSource []byte, edits []SyntaxEdit) error {
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
	return s.Parse(ctx, newRev, newBuf.Bytes(), []SyntaxEdit{edits})
}

type StyleSpan struct {
	Range tree_sitter.Range
	Style lipgloss.Style
}

func newParser() *parser {
	return &parser{
		parser:  tree_sitter.NewParser(),
		cursors: make([]*tree_sitter.QueryCursor, 0),
	}
}

type parser struct {
	parser  *tree_sitter.Parser
	cursors []*tree_sitter.QueryCursor
}

func (p *parser) pushCursor(cursor *tree_sitter.QueryCursor) {
	p.cursors = append(p.cursors, cursor)
}

func (p *parser) popCursor() *tree_sitter.QueryCursor {
	if len(p.cursors) == 0 {
		return nil
	}

	var cursor *tree_sitter.QueryCursor
	cursor, p.cursors = p.cursors[len(p.cursors)-1], p.cursors[:len(p.cursors)-1]
	return cursor
}

func NewSyntaxLayers(source []byte, config HighlightConfiguration) (*SyntaxLayers, error) {
	rootLayer := &LanguageLayer{
		Config: config,
		Tree:   nil,
		Ranges: []tree_sitter.Range{
			{
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
		},
		Depth:  0,
		parent: nil,
		rev:    0,
	}

	layers := slotmap.New[*LanguageLayer]()
	root := layers.Insert(rootLayer)

	syntax := &SyntaxLayers{
		parser: newParser(),
		layers: layers,
		root:   root,
	}

	if err := syntax.Update(context.Background(), 0, 0, source, nil); err != nil {
		return nil, err
	}

	return syntax, nil
}

type SyntaxLayers struct {
	parser *parser
	layers *slotmap.Map[*LanguageLayer]
	root   slotmap.LayerID
}

type injectionItem struct {
	config HighlightConfiguration
	ranges []tree_sitter.Range
}

type combinedInjectionItem struct {
	languageName    string
	nodes           []tree_sitter.Node
	includeChildren bool
}

func (s *SyntaxLayers) Update(ctx context.Context, currentRev uint64, newRev uint64, source []byte, syntaxEdits []SyntaxEdit) error {
	queue := make([]slotmap.LayerID, 0)
	queue = append(queue, s.root)

	edits := make([]*tree_sitter.InputEdit, 0)
	for _, edit := range syntaxEdits {
		edits = append(edits, edit...)
	}

	layersTable := make(map[uint64]slotmap.LayerID)
	layersHasher := hash.New()

	for layerID, layer := range s.layers.Iter() {
		if layer.Depth == 0 {
			continue
		}

		if len(edits) > 0 {
			for _, r := range layer.Ranges {
				// Roughly based on https://github.com/tree-sitter/tree-sitter/blob/ddeaa0c7f534268b35b4f6cb39b52df082754413/lib/src/subtree.c#L691-L720
				for _, edit := range edits {
					pureInsertion := edit.OldEndByte == edit.StartByte

					// if edit is after range, skip
					if edit.StartByte > r.EndByte {
						continue
					}

					// if edit is before range, shift entire range by len
					if edit.OldEndByte < r.StartByte {
						r.StartByte = edit.NewEndByte + (r.StartByte - edit.OldEndByte)
						r.StartPoint = pointAdd(edit.NewEndPosition, pointSub(r.StartPoint, edit.OldEndPosition))

						r.EndByte = edit.NewEndByte + (r.EndByte - edit.OldEndByte)
						r.EndPoint = pointAdd(edit.NewEndPosition, pointSub(r.EndPoint, edit.OldEndPosition))
					} else if edit.StartByte < r.StartByte {
						// if the edit starts in the space before and extends into the range
						r.StartByte = edit.NewEndByte
						r.StartPoint = edit.NewEndPosition

						r.EndByte = r.EndByte - edit.OldEndByte + edit.NewEndByte
						r.EndPoint = pointAdd(edit.NewEndPosition, pointSub(r.EndPoint, edit.OldEndPosition))
					} else if edit.StartByte == r.StartByte && pureInsertion {
						// If the edit is an insertion at the start of the tree, shift
						r.StartByte = edit.NewEndByte
						r.StartPoint = edit.NewEndPosition
					} else {
						r.EndByte = r.EndByte - edit.OldEndByte + edit.NewEndByte
						r.EndPoint = pointAdd(edit.NewEndPosition, pointSub(r.EndPoint, edit.OldEndPosition))
					}
				}
			}
		}

		h := layersHasher.HashOne(layer)
		layersTable[h] = layerID
	}

	cursor := s.parser.popCursor()
	if cursor == nil {
		cursor = tree_sitter.NewQueryCursor()
	}
	// cursor.SetByteRange(0, ^uint(0))
	// cursor.SetMatchLimit(TreeSitterMatchLimit)

	touched := map[slotmap.LayerID]struct{}{}

	for len(queue) > 0 {
		var layerID slotmap.LayerID
		layerID, queue = queue[0], queue[1:]

		// Mark the layer as touched
		touched[layerID] = struct{}{}

		layer := s.layers.Get(layerID)

		hadEdits := layer.rev == currentRev && len(syntaxEdits) > 0
		// If a tree already exists, notify it of changes.
		if hadEdits {
			if layer.Tree != nil {
				for _, edit := range edits {
					layer.Tree.Edit(edit)
				}
			}
		}

		if err := layer.parse(ctx, s.parser.parser, source, hadEdits); err != nil {
			return err
		}
		layer.rev = newRev

		// Process injections.
		if layer.Tree != nil {
			matches := cursor.Captures(layer.Config.InjectionsQuery, layer.Tree.RootNode(), source)
			combinedInjections := make([]combinedInjectionItem, len(layer.Config.CombinedInjectionsPatterns))
			var injections []injectionItem
			var lastInjectionEnd uint
			for {
				match, _ := matches.Next()
				if match == nil {
					break
				}

				languageName, contentNode, includeChildren := layer.Config.injectionForMatch(layer.Config.InjectionsQuery, match, source)

				// in case this is a combined injection save it for more processing later
				index := slices.IndexFunc(layer.Config.CombinedInjectionsPatterns, func(u uint) bool {
					return match.PatternIndex == u
				})
				if index > -1 {
					if languageName == "" {
						combinedInjections[index].languageName = languageName
					}
					if contentNode != nil {
						combinedInjections[index].nodes = append(combinedInjections[index].nodes, *contentNode)
					}
					combinedInjections[index].includeChildren = includeChildren
					continue
				}

				// Explicitly remove this match so that none of its other captures will remain
				// in the stream of captures.
				match.Remove()

				// If a language is found with the given name, then add a new language layer
				// to the highlighted document.
				if languageName != "" && contentNode != nil {
					nextConfig := injectionCallback(languageName)
					if nextConfig != nil {
						nextRanges := intersectRanges(layer.Ranges, []tree_sitter.Node{*contentNode}, includeChildren)
						if len(nextRanges) > 0 {
							if contentNode.StartByte() < lastInjectionEnd {
								continue
							}

							lastInjectionEnd = contentNode.EndByte()

							injections = append(injections, injectionItem{
								config: *nextConfig,
								ranges: nextRanges,
							})
						}
					} else {
						log.Printf("no language found for injection: %s", languageName)
					}
				}
			}

			for _, combinedInjection := range combinedInjections {
				if combinedInjection.languageName != "" && len(combinedInjection.nodes) > 0 {
					nextConfig := injectionCallback(combinedInjection.languageName)
					if nextConfig != nil {
						nextRanges := intersectRanges(layer.Ranges, combinedInjection.nodes, combinedInjection.includeChildren)
						if len(nextRanges) > 0 {
							injections = append(injections, injectionItem{
								config: *nextConfig,
								ranges: nextRanges,
							})
						}
					} else {
						log.Printf("no language found for combined injection: %s", combinedInjection.languageName)
					}
				}
			}

			depth := layer.Depth + 1
			for _, injection := range injections {
				newLayer := &LanguageLayer{
					Config: injection.config,
					Tree:   nil,
					Ranges: injection.ranges,
					Depth:  depth,
					parent: &layerID,
					rev:    0,
				}

				newLayerHash := layersHasher.HashOne(newLayer)

				nextLayerID, ok := layersTable[newLayerHash]
				if ok {
					oldLayer := s.layers.Get(nextLayerID)
					if !oldLayer.Equals(newLayer) {
						ok = false
					}
				}
				if !ok {
					nextLayerID = s.layers.Insert(newLayer)
				}

				queue = append(queue, nextLayerID)
			}
		}
	}

	s.parser.pushCursor(cursor)

	s.layers.Retain(func(id slotmap.LayerID, _ *LanguageLayer) bool {
		_, ok := touched[id]
		return ok
	})

	return nil
}

func (s *SyntaxLayers) Tree() *tree_sitter.Tree {
	return s.layers.Get(s.root).Tree
}

func (s *SyntaxLayers) HighlightIter(ctx context.Context, source []byte, r *ByteRange) iter.Seq2[HighlightEvent, error] {
	var layers []*highlightIterLayer
	for _, layer := range s.layers.Map() {
		// Reuse a cursor from the pool if available.
		cursor := s.parser.popCursor()
		if cursor == nil {
			cursor = tree_sitter.NewQueryCursor()
		}
		// if reusing cursors & no range this resets to whole range
		if r == nil {
			r = &ByteRange{
				StartByte: 0,
				EndByte:   ^uint(0),
			}
		}

		// cursor.SetByteRange(r.StartByte, r.EndByte)
		// cursor.SetMatchLimit(TreeSitterMatchLimit)

		captures := make([]queryCapture, 0)
		queryCaptures := cursor.Captures(layer.Config.Query, layer.Tree.RootNode(), source)
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

		if len(captures) == 0 {
			continue
		}

		layers = append(layers, &highlightIterLayer{
			Tree:              nil,
			Cursor:            cursor,
			Config:            layer.Config,
			HighlightEndStack: nil,
			ScopeStack: []LocalScope{
				{
					Inherits: false,
					Range: ByteRange{
						StartByte: 0,
						EndByte:   ^uint(0),
					},
					LocalDefs: nil,
				},
			},
			Captures: captures,
			Depth:    layer.Depth,
		})
	}

	hIter := &highlightIter{
		Ctx:                ctx,
		Source:             source,
		ByteOffset:         0,
		Layers:             layers,
		NextEvent:          nil,
		LastHighlightRange: nil,
		Syntax:             s,
	}
	hIter.sortLayers()

	return func(yield func(HighlightEvent, error) bool) {
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

func pointAdd(a tree_sitter.Point, b tree_sitter.Point) tree_sitter.Point {
	if b.Row > 0 {
		return tree_sitter.NewPoint(a.Row+b.Row, b.Column)
	}
	return tree_sitter.NewPoint(0, a.Column+b.Column)
}

func pointSub(a tree_sitter.Point, b tree_sitter.Point) tree_sitter.Point {
	if a.Row > b.Row {
		return tree_sitter.NewPoint(a.Row-b.Row, a.Column)
	}
	return tree_sitter.NewPoint(0, a.Column-b.Column)
}

type LanguageLayer struct {
	Config HighlightConfiguration
	Tree   *tree_sitter.Tree
	Ranges []tree_sitter.Range
	Depth  int
	parent *slotmap.LayerID
	rev    uint64
}

func (l *LanguageLayer) Hash(h *hash.Hasher) {
	_ = binary.Write(h, binary.LittleEndian, l.Depth)
	_ = binary.Write(h, binary.LittleEndian, l.Config.LanguageName)
	_ = binary.Write(h, binary.LittleEndian, l.Ranges)
}

func (l *LanguageLayer) parse(ctx context.Context, parser *tree_sitter.Parser, source []byte, hasEdits bool) error {
	if err := parser.SetIncludedRanges(l.Ranges); err != nil {
		return err
	}
	if err := parser.SetLanguage(l.Config.Language); err != nil {
		return err
	}

	var oldTree *tree_sitter.Tree
	if hasEdits && l.Tree != nil {
		oldTree = l.Tree
	}

	tree := parser.ParseCtx(ctx, source, oldTree)
	if tree == nil {
		return ctx.Err()
	}

	l.Tree = tree

	return nil
}

func (l *LanguageLayer) Equals(layer *LanguageLayer) bool {
	return l.Depth == layer.Depth && l.Config.LanguageName == layer.Config.LanguageName && equalRanges(l.Ranges, layer.Ranges)
}

func generateEdits(oldBuf buffer.Buffer, changeSet ChangeSet) []*tree_sitter.InputEdit {
	var oldPos int
	var edits []*tree_sitter.InputEdit

	if changeSet.IsEmpty() {
		return edits
	}

	changes := changeSet.Changes

	for len(changes) > 0 {
		var change Operation
		change, changes = changes[0], changes[1:]

		var changeLen int
		switch c := change.(type) {
		case Move:
			changeLen = c.N
		case Delete:
			changeLen = c.N
		}

		oldEnd := oldPos + changeLen

		switch c := change.(type) {
		case Delete:
			startByte := oldBuf.ByteIndex(oldPos)
			startPosition := oldBuf.Position(oldPos)

			oldEndByte := oldBuf.ByteIndex(oldEnd)
			oldEndPosition := oldBuf.Position(oldEnd)

			edits = append(edits, &tree_sitter.InputEdit{
				StartByte:      uint(startByte),
				OldEndByte:     uint(oldEndByte),
				NewEndByte:     uint(startByte),
				StartPosition:  startPosition.ToTreeSitter(),
				OldEndPosition: oldEndPosition.ToTreeSitter(),
				NewEndPosition: startPosition.ToTreeSitter(),
			})
		case Insert:
			startByte := oldBuf.ByteIndex(oldPos)
			startPosition := oldBuf.Position(oldPos)

			if len(changes) > 0 {
				nextChange, ok := changes[0].(Delete)
				if ok {
					oldEnd = oldPos + nextChange.N
					oldEndByte := oldBuf.ByteIndex(oldEnd)
					oldEndPosition := oldBuf.Position(oldEnd)

					changes = changes[1:]

					edits = append(edits, &tree_sitter.InputEdit{
						StartByte:      uint(startByte),
						OldEndByte:     uint(oldEndByte),
						NewEndByte:     uint(startByte + len(c.Text)),
						StartPosition:  startPosition.ToTreeSitter(),
						OldEndPosition: oldEndPosition.ToTreeSitter(),
						NewEndPosition: traverseBytes(startPosition, c.Text).ToTreeSitter(),
					})
					continue
				}
			}

			edits = append(edits, &tree_sitter.InputEdit{
				StartByte:      uint(startByte),
				OldEndByte:     uint(startByte),
				NewEndByte:     uint(startByte + len(c.Text)),
				StartPosition:  startPosition.ToTreeSitter(),
				OldEndPosition: startPosition.ToTreeSitter(),
				NewEndPosition: traverseBytes(startPosition, c.Text).ToTreeSitter(),
			})
		}

		oldPos = oldEnd
	}

	return edits
}

func traverseBytes(p buffer.Point, text []byte) buffer.Point {
	row, col := p.Point()

	runes := xbytes.Runes(text)

	for i, r := range runes {
		if r == '\n' || r == '\r' {
			if !(len(runes) > i+1 && runes[i+1] == '\n') {
				row++
				col = 0
			}
		}
		col++
	}

	return buffer.NewPoint(row, col)
}

func equalRanges(a []tree_sitter.Range, b []tree_sitter.Range) bool {
	if len(a) != len(b) {
		return false
	}

	for i, r := range a {
		if r != b[i] {
			return false
		}
	}

	return true
}
