package config

import (
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	gopadConfig           = "gopad.toml"
	keymapConfig          = "keymap.toml"
	languagesConfig       = "languages.toml"
	languageServersConfig = "language_servers.toml"
	themeDir              = "themes"
)

var (
	Path            string
	Gopad           GopadConfig
	Languages       LanguagesConfig
	LanguageServers LanguageServerConfigs
	Keys            KeyMap
	Theme           ThemeConfig
	Themes          []RawThemeConfig
)

func FindHome() (string, error) {
	gopadHome := os.Getenv("GOPAD_CONFIG_HOME")
	if gopadHome == "" {
		xdgHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("error finding your home directory: %w", err)
			}
			xdgHome = filepath.Join(home, ".config")
		}
		gopadHome = filepath.Join(xdgHome, "gopad")
	}

	if err := os.MkdirAll(gopadHome, os.ModePerm); err != nil {
		return "", fmt.Errorf("error creating config directory: %w", err)
	}

	return gopadHome, nil
}

func Load(name string, defaultConfigs embed.FS) error {
	gopad := DefaultGopadConfig()
	keymap := DefaultKeyMapConfig()
	languages := DefaultLanguageConfigs()
	languageServers := DefaultLanguageServerConfigs()
	themes := make([]RawThemeConfig, 0)

	if err := readTOMLFile(filepath.Join(name, gopadConfig), &gopad); err != nil {
		return fmt.Errorf("error reading gopad config: %w", err)
	}

	if err := readTOMLFile(filepath.Join(name, keymapConfig), &keymap); err != nil {
		return fmt.Errorf("error reading keymap config: %w", err)
	}

	if err := readTOMLFile(filepath.Join(name, languagesConfig), &languages); err != nil {
		return fmt.Errorf("error reading languages config: %w", err)
	}

	if err := readTOMLFile(filepath.Join(name, languageServersConfig), &languageServers); err != nil {
		return fmt.Errorf("error reading LanguageServers config: %w", err)
	}

	themeConfigDir := name
	themeFiles, err := os.ReadDir(filepath.Join(name, themeDir))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("error reading themes directory: %w", err)
	}

	if len(themeFiles) == 0 {
		themeConfigDir = "config"
		themeFiles, err = defaultConfigs.ReadDir(filepath.Join("config", themeDir))
		if err != nil {
			return fmt.Errorf("error reading default themes directory: %w", err)
		}
	}

	var theme *RawThemeConfig
	for _, themeFile := range themeFiles {
		if themeFile.IsDir() {
			continue
		}

		themeConfig, err := readTheme(themeConfigDir, themeFile)
		if err != nil {
			return fmt.Errorf("error reading theme file %s: %w", themeFile.Name(), err)
		}

		if themeConfig.Name == gopad.Theme {
			theme = &themeConfig
		}

		index := slices.IndexFunc(themes, func(config RawThemeConfig) bool {
			return config.Name == themeConfig.Name
		})
		if index == -1 {
			themes = append(themes, themeConfig)
			continue
		}

		themes[index] = themeConfig
	}

	if theme == nil {
		var allThemes []string
		for _, t := range themes {
			allThemes = append(allThemes, t.Name)
		}
		return fmt.Errorf("theme %s not found in [%s]", gopad.Theme, strings.Join(allThemes, ", "))
	}

	Path = name
	Gopad = gopad
	Languages = languages.filter()
	LanguageServers = languageServers.filter()
	Keys = keymap.Keys()
	Themes = themes
	Theme = theme.Theme()

	return nil
}

func readTOMLFile(name string, a any) error {
	f, err := os.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer f.Close()

	if err = toml.NewDecoder(f).Decode(a); err != nil {
		return fmt.Errorf("error decoding config file: %w", err)
	}

	return nil
}

func readTheme(name string, theme os.DirEntry) (RawThemeConfig, error) {
	f, err := os.Open(filepath.Join(name, themeDir, theme.Name()))
	if err != nil {
		return RawThemeConfig{}, fmt.Errorf("error opening theme file: %w", err)
	}
	defer f.Close()

	var t RawThemeConfig
	if err = toml.NewDecoder(f).Decode(&t); err != nil {
		return RawThemeConfig{}, fmt.Errorf("error decoding theme file: %w", err)
	}

	return t, nil
}

func Create(name string, defaultConfigs embed.FS) error {
	log.Println("creating config in", name)
	return copyDir("config", name, defaultConfigs)
}

func copyDir(name string, dstName string, defaultConfigs embed.FS) error {
	if err := os.MkdirAll(dstName, os.ModePerm); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	files, err := defaultConfigs.ReadDir(name)
	if err != nil {
		return fmt.Errorf("error reading config directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			if err = copyDir(filepath.Join(name, file.Name()), filepath.Join(dstName, file.Name()), defaultConfigs); err != nil {
				return err
			}
			continue
		}

		if err = copyFile(filepath.Join(name, file.Name()), filepath.Join(dstName, file.Name()), defaultConfigs); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(name string, dstName string, defaultConfigs embed.FS) error {
	if _, err := os.Stat(dstName); !errors.Is(err, fs.ErrNotExist) {
		log.Println("skipping", name, "already exists")
		return nil
	}
	log.Println("copying", name, "to", dstName)

	out, err := os.Create(dstName)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}
	defer out.Close()

	f, err := defaultConfigs.Open(name)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer f.Close()

	if _, err = io.Copy(out, f); err != nil {
		return fmt.Errorf("error copying config file: %w", err)
	}

	return nil
}
