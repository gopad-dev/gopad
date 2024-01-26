package gopad

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor"
	"go.gopad.dev/gopad/gopad/lsp"
	"go.gopad.dev/gopad/internal/bubbles/cursor"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

func New(lspClient *lsp.LSP, version string, args []string) (*Gopad, error) {
	e, err := editor.NewEditor(args)
	if err != nil {
		return nil, err
	}
	e.Focus()

	o := config.NewOverlays()
	n := config.NewNotifications()
	return &Gopad{
		lspClient:     lspClient,
		version:       version,
		editor:        *e,
		overlays:      o,
		notifications: n,
	}, nil
}

type Gopad struct {
	lspClient *lsp.LSP
	version   string

	height int
	width  int

	editor        editor.Editor
	overlays      overlay.Model
	notifications notifications.Model
}

func (g Gopad) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("gopad"),
		cursor.Blink,
		g.editor.Init(),
	)
}

func (g Gopad) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g.width = msg.Width
		g.height = msg.Height

	case overlay.ResetFocusMsg:
		cmds = append(cmds, g.editor.Focus())

	case tea.KeyMsg:
		log.Println("KeyMsg", msg)
		// global keybindings
		switch {
		case key.Matches(msg, config.Keys.Quit):
			if !g.overlays.Has(QuitOverlayID) {
				if g.editor.HasChanges() {
					return g, overlay.Open(NewQuitOverlay())
				}
				return g, tea.Quit
			}
		case key.Matches(msg, config.Keys.Help):
			if !g.overlays.Has(HelpOverlayID) {
				cmds = append(cmds, overlay.Open(NewHelpOverlay()))
			}
		case key.Matches(msg, config.Keys.Terminal):
			return g, Terminal()
		}

		if !g.overlays.Has(KeyMapperOverlayID) {
			switch {
			case key.Matches(msg, config.Keys.Run):
				if !g.overlays.Has(RunOverlayID) {
					cmds = append(cmds, overlay.Open(NewRunOverlay()))
				}

			case key.Matches(msg, config.Keys.KeyMapper):
				if !g.overlays.Has(KeyMapperOverlayID) {
					cmds = append(cmds, overlay.Open(NewKeyMapperOverlay()))
				}
			case key.Matches(msg, config.Keys.Editor.OpenFile):
				if !g.overlays.Has(editor.OpenOverlayID) {
					path, err := os.Getwd()
					if err != nil {
						cmds = append(cmds, notifications.Add(fmt.Sprintf("Error getting current working directory: %s", err)))
						return g, tea.Batch(cmds...)
					}
					cmds = append(cmds, overlay.Open(editor.NewOpenOverlay(path, true, false)))
				}
			case key.Matches(msg, config.Keys.Editor.OpenFolder):
				if !g.overlays.Has(editor.OpenOverlayID) {
					path, err := os.Getwd()
					if err != nil {
						cmds = append(cmds, notifications.Add(fmt.Sprintf("Error getting current working directory: %s", err)))
						return g, tea.Batch(cmds...)
					}
					cmds = append(cmds, overlay.Open(editor.NewOpenOverlay(path, false, true)))
				}
			}
		}
	}

	var cmd tea.Cmd
	if g.notifications, cmd = g.notifications.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if g.overlays, cmd = g.overlays.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if _, ok := msg.(tea.KeyMsg); ok && g.overlays.Focused() {
		return g, tea.Batch(cmds...)
	}

	if g.editor, cmd = g.editor.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return g, tea.Batch(cmds...)
}

func (g Gopad) View() string {
	appBar := g.AppBar()
	codeBar := g.CodeBar()

	codeEditor := g.editor.View(g.width, g.height-lipgloss.Height(appBar)-lipgloss.Height(codeBar))
	view := fmt.Sprintf("%s\n%s\n%s", appBar, codeEditor, codeBar)

	if g.overlays.Focused() {
		view = g.overlays.View(view, g.width, g.height)
	}

	if g.notifications.Active() {
		g.notifications.SetBackground(view)
		view = g.notifications.View()
	}

	return view
}

func (g Gopad) AppBar() string {
	appBar := config.Theme.AppBarTitleStyle.Render("gopad-" + g.version)
	appBar += g.editor.FileTabsView(g.width - lipgloss.Width(appBar))

	return config.Theme.AppBarStyle.Width(g.width).Render(appBar)
}

func (g Gopad) CodeBar() string {
	width := g.width - config.Theme.Editor.CodeBarStyle.GetHorizontalFrameSize()

	file := g.editor.File()
	if file == nil {
		return config.Theme.Editor.CodeBarStyle.Render(strings.Repeat(" ", max(0, width)))
	}

	var infoLine string

	if s := file.Selection(); s != nil {
		infoLine += fmt.Sprintf("[%d:%d-%d:%d] ", s.Start.Row+1, s.Start.Col+1, s.End.Row+1, s.End.Col+1)
	} else {
		cursorRow, cursorCol := file.Cursor()
		infoLine += fmt.Sprintf("[%d:%d] ", cursorRow+1, cursorCol+1)
	}

	if clients := g.lspClient.SupportedClients(file.Name()); len(clients) > 0 {
		var clientNames []string
		for _, client := range clients {
			clientNames = append(clientNames, client.Name())
		}
		infoLine += fmt.Sprintf("lsp:%s ", strings.Join(clientNames, ","))
	}

	if language := file.Language(); language != nil {
		infoLine += fmt.Sprintf("%s ", language.Name)
		if language.Grammar != nil {
			infoLine += fmt.Sprintf("(ts:%s) ", language.Grammar.Name)
		}
	}

	infoLine += fmt.Sprintf("theme:%s %s %s", config.Theme.Name, file.LineEnding(), file.Encoding())

	fileName := file.Name()
	if workspace := g.editor.Workspace(); workspace != "" {
		workspaceName := filepath.Base(workspace)
		maxFileNameWidth := max(0, width-lipgloss.Width(workspaceName)-1-1-lipgloss.Width(infoLine))
		fileName = file.RelativeName(workspace)

		if lipgloss.Width(fileName) > maxFileNameWidth {
			baseName := filepath.Base(fileName)
			dirName := filepath.Dir(fileName)
			for {
				dirName = filepath.Dir(dirName)
				if dirName == "." {
					dirName = ""
					break
				}
				if lipgloss.Width(dirName)+lipgloss.Width(baseName)+4 < maxFileNameWidth {
					break
				}
			}

			if dirName == "" {
				fileName = fmt.Sprintf(".../%s", baseName)
			} else {
				fileName = fmt.Sprintf("%s/.../%s", dirName, baseName)
			}
		}
		fileName = fmt.Sprintf("%s/%s", workspaceName, fileName)
	}

	codeBar := fileName + strings.Repeat(" ", max(1, width-lipgloss.Width(fileName)-lipgloss.Width(infoLine))) + infoLine

	return config.Theme.Editor.CodeBarStyle.Width(g.width).Render(codeBar)
}
