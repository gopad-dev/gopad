# Config

GoPad uses [toml](https://toml.io) for configuration. The configuration directory is located at:

- `$GOPAD_CONFIG_HOME` environment variable
- `$XDG_CONFIG_HOME/gopad` (`~/.config/gopad`)
- `$HOME/.config/gopad`

The configuration file is named `gopad.toml`. The following options are available:

For theme settings see [THEMES](config/THEMES.md).

For keymap settings see [KEYMAPS](config/KEYMAPS.md).

```toml
# Set the theme to use.
theme = 'dark'
# Set the keymap to use.
keympa = 'default'

# Editor configuration
[editor]
tab_size = 4
indent_size = 4
end_of_line = 'lf'
charset = 'utf-8'
trim_trailing_whitespace = true
insert_final_newline = true

# Cursor configuration
[editor.cursor]
mode = 'blink'
blink_interval = '530ms'
shape = 'block'

# File view configuration
[file_view]
open_files_wrap = true
line_numbers = true
word_wrap = false
scroll_past_end = true

# File tree configuration
[file_tree]
ignored = [
    '.gopad',
    '.git',
    '.idea',
    '.vscode',
    'node_modules'
]
```