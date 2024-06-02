package config

import (
	"go.gopad.dev/gopad/gopad/buffer"
)

func DefaultLanguageConfigs() map[string]LanguageConfig {
	return make(map[string]LanguageConfig)
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
	Git  string   `toml:"git"`
	Rev  string   `toml:"rev"`
	Dir  string   `toml:"dir"`
	Cmd  string   `toml:"cmd"`
	Args []string `toml:"args"`
}
