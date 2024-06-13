package lsp

import (
	"errors"
	"io"
	"log"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/config"
)

func New(version string, cfg config.LanguageServerConfigs, w io.Writer) *LSP {
	lsp := &LSP{
		registry: make(map[string]ClientConfig, len(cfg.LanguageServers)),
	}

	for name, serverCfg := range cfg.LanguageServers {
		log.Println("registering lsp client for", name)
		lsp.registry[name] = ClientConfig{
			name: name,
			cfg:  serverCfg,
			new: func(name string, cfg config.LanguageServerConfig, workspace string) (*Client, error) {
				return newClient(name, lsp.send, workspace, version, cfg, w)
			},
		}
	}

	return lsp
}

type LSP struct {
	registry map[string]ClientConfig
	clients  []*Client
	p        *tea.Program
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

func (l *LSP) send(cmd tea.Cmd) {
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

func (l *LSP) updateSupportedClients(name string, msg tea.Msg) []tea.Cmd {
	clients := l.SupportedClients(name)

	var cmds []tea.Cmd
	for _, client := range clients {
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
		for _, registry := range l.registry {
			if !registry.Supported(msg.Workspace) {
				continue
			}

			client, err := registry.New(msg.Workspace)
			if err != nil {
				log.Printf("failed to create client for %s: %v", registry.name, err)
				continue
			}

			l.clients = append(l.clients, client)
		}
	case WorkspaceClosedMsg:
		for _, client := range l.clients {
			if client.workspace == msg.Workspace {
				if err := client.Stop(); err != nil {
					log.Printf("failed to stop client %s: %v", client.Name(), err)
				}
			}
		}
	case GetAutocompletionMsg:
		cmds = append(cmds, l.updateSupportedClients(msg.Name, msg)...)

	case FileOpenedMsg:
		cmds = append(cmds, l.updateSupportedClients(msg.Name, msg)...)

	case FileCreatedMsg:
		cmds = append(cmds, l.updateSupportedClients(msg.Name, msg)...)

	case FileClosedMsg:
		cmds = append(cmds, l.updateSupportedClients(msg.Name, msg)...)

	case FileChangedMsg:
		cmds = append(cmds, l.updateSupportedClients(msg.Name, msg)...)

	case FileSavedMsg:
		cmds = append(cmds, l.updateSupportedClients(msg.Name, msg)...)

	case FileRenamedMsg:
		cmds = append(cmds, l.updateSupportedClients(msg.OldName, msg)...)

	case FileDeletedMsg:
		cmds = append(cmds, l.updateSupportedClients(msg.Name, msg)...)
	case GetInlayHintMsg:
		cmds = append(cmds, l.updateSupportedClients(msg.Name, msg)...)
	}

	return tea.Batch(cmds...)
}
