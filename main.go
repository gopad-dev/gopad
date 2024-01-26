package main

import (
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/pflag"

	"go.gopad.dev/gopad/gopad"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor"
	"go.gopad.dev/gopad/gopad/lsp"
)

const (
	Version = "dev"
	Commit  = "unknown"
)

//go:embed config/*
var defaultConfigs embed.FS

func main() {
	help := pflag.BoolP("help", "h", false, "show help")
	version := pflag.BoolP("version", "v", false, "show version")
	debug := pflag.StringP("debug", "d", "", "enable & set debug log file (use - for stdout)")
	pprof := pflag.StringP("pprof", "p", "", "enable & set pprof address:port")
	configDir := pflag.StringP("config-dir", "c", "", "set config directory")
	createConfig := pflag.String("create-config", "", "create a new config file in the specified directory")
	pflag.Parse()

	if help != nil && *help {
		pflag.PrintDefaults()
		return
	}

	if version != nil && *version {
		fmt.Printf("gopad: %s (%s)", Version, Commit)
		return
	}

	var logFile *os.File
	if debug != nil && *debug != "" {
		if *debug != "-" {
			var err error
			logFile, err = tea.LogToFile(*debug, "gopad")
			if err != nil {
				log.Panicln("failed to open debug log file:", err)
			}
			defer logFile.Close()
		}
		log.Println("debug mode enabled")
	} else {
		log.SetOutput(io.Discard)
	}

	if createConfig != nil && *createConfig != "" {
		if err := config.Create(*createConfig, defaultConfigs); err != nil {
			log.Panicln("failed to create config:", err)
		}
		fmt.Println("created config in", *createConfig)
		return
	}

	if pprof != nil && *pprof != "" {
		log.Println("pprof enabled")
		go func() {
			if err := http.ListenAndServe(*pprof, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Println("failed to start pprof:", err)
			}
		}()
	}

	var actualConfigDir string
	if configDir == nil || *configDir == "" {
		var err error
		actualConfigDir, err = config.FindDir()
		if err != nil {
			log.Panicln("failed to find config dir:", err)
		}
	} else {
		actualConfigDir = *configDir
	}

	if err := config.Load(actualConfigDir, defaultConfigs); err != nil {
		log.Panicln("failed to load config:", err)
	}

	if err := editor.LoadLanguages(defaultConfigs); err != nil {
		log.Panicln("failed to load languages:", err)
	}

	lspClient := lsp.New(Version, config.LSP, logFile)
	e, err := gopad.New(lspClient, Version, pflag.Args())
	if err != nil {
		log.Panicln("failed to start gopad:", err)
	}

	p := tea.NewProgram(e, tea.WithAltScreen(), tea.WithFilter(lspClient.Filter))
	lspClient.SetProgram(p)
	tea.Println("running gopad")
	if _, err = p.Run(); err != nil {
		log.Panicln("error while running gopad:", err)
	}
}
