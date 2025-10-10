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
	"github.com/mattn/go-runewidth"

	"github.com/hymkor/csvi"
	"github.com/hymkor/csvi/legacy"
	"github.com/hymkor/csvi/uncsv"

	"github.com/hymkor/csvi/internal/ansi"
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
	flagTitle         = flag.String("title", "", "Set title string")
	flagReverseVideo  = flag.Bool("rv", false, "Enable reverse-video display (invert foreground and background colors)")
	flagAmbWide       = flag.Bool("aw", false, "Bypass width detection; assume ambiguous-width chars are wide (2 cells)")
	flagAmbNallow     = flag.Bool("an", false, "Bypass width detection; assume ambiguous-width chars are narrow (1 cell)")
	flagDebugBell     = flag.Bool("debug-bell", false, "Enable Debug Bell")
)

func isRevertVideoWithEnv() bool {
	colorFgBg, ok := os.LookupEnv("COLORFGBG")
	if !ok {
		return false
	}
	fgStr, bgStr, ok := strings.Cut(colorFgBg, ";")
	if !ok {
		return false
	}
	fgInt, err := strconv.ParseInt(fgStr, 10, 64)
	if err != nil {
		return false
	}
	bgInt, err := strconv.ParseInt(bgStr, 10, 64)
	if err != nil {
		return false
	}
	return fgInt < bgInt
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
	if *flagReverseVideo || isRevertVideoWithEnv() {
		csvi.RevertColor()
	} else if noColor := os.Getenv("NO_COLOR"); len(noColor) > 0 {
		csvi.MonoChrome()
	}
	if *flagDebugBell {
		csvi.EnableDebugBell(os.Stderr)
	}
	var pilot csvi.Pilot
	if *flagAuto != "" {
		pilot = &_AutoPilot{script: *flagAuto}
		defer pilot.Close()
	} else if *flagAmbWide || *flagAmbNallow {
		runewidth.DefaultCondition.EastAsianWidth = *flagAmbWide
		var err error
		pilot, err = legacy.New()
		if err != nil {
			return err
		}
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
	io.WriteString(out, ansi.CURSOR_OFF)
	defer io.WriteString(out, ansi.CURSOR_ON)

	cw := csvi.NewCellWidth()
	if err := cw.Parse(*flagCellWidth); err != nil {
		return err
	}
	var titles []string
	if *flagTitle != "" {
		titles = []string{*flagTitle}
	}
	_, err := csvi.Config{
		Mode:          mode,
		Pilot:         pilot,
		CellWidth:     cw,
		HeaderLines:   int(*flagHeader),
		FixColumn:     *flagFixColumn,
		ReadOnly:      *flagReadOnly,
		ProtectHeader: *flagProtectHeader,
		Titles:        titles,
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
