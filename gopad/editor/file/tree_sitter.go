package file

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	sitter "go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/ls"
)

type Tree struct {
	Tree     *sitter.Tree
	Language *Language
	SubTrees map[string]*Tree
	Ranges   []sitter.Range
}

func (t *Tree) Copy() *Tree {
	subTrees := make(map[string]*Tree, len(t.SubTrees))
	for key, subTree := range t.SubTrees {
		subTrees[key] = subTree.Copy()
	}

	return &Tree{
		Tree:     t.Tree.Copy(),
		Language: t.Language,
		SubTrees: subTrees,
	}
}

func (t *Tree) String() string {
	return t.string(0)
}

func (t *Tree) string(indent int) string {
	s := strings.Repeat(" ", indent) + fmt.Sprintf("Tree: %s\n", t.Language.Name)
	for _, subTree := range t.SubTrees {
		s += subTree.string(indent+1) + "\n"
	}
	return s
}

func (t *Tree) Print() string {
	s := fmt.Sprintf("Tree: %s\n", t.Language.Name)
	s += t.Tree.Print()

	for _, subTree := range t.SubTrees {
		s += "\n\n" + subTree.Print()
	}

	return s
}

func (t *Tree) FindTree(p buffer.Position) *Tree {
	if len(t.SubTrees) == 0 {
		return t
	}

	for _, subTree := range t.SubTrees {
		if subTree.Tree.RootNode().StartPoint().Row <= uint32(p.Row) && subTree.Tree.RootNode().EndPoint().Row >= uint32(p.Row) {
			return subTree.FindTree(p)
		}
	}
	return t
}

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tree, err := parseTree(ctx, f.buffer, f.tree, f.language, nil)
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

func parseTree(ctx context.Context, buff *buffer.Buffer, oldTree *Tree, language *Language, ranges []sitter.Range) (*Tree, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(language.Grammar.Language)
	if len(ranges) > 0 {
		parser.SetIncludedRanges(ranges)
	}

	var oldSitterTree *sitter.Tree
	if oldTree != nil {
		oldSitterTree = oldTree.Tree
	}

	tree, err := parser.ParseCtx(ctx, oldSitterTree, buff.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error parsing tree: %w", err)
	}

	query := language.Grammar.InjectionsQuery
	if query == nil {
		return &Tree{
			Tree:     tree,
			Language: language,
		}, nil
	}

	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(query.Query, tree.RootNode())

	// first we find all the ranges of the sub languages in the tree
	languageRanges := make(map[*Language][]sitter.Range)
	for {
		match, index, ok := queryCursor.NextCapture()
		if !ok {
			break
		}

		capture := match.Captures[index]
		if capture.Index != query.InjectionContentCaptureID {
			continue
		}

		subLanguage := getLanguageByMatch(language, match)
		if subLanguage == nil || subLanguage.Config.Grammar == nil {
			continue
		}

		start := capture.StartPoint()
		end := capture.EndPoint()
		languageRanges[subLanguage] = append(languageRanges[subLanguage], sitter.Range{
			StartPoint: capture.StartPoint(),
			EndPoint:   capture.EndPoint(),
			StartByte:  uint32(buff.ByteIndex(int(start.Row), int(start.Column))),
			EndByte:    uint32(buff.ByteIndex(int(end.Row), int(end.Column))),
		})
	}

	// then we parse the sub trees recursively using the ranges we found before
	subTrees := make(map[string]*Tree, len(ranges))
	for subLanguage, subRanges := range languageRanges {
		var oldSubTree *Tree
		if oldTree != nil {
			oldSubTree = oldTree.SubTrees[subLanguage.Name]
		}

		subTree, err := parseTree(ctx, buff, oldSubTree, subLanguage, subRanges)
		if err != nil {
			return nil, err
		}

		subTrees[subLanguage.Name] = subTree
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
			diagnostics = append(diagnostics, ls.Diagnostic{
				Type:   ls.DiagnosticTypeTreeSitter,
				Name:   tree.Language.Name,
				Source: "syntax",
				Range: buffer.Range{
					Start: buffer.Position{
						Row: int(node.StartPoint().Row),
						Col: int(node.StartPoint().Column),
					},
					End: buffer.Position{
						Row: int(node.EndPoint().Row),
						Col: int(node.EndPoint().Column),
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

type OutlineItem struct {
	Range buffer.Range
	Text  []OutlineItemChar
}

type OutlineItemChar struct {
	Char string
	Pos  *buffer.Position
}

type outlineBufferRange struct {
	r      byteRange
	isName bool
}

type byteRange struct {
	start int
	end   int
}

func (f *File) OutlineTree() []OutlineItem {
	if f.tree == nil || f.tree.Tree == nil || f.tree.Language.Grammar == nil || f.tree.Language.Grammar.OutlineQuery == nil {
		return nil
	}

	queryConfig := f.tree.Language.Grammar.OutlineQuery
	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(queryConfig.Query, f.tree.Tree.RootNode())

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
