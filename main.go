package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-runewidth"
	"github.com/mattn/go-tty"

	"github.com/zetamatta/go-readline-ny"
)

func cutStrInWidth(s string, cellwidth int) (string, int) {
	w := 0
	for n, c := range s {
		w1 := runewidth.RuneWidth(c)
		if w+w1 > cellwidth {
			return s[:n], w
		}
		w += w1
	}
	return s, w
}

const (
	CURSOR_COLOR     = "\x1B[0;40;37;1;7m"
	CELL1_COLOR      = "\x1B[0;44;37;1m"
	CELL2_COLOR      = "\x1B[0;40;37;1m"
	ERASE_LINE       = "\x1B[0m\x1B[0K"
	ERASE_SCRN_AFTER = "\x1B[0m\x1B[0J"
)

type LineView struct {
	CSV       []string
	CellWidth int
	MaxInLine int
	CursorPos int
	Reverse   bool
	Out       io.Writer
}

var replaceTable = strings.NewReplacer(
	"\r", "\u240A",
	"\x1B", "\u241B",
	"\n", "\u2936", // arrow pointing downwards then curving leftwards
	"\t", "\u21E5") // rightwards arrow to bar (rightward tab)

// See. en.wikipedia.org/wiki/Unicode_control_characters#Control_pictures

func (v LineView) Draw() {
	leftWidth := v.MaxInLine
	for i, s := range v.CSV {
		cw := v.CellWidth
		if cw > leftWidth {
			cw = leftWidth
		}
		s = replaceTable.Replace(s)
		ss, w := cutStrInWidth(s, cw)
		if i == v.CursorPos {
			io.WriteString(v.Out, CURSOR_COLOR)
		} else if ((i & 1) == 0) == v.Reverse {
			io.WriteString(v.Out, CELL1_COLOR)
		} else {
			io.WriteString(v.Out, CELL2_COLOR)
		}
		io.WriteString(v.Out, ss)
		leftWidth -= w
		for i := cw - w; i > 0; i-- {
			v.Out.Write([]byte{' '})
			leftWidth--
		}
		if leftWidth <= 0 {
			break
		}
	}
	io.WriteString(v.Out, ERASE_LINE)
}

type CsvIn interface {
	Read() ([]string, error)
}

var cache = map[int]string{}

const CELL_WIDTH = 12

func view(in CsvIn, csrpos, csrlin, w, h int, out io.Writer) (int, error) {
	reverse := false
	count := 0
	lfCount := 0
	for {
		if count >= h {
			return lfCount, nil
		}
		record, err := in.Read()
		if err == io.EOF {
			return lfCount, nil
		}
		if err != nil {
			return lfCount, err
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
}

type MemoryCsv struct {
	Data   [][]string
	StartX int
	StartY int
}

func (this *MemoryCsv) Read() ([]string, error) {
	if this.StartY >= len(this.Data) {
		return nil, io.EOF
	}
	csv := this.Data[this.StartY]
	if this.StartX <= len(csv) {
		csv = csv[this.StartX:]
	} else {
		csv = []string{}
	}
	this.StartY++
	return csv, nil
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

func cat(in io.Reader, out io.Writer) {
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		fmt.Fprintln(out, textfilter(sc.Text()))
	}
}

func getIn() io.ReadCloser {
	pin, pout := io.Pipe()
	go func() {
		args := flag.Args()
		if len(args) <= 0 {
			cat(os.Stdin, pout)
		} else {
			for _, arg1 := range args {
				in, err := os.Open(arg1)
				if err != nil {
					fmt.Fprintf(pout, "\"%s\",\"not found\"\n", arg1)
					continue
				}
				cat(in, pout)
				in.Close()
			}
		}
		pout.Close()
	}()
	return pin
}

var optionTsv = flag.Bool("t", false, "use TAB as field-separator")
var optionCsv = flag.Bool("c", false, "use Comma as field-separator")

func searchForward(csvlines [][]string, r, c int, target string) (bool, int, int) {
	c++
	for r < len(csvlines) {
		for c < len(csvlines[r]) {
			if strings.Contains(csvlines[r][c], target) {
				return true, r, c
			}
			c++
		}
		r++
		c = 0
	}
	return false, r, c
}

func searchBackward(csvlines [][]string, r, c int, target string) (bool, int, int) {
	c--
	for {
		for c >= 0 {
			if strings.Contains(csvlines[r][c], target) {
				return true, r, c
			}
			c--
		}
		r--
		if r < 0 {
			return false, r, c
		}
		c = len(csvlines[r]) - 1
	}
}

func getline(out io.Writer, prompt string, defaultStr string) (string, error) {
	editor := readline.Editor{
		Writer:  out,
		Default: defaultStr,
		Cursor:  65535,
		Prompt: func() (int, error) {
			fmt.Fprintf(out, "\r\x1B[0;33;40;1m%s%s", prompt, ERASE_LINE)
			return 2, nil
		},
		LineFeed: func(readline.Result) {},
	}
	defer io.WriteString(out, _ANSI_CURSOR_OFF)
	editor.BindKeySymbol(readline.K_ESCAPE, readline.F_INTR)
	return editor.ReadLine(context.Background())
}

type WriteNopCloser struct {
	io.Writer
}

func (*WriteNopCloser) Close() error {
	return nil
}

func main1() error {
	out := colorable.NewColorableStderr()

	io.WriteString(out, _ANSI_CURSOR_OFF)
	defer io.WriteString(out, _ANSI_CURSOR_ON)

	pin := getIn()
	defer pin.Close()

	in := csv.NewReader(pin)
	in.FieldsPerRecord = -1
	args := flag.Args()
	if len(args) >= 1 && !strings.HasSuffix(strings.ToLower(args[0]), ".csv") {
		in.Comma = '\t'
	}
	if *optionTsv {
		in.Comma = '\t'
	}
	if *optionCsv {
		in.Comma = ','
	}

	csvlines := [][]string{}
	for {
		csv1, err := in.Read()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		csvlines = append(csvlines, csv1)
	}
	tty1, err := tty.Open()
	if err != nil {
		return err
	}

	defer tty1.Close()

	colIndex := 0
	rowIndex := 0
	startRow := 0
	startCol := 0

	lastSearch := searchForward
	lastSearchRev := searchBackward
	lastWord := ""
	var lastWidth, lastHeight int

	message := ""
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

		window := &MemoryCsv{Data: csvlines, StartX: startCol, StartY: startRow}
		lf, err := view(window, colIndex-startCol, rowIndex-startRow, screenWidth-1, screenHeight-1, out)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, "\r") // \r is for Linux & go-tty
		lf++
		if message != "" {
			io.WriteString(out, _ANSI_YELLOW)
			io.WriteString(out, runewidth.Truncate(message, screenWidth-1, ""))
			io.WriteString(out, _ANSI_RESET)
			message = ""
		} else if 0 <= rowIndex && rowIndex < len(csvlines) {
			if 0 <= colIndex && colIndex < len(csvlines[rowIndex]) {
				fmt.Fprintf(out, "\x1B[0;33;1m(%d,%d):%s\x1B[0m",
					rowIndex+1,
					colIndex+1,
					runewidth.Truncate(replaceTable.Replace(csvlines[rowIndex][colIndex]), screenWidth-11, "..."))
			}
		}
		fmt.Fprint(out, ERASE_SCRN_AFTER)
		ch, err := getKey(tty1)
		if err != nil {
			return err
		}
		switch ch {
		case _KEY_CTRL_L:
			cache = map[int]string{}
		case "q", _KEY_ESC:
			io.WriteString(out, _ANSI_YELLOW+"\rQuit Sure ? [y/n]"+ERASE_LINE)
			if ch, err := getKey(tty1); err == nil && ch == "y" {
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
			colIndex = len(csvlines[rowIndex]) - 1
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
			lastWord, err = getline(out, ch, "")
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
			csvlines = append(csvlines, []string{})
			if len(csvlines) >= rowIndex+1 {
				copy(csvlines[rowIndex+1:], csvlines[rowIndex:])
				csvlines[rowIndex] = []string{""}
			}
			break
		case "D":
			if len(csvlines) <= 1 {
				break
			}
			copy(csvlines[rowIndex:], csvlines[rowIndex+1:])
			csvlines = csvlines[:len(csvlines)-1]
			if rowIndex >= len(csvlines) {
				rowIndex--
			}
			break
		case "i":
			text, err := getline(out, "insert cell>", "")
			if err != nil {
				break
			}
			csvlines[rowIndex] = append(csvlines[rowIndex], "")
			copy(csvlines[rowIndex][colIndex+1:], csvlines[rowIndex][colIndex:])
			csvlines[rowIndex][colIndex] = text
			colIndex++
		case "a":
			text, err := getline(out, "append cell>", "")
			if err != nil {
				break
			}
			csvlines[rowIndex] = append(csvlines[rowIndex], "")
			colIndex++
			copy(csvlines[rowIndex][colIndex+1:], csvlines[rowIndex][colIndex:])
			csvlines[rowIndex][colIndex] = text
		case "r", "R", _KEY_F2:
			text, err := getline(out, "replace cell>", csvlines[rowIndex][colIndex])
			if err != nil {
				break
			}
			csvlines[rowIndex][colIndex] = text
		case "d":
			if len(csvlines[rowIndex]) <= 1 {
				csvlines[rowIndex][0] = ""
			} else {
				copy(csvlines[rowIndex][colIndex:], csvlines[rowIndex][colIndex+1:])
				csvlines[rowIndex] = csvlines[rowIndex][:len(csvlines[rowIndex])-1]
			}
		case "w":
			fname := "-"
			var err error
			if len(args) >= 1 {
				fname, err = filepath.Abs(args[0])
				if err != nil {
					message = err.Error()
					break
				}
			}
			fname, err = getline(out, "write to>", fname)
			if err != nil {
				break
			}
			var fd io.WriteCloser
			if fname == "-" {
				fd = &WriteNopCloser{Writer: os.Stdout}
			} else {
				fd, err = os.OpenFile(fname, os.O_EXCL|os.O_CREATE, 0666)
				if err != nil {
					message = err.Error()
					break
				}
			}
			w := csv.NewWriter(fd)
			w.Comma = in.Comma
			w.UseCRLF = true
			w.WriteAll(csvlines)
			w.Flush()
			fd.Close()
		}
		if colIndex >= len(csvlines[rowIndex]) {
			colIndex = len(csvlines[rowIndex]) - 1
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

		if lf > 0 {
			fmt.Fprintf(out, "\r\x1B[%dA", lf)
		} else {
			fmt.Fprint(out, "\r")
		}
	}
}

func main() {
	flag.Parse()
	if err := main1(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}
