package lsp

import (
	"errors"
	"fmt"
	"io"
	"log"

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

func (m *LSP) SetProgram(p *tea.Program) {
	m.p = p
}

func (m *LSP) Close() error {
	var err error
	for _, client := range m.clients {
		if e := client.Stop(); e != nil {
			err = errors.Join(err, e)
		}
	}

	return err
}

func (m *LSP) Send(cmd tea.Cmd) {
	m.p.Send(cmd())
}

func (m *LSP) SupportedClients(name string) []*Client {
	var clients []*Client
	for _, client := range m.clients {
		if client.SupportedFile(name) {
			clients = append(clients, client)
		}
	}

	return clients
}

func (m *LSP) Filter(_ tea.Model, msg tea.Msg) tea.Msg {
	if cmd := m.update(msg); cmd != nil {
		return cmd()
	}
	return msg
}

func (m *LSP) update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case WorkspaceOpenedMsg:
		for _, client := range m.clients {
			cmds = append(cmds, client.Update(msg))
		}
		return tea.Batch(cmds...)
	case WorkspaceClosedMsg:
		for _, client := range m.clients {
			cmds = append(cmds, client.Update(msg))
		}
	case GetAutocompletionMsg:
		clients := m.SupportedClients(msg.File)
		if len(clients) == 0 {
			break
		}

		for _, client := range clients {
			cmds = append(cmds, client.Update(msg))
		}
	case FileOpenedMsg:
		clients := m.SupportedClients(msg.Name)
		if len(clients) == 0 {
			log.Println("No LSP client for", msg.Name)
			break
		}

		for _, client := range clients {
			if err := client.EnsureRunning(m.Send); err != nil {
				log.Println("Failed to start LSP for", msg.Name, err)
				cmds = append(cmds, notifications.Add(fmt.Sprintf("failed to start LSP for %s: %s", msg.Name, err)))
				continue
			}
			cmds = append(cmds, client.Update(msg))
		}
	case FileCreatedMsg:
		clients := m.SupportedClients(msg.Name)
		if len(clients) == 0 {
			break
		}

		for _, client := range clients {
			cmds = append(cmds, client.Update(msg))
		}
	case FileClosedMsg:
		clients := m.SupportedClients(msg.Name)
		if len(clients) == 0 {
			break
		}

		for _, client := range clients {
			cmds = append(cmds, client.Update(msg))
		}
	case FileChangedMsg:
		clients := m.SupportedClients(msg.Name)
		if len(clients) == 0 {
			break
		}

		for _, client := range clients {
			cmds = append(cmds, client.Update(msg))
		}
	case FileSavedMsg:
		clients := m.SupportedClients(msg.Name)
		if len(clients) == 0 {
			break
		}

		for _, client := range clients {
			cmds = append(cmds, client.Update(msg))
		}
	case FileRenamedMsg:
		clients := m.SupportedClients(msg.OldName)
		if len(clients) == 0 {
			break
		}

		for _, client := range clients {
			cmds = append(cmds, client.Update(msg))
		}
	case FileDeletedMsg:
		clients := m.SupportedClients(msg.Name)
		if len(clients) == 0 {
			break
		}

		for _, client := range clients {
			cmds = append(cmds, client.Update(msg))
		}
	}

	return tea.Batch(cmds...)
}
