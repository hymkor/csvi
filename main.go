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

	"github.com/hymkor/csview/unbreakable-csv"
)

var (
	flagTsv  = flag.Bool("t", false, "use TAB as field-separator")
	flagCsv  = flag.Bool("c", false, "use Comma as field-separator")
	flagIana = flag.String("iana", "", "IANA-registered-name to decode/encode NonUTF8 text(for example: Shift_JIS,EUC-JP... )")
)

const (
	CURSOR_COLOR     = "\x1B[0;40;37;1;7m"
	CELL1_COLOR      = "\x1B[0;48;5;235;37;1m"
	CELL2_COLOR      = "\x1B[0;40;37;1m"
	ERASE_LINE       = "\x1B[0m\x1B[0K"
	ERASE_SCRN_AFTER = "\x1B[0m\x1B[0J"
)

type LineView struct {
	CSV       []csv.Cell
	CellWidth int
	MaxInLine int
	CursorPos int
	Reverse   bool
	Out       io.Writer
}

var replaceTable = strings.NewReplacer(
	"\r", "\u240D",
	"\x1B", "\u241B",
	"\n", "\u240A",
	"\t", "\u2409")

// See. en.wikipedia.org/wiki/Unicode_control_characters#Control_pictures

func (v LineView) Draw() {
	leftWidth := v.MaxInLine
	i := 0
	csvs := v.CSV
	for len(csvs) > 0 {
		cursor := csvs[0]
		text := cursor.Text()
		csvs = csvs[1:]
		nextI := i + 1

		cw := v.CellWidth
		for len(csvs) > 0 && csvs[0].Text() == "" && nextI != v.CursorPos {
			cw += v.CellWidth
			csvs = csvs[1:]
			nextI++
		}
		if cw > leftWidth || len(csvs) <= 0 {
			cw = leftWidth
		}
		text = replaceTable.Replace(text)
		ss, w := cutStrInWidth(text, cw)
		if i == v.CursorPos {
			io.WriteString(v.Out, CURSOR_COLOR)
		} else if v.Reverse {
			io.WriteString(v.Out, CELL2_COLOR)
		} else {
			io.WriteString(v.Out, CELL1_COLOR)
		}
		if cursor.Modified() {
			io.WriteString(v.Out, "\x1B[4m")
		}
		io.WriteString(v.Out, ss)
		if cursor.Modified() {
			io.WriteString(v.Out, "\x1B[24m")
		}
		leftWidth -= w
		for j := cw - w; j > 0; j-- {
			v.Out.Write([]byte{' '})
			leftWidth--
		}
		if leftWidth <= 0 {
			break
		}
		i = nextI
	}
	io.WriteString(v.Out, ERASE_LINE)
}

var cache = map[int]string{}

const CELL_WIDTH = 14

func view(read func() ([]csv.Cell, error), csrpos, csrlin, w, h int, out io.Writer) (func(), error) {
	reverse := false
	count := 0
	lfCount := 0
	for {
		if count >= h {
			break
		}
		record, err := read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return func() {}, err
		}
		if count > 0 {
			lfCount++
			fmt.Fprintln(out, "\r") // "\r" is for Linux and go-tty
		}
		var buffer strings.Builder
		v := LineView{
			CSV:       record,
			CellWidth: CELL_WIDTH,
			MaxInLine: w,
			Reverse:   reverse,
			Out:       &buffer,
		}
		if count == csrlin {
			v.CursorPos = csrpos
		} else {
			v.CursorPos = -1
		}

		v.Draw()
		line := buffer.String()
		if f := cache[count]; f != line {
			io.WriteString(out, line)
			cache[count] = line
		}
		reverse = !reverse
		count++
	}
	fmt.Fprintln(out, "\r") // \r is for Linux & go-tty
	lfCount++
	return func() {
		if lfCount > 0 {
			fmt.Fprintf(out, "\r\x1B[%dA", lfCount)
		} else {
			fmt.Fprint(out, "\r")
		}
	}, nil
}

type MemoryCsv struct {
	Data   []csv.Row
	StartX int
	StartY int
}

func (M *MemoryCsv) Read() ([]csv.Cell, error) {
	if M.StartY >= len(M.Data) {
		return nil, io.EOF
	}
	row := M.Data[M.StartY]
	var cells []csv.Cell
	if M.StartX <= len(row.Cell) {
		cells = row.Cell[M.StartX:]
	} else {
		cells = []csv.Cell{}
	}
	M.StartY++
	return cells, nil
}

func drawView(csvlines []csv.Row, startRow, startCol, rowIndex, colIndex, screenHeight, screenWidth int, out io.Writer) (func(), error) {
	window := &MemoryCsv{Data: csvlines, StartX: startCol, StartY: startRow}
	return view(window.Read, colIndex-startCol, rowIndex-startRow, screenWidth-1, screenHeight-1, out)
}

const (
	_ANSI_CURSOR_OFF = "\x1B[?25l"
	_ANSI_CURSOR_ON  = "\x1B[?25h"
	_ANSI_YELLOW     = "\x1B[0;33;1m"
	_ANSI_RESET      = "\x1B[0m"
)

const (
	_KEY_CTRL_A = "\x01"
	_KEY_CTRL_B = "\x02"
	_KEY_CTRL_E = "\x05"
	_KEY_CTRL_F = "\x06"
	_KEY_CTRL_L = "\x0C"
	_KEY_CTRL_N = "\x0E"
	_KEY_CTRL_P = "\x10"
	_KEY_DOWN   = "\x1B[B"
	_KEY_ESC    = "\x1B"
	_KEY_LEFT   = "\x1B[D"
	_KEY_RIGHT  = "\x1B[C"
	_KEY_UP     = "\x1B[A"
	_KEY_F2     = "\x1B[OQ"
)

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

func getfilename(out io.Writer, prompt, defaultStr string) (string, error) {
	skkInit()
	editor := &readline.Editor{
		Writer:  out,
		Default: defaultStr,
		Cursor:  65535,
		PromptWriter: func(w io.Writer) (int, error) {
			return fmt.Fprintf(w, "\r\x1B[0;33;40;1m%s%s", prompt, ERASE_LINE)
		},
		LineFeedWriter: func(readline.Result, io.Writer) (int, error) {
			return 0, nil
		},
		Coloring: &skk.Coloring{},
	}
	editor.BindKey(keys.CtrlI, completion.CmdCompletionOrList{
		Completion: completion.File{},
	})

	defer io.WriteString(out, _ANSI_CURSOR_OFF)
	editor.BindKey(keys.Escape, readline.CmdInterrupt)
	return editor.ReadLine(context.Background())
}

func getline(out io.Writer, prompt, defaultStr string, c candidate) (string, error) {
	skkInit()

	editor := &readline.Editor{
		Writer:  out,
		Default: defaultStr,
		History: c,
		Cursor:  65535,
		PromptWriter: func(w io.Writer) (int, error) {
			return fmt.Fprintf(w, "\r\x1B[0;33;40;1m%s%s", prompt, ERASE_LINE)
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
	fmt.Fprintf(out, "%s\r%s%s", _ANSI_YELLOW, message, ERASE_LINE)
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

	var csvlines []csv.Row
	mode := &csv.Mode{}

	if *flagIana != "" {
		if err := mode.SetEncoding(*flagIana); err != nil {
			return fmt.Errorf("-iana %w", err)
		}
	}
	if len(flag.Args()) <= 0 && term.IsTerminal(int(os.Stdin.Fd())) {
		// Start with one empty line
		csvlines = []csv.Row{csv.NewRow(mode)}
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
		csvlines, err = csv.ReadAll(multiFileReader(args...), mode)
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
		if lastWidth != screenWidth || lastHeight != screenHeight {
			cache = map[int]string{}
			lastWidth = screenWidth
			lastHeight = screenHeight
			io.WriteString(out, _ANSI_CURSOR_OFF)
		}
		cols := (screenWidth - 1) / CELL_WIDTH

		rewind, err := drawView(csvlines, startRow, startCol, rowIndex, colIndex, screenHeight, screenWidth, out)
		if err != nil {
			return err
		}
		repaint := func() error {
			rewind()
			rewind, err = drawView(csvlines, startRow, startCol, rowIndex, colIndex, screenHeight, screenWidth, out)
			return err
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
			if mode.DefaultTerm == "\r\n" {
				n += first(io.WriteString(out, "[CRLF]"))
			} else {
				n += first(io.WriteString(out, "[LF]"))
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
				buffer.WriteString(csvlines[rowIndex].Cell[colIndex].ReadableSource(mode))
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
		io.WriteString(out, ERASE_SCRN_AFTER)

		ch, err := readline.GetKey(tty1)
		if err != nil {
			return err
		}

		getline := func(out io.Writer, prompt string, defaultStr string, c candidate) (string, error) {
			text, err := getline(out, prompt, defaultStr, c)
			cache = map[int]string{}
			return text, err
		}

		switch ch {
		case _KEY_CTRL_L:
			cache = map[int]string{}
		case "q", _KEY_ESC:
			io.WriteString(out, _ANSI_YELLOW+"\rQuit Sure ? [y/n]"+ERASE_LINE)
			if ch, err := readline.GetKey(tty1); err == nil && ch == "y" {
				io.WriteString(out, "\n")
				return nil
			}
		case "j", _KEY_DOWN, _KEY_CTRL_N:
			if rowIndex < len(csvlines)-1 {
				rowIndex++
			}
		case "k", _KEY_UP, _KEY_CTRL_P:
			if rowIndex > 0 {
				rowIndex--
			}
		case "h", _KEY_LEFT, _KEY_CTRL_B:
			if colIndex > 0 {
				colIndex--
			}
		case "l", _KEY_RIGHT, _KEY_CTRL_F:
			colIndex++
		case "0", "^", _KEY_CTRL_A:
			colIndex = 0
		case "$", _KEY_CTRL_E:
			colIndex = len(csvlines[rowIndex].Cell) - 1
		case "<":
			rowIndex = 0
		case ">":
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
			csvlines = slices.Insert(csvlines, rowIndex, csv.NewRow(mode))
			if err := repaint(); err != nil {
				return err
			}
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
				if err := repaint(); err != nil {
					return err
				}
				text, err := getline(out, "append cell>", "", makeCandidate(rowIndex+1, colIndex+1, csvlines))
				if err != nil {
					colIndex--
					break
				}
				csvlines[rowIndex].Replace(colIndex, text, mode)
			}
		case "r", "R", _KEY_F2:
			text, err := getline(out, "replace cell>", csvlines[rowIndex].Cell[colIndex].Text(), makeCandidate(rowIndex-1, colIndex, csvlines))
			if err != nil {
				break
			}
			csvlines[rowIndex].Replace(colIndex, text, mode)
		case "u":
			csvlines[rowIndex].Cell[colIndex].Undo(mode)
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
			if cursor.Quoted() {
				csvlines[rowIndex].Replace(colIndex, cursor.Text(), mode)
			} else {
				*cursor = cursor.Quote(mode)
			}
		case "w":
			if err := cmdWrite(csvlines, mode, tty1, out); err != nil {
				message = err.Error()
			}
			cache = map[int]string{}
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
		rewind()
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
