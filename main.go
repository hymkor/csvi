package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-runewidth"
	"github.com/mattn/go-tty"
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
	CURSOR_COLOR = "\x1B[0;40;37;1;7m"
	CELL1_COLOR  = "\x1B[0;44;37;1m"
	CELL2_COLOR  = "\x1B[0;40;37;1m"
	ERASE_LINE   = "\x1B[0m\x1B[0K"
)

type View struct {
	CSV       []string
	CellWidth int
	MaxInLine int
	CursorPos int
	Reverse   bool
	Out       io.Writer
}

var replacer = strings.NewReplacer("\n", "\u2936")

func (v View) Draw() {
	leftWidth := v.MaxInLine
	for i, s := range v.CSV {
		cw := v.CellWidth
		if cw > leftWidth {
			cw = leftWidth
		}
		s = replacer.Replace(s)
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
		v := View{
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
)

const (
	_KEY_CTRL_A = "\x01"
	_KEY_CTRL_B = "\x02"
	_KEY_CTRL_E = "\x05"
	_KEY_CTRL_F = "\x06"
	_KEY_CTRL_N = "\x0E"
	_KEY_CTRL_P = "\x10"
	_KEY_DOWN   = "\x1B[B"
	_KEY_ESC    = "\x1B"
	_KEY_LEFT   = "\x1B[D"
	_KEY_RIGHT  = "\x1B[C"
	_KEY_UP     = "\x1B[A"
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

var optionTsv = flag.Bool("t", false, "use TAB as field-seperator")

func main1() error {
	out := colorable.NewColorableStdout()

	io.WriteString(out, _ANSI_CURSOR_OFF)
	defer io.WriteString(out, _ANSI_CURSOR_ON)

	pin := getIn()
	defer pin.Close()

	in := csv.NewReader(pin)
	in.FieldsPerRecord = -1
	if *optionTsv {
		in.Comma = '\t'
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
	disposer := ignoreSigwinch(tty1)

	defer func() {
		tty1.Close()
		disposer()
	}()

	screenWidth, screenHeight, err := tty1.Size()
	if err != nil {
		return err
	}
	clean, err := tty1.Raw()
	if err != nil {
		return err
	}
	defer clean()

	colIndex := 0
	rowIndex := 0
	startRow := 0
	startCol := 0
	cols := (screenWidth - 1) / CELL_WIDTH

	for {
		window := &MemoryCsv{Data: csvlines, StartX: startCol, StartY: startRow}
		lf, err := view(window, colIndex-startCol, rowIndex-startRow, screenWidth-1, screenHeight-1, out)
		if err != nil {
			return err
		}
		fmt.Fprintln(out, "\r") // \r is for Linux & go-tty
		lf++
		if 0 <= rowIndex && rowIndex < len(csvlines) {
			if 0 <= colIndex && colIndex < len(csvlines[rowIndex]) {
				fmt.Fprintf(out, "\x1B[0;33;1m(%d,%d):%s\x1B[0m",
					rowIndex+1,
					colIndex+1,
					replacer.Replace(csvlines[rowIndex][colIndex]))
			}
		}
		fmt.Fprint(out, ERASE_LINE)
		ch, err := getKey(tty1)
		if err != nil {
			return err
		}
		switch ch {
		case "q", _KEY_ESC:
			fmt.Fprintln(out, "\r")
			return nil
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
