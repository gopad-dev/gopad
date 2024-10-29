package file

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea/v2"
	"go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/buffer"
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
	s := fmt.Sprintf("grammar: %s\n", t.Language.Name)
	s += t.Tree.Print()

	for _, subTree := range t.SubTrees {
		s += "\n\n" + subTree.Print()
	}

	return s
}

func (t *Tree) FindTree(p buffer.Point) *Tree {
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

func (f *File) InitTree() tea.Cmd {
	if f.Language == nil || f.Language.Grammar == nil {
		return nil
	}

	if err := f.updateTree(); err != nil {
		return notifications.Addf("Error updating tree sitter tree: %s", err.Error())
	}

	name := f.Buffer.Name()
	version := f.Buffer.Version()

	return tea.Batch(
		HighlightTree(name, version, f.Tree.Copy(), f.Buffer.LinesLen()),
		ValidateTree(name, version, f.Tree.Copy()),
	)
}

func (f *File) UpdateTree(edit sitter.EditInput) tea.Cmd {
	now := time.Now()
	defer func() {
		log.Println("Update tree time: ", time.Since(now))
	}()

	if f.Language == nil || f.Language.Grammar == nil {
		return nil
	}

	editTree(f.Tree, edit)

	if err := f.updateTree(); err != nil {
		return notifications.Addf("Error updating tree sitter tree: %s", err.Error())
	}

	name := f.Buffer.Name()
	version := f.Buffer.Version()

	return tea.Batch(
		HighlightTree(name, version, f.Tree.Copy(), f.Buffer.LinesLen()),
		ValidateTree(name, version, f.Tree.Copy()),
	)
}

func (f *File) updateTree() error {
	now := time.Now()
	defer func() {
		log.Println("update tree time: ", time.Since(now))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tree, err := parseTree(ctx, f.Buffer, f.Tree, f.Language, nil)
	if err != nil {
		return err
	}

	f.Tree = tree
	return nil
}

func parseTree(ctx context.Context, buff buffer.Buffer, oldTree *Tree, language *Language, ranges []sitter.Range) (*Tree, error) {
	now := time.Now()
	defer func() {
		log.Println("parse tree time: ", time.Since(now))
	}()

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
			StartByte: uint32(buff.ByteIndex(buffer.Point{
				Row: int(start.Row),
				Col: int(start.Column),
			})),
			EndByte: uint32(buff.ByteIndex(buffer.Point{
				Row: int(end.Row),
				Col: int(end.Column),
			})),
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
