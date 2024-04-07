package main

import (
	"bufio"
	"container/list"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/term"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-runewidth"
	"github.com/mattn/go-tty"

	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/completion"
	"github.com/nyaosorg/go-readline-ny/keys"
	"github.com/nyaosorg/go-readline-skk"

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

	_ANSI_BLINK_ON  = "\x1B[6m"
	_ANSI_BLINK_OFF = "\x1B[25m"

	_ANSI_UNDERLINE_ON  = "\x1B[4m"
	_ANSI_UNDERLINE_OFF = "\x1B[24m"
)

type _ColorStyle struct {
	Cursor [2]string
	Even   [2]string
	Odd    [2]string
}

var bodyColorStyle = _ColorStyle{
	Cursor: [...]string{"\x1B[107;30m", "\x1B[40;37m"},
	Even:   [...]string{"\x1B[48;5;235;37;1m", "\x1B[22;40m"},
	Odd:    [...]string{"\x1B[40;37;1m", "\x1B[22m"},
}

var headColorStyle = _ColorStyle{
	Cursor: [...]string{"\x1B[107;30m", "\x1B[40;36m"},
	Even:   [...]string{"\x1B[48;5;235;36;1m", "\x1B[22;40m"},
	Odd:    [...]string{"\x1B[40;36;1m", "\x1B[22m"},
}

var (
	flagCellWidth = flag.Uint("w", 14, "set the width of cell")
	flagHeader    = flag.Uint("h", 1, "the number of row-header")
	flagTsv       = flag.Bool("t", false, "use TAB as field-separator")
	flagCsv       = flag.Bool("c", false, "use Comma as field-separator")
	flagIana      = flag.String("iana", "", "IANA-registered-name to decode/encode NonUTF8 text(for example: Shift_JIS,EUC-JP... )")
)

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
		var off string
		if i == cursorPos {
			io.WriteString(out, style.Cursor[0])
			off = style.Cursor[1]
		} else if reverse {
			io.WriteString(out, style.Odd[0])
			off = style.Odd[1]
		} else {
			io.WriteString(out, style.Even[0])
			off = style.Even[1]
		}
		if cursor.Modified() {
			io.WriteString(out, _ANSI_UNDERLINE_ON)
		}
		io.WriteString(out, ss)
		if cursor.Modified() {
			io.WriteString(out, _ANSI_UNDERLINE_OFF)
		}
		screenWidth -= cw
		io.WriteString(out, "\x1B[K")
		io.WriteString(out, off)
		if screenWidth <= 0 {
			break
		}
		fmt.Fprintf(out, "\x1B[%dG", nextI*cellWidth+1)
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

func drawPage(page func(func([]uncsv.Cell) bool), csrpos, csrlin, w, h int, style *_ColorStyle, cache map[int]string, out io.Writer) int {
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
		drawLine(record, int(*flagCellWidth), w, cursorPos, reverse, style, &buffer)
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

var (
	headCache = map[int]string{}
	bodyCache = map[int]string{}
)

func clearCache() {
	clear(headCache)
	clear(bodyCache)
}

func drawView(header, pointor *list.Element, startRow, startCol, cursorRow, cursorCol, screenHeight, screenWidth int, out io.Writer) int {
	// print header
	lfCount := 0
	if h := int(*flagHeader); h > 0 {
		enum := func(callback func([]uncsv.Cell) bool) {
			for i := 0; i < h && header != nil; i++ {
				row := header.Value.(*uncsv.Row)
				if !callback(cellsAfter(row.Cell, startCol)) {
					return
				}
				header = header.Next()
			}
		}
		lfCount = drawPage(enum, cursorCol-startCol, cursorRow, screenWidth-1, h, &headColorStyle, headCache, out)
	}
	if h := int(*flagHeader); startRow < h {
		startRow = h
		for i := 0; i < h; i++ {
			pointor = pointor.Next()
		}
	}
	// print body
	enum := func(callback func([]uncsv.Cell) bool) {
		for pointor != nil {
			row := pointor.Value.(*uncsv.Row)
			if !callback(cellsAfter(row.Cell, startCol)) {
				return
			}
			pointor = pointor.Next()
			startRow++
		}
	}
	style := &bodyColorStyle
	if *flagHeader%2 == 1 {
		style = &_ColorStyle{
			Cursor: bodyColorStyle.Cursor,
			Even:   bodyColorStyle.Odd,
			Odd:    bodyColorStyle.Even,
		}
	}
	return lfCount + drawPage(enum, cursorCol-startCol, cursorRow-startRow, screenWidth-1, screenHeight-1, style, bodyCache, out)
}

var skkInit = sync.OnceFunc(func() {
	env := os.Getenv("GOREADLINESKK")
	if env != "" {
		_, err := skk.Config{
			MiniBuffer: skk.MiniBufferOnCurrentLine{},
		}.SetupWithString(env)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
})

func getline(out io.Writer, prompt, defaultStr string, c candidate) (string, error) {
	skkInit()
	clearCache()
	editor := &readline.Editor{
		Writer:  out,
		Default: defaultStr,
		History: c,
		Cursor:  65535,
		PromptWriter: func(w io.Writer) (int, error) {
			return fmt.Fprintf(w, "\r\x1B[0;33;40;1m%s%s", prompt, _ANSI_ERASE_LINE)
		},
		LineFeedWriter: func(readline.Result, io.Writer) (int, error) {
			return 0, nil
		},
		Coloring: &skk.Coloring{},
	}
	if len(c) > 0 {
		editor.BindKey(keys.CtrlI, completion.CmdCompletion{
			Completion: c,
		})
	}

	defer io.WriteString(out, _ANSI_CURSOR_OFF)
	editor.BindKey(keys.Escape, readline.CmdInterrupt)
	return editor.ReadLine(context.Background())
}

func yesNo(tty1 *tty.TTY, out io.Writer, message string) bool {
	fmt.Fprintf(out, "%s\r%s%s", _ANSI_YELLOW, message, _ANSI_ERASE_LINE)
	ch, err := readline.GetKey(tty1)
	return err == nil && ch == "y"
}

func first[T any](value T, _ error) T {
	return value
}

func mains() error {
	out := colorable.NewColorableStdout()

	io.WriteString(out, _ANSI_CURSOR_OFF)
	defer io.WriteString(out, _ANSI_CURSOR_ON)

	csvlines := list.New()
	mode := &uncsv.Mode{}

	if *flagIana != "" {
		if err := mode.SetEncoding(*flagIana); err != nil {
			return fmt.Errorf("-iana %w", err)
		}
	}
	var reader *bufio.Reader
	if len(flag.Args()) <= 0 && term.IsTerminal(int(os.Stdin.Fd())) {
		// Start with one empty line
		newRow := uncsv.NewRow(mode)
		csvlines.PushBack(&newRow)
		mode.Comma = '\t'
	} else {
		mode.Comma = ','
		args := flag.Args()
		if len(args) >= 1 && !strings.HasSuffix(strings.ToLower(args[0]), ".csv") {
			mode.Comma = '\t'
		}
		if *flagTsv {
			mode.Comma = '\t'
		}
		if *flagCsv {
			mode.Comma = ','
		}
		reader = bufio.NewReader(multiFileReader(args...))
		for i := 0; i < 100; i++ {
			row, err := uncsv.ReadLine(reader, mode)
			if err != nil && err != io.EOF {
				return err
			}
			csvlines.PushBack(row)
			if err == io.EOF {
				reader = nil
				break
			}
		}
		if csvlines.Len() <= 0 {
			return io.EOF
		}
	}
	tty1, err := tty.Open()
	if err != nil {
		return err
	}
	defer tty1.Close()

	if _, ok := out.(*os.File); ok {
		if err := initAmbiguousWidth(tty1); err != nil {
			return err
		}
	}

	colIndex := 0
	rowIndex := 0
	startRow := 0
	startCol := 0
	startP := csvlines.Front()
	cursorP := startP

	lastSearch := searchForward
	lastSearchRev := searchBackward
	lastWord := ""
	var lastWidth, lastHeight int

	keyWorker := nonblock.New(func() (string, error) { return readline.GetKey(tty1) })
	defer keyWorker.Close()

	message := ""
	var killbuffer string
	for {
		screenWidth, screenHeight, err := tty1.Size()
		if err != nil {
			return err
		}
		screenHeight -= int(*flagHeader)
		if lastWidth != screenWidth || lastHeight != screenHeight {
			clearCache()
			lastWidth = screenWidth
			lastHeight = screenHeight
			io.WriteString(out, _ANSI_CURSOR_OFF)
		}
		cols := (screenWidth - 1) / int(*flagCellWidth)

		lfCount := drawView(csvlines.Front(), startP, startRow, startCol, rowIndex, colIndex, screenHeight, screenWidth, out)
		repaint := func() {
			up(lfCount, out)
			lfCount = drawView(csvlines.Front(), startP, startRow, startCol, rowIndex, colIndex, screenHeight, screenWidth, out)
		}

		io.WriteString(out, _ANSI_YELLOW)
		cursorRow := cursorP.Value.(*uncsv.Row)
		if message != "" {
			io.WriteString(out, runewidth.Truncate(message, screenWidth-1, ""))
			message = ""
		} else if 0 <= rowIndex && rowIndex < csvlines.Len() {
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
			if 0 <= colIndex && colIndex < len(cursorRow.Cell) {
				n += first(fmt.Fprintf(out, "(%d,%d/%d): ",
					colIndex+1,
					rowIndex+1,
					csvlines.Len()))
				var buffer strings.Builder
				buffer.WriteString(cursorRow.Cell[colIndex].SourceText(mode))
				if colIndex < len(cursorRow.Cell)-1 {
					buffer.WriteByte(mode.Comma)
				} else if term := cursorRow.Term; term != "" {
					buffer.WriteString(term)
				} else { // EOF
					buffer.WriteString("\u2592")
				}
				io.WriteString(out, runewidth.Truncate(replaceTable.Replace(buffer.String()), screenWidth-n, "..."))
			}
		}
		io.WriteString(out, _ANSI_RESET)
		io.WriteString(out, _ANSI_ERASE_SCRN_AFTER)

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
			return err != io.EOF
		})
		if err != nil {
			return err
		}

		switch ch {
		case keys.CtrlL:
			clearCache()
		case "q", keys.Escape:
			io.WriteString(out, _ANSI_YELLOW+"\rQuit Sure ? [y/n]"+_ANSI_ERASE_LINE)
			if ch, err := readline.GetKey(tty1); err == nil && ch == "y" {
				io.WriteString(out, "\n")
				return nil
			}
		case "j", keys.Down, keys.CtrlN, keys.Enter:
			if rowIndex < csvlines.Len()-1 {
				rowIndex++
				cursorP = cursorP.Next()
			}
		case "k", keys.Up, keys.CtrlP:
			if rowIndex > 0 {
				rowIndex--
				cursorP = cursorP.Prev()
			}
		case "h", keys.Left, keys.CtrlB, keys.ShiftTab:
			if colIndex > 0 {
				colIndex--
			}
		case "l", keys.Right, keys.CtrlF, keys.CtrlI:
			colIndex++
		case "0", "^", keys.CtrlA:
			colIndex = 0
		case "$", keys.CtrlE:
			colIndex = len(cursorRow.Cell) - 1
		case "<":
			rowIndex = 0
			startRow = 0
			colIndex = 0
			startCol = 0
			startP = csvlines.Front()
			cursorP = startP
		case ">", "G":
			rowIndex = csvlines.Len() - 1
			cursorP = csvlines.Back()
		case "n":
			if lastWord == "" {
				break
			}
			foundP, r, c := lastSearch(cursorP, rowIndex, colIndex, lastWord)
			if foundP == nil {
				message = fmt.Sprintf("%s: not found", lastWord)
				break
			}
			rowIndex = r
			colIndex = c
			cursorP = foundP
		case "N":
			if lastWord == "" {
				break
			}
			foundP, r, c := lastSearchRev(cursorP, rowIndex, colIndex, lastWord)
			if foundP == nil {
				message = fmt.Sprintf("%s: not found", lastWord)
				break
			}
			rowIndex = r
			colIndex = c
			cursorP = foundP
		case "/", "?":
			var err error
			lastWord, err = getline(out, ch, "", nil)
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
			foundP, r, c := lastSearch(cursorP, rowIndex, colIndex, lastWord)
			if foundP == nil {
				message = fmt.Sprintf("%s: not found", lastWord)
				break
			}
			rowIndex = r
			colIndex = c
			cursorP = foundP
		case "o":
			newRow := uncsv.NewRow(mode)
			newRow.Term = cursorRow.Term
			if cursorRow.Term == "" {
				cursorRow.Term = mode.DefaultTerm
			}
			cursorP = csvlines.InsertAfter(&newRow, cursorP)
			rowIndex++
			cursorRow = cursorP.Value.(*uncsv.Row)
			repaint()
			text, _ := getline(out, "new line>", "", makeCandidate(rowIndex-1, colIndex, cursorP))
			cursorRow.Replace(0, text, mode)

		case "O":
			startPrevP := startP.Prev()
			newRow := uncsv.NewRow(mode)
			cursorP = csvlines.InsertBefore(&newRow, cursorP)
			if startPrevP != nil {
				startP = startPrevP.Next()
			} else {
				startP = csvlines.Front()
			}
			cursorRow = cursorP.Value.(*uncsv.Row)
			repaint()
			text, _ := getline(out, "new line>", "", makeCandidate(rowIndex-1, colIndex, cursorP))
			cursorRow.Replace(0, text, mode)
		case "D":
			if csvlines.Len() <= 1 {
				break
			}
			startPrevP := startP.Prev()
			prevP := cursorP.Prev()
			removedRow := csvlines.Remove(cursorP).(*uncsv.Row)
			if prevP == nil {
				cursorP = csvlines.Front()
				rowIndex = 0
			} else if next := prevP.Next(); next != nil {
				cursorP = next
			} else {
				cursorP = prevP
				cursorP.Value.(*uncsv.Row).Term = removedRow.Term
				rowIndex--
			}
			if startPrevP == nil {
				startP = csvlines.Front()
			} else {
				startP = startPrevP.Next()
			}
		case "i":
			text, err := getline(out, "insert cell>", "", makeCandidate(rowIndex, colIndex, cursorP))
			if err != nil {
				break
			}
			if cells := cursorRow.Cell; len(cells) == 1 && cells[0].Text() == "" {
				cursorRow.Replace(colIndex, text, mode)
			} else {
				cursorRow.Insert(colIndex, text, mode)
				colIndex++
			}
		case "a":
			if cells := cursorRow.Cell; len(cells) == 1 && cells[0].Text() == "" {
				// current column is the last one and it is empty
				text, err := getline(out, "append cell>", "", makeCandidate(rowIndex, colIndex+1, cursorP))
				if err != nil {
					break
				}
				cursorRow.Replace(colIndex, text, mode)
			} else {
				colIndex++
				cursorRow.Insert(colIndex, "", mode)
				repaint()
				text, err := getline(out, "append cell>", "", makeCandidate(rowIndex+1, colIndex+1, cursorP))
				if err != nil {
					colIndex--
					break
				}
				cursorRow.Replace(colIndex, text, mode)
			}
		case "r", "R", keys.F2:
			cursor := &cursorRow.Cell[colIndex]
			q := cursor.IsQuoted()
			text, err := getline(out, "replace cell>", cursor.Text(), makeCandidate(rowIndex-1, colIndex, cursorP))
			if err != nil {
				break
			}
			cursorRow.Replace(colIndex, text, mode)
			if q {
				*cursor = cursor.Quote(mode)
			}
		case "u":
			cursorRow.Cell[colIndex].Restore(mode)
		case "y":
			killbuffer = cursorRow.Cell[colIndex].Text()
			message = "yanked the current cell: " + killbuffer
		case "p":
			cursorRow.Replace(colIndex, killbuffer, mode)
			message = "pasted: " + killbuffer
		case "d", "x":
			if len(cursorRow.Cell) <= 1 {
				cursorRow.Replace(0, "", mode)
			} else {
				cursorRow.Delete(colIndex)
			}
		case "\"":
			cursor := &cursorRow.Cell[colIndex]
			if cursor.IsQuoted() {
				cursorRow.Replace(colIndex, cursor.Text(), mode)
			} else {
				*cursor = cursor.Quote(mode)
			}
		case "w":
			if reader != nil {
				io.WriteString(out, _ANSI_YELLOW+"\rw: Wait a moment for reading all data..."+_ANSI_ERASE_LINE)
				for {
					row, err := uncsv.ReadLine(reader, mode)
					if err != nil && err != io.EOF {
						return err
					}
					csvlines.PushBack(row)
					if err == io.EOF {
						break
					}
				}
			}
			if err := cmdWrite(csvlines, mode, tty1, out); err != nil {
				message = err.Error()
			}
			clearCache()
		}
		if L := len(cursorRow.Cell); L <= 0 {
			colIndex = 0
		} else if colIndex >= L {
			colIndex = L - 1
		}
		if rowIndex < startRow {
			startRow = rowIndex
			startP = cursorP
		} else if rowIndex >= startRow+screenHeight-1 {
			startRow = rowIndex - (screenHeight - 1) + 1
			var i int
			for startP, i = cursorP, rowIndex; i > startRow; i-- {
				startP = startP.Prev()
			}
		}
		if colIndex < startCol {
			startCol = colIndex
		} else if colIndex >= startCol+cols {
			startCol = colIndex - cols + 1
		}
		up(lfCount, out)
	}
}

var version string

func main() {
	fmt.Printf("%s %s-%s-%s by %s\n",
		os.Args[0], version, runtime.GOOS, runtime.GOARCH, runtime.Version())

	disable := colorable.EnableColorsStdout(nil)
	if disable != nil {
		defer disable()
	}

	flag.Parse()
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
