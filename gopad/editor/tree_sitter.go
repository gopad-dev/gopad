package editor

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	sitter "go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/ls"
)

func (f *File) InitTree() error {
	if f.language == nil || f.language.Grammar == nil {
		return nil
	}

	return f.updateTree()
}

func (f *File) UpdateTree(edit sitter.EditInput) error {
	if f.language == nil || f.language.Grammar == nil {
		return nil
	}

	editTree(f.tree, edit)

	return f.updateTree()
}

func (f *File) updateTree() error {
	ctx, cancel := context.WithTimeout(context.Background(), f.language.Grammar.ParseTimeout)
	defer cancel()

	tree, err := parseTree(ctx, f.buffer.Bytes(), f.tree, *f.language, 0, 0)
	if err != nil {
		return err
	}

	f.tree = tree

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		f.ValidateTree()
	}()

	go func() {
		defer wg.Done()
		f.HighlightTree()
	}()

	wg.Wait()

	return nil
}

func parseTree(ctx context.Context, content []byte, oldTree *Tree, language Language, rowOffset int, colOffset int) (*Tree, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(language.Grammar.Language)

	var oldSitterTree *sitter.Tree
	if oldTree != nil {
		oldSitterTree = oldTree.Tree
	}

	tree, err := parser.ParseCtx(ctx, oldSitterTree, content)
	if err != nil {
		return nil, fmt.Errorf("error parsing tree: %w", err)
	}

	query := language.Grammar.injectionsQuery
	if query == nil {
		return &Tree{
			Tree:     tree,
			Language: language,
		}, nil
	}

	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(query, tree.RootNode())

	var subTrees []*Tree
	var subTreeIndex int
	for {
		match, index, ok := queryCursor.NextCapture()
		if !ok {
			break
		}

		capture := match.Captures[index]

		node := capture.Node
		if query.CaptureNameForID(capture.Index) == "injection.content" {
			subLanguage := getLanguageByMatch(&language, match)
			if subLanguage == nil || subLanguage.Config.Grammar == nil {
				continue
			}

			newRowOffset := int(node.StartPoint().Row) + rowOffset
			newColOffset := int(node.StartPoint().Column) + colOffset

			var oldSubTree *Tree
			// TODO: figure out how to match old sub trees with new ones
			if oldTree != nil && subTreeIndex < len(oldTree.SubTrees) {
				subTree := oldTree.SubTrees[subTreeIndex]
				if subTree.Language.Name == subLanguage.Name {
					oldSubTree = subTree
				}
			}

			subTree, err := parseTree(ctx, content[node.StartByte():node.EndByte()], oldSubTree, *subLanguage, newRowOffset, newColOffset)
			if err != nil {
				return nil, err
			}
			subTree.OffsetRow = newRowOffset
			subTree.OffsetCol = newColOffset

			subTrees = append(subTrees, subTree)
			subTreeIndex++
		}
	}

	return &Tree{
		Tree:     tree,
		SubTrees: subTrees,
		Language: language,
	}, nil
}

func getLanguageByMatch(parentLanguage *Language, match *sitter.QueryMatch) *Language {
	if subLanguageName, ok := match.Properties["injection.language"]; ok {
		if language := GetLanguage(subLanguageName); language != nil {
			return language
		}
	}

	if subFileName, ok := match.Properties["injection.filename"]; ok {
		if language := GetLanguageByFilename(subFileName); language != nil {
			return language
		}
	}

	if subMIMEType, ok := match.Properties["injection.mimetype"]; ok {
		if language := GetLanguageByMIMEType(subMIMEType); language != nil {
			return language
		}
	}

	if _, ok := match.Properties["injection.parent"]; ok {
		if parentLanguage != nil {
			return parentLanguage
		}
	}

	return nil
}

func editTree(tree *Tree, edit sitter.EditInput) {
	tree.Tree.Edit(edit)

	for _, subTree := range tree.SubTrees {
		editTree(subTree, edit)
	}
}

type Tree struct {
	Tree      *sitter.Tree
	Language  Language
	SubTrees  []*Tree
	OffsetRow int
	OffsetCol int
}

func (t *Tree) Copy() *Tree {
	var subTrees []*Tree
	for _, subTree := range t.SubTrees {
		subTrees = append(subTrees, subTree.Copy())
	}

	return &Tree{
		Tree:      t.Tree.Copy(),
		Language:  t.Language,
		SubTrees:  subTrees,
		OffsetRow: t.OffsetRow,
		OffsetCol: t.OffsetCol,
	}
}

func (t *Tree) String() string {
	return t.string(0)
}

func (t *Tree) string(indent int) string {
	s := strings.Repeat(" ", indent) + fmt.Sprintf("Tree: %s: %d|%d \n", t.Language.Name, t.OffsetRow, t.OffsetCol)
	for _, subTree := range t.SubTrees {
		s += subTree.string(indent+1) + "\n"
	}
	return s
}

func (t *Tree) Print() string {
	s := fmt.Sprintf("Tree: %s: %d|%d\n", t.Language.Name, t.OffsetRow, t.OffsetCol)
	s += t.Tree.Print()

	for _, subTree := range t.SubTrees {
		s += "\n\n" + subTree.Print()
	}

	return s
}

func (t *Tree) Range() buffer.Range {
	start := t.Tree.RootNode().StartPoint()
	end := t.Tree.RootNode().EndPoint()

	endCol := int(end.Column)
	if end.Row == start.Row {
		endCol += t.OffsetCol
	}
	return buffer.Range{
		Start: buffer.Position{
			Row: int(start.Row) + t.OffsetRow,
			Col: int(start.Column) + t.OffsetCol,
		},
		End: buffer.Position{
			Row: int(end.Row) + t.OffsetRow,
			Col: endCol,
		},
	}
}

func (t *Tree) FindTree(p buffer.Position) *Tree {
	if len(t.SubTrees) == 0 {
		return t
	}

	for _, subTree := range t.SubTrees {
		if subTree.Range().Contains(p) {
			return subTree.FindTree(p)
		}
	}

	return t
}

func (f *File) ValidateTree() {
	if f.tree == nil || f.tree.Tree == nil {
		return
	}
	version := f.Version()

	diagnostics := validateTree(f.tree.Copy())

	f.SetDiagnostic(ls.DiagnosticTypeTreeSitter, version, diagnostics)
}

func validateTree(tree *Tree) []ls.Diagnostic {
	iter := sitter.NewIterator(tree.Tree.RootNode(), sitter.BFSMode)
	var diagnostics []ls.Diagnostic
	for {
		node, err := iter.Next()
		if err != nil {
			break
		}

		if node.IsError() {
			startRow := int(node.StartPoint().Row) + tree.OffsetRow
			startCol := int(node.StartPoint().Column) + tree.OffsetCol

			endRow := int(node.EndPoint().Row) + tree.OffsetRow
			endCol := int(node.EndPoint().Column)
			if endRow == startRow {
				endCol += tree.OffsetCol
			}

			diagnostics = append(diagnostics, ls.Diagnostic{
				Type:   ls.DiagnosticTypeTreeSitter,
				Name:   tree.Language.Name,
				Source: "syntax",
				Range: buffer.Range{
					Start: buffer.Position{
						Row: startRow,
						Col: startCol,
					},
					End: buffer.Position{
						Row: endRow,
						Col: endCol,
					},
				},
				Severity: ls.DiagnosticSeverityError,
				Message:  "Syntax error",
				Priority: 100,
			})
		}
	}

	for _, subTree := range tree.SubTrees {
		diagnostics = append(diagnostics, validateTree(subTree)...)
	}

	return diagnostics
}

type Match struct {
	Range    buffer.Range
	Type     string
	Priority int
	Source   string
}

func (f *File) HighlightTree() {
	if f.tree == nil || f.tree.Tree == nil {
		return
	}
	version := f.Version()

	matches := highlightTree(f.tree.Copy())
	//slices.SortFunc(matches, func(a, b Match) int {
	//	return b.Priority - a.Priority
	//})

	f.SetMatches(version, matches)
}

func highlightTree(tree *Tree) []Match {
	query := tree.Language.Grammar.highlightsQuery
	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(query, tree.Tree.RootNode())

	var matches []Match
	for {
		match, index, ok := queryCursor.NextCapture()
		if !ok {
			break
		}

		capture := match.Captures[index]

		realRow := int(capture.StartPoint().Row) + tree.OffsetRow
		realCol := int(capture.StartPoint().Column)
		realEndRow := int(capture.EndPoint().Row) + tree.OffsetRow
		realEndCol := int(capture.EndPoint().Column)
		if realRow == tree.OffsetRow {
			realCol += tree.OffsetCol
			realEndCol += tree.OffsetCol
		}

		priority := 100
		if priorityStr, ok := match.Properties["priority"]; ok {
			if parsedPriority, err := strconv.Atoi(priorityStr); err == nil {
				priority = parsedPriority
			}
		}

		matches = append(matches, Match{
			Range: buffer.Range{
				Start: buffer.Position{Row: realRow, Col: realCol},
				End:   buffer.Position{Row: realEndRow, Col: realEndCol - 1}, // -1 to exclude the last character idk why this is like this tbh
			},
			Type:     query.CaptureNameForID(capture.Index),
			Priority: priority,
			Source:   tree.Language.Name,
		})
	}

	for _, subTree := range tree.SubTrees {
		subMatches := highlightTree(subTree)
		matches = append(matches, subMatches...)
	}

	return matches
}

type OutlineItem struct {
	Range buffer.Range
	Text  []OutlineItemChar
}

type OutlineItemChar struct {
	Char string
	Pos  *buffer.Position
}

func (f *File) OutlineTree() []OutlineItem {
	if f.tree == nil || f.tree.Tree == nil || f.tree.Language.Grammar == nil || f.tree.Language.Grammar.outlineQuery == nil {
		return nil
	}

	return f.outlineTree(f.tree)
}

type outlineBufferRange struct {
	r      byteRange
	isName bool
}

type byteRange struct {
	start int
	end   int
}

func (f *File) outlineTree(tree *Tree) []OutlineItem {
	queryConfig := tree.Language.Grammar.OutlineQuery()
	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(queryConfig.Query, tree.Tree.RootNode())

	var items []OutlineItem
	for {
		match, ok := queryCursor.NextMatch()
		if !ok {
			break
		}

		itemNodeIndex := slices.IndexFunc(match.Captures, func(capture sitter.QueryCapture) bool {
			return capture.Index == queryConfig.ItemCaptureID
		})
		if itemNodeIndex < 0 {
			continue
		}

		itemCapture := match.Captures[itemNodeIndex]
		itemRange := buffer.Range{
			Start: buffer.Position{
				Row: int(itemCapture.Node.StartPoint().Row),
				Col: int(itemCapture.Node.StartPoint().Column),
			},
			End: buffer.Position{
				Row: int(itemCapture.Node.EndPoint().Row),
				Col: int(itemCapture.Node.EndPoint().Column),
			},
		}

		var bufferRanges []outlineBufferRange
		for _, capture := range match.Captures {
			var isName bool
			if capture.Index == queryConfig.NameCaptureID {
				isName = true
			} else if (queryConfig.ContextCaptureID != nil && capture.Index == *queryConfig.ContextCaptureID) || (queryConfig.ExtraContextCaptureID != nil && capture.Index == *queryConfig.ExtraContextCaptureID) {
				isName = false
			} else {
				continue
			}

			r := byteRange{
				start: int(capture.Node.StartByte()),
				end:   int(capture.Node.EndByte()),
			}
			start := capture.Node.StartPoint()

			if capture.Node.EndPoint().Row > start.Row {
				r.end = r.start + f.buffer.LineLen(int(start.Row)) - int(start.Column)
			}

			bufferRanges = append(bufferRanges, outlineBufferRange{
				r:      r,
				isName: isName,
			})
		}

		if len(bufferRanges) == 0 {
			continue
		}

		var chars []OutlineItemChar
		var nameRanges []byteRange
		var lastBufferRangeEnd int
		for _, bufferRange := range bufferRanges {
			if len(chars) != 0 && bufferRange.r.start > lastBufferRangeEnd {
				chars = append(chars, OutlineItemChar{
					Char: " ",
				})
			}

			lastBufferRangeEnd = bufferRange.r.end
			if bufferRange.isName {
				start := len(chars)
				end := start + bufferRange.r.end - bufferRange.r.start

				if len(nameRanges) != 0 {
					start -= 1
				}

				nameRanges = append(nameRanges, byteRange{
					start: start,
					end:   end,
				})
			}

			start := f.buffer.Position(bufferRange.r.start)
			end := f.buffer.Position(bufferRange.r.end)

			for i := start.Row; i <= end.Row; i++ {
				line := f.buffer.Line(i)
				var colOffset int
				if i == start.Row && i == end.Row {
					line = line.CutRange(start.Col, end.Col)
					colOffset = start.Col
				} else if i == start.Row {
					line = line.CutStart(start.Col)
					colOffset = start.Col
				} else if i == end.Row {
					line = line.CutEnd(end.Col)
				}

				for j, char := range line.RuneStrings() {
					chars = append(chars, OutlineItemChar{
						Char: char,
						Pos: &buffer.Position{
							Row: i,
							Col: j + colOffset,
						},
					})
				}
			}
		}

		items = append(items, OutlineItem{
			Range: itemRange,
			Text:  chars,
		})
	}

	return items
}

type Local struct {
	Name       string
	Properties map[string]string
}
