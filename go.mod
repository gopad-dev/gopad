module go.gopad.dev/gopad

go 1.22

replace (
	github.com/charmbracelet/lipgloss => github.com/gopad-dev/lipgloss v0.0.0-20240422161526-eea0119313bc
	github.com/muesli/termenv => github.com/gopad-dev/termenv v0.0.0-20240413225005-5f4a43fcdd7b
	go.lsp.dev/protocol => github.com/gopad-dev/protocol v0.0.0-20240529205148-623e5abff393
)

require (
	github.com/atotto/clipboard v0.1.4
	github.com/bmatcuk/doublestar/v4 v4.6.1
	github.com/charmbracelet/bubbles v0.18.0
	github.com/charmbracelet/bubbletea v0.26.4
	github.com/charmbracelet/lipgloss v0.11.0
	github.com/charmbracelet/x/ansi v0.1.2
	github.com/dustin/go-humanize v1.0.1
	github.com/mattn/go-runewidth v0.0.15
	github.com/muesli/reflow v0.3.0
	github.com/muesli/termenv v0.15.2
	github.com/pelletier/go-toml/v2 v2.2.2
	github.com/rivo/uniseg v0.4.7
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.9.0
	go.gopad.dev/fuzzysearch v0.0.0-20240526153819-c12185e04fe2
	go.gopad.dev/go-tree-sitter v0.0.0-20240614175658-13906aaed6af
	go.lsp.dev/jsonrpc2 v0.10.0
	go.lsp.dev/protocol v0.12.1-0.20240203004437-3c0d4339e51f
	golang.org/x/text v0.16.0
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/x/exp/term v0.0.0-20240606154654-7c42867b53c7 // indirect
	github.com/charmbracelet/x/input v0.1.2 // indirect
	github.com/charmbracelet/x/term v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/ebitengine/purego v0.7.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/segmentio/encoding v0.4.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.lsp.dev/pkg v0.0.0-20210717090340-384b27a52fb2 // indirect
	go.lsp.dev/uri v0.3.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
