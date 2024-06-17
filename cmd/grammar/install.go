package grammar

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/gopad/config"
)

const remoteName = "origin"

func LibName(name string) string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("tree-sitter-%s.dll", name)
	default:
		return fmt.Sprintf("libtree-sitter-%s.so", name)
	}
}

func NewInstallCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:               "install [flags]... [languages]...",
		Short:             "Install Tree-Sitter grammars",
		Example:           "gopad grammar install go",
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := ensureGitInstalled(); err != nil {
				return err
			}
			return ensureTreeSitterInstalled()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			languages := filterLanguages(args, true)

			var wg sync.WaitGroup
			for _, language := range languages {
				wg.Add(1)

				go func() {
					defer wg.Done()
					err := installGrammar(cmd.Context(), config.Path, config.Languages.GrammarDir, *language.Grammar)

					var status string
					if err == nil {
						status = successStyle.Render("installed")
						if language.Grammar.Install != nil {
							status += fmt.Sprintf(" (%s)", infoStyle.Render(language.Grammar.Install.Hyperlink()))
						}
					} else {
						status = errorStyle.Render("error installing")
					}

					cmd.Printf("grammar %s: %s\n%s", grammarStyle.Render(language.Grammar.Hyperlink()), status, msgFromErr(err))
				}()
			}

			wg.Wait()

			return nil
		},
	}

	parent.AddCommand(cmd)
}

func installGrammar(ctx context.Context, configDir string, grammarPath string, grammar config.GrammarConfig) error {
	dir, err := os.MkdirTemp("", fmt.Sprintf("tree-sitter-%s-", grammar.Name))
	if err != nil {
		return fmt.Errorf("failed to create temp install dir: %w", err)
	}

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("failed to init temp git repo: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "remote", "add", remoteName, grammar.Install.Git)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git remote: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "fetch", "--depth", "1", remoteName, grammar.Install.Rev)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch from git remote: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "checkout", grammar.Install.Rev)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout git rev: %w", err)
	}

	grammarDir := "."
	if grammar.Install.SubDir != "" {
		grammarDir = grammar.Install.SubDir
	}
	cmd = exec.CommandContext(ctx, "tree-sitter", "build", "--output", LibName(grammar.Name), grammarDir)
	cmd.Dir = dir
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("failed to build tree sitter grammar library: %w", err)
	}

	file, err := os.OpenFile(filepath.Join(dir, LibName(grammar.Name)), os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open tree sitter grammar library: %w", err)
	}

	defer file.Close()

	installPath := grammar.Path
	if installPath == "" {
		if filepath.IsAbs(grammarPath) {
			installPath = grammarPath
		} else {
			installPath = filepath.Join(configDir, grammarPath)
		}
	}

	if err = os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create tree sitter grammar library install dir")
	}

	libFile, err := os.Create(filepath.Join(installPath, LibName(grammar.Name)))
	if err != nil {
		return fmt.Errorf("failed to create tree sitter grammar library: %w", err)
	}

	defer libFile.Close()

	if _, err = io.Copy(libFile, file); err != nil {
		return fmt.Errorf("failed to copy tree sitter grammar library: %w", err)
	}

	return nil
}
