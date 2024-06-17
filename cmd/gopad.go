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
	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/gopad"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor"
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

			if debug != "" {
				if debug != "-" {
					logFile, err := tea.LogToFile(debug, "gopad")
					if err != nil {
						log.Panicln("failed to open debug log file:", err)
					}
					defer logFile.Close()
				}
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
			} else {
				lspLogFile = xio.NopCloser(io.Discard)
			}

			if pprof != "" {
				log.Println("pprof enabled")
				go func() {
					if err := http.ListenAndServe(pprof, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
						log.Println("failed to start pprof:", err)
					}
				}()
			}

			loadConfig(configDir, defaultConfigs)

			if err := editor.LoadLanguages(defaultConfigs); err != nil {
				log.Panicln("failed to load languages:", err)
			}

			editorWorkspace := getWorkspace(workspace, args)
			log.Printf("workspace: %q\n", workspace)
			log.Printf("editorWorkspace: %q\n", editorWorkspace)

			lsClient := ls.New(version, config.LanguageServers, lspLogFile)
			e, err := gopad.New(lsClient, version, editorWorkspace, args)
			if err != nil {
				log.Panicln("failed to start gopad:", err)
			}

			p := tea.NewProgram(e, tea.WithAltScreen(), tea.WithFilter(lsClient.Filter))
			lsClient.SetProgram(p)
			log.Println("running gopad")
			if _, err = p.Run(); err != nil {
				log.Panicln("error while running gopad:", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringP("config-dir", "c", "", "set configuration directory (Default: ./.gopad, $XDG_CONFIG_HOME/gopad or $HOME/.config/gopad)")
	cmd.Flags().StringP("workspace", "w", "", "set workspace directory (Default: first directory argument)")
	cmd.Flags().StringP("debug", "d", "", "set debug log file (use - for stdout)")
	cmd.Flags().StringP("debug-lsp", "l", "", "set debug lsp log file")
	cmd.Flags().StringP("pprof", "p", "", "set pprof address:port")

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
