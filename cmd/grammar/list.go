package grammar

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/gopad/config"
)

func NewListCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "list [flags]... [languages]...",
		Short:   "List configured Tree-Sitter grammars",
		Example: "gopad grammar list go",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			languages := filterLanguages(args, false)

			var wg sync.WaitGroup
			for _, language := range languages {
				wg.Add(1)

				go func() {
					defer wg.Done()
					err := checkInstalledGrammar(config.Path, config.Languages.GrammarDir, *language.Grammar)

					var status string
					if err == nil {
						status = successStyle.Render("installed")
						if language.Grammar.Install != nil {
							status += fmt.Sprintf(" (%s)", infoStyle.Render(language.Grammar.Install.Hyperlink()))
						}
					} else if errors.Is(err, os.ErrNotExist) {
						status = errorStyle.Render("not installed")
						err = nil
					} else {
						status = errorStyle.Render("error checking")
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

func checkInstalledGrammar(configDir string, grammarPath string, grammar config.GrammarConfig) error {
	installPath := grammar.Path
	if installPath == "" {
		if filepath.IsAbs(grammarPath) {
			installPath = grammarPath
		} else {
			installPath = filepath.Join(configDir, grammarPath)
		}
	}

	_, err := os.Stat(filepath.Join(installPath, fmt.Sprintf("libtree-sitter-%s.so", grammar.Name)))
	if err != nil {
		return fmt.Errorf("failed to check %q grammar: %w", grammar.Name, err)
	}
	return nil
}
