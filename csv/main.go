package csv

import (
	"bytes"
	"io"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/nyaosorg/go-windows-mbcs"
)

type Mode struct {
	NonUTF8 bool
	Comma   byte
}

func (m *Mode) toText(s []byte) string {
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

type Cell struct {
	Source string
	text   string
}

type Row struct {
	Cell []Cell
	Term string
}

func ReadLine(br io.ByteReader, mode *Mode) (*Row, error) {
	row := &Row{}
	quoted := false
	source := []byte{}
	for {
		c, err := br.ReadByte()
		if err != nil {
			if len(source) > 0 {
				row.Cell = append(row.Cell, Cell{
					Source: mode.toText(source),
				})
			}
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
					Source: mode.toText(source),
				})
				source = []byte{}
				continue
			case '\n':
				if len(source) > 0 && source[len(source)-1] == '\r' {
					row.Cell = append(row.Cell, Cell{
						Source: mode.toText(source[:len(source)-1]),
					})
					row.Term = "\r\n"
				} else {
					row.Cell = append(row.Cell, Cell{
						Source: mode.toText(source),
					})
					row.Term = "\n"
				}
				return row, nil
			}
		}
		source = append(source, c)
	}
}

func (cell *Cell) Text() string {
	if cell.text != "" {
		return cell.text
	}
	var text strings.Builder

	prevIsQuote := false
	for _, c := range cell.Source {
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
	cell.text = text.String()
	return cell.text
}

func (row *Row) Rebuild(mode *Mode) []byte {
	var buffer bytes.Buffer
	for i, end := 0, len(row.Cell); ; {
		buffer.WriteString(row.Cell[i].Source)
		if i++; i >= end {
			break
		}
		buffer.WriteByte(mode.Comma)
	}
	buffer.WriteString(row.Term)
	if mode.NonUTF8 {
		ansi, err := mbcs.Utf8ToAnsi(buffer.String(), mbcs.ACP)
		if err == nil {
			return ansi
		}
	}
	return buffer.Bytes()
}

func NewCell(text string, mode *Mode) Cell {
	quote := false
	var source bytes.Buffer
	source.WriteByte('"')
	for i, end := 0, len(text); i < end; i++ {
		switch text[i] {
		case '"':
			source.WriteByte('"')
			quote = true
		case '\n', mode.Comma:
			quote = true
		}
		source.WriteByte(text[i])
	}
	if quote {
		source.WriteByte('"')
		return Cell{Source: source.String()}
	} else {
		return Cell{Source: source.String()[1:]}
	}
}

func (row *Row) Insert(i int, text string, mode *Mode) {
	row.Cell = slices.Insert(row.Cell, i, NewCell(text, mode))
}

func (row *Row) Replace(i int, text string, mode *Mode) {
	for i >= len(row.Cell) {
		row.Cell = append(row.Cell, NewCell("", mode))
	}
	row.Cell[i] = NewCell(text, mode)
}

func (row *Row) Delete(i int) {
	row.Cell = slices.Delete(row.Cell, i, i+1)
}
