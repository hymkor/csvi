package uncsv

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"github.com/nyaosorg/go-windows-mbcs"
)

const peekSize = 10

type tristate int

const (
	triNotSet tristate = iota
	triFalse
	triTrue
)

type endian int

const (
	octet endian = iota
	utf16le
	utf16be
)

type Mode struct {
	NonUTF8     bool
	Comma       byte
	DefaultTerm string
	hasBom      tristate
	endian      endian
	decoder     *encoding.Decoder
	encoder     *encoding.Encoder
}

func (m *Mode) IsUTF16LE() bool {
	return m.endian == utf16le
}

func (m *Mode) SetUTF16LE() {
	m.endian = utf16le
	m.setEncoding(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM))
	m.NonUTF8 = true
}

func (m *Mode) IsUTF16BE() bool {
	return m.endian == utf16be
}

func (m *Mode) SetUTF16BE() {
	m.endian = utf16be
	m.setEncoding(unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM))
	m.NonUTF8 = true
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

func (m *Mode) setEncoding(e encoding.Encoding) {
	m.decoder = e.NewDecoder()
	m.encoder = e.NewEncoder()
}

func (m *Mode) SetEncoding(name string) error {
	e, err := ianaindex.IANA.Encoding(name)
	if err != nil {
		return err
	}
	if e == nil {
		return fmt.Errorf("%s: not supported in golang.org/x/text/encoding/ianaindex", name)
	}
	m.setEncoding(e)
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

func (c Cell) Source() []byte {
	return c.source
}

func (c Cell) SourceText(m *Mode) string {
	return m.decode(c.source)
}

func (c Cell) Modified() bool {
	return !bytes.Equal(c.source, c.original)
}

func (c Cell) IsQuoted() bool {
	s := c.source
	return (len(s) > 1 && s[0] == 0 && s[1] == '"') ||
		(len(s) > 0 && s[0] == '"')
}

func (c *Cell) Restore(mode *Mode) {
	c.source = c.original
	c.text = dequote(mode.decode(c.original))
}

func (c *Cell) Original() []byte {
	return c.original
}

type Row struct {
	// Cell must have one or more element at least
	Cell []Cell
	// Term is one of "", "\n", and "\r\n"
	Term string
}

func ReadLine(br *bufio.Reader, mode *Mode) (*Row, error) {
	row := &Row{}
	quoted := false
	source := []byte{}
	if mode.hasBom == triNotSet {
		prefix, err := br.Peek(peekSize)
		if err == nil {
			if bytes.HasPrefix(prefix, []byte{0xEF, 0xBB, 0xBF}) {
				// UTF8
				mode.hasBom = triTrue
				br.Discard(3)
			} else if bytes.HasPrefix(prefix, []byte{0xFF, 0xFE}) {
				mode.hasBom = triTrue
				mode.SetUTF16LE()
				br.Discard(2)
			} else if bytes.HasPrefix(prefix, []byte{0xFE, 0xFF}) {
				mode.hasBom = triTrue
				mode.SetUTF16BE()
				br.Discard(2)
			} else {
				mode.hasBom = triFalse
				if mode.endian != utf16le && mode.endian != utf16be {
					if idx := bytes.IndexByte(prefix, 0); idx >= 0 {
						if idx%2 == 1 {
							mode.SetUTF16LE()
						} else {
							mode.SetUTF16BE()
						}
					}
				}
			}
		}
	}
	if mode.endian == octet {
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
	} else {
		for {
			var buf [2]byte
			n, err := io.ReadFull(br, buf[:])
			if err != nil {
				row.Cell = append(row.Cell, Cell{
					source:   source,
					text:     dequote(mode.decode(source)),
					original: source,
				})
				row.Term = ""
				return row, err
			}
			var c rune
			if mode.endian == utf16le {
				c = rune(buf[0]) | (rune(buf[1]) << 8)
			} else {
				c = rune(buf[1]) | (rune(buf[0]) << 8)
			}
			if c == '"' {
				quoted = !quoted
			}
			if !quoted {
				switch c {
				case rune(mode.Comma):
					row.Cell = append(row.Cell, Cell{
						source:   source,
						text:     dequote(mode.decode(source)),
						original: source,
					})
					source = []byte{}
					continue
				case '\n':
					if bytes.HasSuffix(source, []byte{'\r', 0}) ||
						bytes.HasSuffix(source, []byte{0, '\r'}) {
						source = source[:len(source)-2]
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
			source = append(source, buf[:n]...)
		}
	}
}

func ReadAll(r io.Reader, mode *Mode) ([]Row, error) {
	reader, ok := r.(*bufio.Reader)
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

func writeEndian(w io.Writer, c byte, e endian) {
	switch e {
	case utf16le:
		w.Write([]byte{c, 0})
	case utf16be:
		w.Write([]byte{0, c})
	default:
		w.Write([]byte{c})
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
			writeEndian(&buffer, mode.Comma, mode.endian)
		}
	}
	for i := 0; i < len(row.Term); i++ {
		writeEndian(&buffer, row.Term[i], mode.endian)
	}
	return buffer.Bytes()
}

func (mode *Mode) Dump(rows []Row, w io.Writer) {
	mode.DumpBy(func() *Row {
		if len(rows) <= 0 {
			return nil
		}
		r := &rows[0]
		rows = rows[1:]
		return r
	}, w)
}

func (mode *Mode) DumpBy(fetch func() *Row, w io.Writer) {
	bw := bufio.NewWriter(w)
	if mode.hasBom == triTrue {
		switch mode.endian {
		case utf16le:
			bw.Write([]byte{0xFF, 0xFE})
		case utf16be:
			bw.Write([]byte{0xFE, 0xFF})
		default:
			bw.WriteString("\uFEFF")
		}
	}
	for {
		row := fetch()
		if row == nil {
			break
		}
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
		source = slicesInsert(source, 0, '"')
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

func NewRowFromStringSlice(mode *Mode, texts []string) Row {
	cells := []Cell{}
	for _, t := range texts {
		cells = append(cells, newCell(t, mode))
	}
	return Row{
		Cell: cells,
		Term: mode.DefaultTerm,
	}
}

func (row *Row) Insert(i int, text string, mode *Mode) {
	row.Cell = slicesInsert(row.Cell, i, newCell(text, mode))
}

func (row *Row) Replace(i int, text string, mode *Mode) {
	original := row.Cell[i].original
	row.Cell[i] = newCell(text, mode)
	row.Cell[i].original = original
}

func (row *Row) Delete(i int) {
	row.Cell = slicesDelete(row.Cell, i, i+1)
}

func (row *Row) Restore(mode *Mode) {
	for i, c := range row.Cell {
		c.Restore(mode)
		row.Cell[i] = c
	}
}

func slicesInsert[T any](array []T, at int, val T) []T {
	array = append(array, val)
	copy(array[at+1:], array[at:])
	array[at] = val
	return array
}

func slicesDelete[T any](array []T, from int, to int) []T {
	copy(array[from:], array[to:])
	return array[:len(array)-(to-from)]
}
