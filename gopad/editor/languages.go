package editor

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/bmatcuk/doublestar/v4"
	"go.gopad.dev/go-tree-sitter"

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

var (
	languages []*Language
)

type Language struct {
	Name    string
	Config  config.LanguageConfig
	Grammar *Grammar
}

func (l *Language) Title() string {
	return l.Name
}

func (l *Language) Description() string {
	return ""
}

type Grammar struct {
	Language        *sitter.Language
	HighlightsQuery HighlightsQuery
	InjectionsQuery *InjectionsQuery
	OutlineQuery    *OutlineQuery
}

type HighlightsQuery struct {
	Query *sitter.Query

	HighlightsPatternIndex uint32

	ScopeCaptureID      *uint32
	DefinitionCaptureID *uint32
	ReferenceCaptureID  *uint32
}

type InjectionsQuery struct {
	Query                     *sitter.Query
	InjectionContentCaptureID uint32
}

type OutlineQuery struct {
	Query                 *sitter.Query
	ItemCaptureID         uint32
	NameCaptureID         uint32
	ContextCaptureID      *uint32
	ExtraContextCaptureID *uint32
}

func GetCaptureIndexes(query *sitter.Query, captureNames []string) []*uint32 {
	indexes := make([]*uint32, len(captureNames))
	for id := range query.CaptureCount() {
		name := query.CaptureNameForID(id)
		index := slices.Index(captureNames, name)
		if index >= 0 {
			indexes[index] = &id
		}
	}
	return indexes
}

func LoadLanguages(defaultConfigs embed.FS) error {
	for name, language := range config.Languages.Languages {
		lang := &Language{
			Config: language,
			Name:   name,
		}

		if language.Grammar != nil {
			grammar, err := loadTreeSitterGrammar(name, *language.Grammar, defaultConfigs)
			if err != nil {
				return fmt.Errorf("error loading tree-sitter grammar for %q: %w", name, err)
			}
			if grammar != nil {
				lang.Grammar = grammar
			}
		}

		languages = append(languages, lang)
	}

	return nil
}

func loadTreeSitterGrammar(name string, cfg config.GrammarConfig, defaultConfigs embed.FS) (*Grammar, error) {
	libPath := cfg.Path
	if libPath == "" {
		libPath = filepath.Join(config.Path, "grammars", grammar.LibName(name))
	}

	symbolName := cfg.SymbolName
	if symbolName == "" {
		symbolName = cfg.Name
	}

	_, err := os.Stat(libPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("error checking lib %q: %w", libPath, err)
	}

	tsLang, err := sitter.LoadLanguage(symbolName, libPath)
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

	rawLocalsQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryLocalsFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error reading locals query: %w", err)
	}

	var combinedQuery []byte
	var highlightsQueryOffset int
	if len(rawLocalsQuery) > 0 {
		combinedQuery = rawLocalsQuery
		combinedQuery = append(combinedQuery, '\n')
		highlightsQueryOffset = len(rawLocalsQuery)
	}
	combinedQuery = append(combinedQuery, rawHighlightsQuery...)

	query, err := sitter.NewQuery(combinedQuery, tsLang)
	if err != nil {
		return nil, fmt.Errorf("error parsing combined locals and highlights query: %w", err)
	}

	var highlightsPatternIndex uint32
	for i := range query.PatternCount() {
		patternOffset := query.PatternStartByte(i)
		if int(patternOffset) < highlightsQueryOffset {
			highlightsPatternIndex++
		}
	}

	highlightsQuery := HighlightsQuery{
		Query:                  query,
		HighlightsPatternIndex: highlightsPatternIndex,
	}

	if len(rawLocalsQuery) > 0 {
		indexes := GetCaptureIndexes(query, []string{
			"local.scope",
			"local.definition",
			"local.reference",
		})

		highlightsQuery.ScopeCaptureID = indexes[0]
		highlightsQuery.DefinitionCaptureID = indexes[1]
		highlightsQuery.ReferenceCaptureID = indexes[2]
	}

	rawInjectionsQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryInjectionsFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error reading locals query: %w", err)
	}

	var injectionsQuery *InjectionsQuery
	if len(rawInjectionsQuery) > 0 {
		query, err = sitter.NewQuery(rawInjectionsQuery, tsLang)
		if err != nil {
			return nil, fmt.Errorf("error parsing injections query: %w", err)
		}

		indexes := GetCaptureIndexes(query, []string{
			"injection.content",
		})

		injectionsQuery = &InjectionsQuery{
			Query:                     query,
			InjectionContentCaptureID: *indexes[0],
		}
	}

	rawOutlineQuery, err := readQuery(queriesConfigDir, defaultConfigs, name, queryOutlineFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error reading outline query: %w", err)
	}

	var outlineQuery *OutlineQuery
	if len(rawOutlineQuery) > 0 {
		query, err = sitter.NewQuery(rawOutlineQuery, tsLang)
		if err != nil {
			return nil, fmt.Errorf("error parsing outline query: %w", err)
		}

		indexes := GetCaptureIndexes(query, []string{
			"item",
			"name",
			"context",
			"extra_context",
		})

		outlineQuery = &OutlineQuery{
			Query:                 query,
			ItemCaptureID:         *indexes[0],
			NameCaptureID:         *indexes[1],
			ContextCaptureID:      indexes[2],
			ExtraContextCaptureID: indexes[3],
		}
	}

	return &Grammar{
		Language:        tsLang,
		HighlightsQuery: highlightsQuery,
		InjectionsQuery: injectionsQuery,
		OutlineQuery:    outlineQuery,
	}, nil
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

	defer f.Close()

	return io.ReadAll(f)
}

func GetLanguage(name string) *Language {
	for _, lang := range languages {
		if lang.Name == name || slices.Contains(lang.Config.AltNames, name) {
			return lang
		}
	}

	return nil
}

func GetLanguageByMIMEType(mimeType string) *Language {
	for _, language := range languages {
		if slices.Contains(language.Config.MIMETypes, mimeType) {
			return language
		}
	}
	return nil
}

func GetLanguageByFilename(filename string) *Language {
	ext := filepath.Ext(filename)
	fileName := filepath.Base(filename)

	for _, language := range languages {
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
