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
	}
)

type Language struct {
	config.LanguageConfig
	Name    string
	Grammar *Grammar
}

type Grammar struct {
	Name            string
	TreeSitter      *sitter.Language
	HighlightsQuery []byte
	InjectionsQuery []byte
	LocalsQuery     []byte
	IndentsQuery    []byte
	FoldsQuery      []byte
}

func (l *Language) Title() string {
	return l.Name
}

func (l *Language) Description() string {
	return ""
}

func LoadLanguages(defaultConfigs embed.FS) error {
	languageMap := config.Languages
	for name, language := range languageMap {
		lang := &Language{
			LanguageConfig: language,
			Name:           name,
		}

		if language.TreeSitter != nil {
			grammar, err := loadTreeSitterGrammar(*language.TreeSitter, defaultConfigs)
			if err != nil {
				return fmt.Errorf("error loading tree-sitter grammar for %q: %w", name, err)
			}

			lang.Grammar = grammar
		}

		languages = append(languages, lang)
	}

	return nil
}

func loadTreeSitterGrammar(cfg config.TreeSitterConfig, defaultConfigs embed.FS) (*Grammar, error) {
	tsLang, err := sitter.LoadLanguage(cfg.SymbolName, cfg.Path)
	if err != nil {
		return nil, err
	}

	queriesConfigDir := config.Path
	queryFiles, err := os.ReadDir(filepath.Join(config.Path, queriesDir, cfg.Name))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error reading queries directory: %w", err)
	}

	if len(queryFiles) == 0 {
		queriesConfigDir = "config"
		queryFiles, err = defaultConfigs.ReadDir(filepath.Join("config", queriesDir, cfg.Name))
		if err != nil {
			return nil, fmt.Errorf("error reading default queries directory: %w", err)
		}
	}

	var (
		highlightsQuery []byte
		injectionsQuery []byte
		localsQuery     []byte
		indentsQuery    []byte
		foldsQuery      []byte
	)

	for _, queryFile := range queryFiles {
		if queryFile.IsDir() || !slices.Contains(queryFileNames, queryFile.Name()) {
			continue
		}

		query, err := readQuery(queriesConfigDir, cfg.Name, queryFile)
		if err != nil {
			return nil, fmt.Errorf("error reading query file %s: %w", queryFile.Name(), err)
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
		}
	}

	return &Grammar{
		Name:            cfg.Name,
		TreeSitter:      tsLang,
		HighlightsQuery: highlightsQuery,
		InjectionsQuery: injectionsQuery,
		LocalsQuery:     localsQuery,
		IndentsQuery:    indentsQuery,
		FoldsQuery:      foldsQuery,
	}, nil
}

func readQuery(config string, name string, query os.DirEntry) ([]byte, error) {
	f, err := os.Open(filepath.Join(config, queriesDir, name, query.Name()))
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
