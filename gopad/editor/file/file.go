package file

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/charmbracelet/bubbletea/v2"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/buffer"
	"go.gopad.dev/gopad/internal/xrunes"
)

var wordBreakers = []rune{' ', '{', '}', '(', ')', '[', ']', '<', '>', '.', ',', ';', ':', '\'', '"', '`', '/', '\\', '|', '-', '=', '+', '*', '&', '^', '%', '$', '#', '@', '!', '~', '?', '\t'}

type Mode int

const (
	ModeReadOnly Mode = iota
	ModeWrite
)

type Change struct {
	StartByte   uint32
	OldEndByte  uint32
	NewEndByte  uint32
	StartPoint  buffer.Point
	OldEndPoint buffer.Point
	NewEndPoint buffer.Point

	Text    []byte
	Version uint64
}

func NewFileWithBuffer(b buffer.Buffer, mode Mode) (*File, error) {
	var syntax *Syntax
	if language := GetLanguageByFilename(b.Name()); language != nil {
		layers, err := NewSyntaxLayers(b.Bytes(), language.Grammar.Highlight)
		if err != nil {
			return nil, fmt.Errorf("error creating syntax layers: %w", err)
		}
		syntax, err = NewSyntax(language, layers)
		if err != nil {
			return nil, fmt.Errorf("error creating syntax: %w", err)
		}
	}

	f := &File{
		Buffer:             b,
		Mode:               mode,
		Syntax:             syntax,
		diagnosticVersions: map[ls.DiagnosticType]uint64{},
	}

	f.Autocomplete = NewAutocompleter(f)

	return f, nil
}

func NewFileFromName(name string) (*File, error) {
	stat, err := os.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %w", err)
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("cannot open directory")
	}

	if stat.Mode().Perm()&0400 == 0 {
		return nil, fmt.Errorf("file is not readable")
	}

	b, err := buffer.NewFromFile(name, "UTF-8", buffer.LineEndingAuto)
	if err != nil {
		return nil, err
	}

	mode := ModeWrite
	if stat.Mode().Perm()&0200 == 0 {
		mode = ModeReadOnly
	}

	return NewFileWithBuffer(b, mode)
}

type File struct {
	Buffer       buffer.Buffer
	Mode         Mode
	Syntax       *Syntax
	Autocomplete *Autocompleter

	diagnosticVersions map[ls.DiagnosticType]uint64
	Diagnostics        []ls.Diagnostic

	inlayHintsVersion uint64
	InlayHints        []ls.InlayHint

	Declarations    []ls.FileLocation
	Definitions     []ls.FileLocation
	TypeDefinitions []ls.FileLocation
	Implementations []ls.FileLocation
	References      []ls.FileLocation

	Positions [][]buffer.Point
	Changes   []Change
}

func (f *File) RelativeName(workspace string) string {
	relName, err := filepath.Rel(workspace, f.Buffer.Name())
	if err != nil {
		return f.Buffer.Name()
	}

	return relName
}

func (f *File) SetLanguage(name string) error {
	language := GetLanguage(name)
	if language == nil {
		return fmt.Errorf("language with name %q not found", name)
	}

	layers, err := NewSyntaxLayers(f.Buffer.Bytes(), language.Grammar.Highlight)
	if err != nil {
		return fmt.Errorf("error creating syntax layers: %w", err)
	}

	syntax, err := NewSyntax(language, layers)
	if err != nil {
		return fmt.Errorf("error creating syntax: %w", err)
	}

	f.Syntax = syntax
	return nil
}

func (f *File) Range() buffer.Range {
	return buffer.Range{
		Start: buffer.Point{Row: 0, Col: 0},
		End:   buffer.Point{Row: f.Buffer.LinesLen(), Col: f.Buffer.LineLen(max(f.Buffer.LinesLen()-1, 0))},
	}
}

func (f *File) recordChange(change Change) tea.Cmd {
	now := time.Now()
	defer func() {
		log.Println("record change time: ", time.Since(now))
	}()

	f.Changes = append(f.Changes, change)

	edits := []SyntaxEdit{
		{
			&tree_sitter.InputEdit{
				StartByte:  uint(change.StartByte),
				OldEndByte: uint(change.OldEndByte),
				NewEndByte: uint(change.NewEndByte),
				StartPosition: tree_sitter.Point{
					Row:    uint(change.StartPoint.Row),
					Column: uint(change.StartPoint.Col),
				},
				OldEndPosition: tree_sitter.Point{
					Row:    uint(change.OldEndPoint.Row),
					Column: uint(change.OldEndPoint.Col),
				},
				NewEndPosition: tree_sitter.Point{
					Row:    uint(change.NewEndPoint.Row),
					Column: uint(change.NewEndPoint.Col),
				},
			},
		},
	}
	if f.Syntax != nil {
		ctx := context.Background()
		f.Syntax.Parse(ctx, f.Buffer.Version(), f.Buffer.Bytes(), edits)
	}

	var cmds []tea.Cmd
	cmds = append(cmds, tea.Sequence(
		ls.FileChanged(f.Buffer.Name(), f.Buffer.Version(), change.Text),
		ls.GetInlayHint(f.Buffer.Name(), f.Buffer.Version(), f.Range()),
	))

	return tea.Batch(cmds...)
}

func (f *File) InsertNewLine(p buffer.Point) tea.Cmd {
	startIndex := f.Buffer.ByteIndex(p)
	f.Buffer.InsertNewLine(p)

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + 1),
		NewEndByte: uint32(startIndex + 2),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) Insert(p buffer.Point, text []byte) tea.Cmd {
	text = xrunes.Sanitize(text)
	if len(text) == 0 {
		return nil
	}

	startIndex := f.Buffer.ByteIndex(p)
	f.Buffer.Insert(p, text)

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + 1),
		NewEndByte: uint32(startIndex + len(text) + 1),
		StartPoint: p,
		OldEndPoint: buffer.Point{
			Row: p.Row,
			Col: p.Col,
		},
		NewEndPoint: buffer.Point{
			Row: p.Row,
			Col: p.Col + len(text),
		},
		Text:    f.Buffer.Bytes(),
		Version: f.Buffer.Version(),
	})
}

func (f *File) InsertRunes(p buffer.Point, text []rune) tea.Cmd {
	return f.Insert(p, []byte(string(text)))
}

func (f *File) Replace(r buffer.Range, text []byte) tea.Cmd {
	text = xrunes.Sanitize(text)

	startIndex := f.Buffer.ByteIndex(r.Start)
	endIndex := f.Buffer.ByteIndex(r.End)
	f.Buffer.Replace(r, text)

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(endIndex),
		NewEndByte: uint32(startIndex + len(text)),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) DuplicateLine(row int) tea.Cmd {
	line := f.Buffer.Line(row)
	startIndex := f.Buffer.ByteIndex(buffer.Point{
		Row: row,
		Col: 0,
	})
	f.Buffer.DuplicateLine(row)

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + 1),
		NewEndByte: uint32(startIndex + line.LenBytes() + 1),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) DeleteLine(row int) tea.Cmd {
	line := f.Buffer.Line(row)
	startIndex := f.Buffer.ByteIndex(buffer.Point{
		Row: row,
		Col: 0,
	})

	f.Buffer.DeleteLine(row)

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + line.LenBytes() + 1),
		NewEndByte: uint32(startIndex + 1),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) DeleteBefore(p buffer.Point) tea.Cmd {
	startIndex := f.Buffer.ByteIndex(p)
	f.Buffer.DeleteBefore(p)

	return f.recordChange(Change{
		StartByte:  uint32(startIndex - 1),
		OldEndByte: uint32(startIndex + 1),
		NewEndByte: uint32(startIndex),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) DeleteAfter(p buffer.Point) tea.Cmd {
	startIndex := f.Buffer.ByteIndex(p)

	f.Buffer.DeleteAfter(p)

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + 1),
		NewEndByte: uint32(startIndex + 2),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) DeleteRange(r buffer.Range) tea.Cmd {
	startIndex := f.Buffer.ByteIndex(r.Start)
	endIndex := f.Buffer.ByteIndex(r.End)
	f.Buffer.DeleteRange(r)

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(endIndex),
		NewEndByte: uint32(startIndex),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) DeleteWordLeft(p buffer.Point) tea.Cmd {
	startPoint := f.NextWordLeft(p)
	startIndex := f.Buffer.ByteIndex(startPoint)
	endIndex := f.Buffer.ByteIndex(p)
	f.Buffer.DeleteRange(buffer.Range{
		Start: startPoint,
		End:   p,
	})

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(endIndex),
		NewEndByte: uint32(startIndex),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) DeleteWordRight(p buffer.Point) tea.Cmd {
	endPoint := f.NextWordRight(p)
	startIndex := f.Buffer.ByteIndex(p)
	endIndex := f.Buffer.ByteIndex(endPoint)
	f.Buffer.DeleteRange(buffer.Range{
		Start: p,
		End:   endPoint,
	})

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(endIndex),
		NewEndByte: uint32(startIndex),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) AddTab(row int) tea.Cmd {
	line := f.Buffer.Line(row)
	startIndex := f.Buffer.ByteIndex(buffer.Point{
		Row: row,
		Col: 0,
	})
	f.Buffer.AddTab(row)

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + line.LenBytes() + 1),
		NewEndByte: uint32(startIndex + line.LenBytes() + 2),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) RemoveTab(row int) tea.Cmd {
	line := f.Buffer.Line(row)
	startIndex := f.Buffer.ByteIndex(buffer.Point{
		Row: row,
		Col: 0,
	})
	f.Buffer.RemoveTab(row)

	return f.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + line.LenBytes() + 1),
		NewEndByte: uint32(startIndex + line.LenBytes() - 1),
		Text:       f.Buffer.Bytes(),
		Version:    f.Buffer.Version(),
	})
}

func (f *File) NextWordLeft(p buffer.Point) buffer.Point {
	if p.Col == 0 {
		if p.Row == 0 {
			return p
		}
		p.Row--
		p.Col = f.Buffer.LineLen(p.Row - 1)
		return p
	}

	var ready bool
	for {
		if p.Col == 0 || (ready && slices.Contains(wordBreakers, f.Buffer.Line(p.Row).Rune(p.Col))) {
			break
		}
		p.Col--
		if f.Buffer.Line(p.Row).Rune(p.Col) != ' ' {
			ready = true
		}
	}

	return p
}

func (f *File) NextWordRight(p buffer.Point) buffer.Point {
	if p.Col == f.Buffer.LineLen(p.Row) {
		if p.Row == f.Buffer.LinesLen()-1 {
			return p
		}
		p.Row++
		return p
	}

	var ready bool
	for {
		if p.Col == f.Buffer.LineLen(p.Row) || (ready && slices.Contains(wordBreakers, f.Buffer.Line(p.Row).Rune(p.Col))) {
			break
		}
		p.Col++
		if f.Buffer.Line(p.Row).Rune(p.Col-1) != ' ' {
			ready = true
		}
	}

	return p
}
