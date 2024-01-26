package config

func DefaultLSPConfig() LSPConfig {
	return make(map[string]LSPServerConfig)
}

type LSPConfig map[string]LSPServerConfig

type LSPServerConfig struct {
	Command   string   `toml:"command"`
	Config    any      `toml:"config"`
	Args      []string `toml:"args"`
	FileTypes []string `toml:"file_types"`
	Files     []string `toml:"files"`
}
