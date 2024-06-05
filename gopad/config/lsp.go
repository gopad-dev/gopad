package config

import (
	"slices"
)

func DefaultLSPConfig() LSPConfigs {
	return LSPConfigs{
		LSPs: make(map[string]LSPConfig),
	}
}

type LSPConfigs struct {
	UseServers Use `toml:"use_servers"`

	LSPs map[string]LSPConfig `toml:"lsps"`
}

func (l LSPConfigs) filter() LSPConfigs {
	lsps := make(map[string]LSPConfig)
	for name, lsp := range l.LSPs {
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

	return LSPConfigs{
		UseServers: l.UseServers,
		LSPs:       lsps,
	}
}

type LSPConfig struct {
	Command   string   `toml:"command"`
	Args      []string `toml:"args"`
	Config    any      `toml:"config"`
	FileTypes []string `toml:"file_types"`
	Files     []string `toml:"files"`
}
