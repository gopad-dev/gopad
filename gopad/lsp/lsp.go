package lsp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/internal/bubbles/notifications"

	"go.gopad.dev/gopad/gopad/config"
)

func New(version string, cfg config.LSPConfig, w io.Writer) *LSP {
	clients := make(map[string]*Client, len(cfg))
	for name, serverCfg := range cfg {
		log.Println("Creating lsp client for", name)
		clients[name] = newClient(name, version, serverCfg, w)
	}
	return &LSP{
		clients: clients,
	}
}

type LSP struct {
	clients map[string]*Client
	p       *tea.Program
}

func (l *LSP) SetProgram(p *tea.Program) {
	l.p = p
}

func (l *LSP) Close() error {
	var err error
	for _, client := range l.clients {
		if e := client.Stop(); e != nil {
			err = errors.Join(err, e)
		}
	}

	return err
}

func (l *LSP) Send(cmd tea.Cmd) {
	l.p.Send(cmd())
}

func (l *LSP) SupportedClients(name string) []*Client {
	var clients []*Client
	for _, client := range l.clients {
		if client.SupportedFile(name) {
			clients = append(clients, client)
		}
	}

	slices.SortFunc(clients, func(a, b *Client) int {
		return strings.Compare(a.Name(), b.Name())
	})

	return clients
}

func (l *LSP) UpdateSupportedClients(name string, msg tea.Msg) []tea.Cmd {
	clients := l.SupportedClients(name)

	var cmds []tea.Cmd
	for _, client := range clients {
		cmds = append(cmds, client.Update(msg))
	}

	return cmds
}

func (l *LSP) UpdateClients(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	for _, client := range l.clients {
		cmds = append(cmds, client.Update(msg))
	}
	return cmds
}

func (l *LSP) Filter(_ tea.Model, msg tea.Msg) tea.Msg {
	if cmd := l.update(msg); cmd != nil {
		return cmd()
	}
	return msg
}

func (l *LSP) update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case WorkspaceOpenedMsg:
		for _, client := range l.clients {
			cmds = append(cmds, client.Update(msg))
		}
		return tea.Batch(cmds...)
	case WorkspaceClosedMsg:
		for _, client := range l.clients {
			cmds = append(cmds, client.Update(msg))
		}
	case GetAutocompletionMsg:
		cmds = append(cmds, l.UpdateSupportedClients(msg.Name, msg)...)

	case FileOpenedMsg:
		clients := l.SupportedClients(msg.Name)
		if len(clients) == 0 {
			log.Println("No LSP client for", msg.Name)
			break
		}

		for _, client := range clients {
			if err := client.EnsureRunning(l.Send); err != nil {
				log.Println("Failed to start LSP for", msg.Name, err)
				cmds = append(cmds, notifications.Add(fmt.Sprintf("failed to start LSP for %s: %s", msg.Name, err)))
				continue
			}
			cmds = append(cmds, client.Update(msg))
		}
	case FileCreatedMsg:
		cmds = append(cmds, l.UpdateSupportedClients(msg.Name, msg)...)

	case FileClosedMsg:
		cmds = append(cmds, l.UpdateSupportedClients(msg.Name, msg)...)

	case FileChangedMsg:
		cmds = append(cmds, l.UpdateSupportedClients(msg.Name, msg)...)

	case FileSavedMsg:
		cmds = append(cmds, l.UpdateSupportedClients(msg.Name, msg)...)

	case FileRenamedMsg:
		cmds = append(cmds, l.UpdateSupportedClients(msg.OldName, msg)...)

	case FileDeletedMsg:
		cmds = append(cmds, l.UpdateSupportedClients(msg.Name, msg)...)
	case GetInlayHintMsg:
		cmds = append(cmds, l.UpdateSupportedClients(msg.Name, msg)...)
	}

	return tea.Batch(cmds...)
}
