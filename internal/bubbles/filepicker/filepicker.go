package filepicker

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/internal/bubbles/help"
)

var (
	lastID int
	idMtx  sync.Mutex
)

// Return the next ID we should use on the Model.
func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

// KeyMap defines key bindings for each user action.
type KeyMap struct {
	GoToTop  key.Binding
	GoToEnd  key.Binding
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Back     key.Binding
	Open     key.Binding
	Select   key.Binding
}

func (m KeyMap) HelpView() []help.KeyMapCategory {
	return []help.KeyMapCategory{
		{
			Category: "File Picker",
			Keys: []key.Binding{
				m.GoToTop,
				m.GoToEnd,
				m.Up,
				m.Down,
				m.PageUp,
				m.PageDown,
				m.Back,
				m.Open,
				m.Select,
			},
		},
	}
}

// DefaultKeyMap defines the default keybindings.
var DefaultKeyMap = KeyMap{
	GoToTop: key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("home", "top"),
	),
	GoToEnd: key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", "end"),
	),
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("up", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("down", "down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "page down"),
	),
	Back: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("left", "back"),
	),
	Open: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("right", "open"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
}

// Styles defines the possible customizations for styles in the file picker.
type Styles struct {
	Selected         lipgloss.Style
	DisabledSelected lipgloss.Style

	Symlink        lipgloss.Style
	Directory      lipgloss.Style
	EmptyDirectory lipgloss.Style

	File         lipgloss.Style
	DisabledFile lipgloss.Style
	FileSize     lipgloss.Style

	Permission lipgloss.Style
}

// DefaultStyles defines the default styling for the file picker.
var DefaultStyles = Styles{
	Selected:         lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
	DisabledSelected: lipgloss.NewStyle().Foreground(lipgloss.Color("247")),
	Symlink:          lipgloss.NewStyle().Foreground(lipgloss.Color("36")),
	Directory:        lipgloss.NewStyle().Foreground(lipgloss.Color("99")),
	EmptyDirectory:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	File:             lipgloss.NewStyle(),
	DisabledFile:     lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
	FileSize:         lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(fileSizeWidth).Align(lipgloss.Right),
	Permission:       lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
}

type errorMsg struct {
	err error
}

type readDirMsg struct {
	id      int
	entries []os.DirEntry
	lastDir string
}

const fileSizeWidth = 7

// New returns a new filepicker model with default styling and key bindings.
func New() Model {
	return Model{
		id:               nextID(),
		CurrentDirectory: ".",
		AllowedTypes:     []string{},
		selected:         0,
		offset:           0,
		ShowPermissions:  true,
		ShowSize:         true,
		ShowHidden:       false,
		DirAllowed:       false,
		FileAllowed:      true,
		selectedStack:    newStack(),
		offsetStack:      newStack(),
		KeyMap:           DefaultKeyMap,
		Styles:           DefaultStyles,
	}
}

// Model represents a file picker.
type Model struct {
	id int

	// Path is the path which the user has selected with the file picker.
	Path string

	// CurrentDirectory is the directory that the user is currently in.
	CurrentDirectory string

	// AllowedTypes specifies which file types the user may select.
	// If empty the user may select any file.
	AllowedTypes []string

	KeyMap          KeyMap
	files           []os.DirEntry
	ShowPermissions bool
	ShowSize        bool
	ShowHidden      bool
	DirAllowed      bool
	FileAllowed     bool

	selected      int
	offset        int
	selectedStack stack
	offsetStack   stack

	Styles Styles
}

type stack struct {
	Push   func(int)
	Pop    func() int
	Length func() int
}

func newStack() stack {
	slice := make([]int, 0)
	return stack{
		Push: func(i int) {
			slice = append(slice, i)
		},
		Pop: func() int {
			res := slice[len(slice)-1]
			slice = slice[:len(slice)-1]
			return res
		},
		Length: func() int {
			return len(slice)
		},
	}
}

func (m *Model) pushView(selected int, offset int) {
	m.selectedStack.Push(selected)
	m.offsetStack.Push(offset)
}

func (m *Model) popView() (int, int) {
	return m.selectedStack.Pop(), m.offsetStack.Pop()
}

func (m Model) readDir(path string, showHidden bool, lastDir string) tea.Cmd {
	return func() tea.Msg {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return errorMsg{err}
		}

		sort.Slice(dirEntries, func(i, j int) bool {
			if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
				return dirEntries[i].Name() < dirEntries[j].Name()
			}
			return dirEntries[i].IsDir()
		})

		if showHidden {
			return readDirMsg{
				id:      m.id,
				entries: dirEntries,
				lastDir: lastDir,
			}
		}

		var sanitizedDirEntries []os.DirEntry
		for _, dirEntry := range dirEntries {
			isHidden, _ := IsHidden(dirEntry.Name())
			if isHidden {
				continue
			}
			sanitizedDirEntries = append(sanitizedDirEntries, dirEntry)
		}
		return readDirMsg{
			id:      m.id,
			entries: sanitizedDirEntries,
			lastDir: lastDir,
		}
	}
}

// DidSelect returns whether a user has selected a file (on this msg).
func (m Model) DidSelect(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelect(msg)
	if didSelect && m.canSelect(path) {
		return true, path
	}
	return false, ""
}

// DidSelectDisabled returns whether a user tried to select a disabled file
// (on this msg). This is necessary only if you would like to warn the user that
// they tried to select a disabled file.
func (m Model) DidSelectDisabled(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelect(msg)
	if didSelect && !m.canSelect(path) {
		return true, path
	}
	return false, ""
}

func (m Model) didSelect(msg tea.Msg) (bool, string) {
	if len(m.files) == 0 {
		return false, ""
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// If the msg does not match the Select keymap then this could not have been a selection.
		if !key.Matches(msg, m.KeyMap.Select) {
			return false, ""
		}

		// The key press was a selection, let's confirm whether the current file could
		// be selected or used for navigating deeper into the stack.
		f := m.files[m.selected]
		info, err := f.Info()
		if err != nil {
			return false, ""
		}
		isSymlink := info.Mode()&os.ModeSymlink != 0
		isDir := f.IsDir()

		if isSymlink {
			symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
			info, err = os.Stat(symlinkPath)
			if err != nil {
				break
			}
			if info.IsDir() {
				isDir = true
			}
		}

		if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) && m.Path != "" {
			return true, m.Path
		}

		// If the msg was not a KeyMsg, then the file could not have been selected this iteration.
		// Only a KeyMsg can select a file.
	default:
		return false, ""
	}
	return false, ""
}

func (m Model) canSelect(file string) bool {
	if len(m.AllowedTypes) == 0 {
		return true
	}

	for _, ext := range m.AllowedTypes {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}

func (m *Model) refreshViewOffset(height int) {
	if m.selected-m.offset > height-(1+1) {
		m.offset = max(min(m.selected-height+(1+1), len(m.files)-height), 0)
	} else if m.selected-m.offset < 1 {
		m.offset = max(m.selected-1, 0)
	}
}

// Init initializes the file picker model.
func (m Model) Init() tea.Cmd {
	return m.readDir(m.CurrentDirectory, m.ShowHidden, "")
}

// Update handles user interactions within the file picker model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case readDirMsg:
		if msg.id != m.id {
			break
		}
		m.files = msg.entries
		if msg.lastDir != "" {
			for i, f := range m.files {
				if f.Name() == msg.lastDir {
					m.selected = i
					break
				}
			}
		}
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.GoToTop):
			m.selected = 0
		case key.Matches(msg, m.KeyMap.GoToEnd):
			m.selected = len(m.files) - 1
		case key.Matches(msg, m.KeyMap.Up):
			m.selected--
			if m.selected < 0 {
				m.selected = 0
			}
		case key.Matches(msg, m.KeyMap.Down):
			m.selected++
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
		case key.Matches(msg, m.KeyMap.PageUp):
			m.selected -= 5
			if m.selected < 0 {
				m.selected = 0
			}
		case key.Matches(msg, m.KeyMap.PageDown):
			m.selected += 5
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
		case key.Matches(msg, m.KeyMap.Back):
			base := filepath.Base(m.CurrentDirectory)
			m.CurrentDirectory = filepath.Dir(m.CurrentDirectory)
			if m.selectedStack.Length() > 0 {
				m.selected, m.offset = m.popView()
			} else {
				m.selected = 0
				m.offset = 0
			}
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden, base)
		case key.Matches(msg, m.KeyMap.Select), key.Matches(msg, m.KeyMap.Open):
			if len(m.files) == 0 {
				break
			}

			f := m.files[m.selected]
			info, err := f.Info()
			if err != nil {
				break
			}
			isSymlink := info.Mode()&os.ModeSymlink != 0
			isDir := f.IsDir()

			if isSymlink {
				symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
				info, err = os.Stat(symlinkPath)
				if err != nil {
					break
				}
				if info.IsDir() {
					isDir = true
				}
			}

			if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) {
				if key.Matches(msg, m.KeyMap.Select) {
					// Select the current path as the selection
					m.Path = filepath.Join(m.CurrentDirectory, f.Name())
				}
			}

			if !isDir {
				break
			}

			m.CurrentDirectory = filepath.Join(m.CurrentDirectory, f.Name())
			m.pushView(m.selected, m.offset)
			m.selected = 0
			m.offset = 0
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden, "")
		}
	}
	return m, nil
}

// View returns the view of the file picker.
func (m Model) View(height int) string {
	if len(m.files) == 0 {
		return m.Styles.EmptyDirectory.MaxHeight(height).Render("No Files Found.")
	}
	m.refreshViewOffset(height)
	var s strings.Builder

	for i := 0; i < height; i++ {
		fi := i + m.offset
		if fi >= len(m.files) {
			break
		}
		f := m.files[fi]

		info, _ := f.Info()
		name := f.Name()
		size := strings.Replace(humanize.Bytes(uint64(info.Size())), " ", "", 1)
		disabled := !m.canSelect(name) && !f.IsDir()

		var symlinkPath string
		isSymlink := info.Mode()&os.ModeSymlink != 0
		if isSymlink {
			symlinkPath, _ = filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, name))
		}

		var style lipgloss.Style
		if m.selected == fi {
			if disabled {
				style = m.Styles.DisabledSelected
			} else {
				style = m.Styles.Selected
			}
		}

		var line string
		if m.ShowPermissions {
			line += style.Inherit(m.Styles.Permission).Render(info.Mode().String())
		}
		if m.ShowSize {
			line += style.Inherit(m.Styles.FileSize).Render(size)
		}

		fileStyle := m.Styles.File
		switch {
		case f.IsDir():
			fileStyle = m.Styles.Directory
		case isSymlink:
			fileStyle = m.Styles.Symlink
		case disabled:
			fileStyle = m.Styles.DisabledFile
		}
		if isSymlink {
			name += " → " + symlinkPath
		}
		line += style.Inherit(fileStyle).Render(" ", name)

		s.WriteString(style.Render(" "))
		s.WriteString(line)
		s.WriteString(style.Render(" "))
		s.WriteRune('\n')
	}

	return s.String()
}
