package ls

import (
	"context"
	"errors"
	"io"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/gopad/config"
)

func New(version string, cfg config.LanguageServerConfigs, w io.Writer) *Client {
	c := &Client{
		registry: make(map[string]ServerConfig, len(cfg.LanguageServers)),
	}

	for name, serverCfg := range cfg.LanguageServers {
		log.Println("registering language server for", name)
		c.registry[name] = ServerConfig{
			Name: name,
			Cfg:  serverCfg,
			new: func(name string, cfg config.LanguageServerConfig, workspace string) (*Server, error) {
				return newServer(name, c.send, workspace, version, cfg, w)
			},
		}
	}

	log.Println("registered language servers")

	return c
}

type Client struct {
	registry  map[string]ServerConfig
	servers   []*Server
	serversMu sync.Mutex
	p         *tea.Program
	workspace string
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

func (l *Client) Servers() []ServerConfig {
	servers := make([]ServerConfig, 0, len(l.registry))
	for _, server := range l.registry {
		servers = append(servers, server)
	}

	slices.SortFunc(servers, func(a, b ServerConfig) int {
		return strings.Compare(a.Name, b.Name)
	})

	return servers
}

func (l *Client) SupportedServers(name string) []*Server {
	var servers []*Server

	l.serversMu.Lock()
	defer l.serversMu.Unlock()

	for _, server := range l.servers {
		if server.cfg.SupportsFile(name) {
			servers = append(servers, server)
		}
	}

	for _, serverConfig := range l.registry {
		if !serverConfig.Cfg.SupportsFile(name) {
			continue
		}

		// Check if the server is already running
		if slices.ContainsFunc(l.servers, func(server *Server) bool {
			return server.Name() == serverConfig.Name
		}) {
			continue
		}

		server, err := serverConfig.New(l.workspace)
		if err != nil {
			log.Printf("failed to create client for %s: %v", serverConfig.Name, err)
			continue
		}

		l.servers = append(servers, server)
		servers = append(servers, server)
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

func (l *Client) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case StartServerMsg:
		log.Println("trying to start server", msg.Name)
		serverConfig, ok := l.registry[msg.Name]
		if !ok {
			log.Printf("server %s not found", msg.Name)
			return Err(errors.New("server not found"))
		}

		l.serversMu.Lock()
		server, err := serverConfig.New(msg.Workspace)
		if err != nil {
			l.serversMu.Unlock()
			log.Printf("failed to create server for %s: %v", serverConfig.Name, err)
			return Err(err)
		}

		l.servers = append(l.servers, server)
		l.serversMu.Lock()
	case StopServerMsg:
		for i, server := range l.servers {
			if server.Name() == msg.Name {
				func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					if err := server.Stop(ctx); err != nil {
						log.Printf("failed to stop server %s: %v", server.Name(), err)
					}
				}()

				l.servers = append(l.servers[:i], l.servers[i+1:]...)
				break
			}
		}
	case WorkspaceOpenedMsg:
		l.workspace = msg.Workspace

		l.serversMu.Lock()
		defer l.serversMu.Unlock()
		for _, serverConfig := range l.registry {
			if !serverConfig.Cfg.SupportsWorkspace(msg.Workspace) {
				continue
			}

			// Check if the server is already running
			if slices.ContainsFunc(l.servers, func(server *Server) bool {
				return server.Name() == serverConfig.Name
			}) {
				continue
			}

			server, err := serverConfig.New(msg.Workspace)
			if err != nil {
				log.Printf("failed to create server for %s: %v", serverConfig.Name, err)
				continue
			}

			l.servers = append(l.servers, server)
		}
	case WorkspaceClosedMsg:
		l.workspace = ""
		for i := len(l.servers) - 1; i >= 0; i-- {
			server := l.servers[i]
			if server.workspace != msg.Workspace {
				continue
			}
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := server.Stop(ctx); err != nil {
					log.Printf("failed to stop server %s: %v", server.Name(), err)
				}
			}()
			l.servers = append(l.servers[:i], l.servers[i+1:]...)
		}
	case GetDeclarationMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case GetDefinitionMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case GetTypeDefinitionMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case GetImplementationsMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case GetReferencesMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case GetInlayHintMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case GetAutocompletionMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case FileOpenedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case FileClosedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case FileCreatedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case FileDeletedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case FileRenamedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.OldName, msg)...)
	case FileChangedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	case FileSavedMsg:
		cmds = append(cmds, l.updateSupportedServers(msg.Name, msg)...)
	}

	return tea.Batch(cmds...)
}
