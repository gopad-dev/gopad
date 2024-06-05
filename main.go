package main

import (
	"context"
	"embed"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"go.gopad.dev/gopad/cmd"
)

var (
	Version = "dev"
	Commit  = "unknown"

	//go:embed config/*
	defaultConfigs embed.FS
)

func main() {
	rootCmd := cmd.NewRootCmd(Version, defaultConfigs)
	cmd.NewVersionCmd(rootCmd, Version, Commit)
	cmd.NewConfigCmd(rootCmd, defaultConfigs)
	cmd.NewGrammarCmd(rootCmd)
	cmd.NewCompletionCmd(rootCmd)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
