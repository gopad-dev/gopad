package cmd

import (
	"embed"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/gopad"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor"
	"go.gopad.dev/gopad/gopad/lsp"
	"go.gopad.dev/gopad/internal/xio"
)

func NewRootCmd(version string, defaultConfigs embed.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "gopad [-c dir] [-w dir] [-d file] [-l file] [-p addr:port] [dir | file]",
		Short:                 "gopad",
		Long:                  "",
		Example:               "",
		DisableFlagsInUseLine: true,
		Args:                  cobra.ArbitraryArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			configDir, _ := cmd.Flags().GetString("config-dir")

			initConfig(configDir, defaultConfigs)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, _ := cmd.Flags().GetString("workspace")
			debug, _ := cmd.Flags().GetString("debug")
			debugLSP, _ := cmd.Flags().GetString("debug-lsp")
			pprof, _ := cmd.Flags().GetString("pprof")

			if debug != "" {
				if debug != "-" {
					var err error
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

			if err := editor.LoadLanguages(defaultConfigs); err != nil {
				log.Panicln("failed to load languages:", err)
			}

			lspClient := lsp.New(version, config.LSP, lspLogFile)
			e, err := gopad.New(lspClient, version, workspace, args)
			if err != nil {
				log.Panicln("failed to start gopad:", err)
			}

			p := tea.NewProgram(e, tea.WithAltScreen(), tea.WithFilter(lspClient.Filter))
			lspClient.SetProgram(p)
			log.Println("running gopad")
			if _, err = p.Run(); err != nil {
				log.Panicln("error while running gopad:", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringP("config-dir", "c", "", "set config directory")
	cmd.NonInheritedFlags().StringP("workspace", "w", "", "set workspace directory")
	cmd.NonInheritedFlags().StringP("debug", "d", "", "enable & set debug log file (use - for stdout)")
	cmd.NonInheritedFlags().StringP("debug-lsp", "l", "", "enable & set debug log file for lsp")
	cmd.NonInheritedFlags().StringP("pprof", "p", "", "enable & set pprof address:port")

	return cmd
}
