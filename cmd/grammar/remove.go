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

func NewRemoveCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "Used to remove your tree sitter grammars",
		Example: "gopad grammar remove go",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			languages := filterLanguages(args, false)

			var wg sync.WaitGroup
			for _, language := range languages {
				wg.Add(1)

				go func() {
					defer wg.Done()
					err := removeGrammar(config.Path, config.Languages.GrammarDir, *language.Grammar)

					var status string
					if err == nil {
						status = successStyle.Render("removed")
					} else if errors.Is(err, os.ErrNotExist) {
						status = errorStyle.Render("not installed")
						err = nil
					} else {
						status = errorStyle.Render("error removing")
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

func removeGrammar(configDir string, grammarPath string, grammar config.GrammarConfig) error {
	installPath := grammar.Path
	if installPath == "" {
		if filepath.IsAbs(grammarPath) {
			installPath = grammarPath
		} else {
			installPath = filepath.Join(configDir, grammarPath)
		}
	}

	grammarFilePath := filepath.Join(installPath, fmt.Sprintf("libtree-sitter-%s.so", grammar.Name))

	_, err := os.Stat(grammarFilePath)
	if err != nil {
		return fmt.Errorf("error checking grammar: %w", err)
	}

	if err = os.Remove(grammarFilePath); err != nil {
		return fmt.Errorf("error removing grammar: %w", err)
	}

	return nil
}
