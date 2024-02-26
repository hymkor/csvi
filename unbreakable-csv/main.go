package csv

import (
	"bufio"
	"bytes"
	"io"
	"slices"
	"strings"
	"unicode/utf8"

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
}

func (m *Mode) HasBom() bool {
	return m.hasBom == triTrue
}

func (m *Mode) decode(s []byte) string {
	if !m.NonUTF8 && utf8.Valid(s) {
		return string(s)
	}
	result, err := mbcs.AnsiToUtf8(s, mbcs.ACP)
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
	source []byte
	text   string
}

func (c Cell) Text() string {
	return c.text
}

type Row struct {
	Cell []Cell
	Term string
}

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
				source: source,
				text:   dequote(mode.decode(source)),
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
					source: source,
					text:   dequote(mode.decode(source)),
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
					source: source,
					text:   dequote(mode.decode(source)),
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
		if s, err := mbcs.Utf8ToAnsi(string(source), mbcs.CP_ACP); err == nil {
			source = s
		}
	}
	return Cell{source: source, text: text}
}

func NewRow(mode *Mode) Row {
	return Row{
		Cell: nil,
		Term: mode.DefaultTerm,
	}
}

func (row *Row) Insert(i int, text string, mode *Mode) {
	row.Cell = slices.Insert(row.Cell, i, newCell(text, mode))
}

func (row *Row) Replace(i int, text string, mode *Mode) {
	for i >= len(row.Cell) {
		row.Cell = append(row.Cell, newCell("", mode))
	}
	row.Cell[i] = newCell(text, mode)
}

func (row *Row) Delete(i int) {
	row.Cell = slices.Delete(row.Cell, i, i+1)
}
