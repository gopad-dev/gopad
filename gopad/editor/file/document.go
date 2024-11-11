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
	"github.com/tree-sitter/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/editor/buffer"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/xrunes"
)

var wordBreakers = []rune{' ', '{', '}', '(', ')', '[', ']', '<', '>', '.', ',', ';', ':', '\'', '"', '`', '/', '\\', '|', '-', '=', '+', '*', '&', '^', '%', '$', '#', '@', '!', '~', '?', '\t'}

type Mode int

const (
	ModeReadOnly Mode = iota
	ModeWrite
)

type Changee struct {
	StartByte   int
	OldEndByte  int
	NewEndByte  int
	StartPoint  buffer.Point
	OldEndPoint buffer.Point
	NewEndPoint buffer.Point

	Text    []byte
	Version uint64
}

func NewDocumentWithBuffer(name string, b buffer.Buffer, mode Mode) (*Document, error) {
	var syntax *Syntax
	if language := GetLanguageByFilename(name); language != nil {
		layers, err := NewSyntaxLayers(b.Bytes(), language.Grammar.Highlight)
		if err != nil {
			return nil, fmt.Errorf("error creating syntax layers: %w", err)
		}
		syntax, err = NewSyntax(language, layers)
		if err != nil {
			return nil, fmt.Errorf("error creating syntax: %w", err)
		}
	}

	d := &Document{
		Buffer:             b,
		Mode:               mode,
		Syntax:             syntax,
		diagnosticVersions: map[ls.DiagnosticType]uint64{},
	}

	d.Autocomplete = NewAutocompleter(d)

	return d, nil
}

func NewDocumentFromName(name string) (*Document, error) {
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

	b, err := buffer.NewFromFile(name, buffer.LineEndingAuto)
	if err != nil {
		return nil, err
	}

	mode := ModeWrite
	if stat.Mode().Perm()&0200 == 0 {
		mode = ModeReadOnly
	}

	return NewDocumentWithBuffer(name, b, mode)
}

type Document struct {
	Buffer       buffer.Buffer
	Name         string
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

func (d *Document) RelativeName(workspace string) string {
	relName, err := filepath.Rel(workspace, d.Name)
	if err != nil {
		return d.Name
	}

	return relName
}

func (d *Document) SetLanguage(name string) error {
	language := GetLanguage(name)
	if language == nil {
		return fmt.Errorf("language with name %q not found", name)
	}

	layers, err := NewSyntaxLayers(d.Buffer.Bytes(), language.Grammar.Highlight)
	if err != nil {
		return fmt.Errorf("error creating syntax layers: %w", err)
	}

	syntax, err := NewSyntax(language, layers)
	if err != nil {
		return fmt.Errorf("error creating syntax: %w", err)
	}

	d.Syntax = syntax
	return nil
}

func (d *Document) Range() buffer.Range {
	return buffer.Range{
		Start: buffer.Point{Row: 0, Col: 0},
		End:   buffer.Point{Row: d.Buffer.LinesLen(), Col: d.Buffer.LineLen(max(d.Buffer.LinesLen()-1, 0))},
	}
}

func (d *Document) recordChange(change Change) tea.Cmd {
	now := time.Now()
	defer func() {
		log.Println("record change time: ", time.Since(now))
	}()

	d.Changes = append(d.Changes, change)

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
	if d.Syntax != nil {
		ctx := context.Background()
		d.Syntax.Parse(ctx, d.Buffer.Version(), d.Buffer.Bytes(), edits)
	}

	var cmds []tea.Cmd
	cmds = append(cmds, tea.Sequence(
		ls.FileChanged(d.Name, d.Buffer.Version(), change.Text),
		ls.GetInlayHint(d.Name, d.Buffer.Version(), d.Range()),
	))

	return tea.Batch(cmds...)
}

func (d *Document) InsertNewLine(p buffer.Point) tea.Cmd {
	startIndex := d.Buffer.ByteIndex(p)
	d.Buffer.Insert(p, []byte{'\n'})

	return d.recordChange(Change{
		StartByte:  startIndex,
		OldEndByte: startIndex + 1,
		NewEndByte: startIndex + 2,
		StartPoint: p,
		OldEndPoint: buffer.Point{
			Row: 0,
			Col: 0,
		},
		NewEndPoint: buffer.Point{
			Row: 0,
			Col: 0,
		},
		Text:    d.Buffer.Bytes(),
		Version: d.Buffer.Version(),
	})
}

func (d *Document) Insert(p buffer.Point, text []byte) tea.Cmd {
	text = xrunes.Sanitize(text)
	if len(text) == 0 {
		return nil
	}

	startIndex := d.Buffer.ByteIndex(p)
	d.Buffer.Insert(p, text)

	return d.recordChange(Change{
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
		Text:    d.Buffer.Bytes(),
		Version: d.Buffer.Version(),
	})
}

func (d *Document) InsertRunes(p buffer.Point, text []rune) tea.Cmd {
	return d.Insert(p, []byte(string(text)))
}

func (d *Document) Replace(r buffer.Range, text []byte) tea.Cmd {
	text = xrunes.Sanitize(text)

	startIndex := d.Buffer.ByteIndex(r.Start)
	endIndex := d.Buffer.ByteIndex(r.End)
	d.Buffer.Replace(r, text)

	return d.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(endIndex),
		NewEndByte: uint32(startIndex + len(text)),
		Text:       d.Buffer.Bytes(),
		Version:    d.Buffer.Version(),
	})
}

func (d *Document) DuplicateLine(row int) tea.Cmd {
	line := d.Buffer.Line(row)
	startIndex := d.Buffer.ByteIndex(buffer.Point{
		Row: row,
		Col: 0,
	})
	d.Buffer.DuplicateLine(row)

	return d.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + 1),
		NewEndByte: uint32(startIndex + line.LenBytes() + 1),
		Text:       d.Buffer.Bytes(),
		Version:    d.Buffer.Version(),
	})
}

func (d *Document) DeleteLine(row int) tea.Cmd {
	line := d.Buffer.Line(row)
	startIndex := d.Buffer.ByteIndex(buffer.Point{
		Row: row,
		Col: 0,
	})

	d.Buffer.DeleteLine(row)

	return d.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + line.LenBytes() + 1),
		NewEndByte: uint32(startIndex + 1),
		Text:       d.Buffer.Bytes(),
		Version:    d.Buffer.Version(),
	})
}

func (d *Document) DeleteBefore(p buffer.Point) tea.Cmd {
	startIndex := d.Buffer.ByteIndex(p)
	d.Buffer.DeleteBefore(p)

	return d.recordChange(Change{
		StartByte:  uint32(startIndex - 1),
		OldEndByte: uint32(startIndex + 1),
		NewEndByte: uint32(startIndex),
		Text:       d.Buffer.Bytes(),
		Version:    d.Buffer.Version(),
	})
}

func (d *Document) DeleteAfter(p buffer.Point) tea.Cmd {
	startIndex := d.Buffer.ByteIndex(p)

	d.Buffer.DeleteAfter(p)

	return d.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + 1),
		NewEndByte: uint32(startIndex + 2),
		Text:       d.Buffer.Bytes(),
		Version:    d.Buffer.Version(),
	})
}

func (d *Document) DeleteRange(r buffer.Range) tea.Cmd {
	startIndex := d.Buffer.ByteIndex(r.Start)
	endIndex := d.Buffer.ByteIndex(r.End)
	d.Buffer.DeleteRange(r)

	return d.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(endIndex),
		NewEndByte: uint32(startIndex),
		Text:       d.Buffer.Bytes(),
		Version:    d.Buffer.Version(),
	})
}

func (d *Document) DeleteWordLeft(p buffer.Point) tea.Cmd {
	startPoint := d.NextWordLeft(p)
	startIndex := d.Buffer.ByteIndex(startPoint)
	endIndex := d.Buffer.ByteIndex(p)
	d.Buffer.DeleteRange(buffer.Range{
		Start: startPoint,
		End:   p,
	})

	return d.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(endIndex),
		NewEndByte: uint32(startIndex),
		Text:       d.Buffer.Bytes(),
		Version:    d.Buffer.Version(),
	})
}

func (d *Document) DeleteWordRight(p buffer.Point) tea.Cmd {
	endPoint := d.NextWordRight(p)
	startIndex := d.Buffer.ByteIndex(p)
	endIndex := d.Buffer.ByteIndex(endPoint)
	d.Buffer.DeleteRange(buffer.Range{
		Start: p,
		End:   endPoint,
	})

	return d.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(endIndex),
		NewEndByte: uint32(startIndex),
		Text:       d.Buffer.Bytes(),
		Version:    d.Buffer.Version(),
	})
}

func (d *Document) AddTab(row int) tea.Cmd {
	line := d.Buffer.Line(row)
	startIndex := d.Buffer.ByteIndex(buffer.Point{
		Row: row,
		Col: 0,
	})
	d.Buffer.AddTab(row)

	return d.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + line.LenBytes() + 1),
		NewEndByte: uint32(startIndex + line.LenBytes() + 2),
		Text:       d.Buffer.Bytes(),
		Version:    d.Buffer.Version(),
	})
}

func (d *Document) RemoveTab(row int) tea.Cmd {
	line := d.Buffer.Line(row)
	startIndex := d.Buffer.ByteIndex(buffer.Point{
		Row: row,
		Col: 0,
	})
	d.Buffer.RemoveTab(row)

	return d.recordChange(Change{
		StartByte:  uint32(startIndex),
		OldEndByte: uint32(startIndex + line.LenBytes() + 1),
		NewEndByte: uint32(startIndex + line.LenBytes() - 1),
		Text:       d.Buffer.Bytes(),
		Version:    d.Buffer.Version(),
	})
}

func (d *Document) NextWordLeft(p buffer.Point) buffer.Point {
	if p.Col == 0 {
		if p.Row == 0 {
			return p
		}
		p.Row--
		p.Col = d.Buffer.LineLen(p.Row - 1)
		return p
	}

	var ready bool
	for {
		if p.Col == 0 || (ready && slices.Contains(wordBreakers, d.Buffer.Line(p.Row).Rune(p.Col))) {
			break
		}
		p.Col--
		if d.Buffer.Line(p.Row).Rune(p.Col) != ' ' {
			ready = true
		}
	}

	return p
}

func (d *Document) NextWordRight(p buffer.Point) buffer.Point {
	if p.Col == d.Buffer.LineLen(p.Row) {
		if p.Row == d.Buffer.LinesLen()-1 {
			return p
		}
		p.Row++
		return p
	}

	var ready bool
	for {
		if p.Col == d.Buffer.LineLen(p.Row) || (ready && slices.Contains(wordBreakers, d.Buffer.Line(p.Row).Rune(p.Col))) {
			break
		}
		p.Col++
		if d.Buffer.Line(p.Row).Rune(p.Col-1) != ' ' {
			ready = true
		}
	}

	return p
}
