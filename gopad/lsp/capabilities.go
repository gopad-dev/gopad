package lsp

import (
	"slices"

	"go.lsp.dev/protocol"

	"go.gopad.dev/gopad/gopad/config"
)

func clientCapabilities(cfg config.LanguageServerConfig) protocol.ClientCapabilities {
	var completion *protocol.CompletionTextDocumentClientCapabilities
	if slices.Contains(cfg.Features, config.LanguageServerFeatureCompletion) {
		completion = &protocol.CompletionTextDocumentClientCapabilities{
			DynamicRegistration: false,
			CompletionItem: &protocol.CompletionTextDocumentClientCapabilitiesItem{
				SnippetSupport:          false,
				CommitCharactersSupport: false,
				DocumentationFormat:     nil,
				DeprecatedSupport:       false,
				PreselectSupport:        false,
				TagSupport: &protocol.CompletionTextDocumentClientCapabilitiesItemTagSupport{
					ValueSet: nil,
				},
				InsertReplaceSupport: false,
				ResolveSupport: &protocol.CompletionTextDocumentClientCapabilitiesItemResolveSupport{
					Properties: nil,
				},
				InsertTextModeSupport: &protocol.CompletionTextDocumentClientCapabilitiesItemInsertTextModeSupport{
					ValueSet: nil,
				},
			},
			CompletionItemKind: &protocol.CompletionTextDocumentClientCapabilitiesItemKind{
				ValueSet: nil,
			},
			ContextSupport: false,
		}
	}

	var publishDiagnostics *protocol.PublishDiagnosticsClientCapabilities
	var diagnostic *protocol.DiagnosticClientCapabilities
	if slices.Contains(cfg.Features, config.LanguageServerFeatureDiagnostics) {
		publishDiagnostics = &protocol.PublishDiagnosticsClientCapabilities{
			RelatedInformation: false,
			TagSupport: &protocol.PublishDiagnosticsClientCapabilitiesTagSupport{
				ValueSet: nil,
			},
			VersionSupport:         false,
			CodeDescriptionSupport: true,
			DataSupport:            false,
		}
		diagnostic = &protocol.DiagnosticClientCapabilities{
			DynamicRegistration:    false,
			RelatedDocumentSupport: false,
		}
	}

	var inlayHintWorkspace *protocol.InlayHintWorkspaceClientCapabilities
	var inlayHint *protocol.InlayHintClientCapabilities
	if slices.Contains(cfg.Features, config.LanguageServerFeatureInlayHints) {
		inlayHintWorkspace = &protocol.InlayHintWorkspaceClientCapabilities{
			RefreshSupport: true,
		}
		inlayHint = &protocol.InlayHintClientCapabilities{
			DynamicRegistration: false,
			ResolveSupport:      nil,
		}
	}

	return protocol.ClientCapabilities{
		Workspace: &protocol.WorkspaceClientCapabilities{
			WorkspaceFolders: true,
			InlayHint:        inlayHintWorkspace,
		},
		TextDocument: &protocol.TextDocumentClientCapabilities{
			Completion:         completion,
			PublishDiagnostics: publishDiagnostics,
			InlayHint:          inlayHint,
			Diagnostic:         diagnostic,
		},
	}
}
