<div align="center">

<h1>
<picture>
  <source media="(prefers-color-scheme: dark)" srcset=".github/gopad_dark.svg">
  <source media="(prefers-color-scheme: light)" srcset=".github/gopad_light.svg">
  <img alt="gopad" height="128" src=".github/gopad.svg">
</picture>
</h1>

[![Go Report](https://goreportcard.com/badge/go.gopad.dev/gopad)](https://goreportcard.com/report/go.gopad.dev/gopad)
[![Go Version](https://img.shields.io/github/go-mod/go-version/gopad-dev/gopad)](https://golang.org/doc/devel/release.html)
[![gopad License](https://img.shields.io/github/license/gopad-dev/gopad)](LICENSE)
[![Build status](https://github.com/gopad-dev/gopad/actions/workflows/build.yml/badge.svg)](https://github.com/gopad-dev/gopad/actions)
[![gopad Version](https://img.shields.io/github/v/tag/gopad-dev/gopad?label=release)](https://go.gopad.dev/gopad/releases/latest)

</div>

gopad is a simple terminal-based text editor written in Go. It is inspired mostly by [nano](https://www.nano-editor.org/).

> [!IMPORTANT]
> gopad is still very much wip and not ready for general use.

<details>
<summary>Table of Contents</summary>

- [Installation](#installation)
- [Usage](#usage)
    - [Flags](#flags)
    - [Environment Variables](#environment-variables)
- [Configuration](#configuration)
- [License](#license)

</details>

## Installation

```bash
git clone https://github.com/gopad-dev/gopad.git
  
cd gopad

./install.sh
```

(`go install go.gopad.dev/gopad@latest` is currently not working due to `replace` directives in the `go.mod` file.)

## Usage

```bash
gopad [flags]... [dir | file]...
gopad [command]
```

#### Commands

```bash
completion  Generate the autocompletion script for the specified shell
  bash        Generate the autocompletion script for bash
  fish        Generate the autocompletion script for fish
  powershell  Generate the autocompletion script for powershell
  zsh         Generate the autocompletion script for zsh
config      Create a new config directory with default config files
grammar     Manage Tree-Sitter grammars
  install     Install Tree-Sitter grammars
  list        List configured Tree-Sitter grammars
  remove      Remove installed Tree-Sitter grammars
  update      Check for updates of Tree-Sitter grammars
help        Help about any command
version     Show version information
```

#### Flags

```bash
  -c, --config-dir string   set configuration directory (Default: ./.gopad, $XDG_CONFIG_HOME/gopad or $HOME/.config/gopad)
  -d, --debug string        set debug log file (use - for stdout)
  -l, --debug-lsp string    set debug lsp log file
  -h, --help                help for gopad
  -p, --pprof string        set pprof address:port
  -w, --workspace string    set workspace directory (Default: first directory argument)
```

## Configuration

gopad uses multiple TOML configuration files. See the [default configuration directory](config) for all configuration files.
To create a new configuration directory with default configuration files, run `gopad config`. 

### Environment Variables

- `GOPAD_CONFIG_HOME` - Use the specified directory for configuration files. (Default: `./.gopad`, `$XDG_CONFIG_HOME/gopad` or `$HOME/.config/gopad`)

## License

gopad is licensed under the [Apache License 2.0](LICENSE).
