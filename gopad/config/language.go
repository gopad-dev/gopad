package config

import (
	"slices"

	"go.gopad.dev/gopad/gopad/buffer"
)

type Use struct {
	Only   []string `toml:"only"`
	Except []string `toml:"except"`
}

func DefaultLanguageConfigs() LanguagesConfig {
	return LanguagesConfig{
		GrammarDir: "grammars",
		QueriesDir: "queries",
		Languages:  make(map[string]LanguageConfig),
	}
}

type LanguagesConfig struct {
	GrammarDir  string `toml:"grammar_dir"`
	QueriesDir  string `toml:"queries_dir"`
	UseGrammars Use    `toml:"use_grammars"`

	Languages map[string]LanguageConfig `toml:"languages"`
}

func (l LanguagesConfig) filter() LanguagesConfig {
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

	return LanguagesConfig{
		GrammarDir:  l.GrammarDir,
		UseGrammars: l.UseGrammars,
		Languages:   languages,
	}
}

type LanguageConfig struct {
	Icon               rune                       `toml:"icon"`
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

type GrammarInstallConfig struct {
	Git    string   `toml:"git"`
	Rev    string   `toml:"rev"`
	SubDir string   `toml:"sub_dir"`
	Cmd    string   `toml:"cmd"`
	Args   []string `toml:"args"`
	Path   string   `toml:"path"`
}
