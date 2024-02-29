package csv

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"

	"github.com/nyaosorg/go-windows-mbcs"
)

type tristate int

const (
	triNotSet tristate = iota
	triFalse
	triTrue
)

type Mode struct {
	NonUTF8     bool
	Comma       byte
	DefaultTerm string
	hasBom      tristate
	decoder     *encoding.Decoder
	encoder     *encoding.Encoder
}

func (m *Mode) _decode(s []byte) (string, error) {
	if m.decoder != nil {
		result, _, err := transform.Bytes(m.decoder, s)
		if err != nil {
			return "", err
		}
		return string(result), nil
	}
	return mbcs.AnsiToUtf8(s, mbcs.ACP)
}

func (m *Mode) _encode(s string) ([]byte, error) {
	if m.encoder != nil {
		result, _, err := transform.Bytes(m.encoder, []byte(s))
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return mbcs.Utf8ToAnsi(s, mbcs.ACP)
}

func (m *Mode) SetEncoding(name string) error {
	e, err := ianaindex.IANA.Encoding(name)
	if err != nil {
		return err
	}
	if e == nil {
		return fmt.Errorf("%s: not supported in golang.org/x/text/encoding/ianaindex", name)
	}
	m.decoder = e.NewDecoder()
	m.encoder = e.NewEncoder()
	return nil
}

func (m *Mode) HasBom() bool {
	return m.hasBom == triTrue
}

func (m *Mode) decode(s []byte) string {
	if !m.NonUTF8 && utf8.Valid(s) {
		return string(s)
	}
	result, err := m._decode(s)
	if err != nil {
		return string(s)
	}
	m.NonUTF8 = true
	return result
}

func dequote(raw string) string {
	var text strings.Builder

	prevIsQuote := false
	for _, c := range raw {
		if c == '"' {
			if prevIsQuote {
				text.WriteByte('"')
				prevIsQuote = false
			} else {
				prevIsQuote = true
			}
		} else {
			text.WriteRune(c)
			prevIsQuote = false
		}
	}
	return text.String()
}

type Cell struct {
	source   []byte
	text     string
	original []byte
}

func (c Cell) Text() string {
	return c.text
}

func (c Cell) ReadableSource(m *Mode) string {
	return m.decode(c.source)
}

func (c Cell) Modified() bool {
	return !bytes.Equal(c.source, c.original)
}

func (c Cell) IsQuoted() bool {
	return len(c.source) > 0 && c.source[0] == '"'
}

func (c *Cell) Restore(mode *Mode) {
	c.source = c.original
	c.text = dequote(mode.decode(c.original))
}

type Row struct {
	// Cell must have one or more element at least
	Cell []Cell
	// Term is one of "", "\n", and "\r\n"
	Term string
}

// Reader is assumed to be "bufio".Reader or "strings".Reader
type Reader interface {
	io.RuneScanner
	io.ByteReader
}

func ReadLine(br Reader, mode *Mode) (*Row, error) {
	row := &Row{}
	quoted := false
	source := []byte{}
	if mode.hasBom == triNotSet {
		if ch, n, err := br.ReadRune(); err == nil && n == 3 && ch == '\uFEFF' {
			mode.hasBom = triTrue
		} else {
			mode.hasBom = triFalse
			br.UnreadRune()
		}
	}
	for {
		c, err := br.ReadByte()
		if err != nil {
			row.Cell = append(row.Cell, Cell{
				source:   source,
				text:     dequote(mode.decode(source)),
				original: source,
			})
			row.Term = ""
			return row, err
		}
		if c == '"' {
			quoted = !quoted
		}
		if !quoted {
			switch c {
			case mode.Comma:
				row.Cell = append(row.Cell, Cell{
					source:   source,
					text:     dequote(mode.decode(source)),
					original: source,
				})
				source = []byte{}
				continue
			case '\n':
				if len(source) > 0 && source[len(source)-1] == '\r' {
					source = source[:len(source)-1]
					row.Term = "\r\n"
				} else {
					row.Term = "\n"
				}
				row.Cell = append(row.Cell, Cell{
					source:   source,
					text:     dequote(mode.decode(source)),
					original: source,
				})
				if mode.DefaultTerm == "" {
					mode.DefaultTerm = row.Term
				}
				return row, nil
			}
		}
		source = append(source, c)
	}
}

func ReadAll(r io.Reader, mode *Mode) ([]Row, error) {
	reader, ok := r.(Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}
	rows := []Row{}
	for {
		row, err := ReadLine(reader, mode)
		if err != nil && err != io.EOF {
			return rows, err
		}
		rows = append(rows, *row)
		if err == io.EOF {
			return rows, nil
		}
	}
}

func (row *Row) Rebuild(mode *Mode) []byte {
	var buffer bytes.Buffer
	if len(row.Cell) > 0 {
		for i, end := 0, len(row.Cell); ; {
			buffer.Write(row.Cell[i].source)
			if i++; i >= end {
				break
			}
			buffer.WriteByte(mode.Comma)
		}
	}
	buffer.WriteString(row.Term)
	return buffer.Bytes()
}

func (mode *Mode) Dump(rows []Row, w io.Writer) {
	bw := bufio.NewWriter(w)
	if mode.hasBom == triTrue {
		bw.WriteString("\uFEFF")
	}
	for _, row := range rows {
		bw.Write(row.Rebuild(mode))
	}
	bw.Flush()
}

func newCell(text string, mode *Mode) Cell {
	quote := false
	source := make([]byte, 0, len(text)+4)
	for i, end := 0, len(text); i < end; i++ {
		switch text[i] {
		case '"':
			source = append(source, '"')
			quote = true
		case '\n', mode.Comma:
			quote = true
		}
		source = append(source, text[i])
	}
	if quote {
		source = append(source, '"')
		source = slices.Insert(source, 0, '"')
	}
	if mode.NonUTF8 {
		if s, err := mode._encode(string(source)); err == nil {
			source = s
		}
	}
	return Cell{source: source, text: text, original: nil}
}

func (c Cell) Quote(mode *Mode) Cell {
	text := c.text
	source := make([]byte, 0, len(text))
	source = append(source, '"')
	for i, end := 0, len(text); i < end; i++ {
		if text[i] == '"' {
			source = append(source, '"')
		}
		source = append(source, text[i])
	}
	source = append(source, '"')
	if mode.NonUTF8 {
		if s, err := mode._encode(string(source)); err == nil {
			source = s
		}
	}
	return Cell{source: source, text: text, original: c.original}
}

func NewRow(mode *Mode) Row {
	return Row{
		Cell: []Cell{newCell("", mode)},
		Term: mode.DefaultTerm,
	}
}

func (row *Row) Insert(i int, text string, mode *Mode) {
	row.Cell = slices.Insert(row.Cell, i, newCell(text, mode))
}

func (row *Row) Replace(i int, text string, mode *Mode) {
	original := row.Cell[i].original
	row.Cell[i] = newCell(text, mode)
	row.Cell[i].original = original
}

func (row *Row) Delete(i int) {
	row.Cell = slices.Delete(row.Cell, i, i+1)
}
