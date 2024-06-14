package ls

import (
	"context"
	"errors"
	"io"
	"log"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/config"
)

func New(version string, cfg config.LanguageServerConfigs, w io.Writer) *Client {
	c := &Client{
		registry: make(map[string]ServerConfig, len(cfg.LanguageServers)),
	}

	for name, serverCfg := range cfg.LanguageServers {
		log.Println("registering language server for", name)
		c.registry[name] = ServerConfig{
			name: name,
			cfg:  serverCfg,
			new: func(name string, cfg config.LanguageServerConfig, workspace string) (*Server, error) {
				return newServer(name, c.send, workspace, version, cfg, w)
			},
		}
	}

	return c
}

type Client struct {
	registry map[string]ServerConfig
	servers  []*Server
	p        *tea.Program
}

func (l *Client) SetProgram(p *tea.Program) {
	l.p = p
}

func (l *Client) Close() error {
	var errs []error
	for _, server := range l.servers {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Stop(ctx); err != nil {
				errs = append(errs, err)
				log.Printf("failed to stop server %s: %v", server.Name(), err)
			}
		}()
	}

	return errors.Join(errs...)
}

func (l *Client) send(cmd tea.Cmd) {
	l.p.Send(cmd())
}

func (l *Client) SupportedServers(name string) []*Server {
	var servers []*Server
	for _, server := range l.servers {
		if server.SupportedFile(name) {
			servers = append(servers, server)
		}
	}

	slices.SortFunc(servers, func(a, b *Server) int {
		return strings.Compare(a.Name(), b.Name())
	})

	return servers
}

func (l *Client) updateSupportedServers(name string, msg tea.Msg) []tea.Cmd {
	servers := l.SupportedServers(name)

	var cmds []tea.Cmd
	for _, server := range servers {
		cmds = append(cmds, server.Update(msg))
	}

	return cmds
}

func (l *Client) Filter(_ tea.Model, msg tea.Msg) tea.Msg {
	if cmd := l.update(msg); cmd != nil {
		return cmd()
	}
	return msg
}

func (l *Client) update(msg tea.Msg) tea.Cmd {
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

			l.servers = append(l.servers, client)
		}
	case WorkspaceClosedMsg:
		for _, server := range l.servers {
			if server.workspace == msg.Workspace {
				func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					if err := server.Stop(ctx); err != nil {
						log.Printf("failed to stop server %s: %v", server.Name(), err)
					}
				}()
			}
		}
	case GetAutocompletionMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)

	case FileOpenedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)

	case FileCreatedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)

	case FileClosedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)

	case FileChangedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)

	case FileSavedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)

	case FileRenamedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.OldName, msg)...)

	case FileDeletedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case GetInlayHintMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	}

	return tea.Batch(cmds...)
}
