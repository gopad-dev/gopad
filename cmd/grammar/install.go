package grammar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"

	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/gopad/config"
)

const remoteName = "origin"

func NewInstallCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "install",
		Short:   "Used to install your tree sitter grammars",
		Example: "gopad grammar install go",
		Args:    cobra.ArbitraryArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureGitInstalled(cmd.Context()); err != nil {
				return err
			}
			if err := ensureTreeSitterInstalled(cmd.Context()); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var languages []config.LanguageConfig

			for name, language := range config.Languages.Languages {
				if len(args) > 0 && !slices.Contains(args, name) {
					continue
				}
				if language.Grammar == nil || language.Grammar.Install == nil {
					continue
				}
				languages = append(languages, language)
			}

			var wg sync.WaitGroup
			for _, language := range languages {
				wg.Add(1)

				go func() {
					defer wg.Done()
					installLanguage(cmd.Context(), config.Path, config.Languages.GrammarDir, *language.Grammar)
				}()
			}

			wg.Wait()

			return nil
		},
	}

	parent.AddCommand(cmd)
}

func ensureGitInstalled(ctx context.Context) error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed: %w", err)
	}
	return nil
}

func ensureTreeSitterInstalled(ctx context.Context) error {
	_, err := exec.LookPath("tree-sitter")
	if err != nil {
		return fmt.Errorf("tree-sitter is not installed: %w", err)
	}
	return nil
}

func installLanguage(ctx context.Context, configDir string, grammarPath string, grammar config.GrammarConfig) {
	log.Printf("installing %q grammar\n", grammar.Name)

	dir, err := os.MkdirTemp("", fmt.Sprintf("tree-sitter-%s-", grammar.Name))
	if err != nil {
		log.Printf("failed to create temp install dir for %q grammar: %s\n", grammar.Name, err.Error())
		return
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("failed to cleanup git temp dir for %q grammar at %q\n", grammar.Name, dir)
		}
	}()

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		logCmdErrorf(err, "failed to init temp git repo for %q grammar: %s\n", grammar.Name, err.Error())
		return
	}

	cmd = exec.CommandContext(ctx, "git", "remote", "add", remoteName, grammar.Install.Git)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		logCmdErrorf(err, "failed to set git remote for %q grammar: %s\n", grammar.Name, err.Error())
		return
	}

	cmd = exec.CommandContext(ctx, "git", "fetch", "--depth", "1", remoteName, grammar.Install.Rev)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		logCmdErrorf(err, "failed to fetch from git remote for %q grammar: %s\n", grammar.Name, err.Error())
		return
	}

	cmd = exec.CommandContext(ctx, "git", "checkout", grammar.Install.Rev)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		logCmdErrorf(err, "failed to fetch from git remote for %q grammar: %s\n", grammar.Name, err.Error())
		return
	}

	grammarDir := "."
	if grammar.Install.SubDir != "" {
		grammarDir = grammar.Install.SubDir
	}
	cmd = exec.CommandContext(ctx, "tree-sitter", "build", "--output", "grammar.so", grammarDir)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		logCmdErrorf(err, "failed to build tree sitter grammar library for %q grammar: %s\n", grammar.Name, err.Error())
		return
	}

	file, err := os.OpenFile(filepath.Join(dir, "grammar.so"), os.O_RDONLY, 0)
	if err != nil {
		log.Printf("failed to open tree sitter grammar library for %q grammar: %s\n", grammar.Name, err.Error())
		return
	}

	defer file.Close()

	installPath := grammar.Install.Path
	if installPath == "" {
		if filepath.IsAbs(grammarPath) {
			installPath = grammarPath
		} else {
			installPath = filepath.Join(configDir, grammarPath)
		}
	}

	if err = os.MkdirAll(installPath, 0755); err != nil {
		log.Printf("failed to create tree sitter grammar library install dir for %q grammar: %s\n", grammar.Name, err.Error())
		return
	}

	libFile, err := os.Create(filepath.Join(installPath, fmt.Sprintf("libtree-sitter-%s.so", grammar.Name)))
	if err != nil {
		log.Printf("failed to create tree sitter grammar library for %q grammar: %s\n", grammar.Name, err.Error())
		return
	}

	defer libFile.Close()

	if _, err = io.Copy(libFile, file); err != nil {
		log.Printf("failed to copy tree sitter grammar library for %q grammar: %s\n", grammar.Name, err.Error())
		return
	}
}

func logCmdErrorf(err error, message string, a ...any) {
	var stderr []byte

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		stderr = exitErr.Stderr
	}
	msg := fmt.Sprintf(message, a...)
	if len(stderr) > 0 {
		msg += fmt.Sprintf("stderr: %s\n", stderr)
	}
	log.Println(msg)
}
