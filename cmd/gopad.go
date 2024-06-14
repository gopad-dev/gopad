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

			lsClient := ls.New(version, config.LanguageServers, lspLogFile)
			e, err := gopad.New(lsClient, version, getWorkspace(workspace, args), args)
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

	cmd.PersistentFlags().StringP("config-dir", "c", "", "set config directory")
	cmd.Flags().StringP("workspace", "w", "", "set workspace directory")
	cmd.Flags().StringP("debug", "d", "", "enable & set debug log file (use - for stdout)")
	cmd.Flags().StringP("debug-lsp", "l", "", "enable & set debug log file for lsp")
	cmd.Flags().StringP("pprof", "p", "", "enable & set pprof address:port")

	return cmd
}

func getWorkspace(workspace string, args []string) string {
	if workspace == "" {
		if len(args) > 0 {
			for _, arg := range args {
				stat, err := os.Stat(arg)
				if err == nil && stat.IsDir() {
					workspace = arg
				}
			}
		} else {
			dir, err := os.Getwd()
			if err == nil {
				workspace = dir
			}
		}

		absWorkspace, err := filepath.Abs(workspace)
		if err == nil {
			workspace = absWorkspace
		}
	}

	return workspace
}
