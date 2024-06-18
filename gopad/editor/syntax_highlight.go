package editor

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	sitter "go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
)

func (f *File) SetMatches(version int32, matches []Match) {
	if version < f.matchesVersion {
		return
	}
	if version > f.matchesVersion {
		f.matches = matches
		f.matchesVersion = version
		return
	}
	f.matches = append(f.matches, matches...)
}

func (f *File) Matches() []Match {
	return f.matches
}

func (f *File) MatchesForLineCol(row int, col int) []Match {
	pos := buffer.Position{Row: row, Col: col}

	var matches []Match
	for _, match := range f.matches {
		if match.Range.Contains(pos) {
			matches = append(matches, match)
		}
	}
	return matches
}

func (f *File) HighestMatchStyle(style lipgloss.Style, row int, col int) lipgloss.Style {
	var currentStyle *lipgloss.Style
	var referenceStyle *lipgloss.Style
	for _, match := range f.MatchesForLineCol(row, col) {
		if match.ReferenceType != "" {
			newStyle := getMatchingStyle(match.ReferenceType, f.language.Name)
			if newStyle != nil {
				referenceStyle = newStyle
			}
			continue
		}

		newStyle := getMatchingStyle(match.Type, f.language.Name)
		if newStyle != nil {
			currentStyle = newStyle
		}
	}

	if currentStyle == nil && referenceStyle == nil {
		return style
	}

	if referenceStyle != nil {
		return style.Copy().Inherit(*referenceStyle)
	}

	return style.Copy().Inherit(*currentStyle)
}

func getMatchingStyle(matchType string, name string) *lipgloss.Style {
	var currentStyle *lipgloss.Style

	for {
		codeStyle, ok := config.Theme.Editor.CodeStyles[fmt.Sprintf("%s.%s", matchType, name)]
		if ok {
			currentStyle = &codeStyle
			break
		}
		codeStyle, ok = config.Theme.Editor.CodeStyles[matchType]
		if ok {
			currentStyle = &codeStyle
			break
		}
		lastDot := strings.LastIndex(matchType, ".")
		if lastDot == -1 {
			break
		}
		matchType = matchType[:lastDot]
	}

	return currentStyle
}

type Match struct {
	Range         buffer.Range
	Type          string
	ReferenceType string
	Priority      int
	Source        string
}

type LocalDef struct {
	Name string
	Type string
}

type LocalScope struct {
	Inherits  bool
	Range     buffer.Range
	LocalDefs []*LocalDef
}

func (f *File) HighlightTree() {
	if f.tree == nil || f.tree.Tree == nil || f.tree.Language.Grammar == nil {
		return
	}
	version := f.Version()

	matches := highlightTree(f.tree.Copy())
	// slices.SortFunc(matches, func(a, b Match) int {
	//	return b.Priority - a.Priority
	// })

	f.SetMatches(version, matches)
}

func highlightTree(tree *Tree) []Match {
	query := tree.Language.Grammar.HighlightsQuery
	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(query.Query, tree.Tree.RootNode())

	var matches []Match
	var scopes []*LocalScope
	var localRef *LocalDef
	var localDef *LocalDef
	var lastMatch *sitter.QueryMatch
	for {
		match, index, ok := queryCursor.NextCapture()
		if !ok {
			break
		}
		capture := match.Captures[index]

		log.Println("Capture:", match.ID, capture.Index, query.Query.CaptureNameForID(capture.Index))

		captureRange := buffer.Range{
			Start: buffer.Position{
				Row: int(capture.StartPoint().Row),
				Col: int(capture.StartPoint().Column),
			},
			End: buffer.Position{
				Row: int(capture.EndPoint().Row),
				Col: int(capture.EndPoint().Column),
			},
		}

		for {
			if len(scopes) == 0 {
				break
			}
			lastScope := scopes[len(scopes)-1]
			if captureRange.Start.GreaterThan(lastScope.Range.End) {
				scopes = scopes[:len(scopes)-1]
				continue
			}

			break
		}

		if uint32(match.PatternIndex) < query.HighlightsPatternIndex {
			if query.ScopeCaptureID != nil && capture.Index == *query.ScopeCaptureID {
				log.Println("New scope")
				scopes = append(scopes, &LocalScope{
					Inherits:  true,
					Range:     captureRange,
					LocalDefs: nil,
				})
			} else if query.DefinitionCaptureID != nil && capture.Index == *query.DefinitionCaptureID {
				log.Println("Definition:", capture.Node.Content())
				if len(scopes) > 0 {
					def := &LocalDef{
						Name: capture.Node.Content(),
						Type: "",
					}

					localDef = def

					scope := scopes[len(scopes)-1]
					scope.LocalDefs = append(scope.LocalDefs, def)
				}

			} else if query.ReferenceCaptureID != nil && capture.Index == *query.ReferenceCaptureID {
				log.Println("Reference:", capture.Node.Content())
				for i := len(scopes) - 1; i >= 0; i-- {
					for _, def := range scopes[i].LocalDefs {
						if def.Name == capture.Node.Content() {
							log.Println("Found Definition for Reference:", def)
							localRef = def
							break
						}
					}
					if !scopes[i].Inherits {
						break
					}
				}
			}
			lastMatch = match
			continue
		}

		priority := 100
		if priorityStr, ok := match.Properties["priority"]; ok {
			if parsedPriority, err := strconv.Atoi(priorityStr); err == nil {
				priority = parsedPriority
			}
		}

		if localDef != nil {
			log.Println("Adding definition type:", query.Query.CaptureNameForID(capture.Index))
			localDef.Type = query.Query.CaptureNameForID(capture.Index)
		}

		highlightMatch := Match{
			Range: buffer.Range{
				Start: buffer.Position{Row: int(capture.StartPoint().Row), Col: int(capture.StartPoint().Column)},
				End:   buffer.Position{Row: int(capture.EndPoint().Row), Col: max(0, int(capture.EndPoint().Column)-1)}, // -1 to exclude the last character idk why this is like this tbh
			},
			Type:     query.Query.CaptureNameForID(capture.Index),
			Priority: priority,
			Source:   tree.Language.Name,
		}

		if localRef != nil {
			log.Println("Adding reference type:", *localRef)
			highlightMatch.ReferenceType = localRef.Type
		}

		matches = append(matches, highlightMatch)
		if lastMatch != nil && lastMatch.ID != match.ID {
			localRef = nil
			localDef = nil
		}
		lastMatch = match
	}

	for _, subTree := range tree.SubTrees {
		subMatches := highlightTree(subTree)
		matches = append(matches, subMatches...)
	}

	return matches
}
