package config

import (
	"go.gopad.dev/gopad/gopad/buffer"
)

func DefaultLanguageConfigs() map[string]LanguageConfig {
	return make(map[string]LanguageConfig)
}

type LanguageConfig struct {
	AltNames           []string                   `toml:"alt_names"`
	MIMETypes          []string                   `toml:"mime_types"`
	FileTypes          []string                   `toml:"file_types"`
	Files              []string                   `toml:"files"`
	LineCommentTokens  []string                   `toml:"line_comment_tokens"`
	BlockCommentTokens []buffer.BlockCommentToken `toml:"block_comment_tokens"`
	AutoPairs          []LanguageAutoPairs        `toml:"auto_pairs"`
	TreeSitter         *TreeSitterConfig          `toml:"tree_sitter"`
}

type LanguageAutoPairs struct {
	Open  string `toml:"open"`
	Close string `toml:"close"`
}

type TreeSitterConfig struct {
	Name       string `toml:"name"`
	SymbolName string `toml:"symbol_name"`
	Path       string `toml:"path"`
}
