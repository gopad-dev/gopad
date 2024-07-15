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
	"github.com/lrstanley/bubblezone"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles"
	"go.gopad.dev/gopad/internal/bubbles/cursor"
	"go.gopad.dev/gopad/internal/bubbles/mouse"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const (
	ZoneTheme = "theme"
)

func New(lsClient *ls.Client, version string, workspace string, args []string) *Gopad {
	return &Gopad{
		lsClient:  lsClient,
		version:   version,
		workspace: workspace,
		args:      args,
	}
}

type Gopad struct {
	lsClient  *ls.Client
	version   string
	workspace string
	args      []string

	editor        editor.Editor
	overlays      overlay.Model
	notifications notifications.Model
}

func (g Gopad) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	log.Printf("Initializing gopad, version: %s, color profile: %s\n", g.version, ctx.ColorProfile())
	config.InitTheme(ctx)

	e, err := editor.NewEditor(g.workspace, g.args)
	if err != nil {
		return g, notifications.Add(fmt.Sprintf("Error initializing editor: %s", err))
	}

	cmds := []tea.Cmd{e.Focus()}

	g.editor = *e
	g.overlays = config.NewOverlays()
	g.notifications = config.NewNotifications()

	return g, tea.Batch(append(cmds,
		tea.SetWindowTitle("gopad"),
		cursor.Blink,
		g.editor.Init(ctx),
	)...)
}

func (g Gopad) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case overlay.ResetFocusMsg:
		cmds = append(cmds, g.editor.Focus())
		return g, tea.Batch(cmds...)

	case overlay.TakeFocusMsg:
		g.editor.Blur()
		return g, tea.Batch(cmds...)

	case tea.MouseEvent, tea.MouseDownMsg, tea.MouseMotionMsg:
		log.Printf("MouseMsg: %#v\n", msg)

	case tea.MouseUpMsg:
		log.Printf("MouseMsg: %#v\n", msg)
		switch {
		case mouse.Matches(tea.MouseEvent(msg), ZoneTheme, tea.MouseLeft):
			cmds = append(cmds, overlay.Open(NewSetThemeOverlay()))
			return g, tea.Batch(cmds...)
		}

	case tea.KeyMsg:
		log.Printf("KeyMsg: %#v\n", msg)
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
			case key.Matches(msg, config.Keys.Editor.File.Open):
				if !g.overlays.Has(editor.OpenOverlayID) {
					path, err := os.Getwd()
					if err != nil {
						cmds = append(cmds, notifications.Add(fmt.Sprintf("Error getting current working directory: %s", err)))
						return g, tea.Batch(cmds...)
					}
					cmds = append(cmds, overlay.Open(editor.NewOpenOverlay(path, true, false)))
				}
			case key.Matches(msg, config.Keys.Editor.File.OpenFolder):
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
	if g.notifications, cmd = g.notifications.Update(ctx, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if g.overlays, cmd = g.overlays.Update(ctx, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if bubbles.IsInputMsg(msg) && g.overlays.Focused() {
		return g, tea.Batch(cmds...)
	}

	if g.editor, cmd = g.editor.Update(ctx, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return g, tea.Batch(cmds...)
}

func (g Gopad) View(ctx tea.Context) string {
	// now := time.Now()
	// defer func() {
	//	log.Printf("Render time: %s\n", time.Since(now))
	// }()

	_, height := ctx.WindowSize()

	appBar := g.AppBar(ctx)
	codeBar := g.CodeBar(ctx)
	codeEditor := g.editor.View(ctx, height-lipgloss.Height(appBar)-lipgloss.Height(codeBar))
	view := fmt.Sprintf("%s\n%s\n%s", appBar, codeEditor, codeBar)

	if g.overlays.Focused() {
		view = g.overlays.View(ctx, view)
	}

	if g.notifications.Active() {
		g.notifications.SetBackground(view)
		view = g.notifications.View(ctx)
	}

	return zone.Scan(view)
}

func (g Gopad) AppBar(ctx tea.Context) string {
	width, _ := ctx.WindowSize()
	appBar := config.Theme.UI.AppBar.TitleStyle.Render("gopad-" + g.version)
	appBar += g.editor.FileTabsView(width - lipgloss.Width(appBar))

	return config.Theme.UI.AppBar.Style.Width(width).Render(appBar)
}

func (g Gopad) CodeBar(ctx tea.Context) string {
	width, _ := ctx.WindowSize()
	contentWidth := width - config.Theme.UI.CodeBar.Style.GetHorizontalFrameSize()
	file := g.editor.File()

	infoLine := fmt.Sprintf("%s | ", zone.Mark(ZoneTheme, config.Theme.Name))

	if file != nil {
		if s := file.Selection(); s != nil {
			infoLine += fmt.Sprintf("%d [%d:%d-%d:%d] | ", s.Lines(), s.Start.Row+1, s.Start.Col+1, s.End.Row+1, s.End.Col+1)
		} else {
			cursorRow, cursorCol := file.Cursor()
			infoLine += fmt.Sprintf("[%d:%d] | ", cursorRow+1, cursorCol+1)
		}

		if servers := g.lsClient.SupportedServers(file.Name()); len(servers) > 0 {
			var clientNames []string
			for _, server := range servers {
				clientNames = append(clientNames, server.Name())
			}
			infoLine += fmt.Sprintf("%s | ", strings.Join(clientNames, ","))
		}

		if language := file.Language(); language != nil {
			name := language.Name
			icon := config.Theme.Icons.FileIcon(name).Render()

			name = fmt.Sprintf("%s %s", icon, name)

			if language.Config.Grammar != nil {
				grammarName := language.Config.Grammar.Name
				if language.Grammar == nil {
					grammarName += " (not loaded)"
				}
				name = zone.Mark(editor.ZoneFileLanguage, fmt.Sprintf("%s (ts: %s)", name, grammarName))
			}

			infoLine += fmt.Sprintf("%s | ", name)
		}

		infoLine += fmt.Sprintf("%s | %s", zone.Mark(editor.ZoneFileLineEnding, file.LineEnding().String()), zone.Mark(editor.ZoneFileEncoding, file.Encoding()))
	}
	infoLine = strings.TrimSuffix(infoLine, " | ")

	maxWorkspaceNameWidth := max(0, contentWidth-1-lipgloss.Width(infoLine))
	workspaceName := g.editor.Workspace()
	if workspaceName != "" {
		if file != nil {
			workspaceName = filepath.Join(filepath.Base(workspaceName), file.RelativeName(workspaceName))
		}
	} else if file != nil {
		workspaceName = file.Name()
	}

	if maxWorkspaceNameWidth > 0 && lipgloss.Width(workspaceName) > maxWorkspaceNameWidth {
		dirName := filepath.Dir(workspaceName)
		baseName := filepath.Base(workspaceName)
		for {
			dirName = filepath.Dir(dirName)
			if dirName == "." || dirName == "/" {
				dirName = ""
				break
			}
			if lipgloss.Width(joinPaths(dirName, baseName)) <= maxWorkspaceNameWidth {
				break
			}
		}
		workspaceName = joinPaths(dirName, baseName)
	}

	codeBar := workspaceName + strings.Repeat(" ", max(1, contentWidth-lipgloss.Width(workspaceName)-lipgloss.Width(infoLine))) + infoLine

	return config.Theme.UI.CodeBar.Style.Width(width).Render(codeBar)
}

func joinPaths(dirName string, baseName string) string {
	if dirName == "" {
		return baseName
	}
	return fmt.Sprintf("%s/.../%s", dirName, baseName)
}
