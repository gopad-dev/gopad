package editor

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"go.gopad.dev/go-tree-sitter"

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
	parseTimeout = 10 * time.Second
)

type Language struct {
	config.LanguageConfig
	Name    string
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
	highlightsQuery *sitter.Query
	injectionsQuery *sitter.Query
	localsQuery     *sitter.Query
	indentsQuery    *sitter.Query
	foldsQuery      *sitter.Query
	outlineQuery    *sitter.Query
	ParseTimeout    time.Duration
}

func (g *Grammar) OutlineQuery() OutlineConfig {
	indexes := GetCaptureIndexes(g.outlineQuery, []string{
		"item",
		"name",
		"context",
		"extra_context",
	})

	return OutlineConfig{
		Query:                 g.outlineQuery,
		ItemCaptureID:         *indexes[0],
		NameCaptureID:         *indexes[1],
		ContextCaptureID:      indexes[2],
		ExtraContextCaptureID: indexes[3],
	}
}

type OutlineConfig struct {
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
	languageMap := config.Languages
	for name, language := range languageMap {
		lang := &Language{
			LanguageConfig: language,
			Name:           name,
		}

		if language.TreeSitter != nil {
			grammar, err := loadTreeSitterGrammar(name, *language.TreeSitter, defaultConfigs)
			if err != nil {
				return fmt.Errorf("error loading tree-sitter grammar for %q: %w", name, err)
			}

			lang.Grammar = grammar
		}

		languages = append(languages, lang)
	}

	return nil
}

func loadTreeSitterGrammar(name string, cfg config.TreeSitterConfig, defaultConfigs embed.FS) (*Grammar, error) {
	libPath := cfg.Path
	if libPath == "" {

	}
	symbolName := cfg.SymbolName
	if symbolName == "" {
		symbolName = name
	}
	tsLang, err := sitter.LoadLanguage(symbolName, cfg.Path)
	if err != nil {
		return nil, err
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
		injectionsQuery *sitter.Query
		localsQuery     *sitter.Query
		indentsQuery    *sitter.Query
		foldsQuery      *sitter.Query
		outlineQuery    *sitter.Query
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
			injectionsQuery = query
		case "locals.scm":
			localsQuery = query
		case "indents.scm":
			indentsQuery = query
		case "folds.scm":
			foldsQuery = query
		case "outline.scm":
			outlineQuery = query
		default:
			continue
		}

		log.Printf("Loaded query %s/%s\n", name, queryFile.Name())
	}

	return &Grammar{
		Language:        tsLang,
		highlightsQuery: highlightsQuery,
		injectionsQuery: injectionsQuery,
		localsQuery:     localsQuery,
		indentsQuery:    indentsQuery,
		foldsQuery:      foldsQuery,
		outlineQuery:    outlineQuery,
		ParseTimeout:    parseTimeout,
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
		if lang.Name == name || slices.Contains(lang.AltNames, name) {
			return lang
		}
	}

	return nil
}

func GetLanguageByMIMEType(mimeType string) *Language {
	for _, language := range languages {
		if slices.Contains(language.MIMETypes, mimeType) {
			return language
		}
	}
	return nil
}

func GetLanguageByFilename(filename string) *Language {
	ext := filepath.Ext(filename)
	fileName := filepath.Base(filename)

	for _, language := range languages {
		if slices.Contains(language.FileTypes, ext) || slices.Contains(language.Files, fileName) || matchGlobs(language.Files, filename) {
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