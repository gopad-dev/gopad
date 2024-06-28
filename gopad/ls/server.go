package ls

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbletea"
	"go.lsp.dev/protocol"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
)

type ServerConfig struct {
	name string
	cfg  config.LanguageServerConfig
	new  func(name string, cfg config.LanguageServerConfig, workspace string) (*Server, error)
}

func (c *ServerConfig) New(workspace string) (*Server, error) {
	return c.new(c.name, c.cfg, workspace)
}

func (c *ServerConfig) Supported(workspace string) bool {
	var supports bool
	for _, root := range c.cfg.Roots {
		if _, err := os.Stat(filepath.Join(workspace, root)); err == nil {
			supports = true
			break
		}
	}

	return supports
}

type SendFunc func(msg tea.Cmd)

func newServer(name string, send SendFunc, workspace string, version string, cfg config.LanguageServerConfig, w io.Writer) (*Server, error) {
	c := &Server{
		name:      name,
		workspace: workspace,
		version:   version,
		send:      send,
		cfg:       cfg,
		w:         w,
	}

	if err := c.start(); err != nil {
		return nil, fmt.Errorf("error starting client: %w", err)
	}

	return c, nil
}

type Server struct {
	name      string
	workspace string
	version   string

	send   SendFunc
	cfg    config.LanguageServerConfig
	server protocol.Server
	cmd    *exec.Cmd
	rwc    io.ReadWriteCloser
	w      io.Writer
}

func (c *Server) Name() string {
	return c.name
}

func (c *Server) SupportedFile(name string) bool {
	return slices.Contains(c.cfg.FileTypes, filepath.Ext(name)) || slices.Contains(c.cfg.Files, filepath.Base(name))
}

func (c *Server) start() error {
	var err error
	c.cmd, c.rwc, err = newServerCmdStream(context.Background(), c.w, c.cfg.Command, c.cfg.Args...)
	if err != nil {
		return fmt.Errorf("error creating command stream: %w", err)
	}

	_, c.server, err = newServerConn(context.Background(), c.rwc, c, c.w)
	if err != nil {
		return fmt.Errorf("error creating server: %w", err)
	}

	var workspaceFolders []protocol.WorkspaceFolder
	if c.workspace != "" {
		workspaceFolders = append(workspaceFolders, protocol.WorkspaceFolder{
			URI:  "file://" + c.workspace,
			Name: filepath.Base(c.workspace),
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err = c.server.Initialize(ctx, &protocol.InitializeParams{
		ClientInfo: &protocol.ClientInfo{
			Name:    c.name,
			Version: c.version,
		},
		Locale:                "de",
		InitializationOptions: c.cfg.Config,
		WorkspaceFolders:      workspaceFolders,
		Capabilities:          clientCapabilities(c.cfg),
	}); err != nil {
		return fmt.Errorf("error initializing server: %w", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	if err = c.server.Initialized(ctx2, &protocol.InitializedParams{}); err != nil {
		return fmt.Errorf("error sending initialized: %w", err)
	}

	return nil
}

func (c *Server) Stop(ctx context.Context) error {
	if err := c.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("error sending shutdown: %w", err)
	}

	if err := c.server.Exit(ctx); err != nil {
		return fmt.Errorf("error sending exit: %w", err)
	}

	if err := c.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("error killing process: %w", err)
	}

	if err := c.rwc.Close(); err != nil {
		return fmt.Errorf("error closing rwc: %w", err)
	}

	return nil
}

func (c *Server) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case WorkspaceOpenedMsg:
		c.workspace = msg.Workspace
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
			return Err(err)
		}

	case WorkspaceClosedMsg:
		c.workspace = ""
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
			return Err(err)
		}
	}

	switch msg := msg.(type) {
	case GetDefinitionMsg:
		locations, err := c.server.Definition(context.Background(), &protocol.DefinitionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: protocol.DocumentURI("file://" + msg.Name),
				},
				Position: protocol.Position{
					Line:      uint32(msg.Row),
					Character: uint32(msg.Col),
				},
			},
		})
		if err != nil {
			return Err(err)
		}

		definitions := make([]Definition, 0, len(locations))
		for _, location := range locations {
			definitions = append(definitions, Definition{
				Name:  location.URI.Filename(),
				Range: buffer.ParseRange(location.Range),
			})
		}
		cmds = append(cmds, UpdateDefinition(msg.Name, definitions))
	case GetInlayHintMsg:
		result, err := c.server.InlayHint(context.Background(), &protocol.InlayHintParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI("file://" + msg.Name),
			},
			Range: msg.Range.ToProtocol(),
		})
		if err != nil {
			return Err(err)
		}

		hints := make([]InlayHint, 0, len(result))
		for _, hint := range result {
			kind := InlayHintTypeNone
			if hint.Kind != nil {
				kind = InlayHintType(*hint.Kind)
			}

			var label string
			switch l := hint.Label.(type) {
			case string:
				label = l
			case []protocol.InlayHintLabelPart:
				for _, part := range l {
					label += part.Value
				}
			}

			var tooltip string
			switch t := hint.Tooltip.(type) {
			case string:
				tooltip = t
			case *protocol.MarkupContent:
				tooltip = t.Value
			}
			hints = append(hints, InlayHint{
				Type:         kind,
				Position:     buffer.ParsePosition(hint.Position),
				Label:        label,
				Tooltip:      tooltip,
				PaddingLeft:  hint.PaddingLeft,
				PaddingRight: hint.PaddingRight,
			})
		}
		cmds = append(cmds, UpdateInlayHint(msg.Name, hints))
	case GetAutocompletionMsg:
		result, err := c.server.Completion(context.Background(), &protocol.CompletionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: protocol.DocumentURI("file://" + msg.Name),
				},
				Position: protocol.Position{
					Line:      uint32(msg.Row),
					Character: uint32(msg.Col),
				},
			},
		})
		if err != nil {
			return Err(err)
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
					Range:   buffer.ParseRange(resultItem.TextEdit.Range),
					NewText: resultItem.TextEdit.NewText,
				}
			}

			items = append(items, item)
		}
		cmds = append(cmds, UpdateAutocompletion(msg.Name, items))
	case FileOpenedMsg:
		if err := c.server.DidOpen(context.Background(), &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{
				URI:        protocol.DocumentURI("file://" + msg.Name),
				LanguageID: protocol.GoLanguage,
				Version:    msg.Version,
				Text:       string(msg.Text),
			},
		}); err != nil {
			return Err(err)
		}
	case FileClosedMsg:
		if err := c.server.DidClose(context.Background(), &protocol.DidCloseTextDocumentParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI("file://" + msg.Name),
			},
		}); err != nil {
			return Err(err)
		}
	case FileCreatedMsg:
		if err := c.server.DidCreateFiles(context.Background(), &protocol.CreateFilesParams{
			Files: []protocol.FileCreate{
				{
					URI: "file://" + msg.Name,
				},
			},
		}); err != nil {
			return Err(err)
		}
	case FileDeletedMsg:
		if err := c.server.DidDeleteFiles(context.Background(), &protocol.DeleteFilesParams{
			Files: []protocol.FileDelete{
				{
					URI: "file://" + msg.Name,
				},
			},
		}); err != nil {
			return Err(err)
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
			return Err(err)
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
			return Err(err)
		}
	case FileSavedMsg:
		if err := c.server.DidSave(context.Background(), &protocol.DidSaveTextDocumentParams{
			Text: string(msg.Text),
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI("file://" + msg.Name),
			},
		}); err != nil {
			return Err(err)
		}
	}

	return tea.Batch(cmds...)
}

func (c *Server) Progress(ctx context.Context, params *protocol.ProgressParams) error {
	// log.Println("Progress", params)
	return nil
}

func (c *Server) WorkDoneProgressCreate(ctx context.Context, params *protocol.WorkDoneProgressCreateParams) error {
	// log.Println("WorkDoneProgressCreate", params)
	return nil
}

func (c *Server) LogMessage(ctx context.Context, params *protocol.LogMessageParams) error {
	// log.Println("LogMessage", params)
	// c.gopad.send(notifications.Add(params.Message)())
	return nil
}

func (c *Server) PublishDiagnostics(ctx context.Context, params *protocol.PublishDiagnosticsParams) error {
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
			Type:            DiagnosticTypeLanguageServer,
			Name:            c.Name(),
			Source:          diagnostic.Source,
			Range:           buffer.ParseRange(diagnostic.Range),
			Severity:        DiagnosticSeverity(diagnostic.Severity),
			Code:            code,
			CodeDescription: codeDescription,
			Message:         diagnostic.Message,
			Data:            diagnostic.Data,
			Priority:        110,
		})
	}
	c.send(UpdateFileDiagnostic(params.URI.Filename(), DiagnosticTypeLanguageServer, int32(params.Version), diagnostics))
	return nil
}

func (c *Server) ShowMessage(ctx context.Context, params *protocol.ShowMessageParams) error {
	c.send(notifications.Add(params.Message))
	return nil
}

func (c *Server) ShowMessageRequest(ctx context.Context, params *protocol.ShowMessageRequestParams) (*protocol.MessageActionItem, error) {
	return nil, nil
}

func (c *Server) Telemetry(ctx context.Context, params any) error {
	return nil
}

func (c *Server) RegisterCapability(ctx context.Context, params *protocol.RegistrationParams) error {
	return nil
}

func (c *Server) UnregisterCapability(ctx context.Context, params *protocol.UnregistrationParams) error {
	return nil
}

func (c *Server) ApplyEdit(ctx context.Context, params *protocol.ApplyWorkspaceEditParams) (*protocol.ApplyWorkspaceEditResponse, error) {
	return nil, nil
}

func (c *Server) Configuration(ctx context.Context, params *protocol.ConfigurationParams) ([]any, error) {
	return nil, nil
}

func (c *Server) WorkspaceFolders(ctx context.Context) ([]protocol.WorkspaceFolder, error) {
	var workspaceFolders []protocol.WorkspaceFolder
	if c.workspace != "" {
		workspaceFolders = append(workspaceFolders, protocol.WorkspaceFolder{
			URI:  "file://" + c.workspace,
			Name: filepath.Base(c.workspace),
		})
	}
	return workspaceFolders, nil
}

func (c *Server) InlayHintRefresh(ctx context.Context) error {
	c.send(RefreshInlayHint())
	return nil
}
