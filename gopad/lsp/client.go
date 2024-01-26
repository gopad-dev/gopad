package lsp

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/charmbracelet/bubbletea"
	"go.lsp.dev/protocol"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
)

type SendFunc func(msg tea.Cmd)

func newClient(name string, version string, cfg config.LSPServerConfig, w io.Writer) *Client {
	return &Client{
		name:    name,
		version: version,
		cfg:     cfg,
		w:       w,
	}
}

type Client struct {
	name    string
	version string

	workspace string
	send      SendFunc
	cfg       config.LSPServerConfig
	server    protocol.Server
	cmd       *exec.Cmd
	rwc       io.ReadWriteCloser
	w         io.Writer
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) SupportedFile(name string) bool {
	return slices.Contains(c.cfg.FileTypes, filepath.Ext(name)) || slices.Contains(c.cfg.Files, filepath.Base(name))
}

func (c *Client) Running() bool {
	return c.cmd != nil && c.cmd.ProcessState == nil
}

func (c *Client) EnsureRunning(send SendFunc) error {
	if c.Running() {
		return nil
	}

	if err := c.Start(send); err != nil {
		return fmt.Errorf("error starting lsp client: %w", err)
	}

	return nil
}

func (c *Client) Start(send SendFunc) error {
	c.send = send

	var err error
	ctx := context.Background()
	c.cmd, c.rwc, err = newCmdStream(ctx, c.cfg.Command, c.cfg.Args...)
	if err != nil {
		return fmt.Errorf("error creating command stream: %w", err)
	}

	_, c.server, err = newServer(ctx, c.rwc, c, c.w)
	if err != nil {
		return fmt.Errorf("error creating server: %w", err)
	}

	if _, err = c.server.Initialize(ctx, &protocol.InitializeParams{
		ClientInfo: &protocol.ClientInfo{
			Name:    c.name,
			Version: c.version,
		},
		Locale:                "de",
		InitializationOptions: c.cfg.Config,
		WorkspaceFolders: []protocol.WorkspaceFolder{
			{
				URI:  "file://" + c.workspace,
				Name: filepath.Base(c.workspace),
			},
		},
		Capabilities: protocol.ClientCapabilities{
			Workspace: &protocol.WorkspaceClientCapabilities{
				//		ApplyEdit: true,
				//		WorkspaceEdit: &protocol.WorkspaceClientCapabilitiesWorkspaceEdit{
				//			DocumentChanges:       false,
				//			FailureHandling:       "",
				//			ResourceOperations:    nil,
				//			NormalizesLineEndings: false,
				//			ChangeAnnotationSupport: &protocol.WorkspaceClientCapabilitiesWorkspaceEditChangeAnnotationSupport{
				//				GroupsOnLabel: false,
				//			},
				//		},
				//		DidChangeConfiguration: &protocol.DidChangeConfigurationWorkspaceClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		DidChangeWatchedFiles: &protocol.DidChangeWatchedFilesWorkspaceClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		Symbol: &protocol.WorkspaceSymbolClientCapabilities{
				//			DynamicRegistration: false,
				//			SymbolKind: &protocol.SymbolKindCapabilities{
				//				ValueSet: nil,
				//			},
				//			TagSupport: &protocol.TagSupportCapabilities{
				//				ValueSet: nil,
				//			},
				//		},
				//		ExecuteCommand: &protocol.ExecuteCommandClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				WorkspaceFolders: true,
				//		Configuration:    false,
				//		SemanticTokens: &protocol.SemanticTokensWorkspaceClientCapabilities{
				//			RefreshSupport: false,
				//		},
				//		CodeLens: &protocol.CodeLensWorkspaceClientCapabilities{
				//			RefreshSupport: false,
				//		},
				//		FileOperations: &protocol.WorkspaceClientCapabilitiesFileOperations{
				//			DynamicRegistration: false,
				//			DidCreate:           true,
				//			WillCreate:          false,
				//			DidRename:           false,
				//			WillRename:          false,
				//			DidDelete:           true,
				//			WillDelete:          false,
				//		},
			},
			TextDocument: &protocol.TextDocumentClientCapabilities{
				//		Synchronization: &protocol.TextDocumentSyncClientCapabilities{
				//			DynamicRegistration: false,
				//			WillSave:            false,
				//			WillSaveWaitUntil:   false,
				//			DidSave:             true,
				//		},
				Completion: &protocol.CompletionTextDocumentClientCapabilities{
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
				},
				//		Hover: &protocol.HoverTextDocumentClientCapabilities{
				//			DynamicRegistration: false,
				//			ContentFormat:       nil,
				//		},
				//		SignatureHelp: &protocol.SignatureHelpTextDocumentClientCapabilities{
				//			DynamicRegistration: false,
				//			SignatureInformation: &protocol.TextDocumentClientCapabilitiesSignatureInformation{
				//				DocumentationFormat: nil,
				//				ParameterInformation: &protocol.TextDocumentClientCapabilitiesParameterInformation{
				//					LabelOffsetSupport: false,
				//				},
				//				ActiveParameterSupport: false,
				//			},
				//			ContextSupport: false,
				//		},
				//		Declaration: &protocol.DeclarationTextDocumentClientCapabilities{
				//			DynamicRegistration: false,
				//			LinkSupport:         false,
				//		},
				//		Definition: &protocol.DefinitionTextDocumentClientCapabilities{
				//			DynamicRegistration: false,
				//			LinkSupport:         false,
				//		},
				//		TypeDefinition: &protocol.TypeDefinitionTextDocumentClientCapabilities{
				//			DynamicRegistration: false,
				//			LinkSupport:         false,
				//		},
				//		Implementation: &protocol.ImplementationTextDocumentClientCapabilities{
				//			DynamicRegistration: false,
				//			LinkSupport:         false,
				//		},
				//		References: &protocol.ReferencesTextDocumentClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		DocumentHighlight: &protocol.DocumentHighlightClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		DocumentSymbol: &protocol.DocumentSymbolClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		CodeAction: &protocol.CodeActionClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		CodeLens: &protocol.CodeLensClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		DocumentLink: &protocol.DocumentLinkClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		ColorProvider: &protocol.DocumentColorClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		Formatting: &protocol.DocumentFormattingClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		RangeFormatting: &protocol.DocumentRangeFormattingClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		OnTypeFormatting: &protocol.DocumentOnTypeFormattingClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				PublishDiagnostics: &protocol.PublishDiagnosticsClientCapabilities{
					RelatedInformation: false,
					TagSupport: &protocol.PublishDiagnosticsClientCapabilitiesTagSupport{
						ValueSet: nil,
					},
					VersionSupport:         false,
					CodeDescriptionSupport: true,
					DataSupport:            false,
				},
				//		Rename: &protocol.RenameClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		FoldingRange:   &protocol.FoldingRangeClientCapabilities{},
				//		SelectionRange: &protocol.SelectionRangeClientCapabilities{},
				//		CallHierarchy:  &protocol.CallHierarchyClientCapabilities{},
				//		SemanticTokens: &protocol.SemanticTokensClientCapabilities{},
				//		LinkedEditingRange: &protocol.LinkedEditingRangeClientCapabilities{
				//			DynamicRegistration: false,
				//		},
				//		Moniker: &protocol.MonikerClientCapabilities{
				//			DynamicRegistration: false,
				//		},
			},
		},
	}); err != nil {
		return fmt.Errorf("error initializing server: %w", err)
	}

	if err = c.server.Initialized(ctx, &protocol.InitializedParams{}); err != nil {
		return fmt.Errorf("error sending initialized: %w", err)
	}

	return nil
}

func (c *Client) Stop() error {
	if c.cmd == nil {
		return nil
	}

	if err := c.server.Shutdown(context.Background()); err != nil {
		return err
	}

	if err := c.server.Exit(context.Background()); err != nil {
		return err
	}

	if err := c.cmd.Process.Kill(); err != nil {
		return err
	}

	if err := c.rwc.Close(); err != nil {
		return err
	}

	return nil
}

func (c *Client) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case WorkspaceOpenedMsg:
		c.workspace = msg.Workspace
		if c.server != nil {
			if err := c.server.DidChangeWorkspaceFolders(context.Background(), &protocol.DidChangeWorkspaceFoldersParams{
				Event: protocol.WorkspaceFoldersChangeEvent{
					Added: []protocol.WorkspaceFolder{
						{
							URI:  "file://" + c.workspace,
							Name: filepath.Base(c.workspace),
						},
					},
				},
			}); err != nil {
				return func() tea.Msg {
					return err
				}
			}
		}
	case WorkspaceClosedMsg:
		c.workspace = ""
		if c.server != nil {
			if err := c.server.DidChangeWorkspaceFolders(context.Background(), &protocol.DidChangeWorkspaceFoldersParams{
				Event: protocol.WorkspaceFoldersChangeEvent{
					Removed: []protocol.WorkspaceFolder{
						{
							URI:  "file://" + c.workspace,
							Name: filepath.Base(c.workspace),
						},
					},
				},
			}); err != nil {
				return func() tea.Msg {
					return err
				}
			}
		}
	}

	if c.server == nil {
		return nil
	}

	switch msg := msg.(type) {
	case GetAutocompletionMsg:
		result, err := c.server.Completion(context.Background(), &protocol.CompletionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: protocol.DocumentURI("file://" + msg.File),
				},
				Position: protocol.Position{
					Line:      uint32(msg.Row),
					Character: uint32(msg.Col),
				},
			},
		})
		if err != nil {
			return func() tea.Msg {
				return err
			}
		}

		items := make([]CompletionItem, 0, len(result.Items))
		for _, resultItem := range result.Items {
			item := CompletionItem{
				Label:  resultItem.Label,
				Detail: resultItem.Detail,
				Text:   resultItem.InsertText,
			}

			if resultItem.TextEdit != nil {
				item.Edit = &TextEdit{
					Range: buffer.Range{
						Start: buffer.Position{
							Row: int(resultItem.TextEdit.Range.Start.Line),
							Col: int(resultItem.TextEdit.Range.Start.Character),
						},
						End: buffer.Position{
							Row: int(resultItem.TextEdit.Range.End.Line),
							Col: int(resultItem.TextEdit.Range.End.Character),
						},
					},
					NewText: resultItem.TextEdit.NewText,
				}
			}

			items = append(items, item)
		}
		cmds = append(cmds, UpdateAutocompletion(msg.File, items))
	case FileOpenedMsg:
		if err := c.server.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{
				URI:        protocol.DocumentURI("file://" + msg.Name),
				LanguageID: protocol.GoLanguage,
				Version:    msg.Version,
				Text:       string(msg.Text),
			},
		}); err != nil {
			return func() tea.Msg {
				return err
			}
		}
	case FileClosedMsg:
		if err := c.server.DidClose(context.Background(), &protocol.DidCloseTextDocumentParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI("file://" + msg.Name),
			},
		}); err != nil {
			return func() tea.Msg {
				return err
			}
		}
	case FileCreatedMsg:
		if err := c.server.DidCreateFiles(context.Background(), &protocol.CreateFilesParams{
			Files: []protocol.FileCreate{
				{
					URI: "file://" + msg.Name,
				},
			},
		}); err != nil {
			return func() tea.Msg {
				return err
			}
		}
	case FileDeletedMsg:
		if err := c.server.DidDeleteFiles(context.Background(), &protocol.DeleteFilesParams{
			Files: []protocol.FileDelete{
				{
					URI: "file://" + msg.Name,
				},
			},
		}); err != nil {
			return func() tea.Msg {
				return err
			}
		}
	case FileRenamedMsg:
		if err := c.server.DidRenameFiles(context.Background(), &protocol.RenameFilesParams{
			Files: []protocol.FileRename{
				{
					OldURI: "file://" + msg.OldName,
					NewURI: "file://" + msg.NewName,
				},
			},
		}); err != nil {
			return func() tea.Msg {
				return err
			}
		}
	case FileChangedMsg:
		if err := c.server.DidChange(context.Background(), &protocol.DidChangeTextDocumentParams{
			TextDocument: protocol.VersionedTextDocumentIdentifier{
				TextDocumentIdentifier: protocol.TextDocumentIdentifier{
					URI: protocol.DocumentURI("file://" + msg.Name),
				},
				Version: msg.Version,
			},
			ContentChanges: []protocol.TextDocumentContentChangeEvent{
				{
					Text: string(msg.Text),
				},
			},
		}); err != nil {
			return func() tea.Msg {
				return err
			}
		}
	case FileSavedMsg:
		if err := c.server.DidSave(context.Background(), &protocol.DidSaveTextDocumentParams{
			Text: string(msg.Text),
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI("file://" + msg.Name),
			},
		}); err != nil {
			return func() tea.Msg {
				return err
			}
		}
	}

	return tea.Batch(cmds...)
}

func (c *Client) Progress(ctx context.Context, params *protocol.ProgressParams) error {
	// log.Println("Progress", params)
	return nil
}

func (c *Client) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) error {
	// log.Println("WorkDoneProgressCreate", params)
	return nil
}

func (c *Client) LogMessage(ctx context.Context, params *protocol.LogMessageParams) error {
	// log.Println("LogMessage", params)
	// c.gopad.Send(notifications.Add(params.Message)())
	return nil
}

func (c *Client) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) error {
	diagnostics := make([]Diagnostic, 0, len(params.Diagnostics))
	for _, diagnostic := range params.Diagnostics {
		var code string
		switch dCode := diagnostic.Code.(type) {
		case string:
			code = dCode
		case int32:
			code = strconv.Itoa(int(dCode))
		}

		var codeDescription string
		if diagnostic.CodeDescription != nil {
			codeDescription = string(diagnostic.CodeDescription.Href)
		}

		diagnostics = append(diagnostics, Diagnostic{
			Type:   DiagnosticTypeLanguageServer,
			Source: diagnostic.Source,
			Range: buffer.Range{
				Start: buffer.Position{
					Row: int(diagnostic.Range.Start.Line),
					Col: int(diagnostic.Range.Start.Character),
				},
				End: buffer.Position{
					Row: int(diagnostic.Range.End.Line),
					Col: int(diagnostic.Range.End.Character),
				},
			},
			Severity:        DiagnosticSeverity(diagnostic.Severity),
			Code:            code,
			CodeDescription: codeDescription,
			Message:         diagnostic.Message,
			Data:            diagnostic.Data,
			Priority:        110,
		})
	}
	c.send(UpdateFileDiagnostics(params.URI.Filename(), int32(params.Version), diagnostics))
	return nil
}

func (c *Client) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) error {
	c.send(notifications.Add(params.Message))
	return nil
}

func (c *Client) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (*protocol.MessageActionItem, error) {
	return nil, nil
}

func (c *Client) Telemetry(ctx context.Context, params any) error {
	return nil
}

func (c *Client) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams) error {
	return nil
}

func (c *Client) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) error {
	return nil
}

func (c *Client) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (*protocol.ApplyWorkspaceEditResponse, error) {
	return nil, nil
}

func (c *Client) Configuration(ctx context.Context, params *protocol.ConfigurationParams) ([]any, error) {
	return nil, nil
}

func (c *Client) WorkspaceFolders(ctx context.Context) ([]protocol.WorkspaceFolder, error) {
	return []protocol.WorkspaceFolder{
		{
			URI:  "file://" + c.workspace,
			Name: filepath.Base(c.workspace),
		},
	}, nil
}
