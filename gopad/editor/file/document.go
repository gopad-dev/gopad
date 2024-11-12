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
	"github.com/charmbracelet/lipgloss"

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
		Name:               name,
		Mode:               mode,
		Syntax:             syntax,
		oldState:           nil,
		changes:            NewChangeSetFromBuf(b),
		History:            NewHistory(),
		Autocomplete:       nil,
		version:            0,
		diagnosticVersions: map[ls.DiagnosticType]uint64{},
		Diagnostics:        nil,
		inlayHintsVersion:  0,
		InlayHints:         nil,
		Declarations:       nil,
		Definitions:        nil,
		TypeDefinitions:    nil,
		Implementations:    nil,
		References:         nil,
		Positions:          nil,
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
	Buffer buffer.Buffer
	Name   string
	Mode   Mode
	Syntax *Syntax

	oldState     buffer.Buffer
	changes      ChangeSet
	History      *History
	Autocomplete *Autocompleter
	version      uint64

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
}

func (d *Document) Version() uint64 {
	return d.version
}

func (d *Document) FileName() string {
	return filepath.Base(d.Name)
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

func (d *Document) Apply(t Transaction) tea.Cmd {
	cmd, success := d.applyInner(t)
	if success {
		d.appendChangesToHistory()
	}
	return cmd
}

func (d *Document) applyInner(t Transaction) (tea.Cmd, bool) {
	now := time.Now()
	defer func() {
		log.Println("record change time: ", time.Since(now))
	}()

	if d.changes.IsEmpty() && !t.Changes.IsEmpty() {
		d.oldState = d.Buffer.Clone()
	}

	cmd, success := d.apply(t)
	if !t.Changes.IsEmpty() {
		d.changes = d.changes.Merge(t.Changes)
	}

	return cmd, success
}

func (d *Document) apply(t Transaction) (tea.Cmd, bool) {
	oldBuf := d.Buffer.Clone()
	changes := t.Changes

	success := changes.Apply(d.Buffer)
	if !success {
		log.Printf("error applying changes: %v", changes)
		return nil, false
	}

	if changes.IsEmpty() {
		return nil, true
	}

	d.version++

	cmds := []tea.Cmd{
		ls.FileChanged(d.Name, d.Version(), d.Buffer.Bytes()),
		ls.GetInlayHint(d.Name, d.Version(), d.Range()),
	}

	if d.Syntax != nil {
		ctx := context.Background()
		if err := d.Syntax.Update(ctx, d.version, d.Buffer, oldBuf, t.Changes); err != nil {
			log.Printf("error updating syntax: %v", err)
			d.Syntax = nil
		}
	}

	return tea.Batch(cmds...), true
}

func (d *Document) appendChangesToHistory() {
	if d.changes.IsEmpty() {
		return
	}

	changes := d.changes
	d.changes = NewChangeSetFromBuf(d.Buffer)

	transaction := NewTransactionFrom(changes)
	oldBuf := d.oldState

	d.History.CommitRevision(transaction, oldBuf)
}

func (d *Document) InsertNewLine(p buffer.Point) tea.Cmd {
	startIndex := d.Buffer.ByteIndexByPoint(p)
	transaction := NewTransactionFromInsert(d.Buffer, startIndex, []byte{'\n'})

	return d.Apply(transaction)
}

func (d *Document) Insert(p buffer.Point, text []byte) tea.Cmd {
	text = xrunes.Sanitize(text)
	if len(text) == 0 {
		return nil
	}

	startIndex := d.Buffer.ByteIndexByPoint(p)
	transaction := NewTransactionFromInsert(d.Buffer, startIndex, text)

	log.Printf("inserting text at index %d: %#v", startIndex, transaction)
	return d.Apply(transaction)
}

func (d *Document) Replace(r buffer.Range, text []byte) tea.Cmd {
	text = xrunes.Sanitize(text)

	startIndex := d.Buffer.ByteIndexByPoint(r.Start)
	endIndex := d.Buffer.ByteIndexByPoint(r.End)

	transaction := NewTransactionFromChange(d.Buffer, []Change{{
		From: startIndex,
		To:   endIndex,
		Text: text,
	}})

	return d.Apply(transaction)
}

func (d *Document) DeleteRange(r buffer.Range) tea.Cmd {
	startIndex := d.Buffer.ByteIndexByPoint(r.Start)
	endIndex := d.Buffer.ByteIndexByPoint(r.End)

	transaction := NewTransactionFromDelete(d.Buffer, []Deletion{{
		From: startIndex,
		To:   endIndex,
	}})

	return d.Apply(transaction)
}

func (d *Document) DeleteBefore(p buffer.Point) tea.Cmd {
	startIndex := d.Buffer.ByteIndexByPoint(p)
	if startIndex == 0 {
		return nil
	}

	transaction := NewTransactionFromDelete(d.Buffer, []Deletion{{
		From: startIndex - 1,
		To:   startIndex,
	}})

	return d.Apply(transaction)
}

func (d *Document) DeleteAfter(p buffer.Point) tea.Cmd {
	startIndex := d.Buffer.ByteIndexByPoint(p)
	if startIndex == d.Buffer.Len() {
		return nil
	}

	transaction := NewTransactionFromDelete(d.Buffer, []Deletion{{
		From: startIndex,
		To:   startIndex + 1,
	}})

	return d.Apply(transaction)
}

//func (d *Document) DuplicateLine(row int) tea.Cmd {
//	line := d.Buffer.Line(row)
//	startIndex := d.Buffer.ByteIndex(buffer.Point{
//		Row: row,
//		Col: 0,
//	})
//	d.Buffer.DuplicateLine(row)
//
//	return d.recordChange(Change{
//		StartByte:  uint32(startIndex),
//		OldEndByte: uint32(startIndex + 1),
//		NewEndByte: uint32(startIndex + line.LenBytes() + 1),
//		Text:       d.Buffer.Bytes(),
//		Version:    d.Buffer.Version(),
//	})
//}
//
//func (d *Document) DeleteLine(row int) tea.Cmd {
//	line := d.Buffer.Line(row)
//	startIndex := d.Buffer.ByteIndex(buffer.Point{
//		Row: row,
//		Col: 0,
//	})
//
//	d.Buffer.DeleteLine(row)
//
//	return d.recordChange(Change{
//		StartByte:  uint32(startIndex),
//		OldEndByte: uint32(startIndex + line.LenBytes() + 1),
//		NewEndByte: uint32(startIndex + 1),
//		Text:       d.Buffer.Bytes(),
//		Version:    d.Buffer.Version(),
//	})
//}
//

//

//

//
//func (d *Document) DeleteWordLeft(p buffer.Point) tea.Cmd {
//	startPoint := d.NextWordLeft(p)
//	startIndex := d.Buffer.ByteIndex(startPoint)
//	endIndex := d.Buffer.ByteIndex(p)
//	d.Buffer.DeleteRange(buffer.Range{
//		Start: startPoint,
//		End:   p,
//	})
//
//	return d.recordChange(Change{
//		StartByte:  uint32(startIndex),
//		OldEndByte: uint32(endIndex),
//		NewEndByte: uint32(startIndex),
//		Text:       d.Buffer.Bytes(),
//		Version:    d.Buffer.Version(),
//	})
//}
//
//func (d *Document) DeleteWordRight(p buffer.Point) tea.Cmd {
//	endPoint := d.NextWordRight(p)
//	startIndex := d.Buffer.ByteIndex(p)
//	endIndex := d.Buffer.ByteIndex(endPoint)
//	d.Buffer.DeleteRange(buffer.Range{
//		Start: p,
//		End:   endPoint,
//	})
//
//	return d.recordChange(Change{
//		StartByte:  uint32(startIndex),
//		OldEndByte: uint32(endIndex),
//		NewEndByte: uint32(startIndex),
//		Text:       d.Buffer.Bytes(),
//		Version:    d.Buffer.Version(),
//	})
//}
//
//func (d *Document) AddTab(row int) tea.Cmd {
//	line := d.Buffer.Line(row)
//	startIndex := d.Buffer.ByteIndex(buffer.Point{
//		Row: row,
//		Col: 0,
//	})
//	d.Buffer.AddTab(row)
//
//	return d.recordChange(Change{
//		StartByte:  uint32(startIndex),
//		OldEndByte: uint32(startIndex + line.LenBytes() + 1),
//		NewEndByte: uint32(startIndex + line.LenBytes() + 2),
//		Text:       d.Buffer.Bytes(),
//		Version:    d.Buffer.Version(),
//	})
//}
//
//func (d *Document) RemoveTab(row int) tea.Cmd {
//	line := d.Buffer.Line(row)
//	startIndex := d.Buffer.ByteIndex(buffer.Point{
//		Row: row,
//		Col: 0,
//	})
//	d.Buffer.RemoveTab(row)
//
//	return d.recordChange(Change{
//		StartByte:  uint32(startIndex),
//		OldEndByte: uint32(startIndex + line.LenBytes() + 1),
//		NewEndByte: uint32(startIndex + line.LenBytes() - 1),
//		Text:       d.Buffer.Bytes(),
//		Version:    d.Buffer.Version(),
//	})
//}

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

// TODO: implement
func (d *Document) Save() error {
	return nil
}

// TODO: implement
func (d *Document) Rename(name string) error {
	return nil
}

// TODO: implement
func (d *Document) Delete() error {
	return nil
}

type CharStyle struct {
	Style lipgloss.Style
	End   int
}
