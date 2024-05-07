package csvi

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"

	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/keys"

	"github.com/hymkor/csvi/internal/nonblock"
	"github.com/hymkor/csvi/uncsv"
)

const (
	_ANSI_CURSOR_OFF = "\x1B[?25l"
	_ANSI_CURSOR_ON  = "\x1B[?25h"
	_ANSI_YELLOW     = "\x1B[0;33;1m"
	_ANSI_RESET      = "\x1B[0m"

	_ANSI_ERASE_LINE       = "\x1B[0m\x1B[0K"
	_ANSI_ERASE_SCRN_AFTER = "\x1B[0m\x1B[0J"

	_ANSI_UNDERLINE_ON  = "\x1B[4m"
	_ANSI_UNDERLINE_OFF = "\x1B[24m"
)

type _ColorStyle struct {
	Cursor [2]string
	Even   [2]string
	Odd    [2]string
}

var bodyColorStyle = _ColorStyle{
	Cursor: [...]string{"\x1B[107;30;22m", "\x1B[40;37m"},
	Even:   [...]string{"\x1B[48;5;235;37;1m", "\x1B[22;40m"},
	Odd:    [...]string{"\x1B[40;37;1m", "\x1B[22m"},
}

var headColorStyle = _ColorStyle{
	Cursor: [...]string{"\x1B[107;30;22m", "\x1B[40;36m"},
	Even:   [...]string{"\x1B[48;5;235;36;1m", "\x1B[22;40m"},
	Odd:    [...]string{"\x1B[40;36;1m", "\x1B[22m"},
}

var replaceTable = strings.NewReplacer(
	"\r", "\u240D",
	"\x1B", "\u241B",
	"\n", "\u240A",
	"\t", "\u2409")

// See. en.wikipedia.org/wiki/Unicode_control_characters#Control_pictures

func drawLine(
	csvs []uncsv.Cell,
	cellWidth int,
	screenWidth int,
	cursorPos int,
	reverse bool,
	style *_ColorStyle,
	out io.Writer) {

	if len(csvs) <= 0 && cursorPos >= 0 {
		io.WriteString(out, style.Cursor[0])
		io.WriteString(out, "\x1B[K")
		io.WriteString(out, style.Cursor[1])
		return
	}
	i := 0

	if reverse {
		io.WriteString(out, style.Odd[0])
		defer io.WriteString(out, style.Odd[1])
	} else {
		io.WriteString(out, style.Even[0])
		defer io.WriteString(out, style.Even[1])
	}
	io.WriteString(out, "\x1B[K")

	for len(csvs) > 0 {
		cursor := csvs[0]
		text := cursor.Text()
		csvs = csvs[1:]
		nextI := i + 1

		cw := cellWidth
		for len(csvs) > 0 && csvs[0].Text() == "" && nextI != cursorPos {
			cw += cellWidth
			csvs = csvs[1:]
			nextI++
		}
		if cw > screenWidth || len(csvs) <= 0 {
			cw = screenWidth
		}
		text = replaceTable.Replace(text)
		ss, _ := cutStrInWidth(text, cw)
		if i == cursorPos {
			io.WriteString(out, style.Cursor[0])
		}
		if cursor.Modified() {
			io.WriteString(out, _ANSI_UNDERLINE_ON)
		}
		io.WriteString(out, ss)
		if cursor.Modified() {
			io.WriteString(out, _ANSI_UNDERLINE_OFF)
		}
		if i == cursorPos {
			io.WriteString(out, "\x1B[K")
			if reverse {
				io.WriteString(out, style.Odd[0])
			} else {
				io.WriteString(out, style.Even[0])
			}
		}
		screenWidth -= cw
		if screenWidth <= 0 {
			break
		}
		fmt.Fprintf(out, "\x1B[%dG", nextI*cellWidth+1)
		if i == cursorPos {
			io.WriteString(out, "\x1B[K")
		}
		i = nextI
	}
}

func up(n int, out io.Writer) {
	if n == 0 {
		out.Write([]byte{'\r'})
	} else if n == 1 {
		io.WriteString(out, "\r\x1B[A")
	} else {
		fmt.Fprintf(out, "\r\x1B[%dA", n)
	}
}

func drawPage(page func(func([]uncsv.Cell) bool), cellWidth, csrpos, csrlin, w, h int, style *_ColorStyle, cache map[int]string, out io.Writer) int {
	reverse := false
	count := 0
	lfCount := 0
	page(func(record []uncsv.Cell) bool {
		if count >= h {
			return false
		}
		if count > 0 {
			lfCount++
			io.WriteString(out, "\r\n") // "\r" is for Linux and go-tty
		}
		cursorPos := -1
		if count == csrlin {
			cursorPos = csrpos
		}
		var buffer strings.Builder
		drawLine(record, cellWidth, w, cursorPos, reverse, style, &buffer)
		line := buffer.String()
		if f := cache[count]; f != line {
			io.WriteString(out, line)
			cache[count] = line
		}
		reverse = !reverse
		count++
		return true
	})
	io.WriteString(out, "\r\n") // \r is for Linux & go-tty
	lfCount++
	return lfCount
}

func cellsAfter(cells []uncsv.Cell, n int) []uncsv.Cell {
	if n <= len(cells) {
		return cells[n:]
	} else {
		return []uncsv.Cell{}
	}
}

type _View struct {
	headCache map[int]string
	bodyCache map[int]string
}

func newView() *_View {
	return &_View{
		headCache: map[int]string{},
		bodyCache: map[int]string{},
	}
}

func (v *_View) clearCache() {
	clear(v.headCache)
	clear(v.bodyCache)
}

func (v *_View) Draw(header, startRow, cursorRow *RowPtr, cellWidth, headerLines, startCol, cursorCol, screenHeight, screenWidth int, out io.Writer) int {
	// print header
	lfCount := 0
	if h := headerLines; h > 0 {
		enum := func(callback func([]uncsv.Cell) bool) {
			for i := 0; i < h && header != nil; i++ {
				if !callback(cellsAfter(header.Cell, startCol)) {
					return
				}
				header = header.Next()
			}
		}
		lfCount = drawPage(enum, cellWidth, cursorCol-startCol, cursorRow.lnum, screenWidth-1, h, &headColorStyle, v.headCache, out)
	}
	if startRow.lnum < headerLines {
		for i := 0; i < headerLines && startRow != nil; i++ {
			startRow = startRow.Next()
		}
	}
	if startRow == nil {
		return lfCount
	}
	p := startRow.Clone()
	// print body
	enum := func(callback func([]uncsv.Cell) bool) {
		for p != nil {
			if !callback(cellsAfter(p.Cell, startCol)) {
				return
			}
			p = p.Next()
		}
	}
	style := &bodyColorStyle
	if headerLines%2 == 1 {
		style = &_ColorStyle{
			Cursor: bodyColorStyle.Cursor,
			Even:   bodyColorStyle.Odd,
			Odd:    bodyColorStyle.Even,
		}
	}
	return lfCount + drawPage(enum, cellWidth, cursorCol-startCol, cursorRow.lnum-startRow.lnum, screenWidth-1, screenHeight-1, style, v.bodyCache, out)
}

func yesNo(pilot Pilot, out io.Writer, message string) bool {
	fmt.Fprintf(out, "%s\r%s%s", _ANSI_YELLOW, message, _ANSI_ERASE_LINE)
	io.WriteString(out, _ANSI_CURSOR_ON)
	ch, err := pilot.GetKey()
	io.WriteString(out, _ANSI_CURSOR_OFF)
	return err == nil && ch == "y"
}

func first[T any](value T, _ error) T {
	return value
}

func printStatusLine(out io.Writer, mode *uncsv.Mode, cursorRow *RowPtr, cursorCol int, screenWidth int) {
	n := 0
	if mode.Comma == '\t' {
		n += first(io.WriteString(out, "[TSV]"))
	} else if mode.Comma == ',' {
		n += first(io.WriteString(out, "[CSV]"))
	}
	switch cursorRow.Term {
	case "\r\n":
		n += first(io.WriteString(out, "[CRLF]"))
	case "\n":
		n += first(io.WriteString(out, "[LF]"))
	case "":
		n += first(io.WriteString(out, "[EOF]"))
	}
	if mode.HasBom() {
		n += first(io.WriteString(out, "[BOM]"))
	}
	if mode.NonUTF8 {
		if mode.IsUTF16LE() {
			n += first(io.WriteString(out, "[16LE]"))
		} else if mode.IsUTF16BE() {
			n += first(io.WriteString(out, "[16BE]"))
		} else {
			n += first(io.WriteString(out, "[ANSI]"))
		}
	}
	if 0 <= cursorCol && cursorCol < len(cursorRow.Cell) {
		n += first(fmt.Fprintf(out, "(%d,%d/%d): ",
			cursorCol+1,
			cursorRow.lnum+1,
			cursorRow.list.Len()))
		var buffer strings.Builder
		buffer.WriteString(cursorRow.Cell[cursorCol].SourceText(mode))
		if cursorCol < len(cursorRow.Cell)-1 {
			buffer.WriteByte(mode.Comma)
		} else if term := cursorRow.Term; term != "" {
			buffer.WriteString(term)
		} else { // EOF
			buffer.WriteString("\u2592")
		}
		io.WriteString(out, runewidth.Truncate(replaceTable.Replace(buffer.String()), screenWidth-n, "..."))
	}
}

type Pilot interface {
	Size() (int, int, error)
	Calibrate() error
	GetKey() (string, error)
	ReadLine(io.Writer, string, string, Candidate) (string, error)
	GetFilename(io.Writer, string, string) (string, error)
	Close() error
}

type Config struct {
	CellWidth   int
	HeaderLines int
	Pilot       Pilot
}

func (cfg Config) Main(mode *uncsv.Mode, reader *bufio.Reader, out io.Writer) (*RowPtr, error) {
	pilot := cfg.Pilot
	if pilot == nil {
		var err error
		pilot, err = newManualCtl()
		if err != nil {
			return nil, err
		}
		pilot.Close()
	}
	if _, ok := out.(*os.File); ok {
		if err := pilot.Calibrate(); err != nil {
			return nil, err
		}
	}
	csvlines := list.New()
	if reader != nil {
		for i := 0; i < 100; i++ {
			row, err := uncsv.ReadLine(reader, mode)
			if err != nil && err != io.EOF {
				return nil, err
			}
			csvlines.PushBack(row)
			if err == io.EOF {
				reader = nil
				break
			}
		}
	} else {
		newRow := uncsv.NewRow(mode)
		csvlines.PushBack(&newRow)
	}
	cursorCol := 0
	cursorRow := frontPtr(csvlines)
	startRow := frontPtr(csvlines)
	startCol := 0

	lastSearch := searchForward
	lastSearchRev := searchBackward
	lastWord := ""
	var lastWidth, lastHeight int

	keyWorker := nonblock.New(func() (string, error) {
		return pilot.GetKey()
	})
	defer keyWorker.Close()

	view := newView()

	message := ""
	var killbuffer string
	for {
		screenWidth, screenHeight, err := pilot.Size()
		if err != nil {
			return nil, err
		}
		screenHeight -= cfg.HeaderLines
		if lastWidth != screenWidth || lastHeight != screenHeight {
			view.clearCache()
			lastWidth = screenWidth
			lastHeight = screenHeight
			io.WriteString(out, _ANSI_CURSOR_OFF)
		}
		cols := (screenWidth - 1) / cfg.CellWidth

		lfCount := view.Draw(frontPtr(csvlines), startRow, cursorRow, cfg.CellWidth, cfg.HeaderLines, startCol, cursorCol, screenHeight, screenWidth, out)
		repaint := func() {
			up(lfCount, out)
			lfCount = view.Draw(frontPtr(csvlines), startRow, cursorRow, cfg.CellWidth, cfg.HeaderLines, startCol, cursorCol, screenHeight, screenWidth, out)
		}

		io.WriteString(out, _ANSI_YELLOW)
		if message != "" {
			io.WriteString(out, runewidth.Truncate(message, screenWidth-1, ""))
		} else if 0 <= cursorRow.lnum && cursorRow.lnum < csvlines.Len() {
			printStatusLine(out, mode, cursorRow, cursorCol, screenWidth)
		}
		io.WriteString(out, _ANSI_RESET)
		io.WriteString(out, _ANSI_ERASE_SCRN_AFTER)

		const interval = 4
		displayUpdateTime := time.Now().Add(time.Second / interval)

		ch, err := keyWorker.GetOr(func() bool {
			if reader == nil {
				return false
			}
			row, err := uncsv.ReadLine(reader, mode)
			if err != nil {
				reader = nil
				if err != io.EOF {
					return false
				}
			}
			csvlines.PushBack(row)
			if message == "" && (err == io.EOF || time.Now().After(displayUpdateTime)) {
				io.WriteString(out, "\r"+_ANSI_YELLOW)
				printStatusLine(out, mode, cursorRow, cursorCol, screenWidth)
				io.WriteString(out, _ANSI_RESET)
				io.WriteString(out, _ANSI_ERASE_SCRN_AFTER)
				displayUpdateTime = time.Now().Add(time.Second / interval)
			}
			return err != io.EOF
		})
		if err != nil {
			return nil, err
		}
		message = ""

		switch ch {
		case keys.CtrlL:
			view.clearCache()
		case "q", keys.Escape:
			if yesNo(pilot, out, "Quit Sure ? [y/n]") {
				io.WriteString(out, "\n")
				return frontPtr(csvlines), nil
			}
		case "j", keys.Down, keys.CtrlN, keys.Enter:
			if next := cursorRow.Next(); next != nil {
				cursorRow = next
			}
		case "k", keys.Up, keys.CtrlP:
			if prev := cursorRow.Prev(); prev != nil {
				cursorRow = prev
			}
		case "h", keys.Left, keys.CtrlB, keys.ShiftTab:
			if cursorCol > 0 {
				cursorCol--
			}
		case "l", keys.Right, keys.CtrlF, keys.CtrlI:
			cursorCol++
		case "0", "^", keys.CtrlA:
			cursorCol = 0
		case "$", keys.CtrlE:
			cursorCol = len(cursorRow.Cell) - 1
		case "<":
			cursorRow = frontPtr(csvlines)
			startRow = frontPtr(csvlines)
			cursorCol = 0
			startCol = 0
		case ">", "G":
			cursorRow = backPtr(csvlines)
		case "n":
			if lastWord == "" {
				break
			}
			r, c := lastSearch(cursorRow, cursorCol, lastWord)
			if r == nil {
				message = fmt.Sprintf("%s: not found", lastWord)
				break
			}
			cursorRow = r
			cursorCol = c
		case "N":
			if lastWord == "" {
				break
			}
			r, c := lastSearchRev(cursorRow, cursorCol, lastWord)
			if r == nil {
				message = fmt.Sprintf("%s: not found", lastWord)
				break
			}
			cursorRow = r
			cursorCol = c
		case "/", "?":
			var err error
			view.clearCache()
			lastWord, err = pilot.ReadLine(out, ch, "", nil)
			if err != nil {
				if err != readline.CtrlC {
					message = err.Error()
				}
				break
			}
			if ch == "/" {
				lastSearch = searchForward
				lastSearchRev = searchBackward
			} else {
				lastSearch = searchBackward
				lastSearchRev = searchForward
			}
			r, c := lastSearch(cursorRow, cursorCol, lastWord)
			if r == nil {
				message = fmt.Sprintf("%s: not found", lastWord)
				break
			}
			cursorRow = r
			cursorCol = c
		case "o":
			newRow := uncsv.NewRow(mode)
			newRow.Term = cursorRow.Term
			if cursorRow.Term == "" {
				cursorRow.Term = mode.DefaultTerm
			}
			cursorRow = cursorRow.InsertAfter(&newRow)
			repaint()
			view.clearCache()
			text, _ := pilot.ReadLine(out, "new line>", "", makeCandidate(cursorRow.lnum-1, cursorCol, cursorRow))
			cursorRow.Replace(0, text, mode)

		case "O":
			startPrevP := startRow.Prev()
			newRow := uncsv.NewRow(mode)
			cursorRow = cursorRow.InsertBefore(&newRow)
			if startPrevP != nil {
				startRow = startPrevP.Next()
			} else {
				startRow = frontPtr(csvlines)
			}
			repaint()
			view.clearCache()
			text, _ := pilot.ReadLine(out, "new line>", "", makeCandidate(cursorRow.lnum-1, cursorCol, cursorRow))
			cursorRow.Replace(0, text, mode)
		case "D":
			if csvlines.Len() <= 1 {
				break
			}
			startPrevP := startRow.Prev()
			prevP := cursorRow.Prev()
			removedRow := cursorRow.Remove()
			if prevP == nil {
				cursorRow = frontPtr(csvlines)
			} else if next := prevP.Next(); next != nil {
				cursorRow = next
			} else {
				cursorRow = prevP
				cursorRow.Term = removedRow.Term
			}
			if startPrevP == nil {
				startRow = frontPtr(csvlines)
			} else {
				startRow = startPrevP.Next()
			}
		case "i":
			view.clearCache()
			text, err := pilot.ReadLine(out, "insert cell>", "", makeCandidate(cursorRow.lnum, cursorCol, cursorRow))
			if err != nil {
				break
			}
			if cells := cursorRow.Cell; len(cells) == 1 && cells[0].Text() == "" {
				cursorRow.Replace(cursorCol, text, mode)
			} else {
				cursorRow.Insert(cursorCol, text, mode)
				cursorCol++
			}
		case "a":
			if cells := cursorRow.Cell; len(cells) == 1 && cells[0].Text() == "" {
				// current column is the last one and it is empty
				view.clearCache()
				text, err := pilot.ReadLine(out, "append cell>", "", makeCandidate(cursorRow.lnum, cursorCol+1, cursorRow))
				if err != nil {
					break
				}
				cursorRow.Replace(cursorCol, text, mode)
			} else {
				cursorCol++
				cursorRow.Insert(cursorCol, "", mode)
				repaint()
				view.clearCache()
				text, err := pilot.ReadLine(out, "append cell>", "", makeCandidate(cursorRow.lnum+1, cursorCol+1, cursorRow))
				if err != nil {
					cursorCol--
					break
				}
				cursorRow.Replace(cursorCol, text, mode)
			}
		case "r", "R", keys.F2:
			cursor := &cursorRow.Cell[cursorCol]
			q := cursor.IsQuoted()
			view.clearCache()
			text, err := pilot.ReadLine(out, "replace cell>", cursor.Text(), makeCandidate(cursorRow.lnum-1, cursorCol, cursorRow))
			if err != nil {
				break
			}
			cursorRow.Replace(cursorCol, text, mode)
			if q {
				*cursor = cursor.Quote(mode)
			}
		case "u":
			cursorRow.Cell[cursorCol].Restore(mode)
		case "y":
			killbuffer = cursorRow.Cell[cursorCol].Text()
			message = "yanked the current cell: " + killbuffer
		case "p":
			cursorRow.Replace(cursorCol, killbuffer, mode)
			message = "pasted: " + killbuffer
		case "d", "x":
			if len(cursorRow.Cell) <= 1 {
				cursorRow.Replace(0, "", mode)
			} else {
				cursorRow.Delete(cursorCol)
			}
		case "\"":
			cursor := &cursorRow.Cell[cursorCol]
			if cursor.IsQuoted() {
				cursorRow.Replace(cursorCol, cursor.Text(), mode)
			} else {
				*cursor = cursor.Quote(mode)
			}
		case "w":
			if reader != nil {
				io.WriteString(out, _ANSI_YELLOW+"\rw: Wait a moment for reading all data..."+_ANSI_ERASE_LINE)
				for {
					row, err := uncsv.ReadLine(reader, mode)
					if err != nil && err != io.EOF {
						return nil, err
					}
					csvlines.PushBack(row)
					if err == io.EOF {
						break
					}
				}
			}
			if err := cmdWrite(pilot, csvlines, mode, out); err != nil {
				message = err.Error()
			}
			view.clearCache()
		}
		if L := len(cursorRow.Cell); L <= 0 {
			cursorCol = 0
		} else if cursorCol >= L {
			cursorCol = L - 1
		}
		if cursorRow.lnum < startRow.lnum {
			startRow = cursorRow.Clone()
		} else if cursorRow.lnum >= startRow.lnum+screenHeight-1 {
			goal := cursorRow.lnum - (screenHeight - 1) + 1
			for startRow = cursorRow.Clone(); startRow.lnum > goal; {
				startRow = startRow.Prev()
			}
		}
		if cursorCol < startCol {
			startCol = cursorCol
		} else if cursorCol >= startCol+cols {
			startCol = cursorCol - cols + 1
		}
		up(lfCount, out)
	}
}
