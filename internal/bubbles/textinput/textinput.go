package textinput

import (
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/runeutil"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	rw "github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"

	"go.gopad.dev/gopad/internal/bubbles/cursor"
)

// Internal messages for clipboard operations.
type pasteMsg string
type pasteErrMsg struct{ error }

// EchoMode sets the input behavior of the text input field.
type EchoMode int

const (
	// EchoNormal displays text as is. This is the default behavior.
	EchoNormal EchoMode = iota

	// EchoPassword displays the EchoCharacter mask instead of actual
	// characters. This is commonly used for password fields.
	EchoPassword

	// EchoNone displays nothing as characters are entered. This is commonly
	// seen for password fields on the command line.
	EchoNone
)

// ValidateFunc is a function that returns an error if the input is invalid.
type ValidateFunc func(string) error

// KeyMap is the key bindings for different actions within the textinput.
type KeyMap struct {
	CharacterRight key.Binding
	CharacterLeft  key.Binding

	WordRight key.Binding
	WordLeft  key.Binding

	DeleteLeft  key.Binding
	DeleteRight key.Binding

	LineStart key.Binding
	LineEnd   key.Binding

	Paste key.Binding
}

// DefaultKeyMap is the default set of key bindings for navigating and acting upon the textinput.
var DefaultKeyMap = KeyMap{
	CharacterLeft: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("🠄", "left"),
	),
	CharacterRight: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("🠆", "right"),
	),
	WordLeft: key.NewBinding(
		key.WithKeys("ctrl+left"),
		key.WithHelp("ctrl+🠄", "word left"),
	),
	WordRight: key.NewBinding(
		key.WithKeys("ctrl+right"),
		key.WithHelp("ctrl+🠆", "word right"),
	),
	LineStart: key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("home", "line start"),
	),
	LineEnd: key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", "line end"),
	),
	Paste: key.NewBinding(
		key.WithKeys("ctrl+v"),
		key.WithHelp("ctrl+v", "paste"),
	),
	DeleteLeft: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete left"),
	),
	DeleteRight: key.NewBinding(
		key.WithKeys("delete"),
		key.WithHelp("delete", "delete right"),
	),
}

type Styles struct {
	PromptStyle        lipgloss.Style
	FocusedPromptStyle lipgloss.Style
	TextStyle          lipgloss.Style
	PlaceholderStyle   lipgloss.Style
}

var DefaultStyles = Styles{
	PromptStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	PlaceholderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
}

// New creates a new model with default settings.
func New() Model {
	return Model{
		Prompt:        "> ",
		EchoCharacter: '*',
		CharLimit:     0,
		Styles:        DefaultStyles,
		Cursor:        cursor.New(),
		KeyMap:        DefaultKeyMap,

		value: nil,
		focus: false,
		pos:   0,
	}
}

// Model is the Bubble Tea model for this text input element.
type Model struct {
	Err error

	// General settings.
	Prompt        string
	Placeholder   string
	EchoMode      EchoMode
	EchoCharacter rune
	Cursor        cursor.Model

	// Styles. These will be applied as inline config.Styles.
	//
	// For an introduction to styling with Lip Gloss see:
	// https://github.com/charmbracelet/lipgloss
	Styles Styles

	// CharLimit is the maximum amount of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// Width is the maximum number of characters that can be displayed at once.
	// It essentially treats the text field like a horizontally scrolling
	// viewport. If 0 or less this setting is ignored.
	Width int

	// KeyMap encodes the keybindings recognized by the widget.
	KeyMap KeyMap

	// Underlying text value.
	value []rune

	// focus indicates whether user input focus should be on this input
	// component. When false, ignore keyboard input and hide the cursor.
	focus bool

	// Cursor position.
	pos int

	// Used to emulate a viewport when width is set and the content is
	// overflowing.
	offset      int
	offsetRight int

	// Validate is a function that checks whether or not the text within the
	// input is valid. If it is not valid, the `Err` field will be set to the
	// error returned by the function. If the function is not defined, all
	// input is considered valid.
	Validate ValidateFunc

	// rune sanitizer for input.
	rsan runeutil.Sanitizer
}

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	// Clean up any special characters in the input provided by the
	// caller. This avoids bugs due to e.g. tab characters and whatnot.
	runes := m.san().Sanitize([]rune(s))
	m.setValueInternal(runes)
}

func (m *Model) setValueInternal(runes []rune) {
	if m.Validate != nil {
		if err := m.Validate(string(runes)); err != nil {
			m.Err = err
			return
		}
	}

	empty := len(m.value) == 0
	m.Err = nil

	if m.CharLimit > 0 && len(runes) > m.CharLimit {
		m.value = runes[:m.CharLimit]
	} else {
		m.value = runes
	}
	if (m.pos == 0 && empty) || m.pos > len(m.value) {
		m.SetCursor(len(m.value))
	}
	m.handleOverflow()
}

// Value returns the value of the text input.
func (m Model) Value() string {
	return string(m.value)
}

// Position returns the cursor position.
func (m Model) Position() int {
	return m.pos
}

// SetCursor moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
func (m *Model) SetCursor(pos int) {
	m.pos = clamp(pos, 0, len(m.value))
	m.handleOverflow()
}

// CursorStart moves the cursor to the start of the input field.
func (m *Model) CursorStart() {
	m.SetCursor(0)
}

// CursorEnd moves the cursor to the end of the input field.
func (m *Model) CursorEnd() {
	m.SetCursor(len(m.value))
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model. When the model is in focus it can
// receive keyboard input and the cursor will be shown.
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	return m.Cursor.Focus()
}

// Blur removes the focus state on the model.  When the model is blurred it can
// not receive keyboard input and the cursor will be hidden.
func (m *Model) Blur() {
	m.focus = false
	m.Cursor.Blur()
}

// Reset sets the input to its default state with no input.
func (m *Model) Reset() {
	m.value = nil
	m.SetCursor(0)
}

// rsan initializes or retrieves the rune sanitizer.
func (m *Model) san() runeutil.Sanitizer {
	if m.rsan == nil {
		// Textinput has all its input on a single line so collapse
		// newlines/tabs to single spaces.
		m.rsan = runeutil.NewSanitizer(
			runeutil.ReplaceTabs(" "), runeutil.ReplaceNewlines(" "))
	}
	return m.rsan
}

func (m *Model) insertRunesFromUserInput(v []rune) {
	// Clean up any special characters in the input provided by the
	// clipboard. This avoids bugs due to e.g. tab characters and
	// whatnot.
	paste := m.san().Sanitize(v)

	var availSpace int
	if m.CharLimit > 0 {
		availSpace = m.CharLimit - len(m.value)

		// If the char limit's been reached, cancel.
		if availSpace <= 0 {
			return
		}

		// If there's not enough space to paste the whole thing cut the pasted
		// runes down so they'll fit.
		if availSpace < len(paste) {
			paste = paste[:availSpace]
		}
	}

	// Stuff before and after the cursor
	head := m.value[:m.pos]
	tailSrc := m.value[m.pos:]
	tail := make([]rune, len(tailSrc))
	copy(tail, tailSrc)

	oldPos := m.pos

	// Insert pasted runes
	for _, r := range paste {
		head = append(head, r)
		m.pos++
		if m.CharLimit > 0 {
			availSpace--
			if availSpace <= 0 {
				break
			}
		}
	}

	// Put it all back together
	value := append(head, tail...)
	m.setValueInternal(value)

	if m.Err != nil {
		m.pos = oldPos
	}
}

// If a max width is defined, perform some logic to treat the visible area
// as a horizontally scrolling viewport.
func (m *Model) handleOverflow() {
	if m.Width <= 0 || uniseg.StringWidth(string(m.value)) <= m.Width {
		m.offset = 0
		m.offsetRight = len(m.value)
		return
	}

	// Correct right offset if we've deleted characters
	m.offsetRight = min(m.offsetRight, len(m.value))

	if m.pos < m.offset {
		m.offset = m.pos

		w := 0
		i := 0
		runes := m.value[m.offset:]

		for i < len(runes) && w <= m.Width {
			w += rw.RuneWidth(runes[i])
			if w <= m.Width+1 {
				i++
			}
		}

		m.offsetRight = m.offset + i
	} else if m.pos >= m.offsetRight {
		m.offsetRight = m.pos

		w := 0
		runes := m.value[:m.offsetRight]
		i := len(runes) - 1

		for i > 0 && w < m.Width {
			w += rw.RuneWidth(runes[i])
			if w <= m.Width {
				i--
			}
		}

		m.offset = m.offsetRight - (len(runes) - 1 - i)
	}
}

// wordBackward moves the cursor one word to the left. If input is masked, move
// input to the start so as not to reveal word breaks in the masked input.
func (m *Model) wordBackward() {
	if m.pos == 0 || len(m.value) == 0 {
		return
	}

	if m.EchoMode != EchoNormal {
		m.CursorStart()
		return
	}

	i := m.pos - 1
	for i >= 0 {
		if unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos - 1)
			i--
		} else {
			break
		}
	}

	for i >= 0 {
		if !unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos - 1)
			i--
		} else {
			break
		}
	}
}

// wordForward moves the cursor one word to the right. If the input is masked,
// move input to the end so as not to reveal word breaks in the masked input.
func (m *Model) wordForward() {
	if m.pos >= len(m.value) || len(m.value) == 0 {
		return
	}

	if m.EchoMode != EchoNormal {
		m.CursorEnd()
		return
	}

	i := m.pos
	for i < len(m.value) {
		if unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos + 1)
			i++
		} else {
			break
		}
	}

	for i < len(m.value) {
		if !unicode.IsSpace(m.value[i]) {
			m.SetCursor(m.pos + 1)
			i++
		} else {
			break
		}
	}
}

func (m Model) echoTransform(v string) string {
	switch m.EchoMode {
	case EchoPassword:
		return strings.Repeat(string(m.EchoCharacter), uniseg.StringWidth(v))
	case EchoNone:
		return ""
	case EchoNormal:
		return v
	default:
		return v
	}
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	// Let's remember where the position of the cursor currently is so that if
	// the cursor position changes, we can reset the blink.
	oldPos := m.pos // nolint

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.DeleteLeft):
			m.Err = nil
			if len(m.value) > 0 {
				m.value = append(m.value[:max(0, m.pos-1)], m.value[m.pos:]...)
				if m.pos > 0 {
					m.SetCursor(m.pos - 1)
				}
			}
		case key.Matches(msg, m.KeyMap.WordLeft):
			m.wordBackward()
		case key.Matches(msg, m.KeyMap.CharacterLeft):
			if m.pos > 0 {
				m.SetCursor(m.pos - 1)
			}
		case key.Matches(msg, m.KeyMap.WordRight):
			m.wordForward()
		case key.Matches(msg, m.KeyMap.CharacterRight):
			if m.pos < len(m.value) {
				m.SetCursor(m.pos + 1)
			}

		case key.Matches(msg, m.KeyMap.LineStart):
			m.CursorStart()
		case key.Matches(msg, m.KeyMap.DeleteRight):
			if len(m.value) > 0 && m.pos < len(m.value) {
				m.value = append(m.value[:m.pos], m.value[m.pos+1:]...)
			}
		case key.Matches(msg, m.KeyMap.LineEnd):
			m.CursorEnd()
		case key.Matches(msg, m.KeyMap.Paste):
			return m, Paste
		default:
			// Input one or more regular characters.
			m.insertRunesFromUserInput(msg.Runes)
		}

	case pasteMsg:
		m.insertRunesFromUserInput([]rune(msg))

	case pasteErrMsg:
		m.Err = msg
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.Cursor, cmd = m.Cursor.Update(msg)
	cmds = append(cmds, cmd)

	if oldPos != m.pos && m.Cursor.Mode() == cursor.ModeBlink {
		m.Cursor.Blink = false
		cmds = append(cmds, m.Cursor.BlinkCmd())
	}

	m.handleOverflow()
	return m, tea.Batch(cmds...)
}

// View renders the textinput in its current state.
func (m Model) View() string {
	// Placeholder text
	if len(m.value) == 0 && m.Placeholder != "" {
		return m.placeholderView()
	}

	styleText := m.Styles.TextStyle.Inline(true)

	value := m.value[m.offset:m.offsetRight]
	pos := max(0, m.pos-m.offset)
	v := styleText.Render(m.echoTransform(string(value[:pos])))

	if pos < len(value) {
		char := m.echoTransform(string(value[pos]))
		v += m.Cursor.View(char, styleText)                           // cursor and text under it
		v += styleText.Render(m.echoTransform(string(value[pos+1:]))) // text after cursor
	} else {
		v += m.Cursor.View(" ", styleText)
	}

	// If a max width and background color were set fill the empty spaces with
	// the background color.
	valWidth := uniseg.StringWidth(string(value))
	if m.Width > 0 && valWidth <= m.Width {
		padding := max(0, m.Width-valWidth)
		if valWidth+padding <= m.Width && pos < len(value) {
			padding++
		}
		v += styleText.Render(strings.Repeat(" ", padding))
	}

	return m.viewPrompt() + v
}

func (m Model) viewPrompt() string {
	if m.Focused() {
		return m.Styles.FocusedPromptStyle.Render(m.Prompt)
	}
	return m.Styles.PromptStyle.Render(m.Prompt)
}

func (m Model) placeholderView() string {
	var (
		v     string
		p     = []rune(m.Placeholder)
		style = m.Styles.PlaceholderStyle.Inline(true).Render
	)

	v += m.Cursor.View(string(p[:1]), m.Styles.PlaceholderStyle)

	// If the entire placeholder is already set and no padding is needed, finish
	if m.Width < 1 && len(p) <= 1 {
		return m.viewPrompt() + v
	}

	// If Width is set then size placeholder accordingly
	if m.Width > 0 {
		// available width is width - len + cursor offset of 1
		minWidth := lipgloss.Width(m.Placeholder)
		availWidth := m.Width - minWidth + 1

		// if width < len, 'subtract'(add) number to len and dont add padding
		if availWidth < 0 {
			minWidth += availWidth
			availWidth = 0
		}
		// append placeholder[len] - cursor, append padding
		v += style(string(p[1:minWidth]))
		v += style(strings.Repeat(" ", availWidth))
	} else {
		// if there is no width, the placeholder can be any length
		v += style(string(p[1:]))
	}

	return m.viewPrompt() + v
}

// Blink is a command used to initialize cursor blinking.
func Blink() tea.Msg {
	return cursor.Blink()
}

// Paste is a command for pasting from the clipboard into the text input.
func Paste() tea.Msg {
	str, err := clipboard.ReadAll()
	if err != nil {
		return pasteErrMsg{err}
	}
	return pasteMsg(str)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
