package ls

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
				CommitCharactersSupport: true,
				DocumentationFormat:     []protocol.MarkupKind{protocol.PlainText},
				DeprecatedSupport:       true,
				PreselectSupport:        true,
				TagSupport: &protocol.CompletionTextDocumentClientCapabilitiesItemTagSupport{
					ValueSet: nil,
				},
				InsertReplaceSupport: true,
				ResolveSupport: &protocol.CompletionTextDocumentClientCapabilitiesItemResolveSupport{
					Properties: nil,
				},
				InsertTextModeSupport: &protocol.CompletionTextDocumentClientCapabilitiesItemInsertTextModeSupport{
					ValueSet: []protocol.InsertTextMode{
						protocol.InsertTextModeAsIs,
					},
				},
			},
			CompletionItemKind: &protocol.CompletionTextDocumentClientCapabilitiesItemKind{
				ValueSet: []protocol.CompletionItemKind{
					protocol.CompletionItemKindText,
					protocol.CompletionItemKindMethod,
					protocol.CompletionItemKindFunction,
					protocol.CompletionItemKindConstructor,
					protocol.CompletionItemKindField,
					protocol.CompletionItemKindVariable,
					protocol.CompletionItemKindClass,
					protocol.CompletionItemKindInterface,
					protocol.CompletionItemKindModule,
					protocol.CompletionItemKindProperty,
					protocol.CompletionItemKindUnit,
					protocol.CompletionItemKindValue,
					protocol.CompletionItemKindEnum,
					protocol.CompletionItemKindKeyword,
					protocol.CompletionItemKindSnippet,
					protocol.CompletionItemKindColor,
					protocol.CompletionItemKindFile,
					protocol.CompletionItemKindReference,
					protocol.CompletionItemKindFolder,
					protocol.CompletionItemKindEnumMember,
					protocol.CompletionItemKindConstant,
					protocol.CompletionItemKindStruct,
					protocol.CompletionItemKindEvent,
					protocol.CompletionItemKindOperator,
					protocol.CompletionItemKindTypeParameter,
				},
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

	var definition *protocol.DefinitionTextDocumentClientCapabilities
	if slices.Contains(cfg.Features, config.LanguageServerFeatureGoToDefinition) {
		definition = &protocol.DefinitionTextDocumentClientCapabilities{
			DynamicRegistration: false,
			LinkSupport:         false,
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
			Definition:         definition,
		},
	}
}
