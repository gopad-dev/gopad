package file

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
)

func HighlightTree(name string, version int32, tree *Tree, lines int) tea.Cmd {
	return func() tea.Msg {
		if tree == nil || tree.Tree == nil || tree.Language.Grammar == nil {
			return nil
		}

		return UpdateMatchesMsg{
			Name:    name,
			Version: version,
			Matches: highlightTree(tree, lines),
		}
	}
}

type UpdateMatchesMsg struct {
	Name    string
	Version int32
	Matches [][]*Match
}

func (f *File) SetMatches(version int32, matches [][]*Match) {
	log.Println("setting matches", version, len(matches))
	if version < f.matchesVersion {
		log.Printf("skipping outdated matches: %d < %d", version, f.matchesVersion)
		return
	}
	if version > f.matchesVersion {
		f.matches = matches
		f.matchesVersion = version
		return
	}
	f.matches = append(f.matches, matches...)
}

func (f *File) MatchesForLineCol(row int, col int) []*Match {
	pos := buffer.Position{Row: row, Col: col}

	if len(f.matches) <= row {
		return nil
	}

	lineMatches := f.matches[row]

	var matches []*Match
	for _, match := range lineMatches {
		if match.Range.Contains(pos) {
			matches = append(matches, match)
		}
	}

	return matches
}

func (f *File) HighestMatchStyle(style lipgloss.Style, row int, col int) lipgloss.Style {
	var (
		currentStyle   *lipgloss.Style
		referenceStyle *lipgloss.Style
	)
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

	if referenceStyle != nil {
		return style.Inherit(*referenceStyle)
	}

	if currentStyle != nil {
		return style.Inherit(*currentStyle)
	}

	return style
}

func getMatchingStyle(matchType string, name string) *lipgloss.Style {
	var currentStyle *lipgloss.Style

	for {
		codeStyle, ok := config.Theme.CodeStyles[fmt.Sprintf("%s.%s", matchType, name)]
		if ok {
			currentStyle = &codeStyle
			break
		}
		codeStyle, ok = config.Theme.CodeStyles[matchType]
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

func highlightTree(tree *Tree, lines int) [][]*Match {
	now := time.Now()
	defer func() {
		log.Println("highlightTree took", time.Since(now))
	}()

	query := tree.Language.Grammar.HighlightsQuery
	queryCursor := sitter.NewQueryCursor()
	queryCursor.Exec(query.Query, tree.Tree.RootNode())

	lineMatches := make([][]*Match, lines)
	var lineMatchesMu sync.Mutex
	var scopes []*LocalScope
	var lastDef *LocalDef
	var lastRef *LocalDef
	var lastCapture *sitter.QueryCapture

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			match, index, ok := queryCursor.NextCapture()
			if !ok {
				break
			}
			capture := match.Captures[index]

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
					lastRef = nil
					lastDef = nil
					continue
				}

				break
			}

			// if lastCapture != nil && !capture.Node.Equal(lastCapture.Node) {
			if lastCapture != nil && lastCapture.Node.Content() != capture.Node.Content() {
				lastDef = nil
				lastRef = nil
			}

			if uint32(match.PatternIndex) < query.HighlightsPatternIndex {
				if query.ScopeCaptureID != nil && capture.Index == *query.ScopeCaptureID {
					scopes = append(scopes, &LocalScope{
						Inherits:  true,
						Range:     captureRange,
						LocalDefs: nil,
					})
				} else if query.DefinitionCaptureID != nil && capture.Index == *query.DefinitionCaptureID {
					if len(scopes) > 0 {
						def := &LocalDef{
							Name: capture.Node.Content(),
							Type: "",
						}

						lastDef = def

						scope := scopes[len(scopes)-1]
						scope.LocalDefs = append(scope.LocalDefs, def)
					}
				} else if query.ReferenceCaptureID != nil && capture.Index == *query.ReferenceCaptureID {
					for i := len(scopes) - 1; i >= 0; i-- {
						for ii := len(scopes[i].LocalDefs) - 1; ii >= 0; ii-- {
							def := scopes[i].LocalDefs[ii]

							if def.Type != "" && def.Name == capture.Node.Content() {
								lastRef = def
								break
							}
						}
						if !scopes[i].Inherits {
							break
						}
					}
				}

				lastCapture = &capture
				continue
			}

			if lastDef != nil {
				lastDef.Type = query.Query.CaptureNameForID(capture.Index)
			}

			var refType string
			if lastRef != nil {
				refType = lastRef.Type
			}

			lineMatch := &Match{
				Range: buffer.Range{
					Start: buffer.Position{Row: int(capture.StartPoint().Row), Col: int(capture.StartPoint().Column)},
					End:   buffer.Position{Row: int(capture.EndPoint().Row), Col: max(0, int(capture.EndPoint().Column)-1)}, // -1 to exclude the last character idk why this is like this tbh
				},
				Type:          query.Query.CaptureNameForID(capture.Index),
				ReferenceType: refType,
				Priority:      getPriority(match),
				Source:        tree.Language.Name,
			}

			for row := captureRange.Start.Row; row <= captureRange.End.Row; row++ {
				lineMatchesMu.Lock()
				lineMatches[row] = append(lineMatches[row], lineMatch)
				lineMatchesMu.Unlock()
			}

			lastCapture = &capture
		}
	}()

	for _, subTree := range tree.SubTrees {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for row, matches := range highlightTree(subTree, lines) {
				if len(matches) == 0 {
					continue
				}
				lineMatchesMu.Lock()
				lineMatches[row] = append(lineMatches[row], matches...)
				lineMatchesMu.Unlock()
			}
		}()
	}

	wg.Wait()

	return lineMatches
}

func getPriority(match *sitter.QueryMatch) int {
	priority := 100
	if priorityStr, ok := match.Properties["priority"]; ok {
		if parsedPriority, err := strconv.Atoi(priorityStr); err == nil {
			priority = parsedPriority
		}
	}
	return priority
}
