package config

import (
	"time"

	"go.gopad.dev/gopad/internal/bubbles/cursor"
)

type GopadConfig struct {
	Theme    string         `toml:"theme"`
	Editor   EditorConfig   `toml:"editor"`
	FileView FileViewConfig `toml:"file_view"`
	FileTree FileTreeConfig `toml:"file_tree"`
}

func DefaultGopadConfig() GopadConfig {
	return GopadConfig{
		Theme: "dark",
		Editor: EditorConfig{
			TabSize:                4,
			IndentSize:             4,
			EndOfLine:              "lf",
			Charset:                "utf-8",
			TrimTrailingWhitespace: true,
			InsertFinalNewline:     true,
			Theme:                  "default",
			Cursor: CursorConfig{
				Mode:          cursor.ModeBlink,
				BlinkInterval: Duration(cursor.DefaultBlinkInterval),
				Shape:         cursor.ShapeBlock,
			},
		},
		FileView: FileViewConfig{
			OpenFilesWrap:   false,
			ShowLineNumbers: true,
			WordWrap:        false,
		},
		FileTree: FileTreeConfig{
			Ignored: nil,
		},
	}
}

type EditorConfig struct {
	TabSize                int          `toml:"tab_size"`
	IndentSize             int          `toml:"indent_size"`
	EndOfLine              string       `toml:"end_of_line"`
	Charset                string       `toml:"charset"`
	TrimTrailingWhitespace bool         `toml:"trim_trailing_whitespace"`
	InsertFinalNewline     bool         `toml:"insert_final_newline"`
	Theme                  string       `toml:"theme"`
	Cursor                 CursorConfig `toml:"cursor"`
}

type CursorConfig struct {
	Mode          cursor.Mode  `toml:"mode"`
	BlinkInterval Duration     `toml:"blink_interval"`
	Shape         cursor.Shape `toml:"shape"`
}

type Duration time.Duration

func (d *Duration) UnmarshalText(text []byte) error {
	duration, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

type FileViewConfig struct {
	OpenFilesWrap   bool `toml:"open_files_wrap"`
	ShowLineNumbers bool `toml:"show_line_numbers"`
	WordWrap        bool `toml:"word_wrap"`
}

type FileTreeConfig struct {
	Ignored []string `toml:"ignored"`
}
