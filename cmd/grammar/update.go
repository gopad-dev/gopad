package grammar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
	"sync"

	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/gopad/config"
)

func NewUpdateCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Used to update your tree sitter grammars",
		Example: "gopad grammar update go",
		Args:    cobra.ArbitraryArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureGitInstalled()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			languages := filterLanguages(args, true)

			var wg sync.WaitGroup
			for _, language := range languages {
				wg.Add(1)

				go func() {
					defer wg.Done()
					newRefs, err := checkUpdatedGrammar(cmd.Context(), *language.Grammar)

					var status string
					if err == nil {
						status = successStyle.Render("updated")
					} else if newRefs != nil {
						status = fmt.Sprintf("%s (%s -> %s)", errorStyle.Render("outdated"), errorStyle.Render(newRefs.rev), errorStyle.Render(newRefs.newRev))
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

type refs struct {
	rev    string
	newRev string
}

func checkUpdatedGrammar(ctx context.Context, grammar config.GrammarConfig) (*refs, error) {
	switch grammar.Install.RefType {
	case config.RefTypeCommit:
		return checkUpdatedGrammarCommit(ctx, grammar)
	case config.RefTypeTag:
		return checkUpdatedGrammarTag(ctx, grammar)
	}

	return nil, fmt.Errorf("unknown ref type %s", grammar.Install.RefType)
}

func checkUpdatedGrammarCommit(ctx context.Context, grammar config.GrammarConfig) (*refs, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--head", grammar.Install.Git)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating pipe: %w", err)
	}

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting command: %w", err)
	}

	data, err := io.ReadAll(pipe)
	if err != nil {
		return nil, fmt.Errorf("error reading pipe: %w", err)
	}

	if err = cmd.Wait(); err != nil {
		return nil, fmt.Errorf("error waiting for command: %w", err)
	}

	var revLine string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasSuffix(line, fmt.Sprintf("refs/heads/%s", grammar.Install.Ref)) {
			revLine = line
			break
		}
	}

	if revLine == "" {
		return nil, fmt.Errorf("ref %s of type %s not found", grammar.Install.Ref, grammar.Install.RefType)
	}

	fields := strings.Fields(revLine)
	if len(fields) < 1 {
		return nil, fmt.Errorf("error parsing ref line %s", revLine)
	}

	rev := fields[0]

	if rev != grammar.Install.Rev {
		return &refs{
			rev:    grammar.Install.Hyperlink(),
			newRev: termenv.Hyperlink(grammar.Install.Git+"/commit/"+rev, rev),
		}, errors.New("outdated")
	}

	return nil, nil
}

func checkUpdatedGrammarTag(ctx context.Context, grammar config.GrammarConfig) (*refs, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--tags", "--refs", grammar.Install.Git)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating pipe: %w", err)
	}

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting command: %w", err)
	}

	data, err := io.ReadAll(pipe)
	if err != nil {
		return nil, fmt.Errorf("error reading pipe: %w", err)
	}

	if err = cmd.Wait(); err != nil {
		return nil, fmt.Errorf("error waiting for command: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	slices.Reverse(lines)
	var revLine string
	for _, line := range lines {
		if len(line) > 0 {
			revLine = line
			break
		}
	}

	if revLine == "" {
		return nil, fmt.Errorf("ref %s of type %s not found", grammar.Install.Ref, grammar.Install.RefType)
	}

	fields := strings.Fields(revLine)
	if len(fields) < 1 {
		return nil, fmt.Errorf("error parsing ref line %s", revLine)
	}

	rev := fields[0]
	ref := strings.TrimPrefix(fields[1], "refs/tags/")

	if rev != grammar.Install.Rev || ref != grammar.Install.Ref {
		return &refs{
			rev:    grammar.Install.Hyperlink(),
			newRev: termenv.Hyperlink(grammar.Install.Git+"/releases/tag/"+ref, ref),
		}, errors.New("outdated")
	}

	return nil, nil
}
