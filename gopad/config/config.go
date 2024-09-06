package config

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

	"github.com/pelletier/go-toml/v2"
)

const (
	gopadConfig           = "gopad.toml"
	keymapConfig          = "keymap.toml"
	languagesConfig       = "languages.toml"
	languageServersConfig = "language_servers.toml"
	configDir             = "config"
	themeDir              = "themes"
)

var (
	Path            string
	Gopad           GopadConfig
	Languages       LanguageConfigs
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
	var (
		gopad           GopadConfig
		keymap          KeyMapConfig
		languages       LanguageConfigs
		languageServers LanguageServerConfigs
	)

	if err := readTOMLFile(name, gopadConfig, defaultConfigs, &gopad); err != nil {
		return fmt.Errorf("error reading gopad config: %w", err)
	}
	if err := readTOMLFile(name, keymapConfig, defaultConfigs, &keymap); err != nil {
		return fmt.Errorf("error reading keymap config: %w", err)
	}
	if err := readTOMLFile(name, languagesConfig, defaultConfigs, &languages); err != nil {
		return fmt.Errorf("error reading languages config: %w", err)
	}
	if err := readTOMLFile(name, languageServersConfig, defaultConfigs, &languageServers); err != nil {
		return fmt.Errorf("error reading language servers config: %w", err)
	}

	themes, err := loadThemes(name, defaultConfigs)
	if err != nil {
		return fmt.Errorf("error loading themes: %w", err)
	}

	Path = name
	Gopad = gopad
	Languages = languages.filter()
	LanguageServers = languageServers.filter()
	Keys = keymap.Keys()
	Themes = themes

	var theme RawThemeConfig
	for _, t := range Themes {
		theme = t
		if t.Name == Gopad.Theme {
			break
		}
	}
	Theme = theme.Theme()

	return nil
}

func loadThemes(name string, defaultConfigs embed.FS) ([]RawThemeConfig, error) {
	themes := make([]RawThemeConfig, 0)

	themeFiles, err := os.ReadDir(filepath.Join(name, themeDir))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error reading themes directory: %w", err)
	}
	for _, themeFile := range themeFiles {
		themeConfig, err := readTheme(name, themeFile, nil)
		if err != nil {
			return nil, fmt.Errorf("error reading theme file %s: %w", themeFile.Name(), err)
		}

		if themeConfig == nil {
			continue
		}
		themes = append(themes, *themeConfig)
	}

	defaultThemeFiles, err := defaultConfigs.ReadDir(filepath.Join(configDir, themeDir))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error reading default themes directory: %w", err)
	}

	for _, themeFile := range defaultThemeFiles {
		themeConfig, err := readTheme("", themeFile, &defaultConfigs)
		if err != nil {
			return nil, fmt.Errorf("error reading default theme file %s: %w", themeFile.Name(), err)
		}

		if themeConfig == nil || slices.ContainsFunc(themes, func(theme RawThemeConfig) bool {
			return theme.Name == themeConfig.Name
		}) {
			continue
		}
		themes = append(themes, *themeConfig)
	}

	return themes, nil
}

func readTheme(name string, entry os.DirEntry, defaultConfigs *embed.FS) (*RawThemeConfig, error) {
	if entry.IsDir() {
		return &RawThemeConfig{}, nil
	}

	var (
		f   fs.File
		err error
	)
	if defaultConfigs != nil {
		f, err = defaultConfigs.Open(filepath.Join(configDir, themeDir, entry.Name()))
	} else {
		f, err = os.Open(filepath.Join(name, themeDir, entry.Name()))
	}
	if err != nil {
		return nil, fmt.Errorf("error opening theme file: %w", err)
	}
	defer f.Close()

	var themeConfig RawThemeConfig
	if err = toml.NewDecoder(f).Decode(&themeConfig); err != nil {
		return nil, fmt.Errorf("error decoding theme file: %w", err)
	}

	return &themeConfig, nil
}

func readTOMLFile(localConfigDir string, name string, defaultConfigs embed.FS, a any) error {
	var (
		f   fs.File
		err error
	)
	_, err = os.Stat(filepath.Join(localConfigDir, name))
	if err != nil {
		f, err = defaultConfigs.Open(filepath.Join(configDir, name))
	} else {
		f, err = os.Open(filepath.Join(localConfigDir, name))
	}

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

func Create(name string, defaultConfigs embed.FS) error {
	log.Println("creating config in", name)
	return copyDir(configDir, name, defaultConfigs)
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
