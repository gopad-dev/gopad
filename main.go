package main

import (
	"context"
	"embed"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell/v2"

	"go.gopad.dev/gopad/cmd"
)

var (
	Version = "dev"
	Commit  = "unknown"

	//go:embed config/*
	defaultConfigs embed.FS
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	screen.Init()

	screen.SetStyle(tcell.StyleDefault)

	screen.Sync()

	rootCmd := cmd.NewRootCmd(Version, defaultConfigs)
	cmd.NewVersionCmd(rootCmd, Version, Commit)
	cmd.NewConfigCmd(rootCmd, defaultConfigs)
	cmd.NewGrammarCmd(rootCmd, defaultConfigs)
	cmd.NewCompletionCmd(rootCmd)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
