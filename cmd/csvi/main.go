package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"

	"github.com/hymkor/csvi"
	"github.com/hymkor/csvi/uncsv"
)

var (
	flagCellWidth = flag.Uint("w", 14, "set the width of cell")
	flagHeader    = flag.Uint("h", 1, "the number of row-header")
	flagTsv       = flag.Bool("t", false, "use TAB as field-separator")
	flagCsv       = flag.Bool("c", false, "use Comma as field-separator")
	flagSemicolon = flag.Bool("semicolon", false, "use Semicolon as field-separator")
	flagIana      = flag.String("iana", "", "IANA-registered-name to decode/encode NonUTF8 text(for example: Shift_JIS,EUC-JP... )")
	flagNonUTF8   = flag.Bool("nonutf8", false, "do not judge as utf8")
	flagHelp      = flag.Bool("help", false, "this help")
	flagAuto      = flag.String("auto", "", "autopilot")
	flag16le      = flag.Bool("16le", false, "Force read/write as UTF-16LE")
	flag16be      = flag.Bool("16be", false, "Force read/write as UTF-16BE")
	flagFixColumn = flag.Bool("fixcol", false, "Do not insert/delete a column")
)

const (
	_ANSI_CURSOR_OFF = "\x1B[?25l"
	_ANSI_CURSOR_ON  = "\x1B[?25h"
)

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

	_, err := csvi.Config{
		Pilot:       pilot,
		CellWidth:   int(*flagCellWidth),
		HeaderLines: int(*flagHeader),
		FixColumn:   *flagFixColumn,
	}.Main(mode, reader, out)

	return err
}

var version string

func main() {
	fmt.Fprintf(os.Stderr, "%s %s-%s-%s by %s\n",
		os.Args[0], version, runtime.GOOS, runtime.GOARCH, runtime.Version())

	flag.Parse()
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
