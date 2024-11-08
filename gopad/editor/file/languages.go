package file

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"unsafe"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/ebitengine/purego"
	"github.com/tree-sitter/go-tree-sitter"

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

var Languages []*Language

type Language struct {
	Name    string
	Config  config.LanguageConfig
	Grammar *GrammarConfig
}

func (l *Language) Title() string {
	return l.Name
}

func (l *Language) Description() string {
	return ""
}

type GrammarConfig struct {
	Highlight HighlightConfig
	Outline   *OutlineQueryConfig
}

type OutlineQueryConfig struct {
	Query                 *tree_sitter.Query
	ItemCaptureID         uint
	NameCaptureID         uint
	ContextCaptureID      *uint
	ExtraContextCaptureID *uint
}

func LoadLanguages(defaultConfigs embed.FS) error {
	for name, language := range config.Languages.Languages {
		lang := &Language{
			Config: language,
			Name:   name,
		}

		if language.Grammar != nil {
			g, err := newHighlightConfig(name, *language.Grammar, defaultConfigs)
			if err != nil {
				return fmt.Errorf("error loading tree-sitter grammar for %q: %w", name, err)
			}
			if g != nil {
				lang.Grammar = g
			}
		}

		Languages = append(Languages, lang)
	}

	return nil
}

func newHighlightConfig(languageName string, cfg config.GrammarConfig, defaultConfigs embed.FS) (*GrammarConfig, error) {
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

	language, err := loadLanguage(symbolName, libPath)
	if err != nil {
		return nil, fmt.Errorf("error loading lib %q: %w", libPath, err)
	}

	queriesConfigDir := cfg.QueriesDir
	if queriesConfigDir == "" {
		queriesConfigDir = filepath.Join(config.Path, queriesDir, name)
	}

	highlightsQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryHighlightsFileName)
	if err != nil {
		return nil, fmt.Errorf("error reading highlights query: %w", err)
	}

	injectionQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryInjectionsFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error reading injection query: %w", err)
	}

	localsQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryLocalsFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error reading locals query: %w", err)
	}

	highlightConfig, err := NewHighlightConfig(language, languageName, highlightsQuery, injectionQuery, localsQuery)
	if err != nil {
		return nil, fmt.Errorf("error creating highlight config: %w", err)
	}

	outlineQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryOutlineFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error reading outline query: %w", err)
	}

	var outlineQueryConfig *OutlineQueryConfig
	if len(outlineQuery) > 0 {
		query, err := tree_sitter.NewQuery(language, string(outlineQuery))
		if err != nil {
			return nil, fmt.Errorf("error parsing outline query: %w", err)
		}

		indexes := getCaptureIndexes(query, []string{
			"item",
			"name",
			"context",
			"extra_context",
		})

		outlineQueryConfig = &OutlineQueryConfig{
			Query:                 query,
			ItemCaptureID:         *indexes[0],
			NameCaptureID:         *indexes[1],
			ContextCaptureID:      indexes[2],
			ExtraContextCaptureID: indexes[3],
		}
	}

	return &GrammarConfig{
		Highlight: *highlightConfig,
		Outline:   outlineQueryConfig,
	}, nil
}

func loadLanguage(symbolName string, path string) (*tree_sitter.Language, error) {
	lib, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, fmt.Errorf("failed to open language library: %w", err)
	}

	var newTreeSitter func() uintptr
	purego.RegisterLibFunc(&newTreeSitter, lib, "tree_sitter_"+symbolName)

	return tree_sitter.NewLanguage(unsafe.Pointer(newTreeSitter())), nil
}

func getCaptureIndexes(query *tree_sitter.Query, captureNames []string) []*uint {
	indexes := make([]*uint, len(captureNames))
	for i, name := range captureNames {
		id, ok := query.CaptureIndexForName(name)
		if !ok {
			continue
		}
		indexes[i] = &id
	}
	return indexes
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

func GetLanguage(name string) *Language {
	for _, lang := range Languages {
		if lang.Name == name || slices.Contains(lang.Config.AltNames, name) {
			return lang
		}
	}

	return nil
}

func GetLanguageByFilename(filename string) *Language {
	ext := filepath.Ext(filename)
	fileName := filepath.Base(filename)

	for _, language := range Languages {
		if slices.Contains(language.Config.FileTypes, ext) || slices.Contains(language.Config.Files, fileName) || matchGlobs(language.Config.Files, filename) {
			return language
		}
	}
	return nil
}

func matchGlobs(globs []string, filename string) bool {
	for _, glob := range globs {
		if ok, _ := doublestar.PathMatch(glob, filename); ok {
			return true
		}
	}
	return false
}
