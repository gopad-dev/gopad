module go.gopad.dev/gopad

go 1.23

replace (
	github.com/charmbracelet/bubbletea/v2 => github.com/gopad-dev/bubbletea/v2 v2.0.0-20240919124401-5a057e723e33
	github.com/charmbracelet/lipgloss => github.com/gopad-dev/lipgloss v0.0.0-20240906153413-0bcc656d0482
	github.com/lrstanley/bubblezone => github.com/gopad-dev/bubblezone v0.0.0-20240919125415-44caa82cfbd5
	go.lsp.dev/protocol => github.com/gopad-dev/protocol v0.0.0-20240916085830-4815610e4100
)

require (
	github.com/atotto/clipboard v0.1.4
	github.com/bmatcuk/doublestar/v4 v4.6.1
	github.com/charmbracelet/bubbletea/v2 v2.0.0-alpha.1
	github.com/charmbracelet/lipgloss v0.13.0
	github.com/charmbracelet/x/ansi v0.3.2
	github.com/dustin/go-humanize v1.0.1
	github.com/lrstanley/bubblezone v0.0.0-20240624011428-67235275f80c
	github.com/muesli/reflow v0.3.0
	github.com/pelletier/go-toml/v2 v2.2.3
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.9.0
	go.gopad.dev/fuzzysearch v0.0.0-20240526153819-c12185e04fe2
	go.gopad.dev/go-tree-sitter v0.0.0-20240620185356-89c6dfd0fb37
	go.lsp.dev/jsonrpc2 v0.10.0
	go.lsp.dev/protocol v0.12.1-0.20240203004437-3c0d4339e51f
	golang.org/x/text v0.18.0
)

require (
	github.com/charmbracelet/x/term v0.2.0 // indirect
	github.com/charmbracelet/x/windows v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ebitengine/purego v0.7.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/segmentio/encoding v0.4.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.lsp.dev/pkg v0.0.0-20210717090340-384b27a52fb2 // indirect
	go.lsp.dev/uri v0.3.0 // indirect
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
