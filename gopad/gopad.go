package gopad

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	height int
	width  int

	editor        editor.Editor
	overlays      overlay.Model
	notifications notifications.Model
}

func (g Gopad) Init() (tea.Model, tea.Cmd) {
	log.Printf("Initializing gopad, version: %s\n", g.version)

	var err error
	g.editor, err = editor.NewEditor(g.workspace, g.args)
	if err != nil {
		return g, notifications.Add(fmt.Sprintf("Error initializing editor: %s", err))
	}

	var cmd tea.Cmd
	g.editor, cmd = g.editor.Init()

	cmds := []tea.Cmd{
		cmd,
		g.editor.Focus(),
		tea.SetWindowTitle("gopad"),
		cursor.Blink,
	}

	g.overlays = config.NewOverlays()
	g.notifications = config.NewNotifications()

	return g, tea.Batch(cmds...)
}

func (g Gopad) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	now := time.Now()
	defer func() {
		log.Printf("Update time: %s\n", time.Since(now))
	}()

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g.height = msg.Height
		g.width = msg.Width
		return g, tea.Batch(cmds...)

	case overlay.ResetFocusMsg:
		cmds = append(cmds, g.editor.Focus())
		return g, tea.Batch(cmds...)

	case overlay.TakeFocusMsg:
		g.editor.Blur()
		return g, tea.Batch(cmds...)

	case tea.MouseMsg:
		log.Printf("MouseMsg: %#v\n", msg)
		switch {
		case mouse.Matches(msg, ZoneTheme, tea.MouseLeft):
			cmds = append(cmds, overlay.Open(NewSetThemeOverlay()))
			return g, tea.Batch(cmds...)
		}

	case tea.KeyMsg:
		log.Printf("KeyMsg: %s: %#v\n", msg.String(), msg)
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
	if g.notifications, cmd = g.notifications.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if focused := g.overlays.Focused(); focused || !bubbles.IsKeyMsg(msg) {
		if g.overlays, cmd = g.overlays.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if bubbles.IsKeyMsg(msg) && focused {
			return g, tea.Batch(cmds...)
		}
	}

	if g.editor.Focused() || !bubbles.IsKeyMsg(msg) {
		if g.editor, cmd = g.editor.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return g, tea.Batch(cmds...)
}

func (g Gopad) View() string {
	now := time.Now()
	defer func() {
		log.Printf("Render time: %s\n", time.Since(now))
	}()

	height := g.height

	appBar := g.AppBar()
	codeBar := g.CodeBar()
	codeEditor := g.editor.View(g.width, height-lipgloss.Height(appBar)-lipgloss.Height(codeBar))
	view := fmt.Sprintf("%s\n%s\n%s", appBar, codeEditor, codeBar)

	if g.overlays.Focused() {
		view = g.overlays.View(g.width, g.height, view)
	}

	if g.notifications.Active() {
		g.notifications.SetBackground(view)
		view = g.notifications.View(g.width, g.height)
	}

	return zone.Scan(view)
}

func (g Gopad) AppBar() string {
	width := g.width
	appBar := config.Theme.UI.AppBar.TitleStyle.Render("gopad-" + g.version)
	appBar += g.editor.FileTabsView(width - lipgloss.Width(appBar))

	return config.Theme.UI.AppBar.Style.Width(width).Render(appBar)
}

func (g Gopad) CodeBar() string {
	width := g.width
	contentWidth := width - config.Theme.UI.CodeBar.Style.GetHorizontalFrameSize()
	file := g.editor.File()

	barStyle := config.Theme.UI.CodeBar.Style
	inlineBarStyle := barStyle.Inline(true).Render

	var infoLine []string
	infoLine = append(infoLine, zone.Mark(ZoneTheme, inlineBarStyle(config.Theme.Name)))

	if file != nil {
		if s := file.Selection(); s != nil {
			infoLine = append(infoLine, zone.Mark(editor.ZoneFileGoTo, inlineBarStyle(fmt.Sprintf("%d lines | [%d:%d-%d:%d]", s.Lines(), s.Start.Row+1, s.Start.Col+1, s.End.Row+1, s.End.Col+1))))
		} else {
			cursorRow, cursorCol := file.Cursor()
			infoLine = append(infoLine, zone.Mark(editor.ZoneFileGoTo, inlineBarStyle(fmt.Sprintf("[%d:%d]", cursorRow+1, cursorCol+1))))
		}

		if servers := g.lsClient.SupportedServers(file.Name()); len(servers) > 0 {
			var clientNames []string
			for _, server := range servers {
				clientNames = append(clientNames, server.Name())
			}
			infoLine = append(infoLine, inlineBarStyle(strings.Join(clientNames, ",")))
		}

		if language := file.Language(); language != nil {
			name := language.Name
			icon := config.Theme.Icons.FileIcon(name).Render()

			name = icon + inlineBarStyle(" "+name)

			if language.Config.Grammar != nil {
				grammarName := language.Config.Grammar.Name
				if language.Grammar == nil {
					grammarName += " (not loaded)"
				}
				name += inlineBarStyle(fmt.Sprintf(" (ts: %s)", grammarName))
			}

			infoLine = append(infoLine, zone.Mark(editor.ZoneFileLanguage, inlineBarStyle(name)))
		}

		infoLine = append(infoLine,
			zone.Mark(editor.ZoneFileLineEnding, inlineBarStyle(file.LineEnding().String())),
			zone.Mark(editor.ZoneFileEncoding, inlineBarStyle(file.Encoding())),
		)
	}
	infoLineStr := strings.Join(infoLine, inlineBarStyle(" | "))

	maxWorkspaceNameWidth := max(0, contentWidth-1-lipgloss.Width(infoLineStr))
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

	codeBar := workspaceName + strings.Repeat(" ", max(1, contentWidth-lipgloss.Width(workspaceName)-lipgloss.Width(infoLineStr))) + infoLineStr

	return config.Theme.UI.CodeBar.Style.Width(width).Render(codeBar)
}

func joinPaths(dirName string, baseName string) string {
	if dirName == "" {
		return baseName
	}
	return fmt.Sprintf("%s/.../%s", dirName, baseName)
}
