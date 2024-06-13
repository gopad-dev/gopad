package config

import (
	"slices"
)

func DefaultLanguageServerConfigs() LanguageServerConfigs {
	return LanguageServerConfigs{
		LanguageServers: make(map[string]LanguageServerConfig),
	}
}

type LanguageServerConfigs struct {
	UseServers Use `toml:"use_servers"`

	LanguageServers map[string]LanguageServerConfig `toml:"language_servers"`
}

func (l LanguageServerConfigs) filter() LanguageServerConfigs {
	lsps := make(map[string]LanguageServerConfig)
	for name, lsp := range l.LanguageServers {
		if len(l.UseServers.Only) > 0 {
			if !slices.Contains(l.UseServers.Only, name) {
				continue
			}
		} else if len(l.UseServers.Except) > 0 {
			if slices.Contains(l.UseServers.Except, name) {
				continue
			}
		}

		lsps[name] = lsp
	}

	return LanguageServerConfigs{
		UseServers:      l.UseServers,
		LanguageServers: lsps,
	}
}

type LanguageServerConfig struct {
	Command   string                  `toml:"command"`
	Args      []string                `toml:"args"`
	Config    any                     `toml:"config"`
	FileTypes []string                `toml:"file_types"`
	Files     []string                `toml:"files"`
	Roots     []string                `toml:"roots"`
	Features  []LanguageServerFeature `toml:"features"`
}

type LanguageServerFeature string

const (
	LanguageServerFeatureCompletion  LanguageServerFeature = "completion"
	LanguageServerFeatureDiagnostics LanguageServerFeature = "diagnostics"
	LanguageServerFeatureInlayHints  LanguageServerFeature = "inlay_hints"
)
