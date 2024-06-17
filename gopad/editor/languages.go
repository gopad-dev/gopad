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

const queriesDir = "queries"

var (
	languages      []*Language
	queryFileNames = []string{
		"highlights.scm",
		"injections.scm",
		"locals.scm",
		"indents.scm",
		"folds.scm",
		"outline.scm",
	}
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
	HighlightsQuery *sitter.Query
	InjectionsQuery *InjectionsQuery
	LocalsQuery     *LocalsQuery
	OutlineQuery    *OutlineQuery
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

type LocalsQuery struct {
	Query *sitter.Query
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
	queryFiles, err := os.ReadDir(queriesConfigDir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error reading queries directory: %w", err)
	}

	if len(queryFiles) == 0 {
		queriesConfigDir = filepath.Join("config", queriesDir, name)
		queryFiles, err = defaultConfigs.ReadDir(queriesConfigDir)
		if err != nil {
			return nil, fmt.Errorf("error reading default queries directory: %w", err)
		}
	}

	var (
		highlightsQuery *sitter.Query
		injectionsQuery *InjectionsQuery
		localsQuery     *LocalsQuery
		outlineQuery    *OutlineQuery
	)

	for _, queryFile := range queryFiles {
		if queryFile.IsDir() || !slices.Contains(queryFileNames, queryFile.Name()) {
			continue
		}

		var rawQuery []byte
		rawQuery, err = readQuery(queriesConfigDir, queryFile)
		if err != nil {
			return nil, fmt.Errorf("error reading query file %s: %w", queryFile.Name(), err)
		}

		var query *sitter.Query
		query, err = sitter.NewQuery(rawQuery, tsLang)
		if err != nil {
			return nil, fmt.Errorf("error parsing query file %s: %w", queryFile.Name(), err)
		}

		switch queryFile.Name() {
		case "highlights.scm":
			highlightsQuery = query
		case "injections.scm":
			indexes := GetCaptureIndexes(query, []string{
				"injection.content",
			})

			if indexes[0] == nil {
				return nil, fmt.Errorf("injection.content capture not found in %s", queryFile.Name())
			}

			injectionsQuery = &InjectionsQuery{
				Query:                     query,
				InjectionContentCaptureID: *indexes[0],
			}
		case "locals.scm":
			_ = GetCaptureIndexes(query, []string{})

			localsQuery = &LocalsQuery{
				Query: query,
			}
		case "outline.scm":
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
		default:
			continue
		}
	}

	return &Grammar{
		Language:        tsLang,
		HighlightsQuery: highlightsQuery,
		InjectionsQuery: injectionsQuery,
		LocalsQuery:     localsQuery,
		OutlineQuery:    outlineQuery,
	}, nil
}

func readQuery(config string, query os.DirEntry) ([]byte, error) {
	f, err := os.Open(filepath.Join(config, query.Name()))
	if err != nil {
		return nil, fmt.Errorf("error opening theme file: %w", err)
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
