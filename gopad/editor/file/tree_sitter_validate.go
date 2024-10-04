package file

import (
	"github.com/charmbracelet/bubbletea/v2"
	"go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/ls"
)

func ValidateTree(name string, version int32, tree *Tree) tea.Cmd {
	return func() tea.Msg {
		if tree == nil || tree.Tree == nil {
			return nil
		}

		return ls.UpdateFileDiagnosticMsg{
			Name:        name,
			Type:        ls.DiagnosticTypeTreeSitter,
			Version:     version,
			Diagnostics: validateTree(tree),
		}
	}
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
