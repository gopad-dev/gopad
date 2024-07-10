package cmd

import (
	"embed"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/bubblezone"
	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/gopad"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/xio"
)

func NewRootCmd(version string, defaultConfigs embed.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "gopad [flags]... [dir | file]...",
		Short:                 "A terminal-based text editor with Tree-sitter and LSP support.",
		Long:                  "",
		Example:               "",
		DisableFlagsInUseLine: true,
		Args:                  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir, _ := cmd.Flags().GetString("config-dir")
			workspace, _ := cmd.Flags().GetString("workspace")
			debug, _ := cmd.Flags().GetString("debug")
			debugLSP, _ := cmd.Flags().GetString("debug-lsp")
			pprof, _ := cmd.Flags().GetString("pprof")
			mouse, _ := cmd.Flags().GetBool("mouse")

			if debug != "" {
				logFile, err := tea.LogToFile(debug, "gopad")
				if err != nil {
					log.Panicln("failed to open debug log file:", err)
				}
				defer logFile.Close()

				log.Println("debug mode enabled")
			} else {
				log.SetOutput(io.Discard)
			}

			var lspLogFile io.WriteCloser
			if debugLSP != "" {
				var err error
				lspLogFile, err = os.OpenFile(debugLSP, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
				if err != nil {
					log.Panicln("failed to open debug lsp log file:", err)
				}
				defer lspLogFile.Close()

				log.Println("debug lsp mode enabled")
			} else {
				lspLogFile = xio.NopCloser(io.Discard)
			}

			if pprof != "" {
				go func() {
					if err := http.ListenAndServe(pprof, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
						log.Println("failed to start pprof:", err)
					}
				}()
				log.Println("pprof enabled")
			}

			loadConfig(configDir, defaultConfigs)
			if err := file.LoadLanguages(defaultConfigs); err != nil {
				return err
			}

			lsClient := ls.New(version, config.LanguageServers, lspLogFile)
			e := gopad.New(lsClient, version, getWorkspace(workspace, args), args)

			opts := []tea.ProgramOption{
				tea.WithAltScreen(),
				tea.WithFilter(lsClient.Filter),
			}
			if mouse {
				zone.NewGlobal()
				defer zone.Close()
				opts = append(opts, tea.WithMouseCellMotion())
			}
			p := tea.NewProgram(e, opts...)
			lsClient.SetProgram(p)
			log.Println("running gopad")
			if _, err := p.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringP("config-dir", "c", "", "set configuration directory (Default: ./.gopad, $XDG_CONFIG_HOME/gopad or $HOME/.config/gopad)")
	cmd.Flags().StringP("workspace", "w", "", "set workspace directory (Default: first directory argument)")
	cmd.Flags().StringP("debug", "d", "", "set debug log file")
	cmd.Flags().StringP("debug-lsp", "l", "", "set debug lsp log file")
	cmd.Flags().StringP("pprof", "p", "", "set pprof address:port")
	cmd.Flags().BoolP("mouse", "m", true, "enable mouse support (Default: true)")

	return cmd
}

func getWorkspace(workspace string, args []string) string {
	if workspace == "" {
		for _, arg := range args {
			stat, err := os.Stat(arg)
			if err == nil && stat.IsDir() {
				workspace = arg
				break
			}
		}

		if workspace == "" {
			return ""
		}
	}

	if absWorkspace, err := filepath.Abs(workspace); err == nil {
		workspace = absWorkspace
	}

	return workspace
}
