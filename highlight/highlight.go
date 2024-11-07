package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"go.gopad.dev/go-tree-sitter"
	"go.gopad.dev/go-tree-sitter/highlight"

	"go.gopad.dev/gopad/cmd/grammar"
	"go.gopad.dev/gopad/gopad/config"
)

const (
	configDir  = "config"
	queriesDir = "queries"

	queryHighlightsFileName = "highlights.scm"
	queryInjectionsFileName = "injections.scm"
	queryLocalsFileName     = "locals.scm"
	queryOutlineFileName    = "outline.scm"
)

func NewHighlightConfig(languageName string, cfg config.GrammarConfig, defaultConfigs embed.FS) (*highlight.HighlightConfig, error) {
	name := cfg.Name
	if name == "" {
		name = languageName
	}

	symbolName := cfg.SymbolName
	if symbolName == "" {
		symbolName = cfg.Name
	}

	libPath := cfg.Path
	if libPath == "" {
		libPath = filepath.Join(config.Path, "grammars", grammar.LibName(name))
	}

	// compiled tree-sitter grammars are stored on disk only
	_, err := os.Stat(libPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("error checking lib %q: %w", libPath, err)
	}

	language, err := sitter.LoadLanguage(symbolName, libPath)
	if err != nil {
		return nil, fmt.Errorf("error loading lib %q: %w", libPath, err)
	}

	queriesConfigDir := cfg.QueriesDir
	if queriesConfigDir == "" {
		queriesConfigDir = filepath.Join(config.Path, queriesDir, name)
	}

	rawHighlightsQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryHighlightsFileName)
	if err != nil {
		return nil, fmt.Errorf("error reading highlights query: %w", err)
	}

	rawInjectionQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryInjectionsFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error reading injections query: %w", err)
	}

	rawLocalsQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryLocalsFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error reading locals query: %w", err)
	}

	highlight.NewHighlightConfig()
}

func readQuery(config string, defaultConfigs embed.FS, name string, query string) ([]byte, error) {
	_, err := os.Stat(filepath.Join(config, query))

	var f fs.File
	if errors.Is(err, os.ErrNotExist) {
		f, err = defaultConfigs.Open(filepath.Join(configDir, queriesDir, name, query))
	} else if err == nil {
		f, err = os.Open(filepath.Join(config, query))
	}

	if err != nil {
		return nil, fmt.Errorf("error opening query %q: %w", query, err)
	}

	defer func() {
		_ = f.Close()
	}()

	return io.ReadAll(f)
}
