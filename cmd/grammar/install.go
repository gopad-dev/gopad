package grammar

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"

	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/gopad/config"
)

func NewInstallCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "install",
		Short:   "Used to install your tree sitter grammars",
		Example: "gopad grammar install go",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var languages []config.LanguageConfig

			for name, language := range config.Languages {
				if len(args) > 0 && (!slices.Contains(args, name) || language.Grammar == nil || language.Grammar.Install == nil) {
					continue
				}
				languages = append(languages, language)
			}

			var wg sync.WaitGroup

			for _, language := range languages {
				wg.Add(1)

				go installLanguage(cmd.Context(), *language.Grammar)
			}

			wg.Wait()

			return nil
		},
	}

	parent.AddCommand(cmd)
}

const remoteName = "origin"

func installLanguage(ctx context.Context, grammar config.GrammarConfig) {
	dir, err := os.MkdirTemp("", grammar.Name)
	if err != nil {
		log.Printf("failed to create temp install dir for %q grammar", grammar.Name)
		return
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("failed to cleanup git temp dir for %q grammar at %Q\n", grammar.Name, dir)
		}
	}()

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		log.Printf("failed to init temp git repo for %q grammar\n", grammar.Name)
		return
	}

	cmd = exec.CommandContext(ctx, "git", "remote", "set", remoteName, grammar.Install.Git)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		log.Printf("failed to set git remote for %q grammar\n", grammar.Name)
		return
	}

	cmd = exec.CommandContext(ctx, "git", "fetch", "--depth", "1", remoteName, grammar.Install.Git, grammar.Install.Rev)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		log.Printf("failed to fetch from git remote for %q grammar\n", grammar.Name)
		return
	}

	cmd = exec.CommandContext(ctx, "git", "checkout", grammar.Install.Rev)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		log.Printf("failed to fetch from git remote for %q grammar\n", grammar.Name)
		return
	}

	if grammar.Install.Dir != "" {
		dir = filepath.Join(dir, grammar.Install.Dir)
	}

	cmd = exec.CommandContext(ctx, "tree-sitter", "build")
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		log.Printf("failed to build tree sitter grammar library for %q grammar\n", grammar.Name)
		return
	}

	os.
}
