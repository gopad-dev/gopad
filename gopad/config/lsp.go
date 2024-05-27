package config

func DefaultLSPConfig() LSPConfig {
	return make(map[string]LSPServerConfig)
}

type LSPConfig map[string]LSPServerConfig

type LSPServerConfig struct {
	Command   string   `toml:"command"`
	Args      []string `toml:"args"`
	Config    any      `toml:"config"`
	FileTypes []string `toml:"file_types"`
	Files     []string `toml:"files"`
}
