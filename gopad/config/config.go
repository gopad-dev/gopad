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
	"go.opentelemetry.io/otel/trace"
)

const (
	gopadConfig           = "gopad.toml"
	languagesConfig       = "languages.toml"
	languageServersConfig = "language_servers.toml"
	configDir             = "config"
	keymapsDir            = "keymaps"
	themesDir             = "themes"
)

var (
	Path            string
	Gopad           GopadConfig
	Languages       LanguageConfigs
	LanguageServers LanguageServerConfigs
	Keys            Keymap
	Keymaps         []KeymapConfig
	Theme           ThemeStyles
	Themes          []ThemeConfig
	Tracer          trace.Tracer
)

type Identifiable interface {
	ID() string
}

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
	gopad, err := readFileFallback[GopadConfig](name, gopadConfig, defaultConfigs)
	if err != nil {
		return fmt.Errorf("error reading gopad config: %w", err)
	}

	languages, err := readFileFallback[LanguageConfigs](name, languagesConfig, defaultConfigs)
	if err != nil {
		return fmt.Errorf("error reading languages config: %w", err)
	}

	languageServers, err := readFileFallback[LanguageServerConfigs](name, languageServersConfig, defaultConfigs)
	if err != nil {
		return fmt.Errorf("error reading language servers config: %w", err)
	}

	keymaps, err := readDir[KeymapConfig](name, keymapsDir, defaultConfigs)
	if err != nil {
		return fmt.Errorf("error loading keymaps: %w", err)
	}

	themes, err := readDir[ThemeConfig](name, themesDir, defaultConfigs)
	if err != nil {
		return fmt.Errorf("error loading themes: %w", err)
	}

	Path = name
	Gopad = gopad
	Languages = languages.filter()
	LanguageServers = languageServers.filter()

	Keymaps = keymaps
	var keymap KeymapConfig
	for _, k := range Keymaps {
		keymap = k
		if k.Name == Gopad.Keymap {
			break
		}
	}
	Keys = keymap.KeyMap()

	Themes = themes
	var theme ThemeConfig
	for _, t := range Themes {
		theme = t
		if t.Name == Gopad.Theme {
			break
		}
	}
	Theme = theme.Theme()

	return nil
}

func readDir[T Identifiable](name string, dir string, defaultConfigs embed.FS) ([]T, error) {
	files := make([]T, 0)

	configFiles, err := os.ReadDir(filepath.Join(name, dir))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	for _, configFile := range configFiles {
		if configFile.IsDir() {
			continue
		}

		config, err := readFile[T](dir, name, configFile, nil)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %w", configFile.Name(), err)
		}

		files = append(files, config)
	}

	defaultConfigFiles, err := defaultConfigs.ReadDir(filepath.Join(configDir, dir))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error reading default directory: %w", err)
	}

	for _, configFile := range defaultConfigFiles {
		if configFile.IsDir() {
			continue
		}

		config, err := readFile[T](dir, "", configFile, &defaultConfigs)
		if err != nil {
			return nil, fmt.Errorf("error reading default file %s: %w", configFile.Name(), err)
		}

		if slices.ContainsFunc(files, func(i T) bool {
			return i.ID() == config.ID()
		}) {
			continue
		}

		files = append(files, config)
	}

	return files, nil
}

func readFile[T Identifiable](dir string, name string, entry os.DirEntry, defaultConfigs *embed.FS) (T, error) {
	var (
		f   fs.File
		err error
	)
	if defaultConfigs != nil {
		f, err = defaultConfigs.Open(filepath.Join(configDir, dir, entry.Name()))
	} else {
		f, err = os.Open(filepath.Join(name, dir, entry.Name()))
	}
	if err != nil {
		var zero T
		return zero, fmt.Errorf("error opening config file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	var config T
	if err = toml.NewDecoder(f).Decode(&config); err != nil {
		var zero T
		return zero, fmt.Errorf("error decoding config file: %w", err)
	}

	return config, nil
}

func readFileFallback[T any](localConfigDir string, name string, defaultConfigs embed.FS) (T, error) {
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
		var zero T
		return zero, fmt.Errorf("error opening config file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	var config T
	if err = toml.NewDecoder(f).Decode(&config); err != nil {
		var zero T
		return zero, fmt.Errorf("error decoding config file: %w", err)
	}

	return config, nil
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
	defer func() {
		_ = out.Close()
	}()

	f, err := defaultConfigs.Open(name)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if _, err = io.Copy(out, f); err != nil {
		return fmt.Errorf("error copying config file: %w", err)
	}

	return nil
}
