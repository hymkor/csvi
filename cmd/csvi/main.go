package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"

	"github.com/hymkor/csvi"
	"github.com/hymkor/csvi/uncsv"
)

var (
	flagCellWidth     = flag.String("w", "14", "set the `widths` of cells like '-w DefaultWidth,COL0:WIDTH0,COL1:WIDTH1,...'. COLn is the index starting from 0")
	flagHeader        = flag.Uint("h", 1, "the number of row-header")
	flagTsv           = flag.Bool("t", false, "use TAB as field-separator")
	flagCsv           = flag.Bool("c", false, "use Comma as field-separator")
	flagSemicolon     = flag.Bool("semicolon", false, "use Semicolon as field-separator")
	flagIana          = flag.String("iana", "", "IANA-registered-name to decode/encode NonUTF8 text(for example: Shift_JIS,EUC-JP... )")
	flagNonUTF8       = flag.Bool("nonutf8", false, "do not judge as utf8")
	flagHelp          = flag.Bool("help", false, "this help")
	flagAuto          = flag.String("auto", "", "autopilot")
	flag16le          = flag.Bool("16le", false, "Force read/write as UTF-16LE")
	flag16be          = flag.Bool("16be", false, "Force read/write as UTF-16BE")
	flagFixColumn     = flag.Bool("fixcol", false, "Do not insert/delete a column")
	flagReadOnly      = flag.Bool("readonly", false, "Read Only Mode")
	flagProtectHeader = flag.Bool("p", false, "Protect the header line")
)

const (
	_ANSI_CURSOR_OFF = "\x1B[?25l"
	_ANSI_CURSOR_ON  = "\x1B[?25h"
)

type CellWidth struct {
	Default int
	Option  map[int]int
}

func NewCellWidth() *CellWidth {
	cw := &CellWidth{
		Default: 14,
		Option:  map[int]int{},
	}
	return cw
}

func (cw *CellWidth) Get(n int) int {
	if val, ok := cw.Option[n]; ok {
		return val
	}
	return cw.Default
}

func (cw *CellWidth) Parse(s string) error {
	var p string
	cont := true
	for cont {
		p, s, cont = strings.Cut(s, ",")
		left, right, ok := strings.Cut(p, ":")
		if ok {
			leftN, err := strconv.ParseUint(left, 10, 64)
			if err != nil {
				return err
			}
			rightN, err := strconv.ParseUint(right, 10, 64)
			if err != nil {
				return err
			}
			cw.Option[int(leftN)] = int(rightN)
		} else {
			value, err := strconv.ParseUint(p, 10, 64)
			if err != nil {
				return err
			}
			cw.Default = int(value)
		}
	}
	return nil
}

func mains() error {
	if *flagHelp {
		flag.Usage()
		return nil
	}

	disable := colorable.EnableColorsStdout(nil)
	if disable != nil {
		defer disable()
	}

	var pilot csvi.Pilot
	if *flagAuto != "" {
		pilot = &_AutoPilot{script: *flagAuto}
		defer pilot.Close()
	}
	mode := &uncsv.Mode{}
	if *flagIana != "" {
		if err := mode.SetEncoding(*flagIana); err != nil {
			return fmt.Errorf("-iana %w", err)
		}
	}
	if *flagNonUTF8 {
		mode.NonUTF8 = true
	}
	if *flag16le {
		mode.SetUTF16LE()
	}
	if *flag16be {
		mode.SetUTF16BE()
	}

	var out io.Writer
	var reader io.Reader
	if len(flag.Args()) <= 0 {
		out = colorable.NewColorableStderr()
	} else {
		out = colorable.NewColorableStdout()
	}
	if len(flag.Args()) <= 0 && isatty.IsTerminal(uintptr(os.Stdin.Fd())) {
		// Start with one empty line
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
		if *flagSemicolon {
			mode.Comma = ';'
		}
		reader = multiFileReader(args...)
	}
	io.WriteString(out, _ANSI_CURSOR_OFF)
	defer io.WriteString(out, _ANSI_CURSOR_ON)

	cw := NewCellWidth()
	if err := cw.Parse(*flagCellWidth); err != nil {
		return err
	}
	_, err := csvi.Config{
		Mode:          mode,
		Pilot:         pilot,
		CellWidth:     cw.Get,
		HeaderLines:   int(*flagHeader),
		FixColumn:     *flagFixColumn,
		ReadOnly:      *flagReadOnly,
		ProtectHeader: *flagProtectHeader,
	}.Edit(reader, out)

	return err
}

var version string

func main() {
	fmt.Fprintf(os.Stderr, "%s %s-%s-%s built with %s\n",
		os.Args[0], version, runtime.GOOS, runtime.GOARCH, runtime.Version())

	flag.Parse()
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
