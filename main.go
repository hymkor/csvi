package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"
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

	"github.com/hymkor/csview/uncsv"
)

const (
	_ANSI_CURSOR_OFF = "\x1B[?25l"
	_ANSI_CURSOR_ON  = "\x1B[?25h"
	_ANSI_YELLOW     = "\x1B[0;33;1m"
	_ANSI_RESET      = "\x1B[0m"

	_ANSI_ERASE_LINE       = "\x1B[0m\x1B[0K"
	_ANSI_ERASE_SCRN_AFTER = "\x1B[0m\x1B[0J"
)

type _ColorStyle struct {
	Cursor [2]string
	Even   [2]string
	Odd    [2]string
}

var bodyColorStyle = _ColorStyle{
	Cursor: [...]string{"\x1B[40;37;1;7m", "\x1B[27;22m"},
	Even:   [...]string{"\x1B[48;5;235;37;1m", "\x1B[22;40m"},
	Odd:    [...]string{"\x1B[40;37;1m", "\x1B[22m"},
}

var headColorStyle = _ColorStyle{
	Cursor: [...]string{"\x1B[40;36;1;7m", "\x1B[27;22m"},
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
		ss, w := cutStrInWidth(text, cw)
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
			io.WriteString(out, "\x1B[4m")
		}
		io.WriteString(out, ss)
		if cursor.Modified() {
			io.WriteString(out, "\x1B[24m")
		}
		screenWidth -= w
		for j := cw - w; j > 0; j-- {
			out.Write([]byte{' '})
			screenWidth--
		}
		io.WriteString(out, off)
		if screenWidth <= 0 {
			break
		}
		i = nextI
	}
	io.WriteString(out, _ANSI_ERASE_LINE)
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
	for record := range page {
		if count >= h {
			break
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
	}
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

func drawView(csvlines []uncsv.Row, startRow, startCol, cursorRow, cursorCol, screenHeight, screenWidth int, out io.Writer) int {
	// print header
	lfCount := 0
	if h := int(*flagHeader); h > 0 {
		enum := func(callback func([]uncsv.Cell) bool) {
			for i := 0; i < h; i++ {
				if !callback(cellsAfter(csvlines[i].Cell, startCol)) {
					return
				}
			}
		}
		lfCount = drawPage(enum, cursorCol-startCol, cursorRow, screenWidth-1, h, &headColorStyle, headCache, out)
	}
	if h := int(*flagHeader); startRow < h {
		startRow = h
	}
	// print body
	enum := func(callback func([]uncsv.Cell) bool) {
		for startRow < len(csvlines) {
			if !callback(cellsAfter(csvlines[startRow].Cell, startCol)) {
				return
			}
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

	var csvlines []uncsv.Row
	mode := &uncsv.Mode{}

	if *flagIana != "" {
		if err := mode.SetEncoding(*flagIana); err != nil {
			return fmt.Errorf("-iana %w", err)
		}
	}
	if len(flag.Args()) <= 0 && term.IsTerminal(int(os.Stdin.Fd())) {
		// Start with one empty line
		csvlines = []uncsv.Row{uncsv.NewRow(mode)}
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
		var err error
		csvlines, err = uncsv.ReadAll(multiFileReader(args...), mode)
		if err != nil {
			return err
		}
		if len(csvlines) <= 0 {
			return io.EOF
		}
	}
	tty1, err := tty.Open()
	if err != nil {
		return err
	}
	defer tty1.Close()

	if err := initAmbiguousWidth(tty1); err != nil {
		return err
	}

	colIndex := 0
	rowIndex := 0
	startRow := 0
	startCol := 0

	lastSearch := searchForward
	lastSearchRev := searchBackward
	lastWord := ""
	var lastWidth, lastHeight int

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

		lfCount := drawView(csvlines, startRow, startCol, rowIndex, colIndex, screenHeight, screenWidth, out)
		repaint := func() {
			up(lfCount, out)
			lfCount = drawView(csvlines, startRow, startCol, rowIndex, colIndex, screenHeight, screenWidth, out)
		}

		io.WriteString(out, _ANSI_YELLOW)
		if message != "" {
			io.WriteString(out, runewidth.Truncate(message, screenWidth-1, ""))
			message = ""
		} else if 0 <= rowIndex && rowIndex < len(csvlines) {
			n := 0
			if mode.Comma == '\t' {
				n += first(io.WriteString(out, "[TSV]"))
			} else if mode.Comma == ',' {
				n += first(io.WriteString(out, "[CSV]"))
			}
			switch csvlines[rowIndex].Term {
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
				n += first(io.WriteString(out, "[ANSI]"))
			}
			if 0 <= colIndex && colIndex < len(csvlines[rowIndex].Cell) {
				n += first(fmt.Fprintf(out, "(%d,%d): ",
					rowIndex+1,
					colIndex+1))

				var buffer strings.Builder
				buffer.WriteString(csvlines[rowIndex].Cell[colIndex].SourceText(mode))
				if colIndex < len(csvlines[rowIndex].Cell)-1 {
					buffer.WriteByte(mode.Comma)
				} else if term := csvlines[rowIndex].Term; term != "" {
					buffer.WriteString(term)
				} else { // EOF
					buffer.WriteString("\u2592")
				}
				io.WriteString(out, runewidth.Truncate(replaceTable.Replace(buffer.String()), screenWidth-n, "..."))
			}
		}
		io.WriteString(out, _ANSI_RESET)
		io.WriteString(out, _ANSI_ERASE_SCRN_AFTER)

		ch, err := readline.GetKey(tty1)
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
			if rowIndex < len(csvlines)-1 {
				rowIndex++
			}
		case "k", keys.Up, keys.CtrlP:
			if rowIndex > 0 {
				rowIndex--
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
			colIndex = len(csvlines[rowIndex].Cell) - 1
		case "<":
			rowIndex = 0
			startRow = 0
			colIndex = 0
			startCol = 0
		case ">", "G":
			rowIndex = len(csvlines) - 1
		case "n":
			if lastWord == "" {
				break
			}
			found, r, c := lastSearch(csvlines, rowIndex, colIndex, lastWord)
			if !found {
				message = fmt.Sprintf("%s: not found", lastWord)
				break
			}
			rowIndex = r
			colIndex = c
		case "N":
			if lastWord == "" {
				break
			}
			found, r, c := lastSearchRev(csvlines, rowIndex, colIndex, lastWord)
			if !found {
				message = fmt.Sprintf("%s: not found", lastWord)
				break
			}
			rowIndex = r
			colIndex = c
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
			found, r, c := lastSearch(csvlines, rowIndex, colIndex, lastWord)
			if !found {
				message = fmt.Sprintf("%s: not found", lastWord)
				break
			}
			rowIndex = r
			colIndex = c
		case "o":
			if rowIndex > len(csvlines)-1 {
				break
			}
			rowIndex++
			fallthrough
		case "O":
			csvlines = slices.Insert(csvlines, rowIndex, uncsv.NewRow(mode))
			repaint()
			text, _ := getline(out, "new line>", "", makeCandidate(rowIndex-1, colIndex, csvlines))
			csvlines[rowIndex].Replace(0, text, mode)
		case "D":
			if len(csvlines) <= 1 {
				break
			}
			csvlines = slices.Delete(csvlines, rowIndex, rowIndex+1)
			if rowIndex >= len(csvlines) {
				rowIndex--
			}
		case "i":
			text, err := getline(out, "insert cell>", "", makeCandidate(rowIndex, colIndex, csvlines))
			if err != nil {
				break
			}
			if cells := csvlines[rowIndex].Cell; len(cells) == 1 && cells[0].Text() == "" {
				csvlines[rowIndex].Replace(colIndex, text, mode)
			} else {
				csvlines[rowIndex].Insert(colIndex, text, mode)
				colIndex++
			}
		case "a":
			if cells := csvlines[rowIndex].Cell; len(cells) == 1 && cells[0].Text() == "" {
				// current column is the last one and it is empty
				text, err := getline(out, "append cell>", "", makeCandidate(rowIndex, colIndex+1, csvlines))
				if err != nil {
					break
				}
				csvlines[rowIndex].Replace(colIndex, text, mode)
			} else {
				colIndex++
				csvlines[rowIndex].Insert(colIndex, "", mode)
				repaint()
				text, err := getline(out, "append cell>", "", makeCandidate(rowIndex+1, colIndex+1, csvlines))
				if err != nil {
					colIndex--
					break
				}
				csvlines[rowIndex].Replace(colIndex, text, mode)
			}
		case "r", "R", keys.F2:
			text, err := getline(out, "replace cell>", csvlines[rowIndex].Cell[colIndex].Text(), makeCandidate(rowIndex-1, colIndex, csvlines))
			if err != nil {
				break
			}
			csvlines[rowIndex].Replace(colIndex, text, mode)
		case "u":
			csvlines[rowIndex].Cell[colIndex].Restore(mode)
		case "y":
			killbuffer = csvlines[rowIndex].Cell[colIndex].Text()
			message = "yanked the current cell: " + killbuffer
		case "p":
			csvlines[rowIndex].Replace(colIndex, killbuffer, mode)
			message = "pasted: " + killbuffer
		case "d", "x":
			if len(csvlines[rowIndex].Cell) <= 1 {
				csvlines[rowIndex].Replace(0, "", mode)
			} else {
				csvlines[rowIndex].Delete(colIndex)
			}
		case "\"":
			cursor := &csvlines[rowIndex].Cell[colIndex]
			if cursor.IsQuoted() {
				csvlines[rowIndex].Replace(colIndex, cursor.Text(), mode)
			} else {
				*cursor = cursor.Quote(mode)
			}
		case "w":
			if err := cmdWrite(csvlines, mode, tty1, out); err != nil {
				message = err.Error()
			}
			clearCache()
		}
		if L := len(csvlines[rowIndex].Cell); L <= 0 {
			colIndex = 0
		} else if colIndex >= L {
			colIndex = L - 1
		}
		if rowIndex < startRow {
			startRow = rowIndex
		} else if rowIndex >= startRow+screenHeight-1 {
			startRow = rowIndex - (screenHeight - 1) + 1
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
	fmt.Printf("csview %s-%s-%s by %s\n",
		version, runtime.GOOS, runtime.GOARCH, runtime.Version())

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
