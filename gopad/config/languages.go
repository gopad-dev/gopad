package config

import (
	"slices"

	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/buffer"
)

type Use struct {
	Only   []string `toml:"only"`
	Except []string `toml:"except"`
}

type LanguageConfigs struct {
	GrammarDir  string `toml:"grammar_dir"`
	QueriesDir  string `toml:"queries_dir"`
	UseGrammars Use    `toml:"use_grammars"`

	Languages map[string]LanguageConfig `toml:"languages"`
}

func (l LanguageConfigs) filter() LanguageConfigs {
	languages := make(map[string]LanguageConfig)
	for name, language := range l.Languages {
		if len(l.UseGrammars.Only) > 0 {
			if !slices.Contains(l.UseGrammars.Only, name) {
				continue
			}
		} else if len(l.UseGrammars.Except) > 0 {
			if slices.Contains(l.UseGrammars.Except, name) {
				continue
			}
		}

		languages[name] = language
	}

	return LanguageConfigs{
		GrammarDir:  l.GrammarDir,
		UseGrammars: l.UseGrammars,
		Languages:   languages,
	}
}

type LanguageConfig struct {
	AltNames           []string                   `toml:"alt_names"`
	MIMETypes          []string                   `toml:"mime_types"`
	FileTypes          []string                   `toml:"file_types"`
	Files              []string                   `toml:"files"`
	LineCommentTokens  []string                   `toml:"line_comment_tokens"`
	BlockCommentTokens []buffer.BlockCommentToken `toml:"block_comment_tokens"`
	AutoPairs          []LanguageAutoPairs        `toml:"auto_pairs"`
	Grammar            *GrammarConfig             `toml:"grammar"`
}

type LanguageAutoPairs struct {
	Open  string `toml:"open"`
	Close string `toml:"close"`
}

type GrammarConfig struct {
	Name       string                `toml:"name"`
	SymbolName string                `toml:"symbol_name"`
	QueriesDir string                `toml:"queries_dir"`
	Path       string                `toml:"path"`
	Install    *GrammarInstallConfig `toml:"install"`
}

func (c GrammarConfig) Hyperlink() string {
	if c.Install == nil {
		return c.Name
	}
	return lipgloss.Hyperlink(c.Install.Git, c.Name)
}

type GrammarInstallConfig struct {
	Git     string  `toml:"git"`
	Rev     string  `toml:"rev"`
	Ref     string  `toml:"ref"`
	RefType RefType `toml:"ref_type"`
	SubDir  string  `toml:"sub_dir"`
}

func (c GrammarInstallConfig) Hyperlink() string {
	switch c.RefType {
	case RefTypeCommit:
		return lipgloss.Hyperlink(c.Git+"/commit/"+c.Rev, c.Rev)
	case RefTypeTag:
		return lipgloss.Hyperlink(c.Git+"/releases/tag/"+c.Ref, c.Ref)
	}
	return c.Ref
}

type RefType string

const (
	RefTypeCommit RefType = "commit"
	RefTypeTag    RefType = "tag"
)
