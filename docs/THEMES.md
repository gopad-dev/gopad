# Themes

To choose a theme change the `theme` variable in the `gopad.toml` file.

```toml
theme = 'dark'
```

## Creating your own theme

To get started check the `themes` directory in the root of the project. You can copy one of the existing themes and modify it to your liking.

### Scopes

We use a similar set of scopes as [Helix](https://helix-editor.com/), [Sublime Text](https://www.sublimetext.com/docs/scope_naming.html)
and [TextMate](https://macromates.com/manual/en/language_grammars).

The following scopes are available for styling:

- `attribute` - Class attributes, HTML tag attributes
- `property` - Object properties
- `tag` - Tags (e.g. `<body>` in HTML)
    - `builtin` - Built-in tags (e.g. `div`, `span`, etc.)
- `label` - Labels (e.g. `Label:` to break out of a loop in Go)
- `operator` - `||`, `+=`, `>`, `&&`, etc.
- `namespace`
- `special`

- `keyword`
    - `control`
        - `conditional` - `if`, `else`
        - `repeat` - `for`, `while`, `loop`
        - `import` - `import`, `export`
        - `return`
        - `exception` - `try`, `catch`, `throw`
    - `operator` - `or`, `in`
    - `directive` - Preprocessor directives (`#if` in C)
    - `function` - `fn`, `func`
    - `storage` - Keywords describing how things are stored
        - `type` - The type of something, `class`, `function`, `var`, `let`, etc.
        - `modifier` - Storage modifiers like `static`, `mut`, `const`, `ref`, etc.

- `function`
    - `builtin` - Built-in functions (`len`, `print`, etc.)
    - `method`
        - `private` - Private methods
    - `macro`
    - `special` - (preprocessor in C)

- `variable` - Variables
    - `builtin` - Reserved language variables (`self`, `this`, `super`, etc.)
    - `parameter` - Function parameters
    - `other`
        - `member` - Fields of composite data types (e.g. structs, unions)
            - `private` - Private fields that use a unique syntax

- `type` - Types
    - `builtin` - Primitive types provided by the language (`int`, `uint`, `bool`, etc.)
    - `parameter` - Generic type parameters (`T`)
    - `enum`
        - `variant`
- `constructor` - Constructors

- `constant` - Constants
    - `builtin` Special constants provided by the language (`true`, `false`, `nil`, etc.)
        - `boolean`
    - `character`
        - `escape` - Escape sequences in strings (`\n`, `\t`, etc.)
    - `numeric` (numbers)
        - `integer`
        - `float`
    - `other` - Other constants
        - `placeholder` - Placeholders like `%v` in Go

- `string` - Strings
    - `regexp` - Regular expressions
    - `special`
        - `path` - File paths
        - `url` - URLs
        - `symbol` - Erlang/Elixir atoms, Ruby symbols, Clojure keywords

- `comment` - Code comments
    - `line` - Single line comments (`//`)
    - `block` - Block comments (e.g. (`/* */`)
        - `documentation` - Documentation comments (e.g. `///` in Rust)
    - `todo` - TODO comments (e.g. `TODO:`, `FIXME:`, etc.)
    - `note` - NOTE comments (e.g. `NOTE:`, `INFO:`, etc.)
    - `warning` - WARNING comments (e.g. `WARNING:`, `CAUTION:`, etc.)
    - `error` - ERROR comments (e.g. `ERROR:`, `BUG:`, etc.)

- `punctuation`
    - `delimiter` - Commas, colons
    - `bracket` - Parentheses, angle brackets, etc.
    - `special` - String interpolation brackets.

- `markup`
    - `heading` - Headings
        - `marker` - The `#` in Markdown headings
        - `1`, `2`, `3`, `4`, `5`, `6` - heading text for h1 through h6
    - `list` - Lists
        - `unnumbered` - Bullet lists
        - `numbered` - Numbered lists
        - `checked` - Checked list items
        - `unchecked` - Unchecked list items
    - `bold` - Bold text
    - `italic` - Italic text
    - `strikethrough` - Strikethrough text
    - `link`
        - `url` - URLs pointed to by links
        - `label` - non-URL link references
        - `text` - URL and image descriptions in links
    - `quote` - Block quotes
    - `raw` - Raw text
        - `inline` - Inline code blocks
        - `block` - Block code blocks

- `diff` - version control changes
    - `plus` - additions
        - `gutter` - gutter indicator
    - `minus` - deletions
        - `gutter` - gutter indicator
    - `delta` - modifications
        - `moved` - renamed or moved files/changes
        - `conflict` - merge conflicts
        - `gutter` - gutter indicator
