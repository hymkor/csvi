package startup

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"

	"github.com/hymkor/struct2flag"

	"github.com/hymkor/csvi"

	"github.com/hymkor/csvi/uncsv"

	"github.com/hymkor/csvi/internal/ansi"
)

type Flag struct {
	CellWidth     string `flag:"w,set the \x60widths\x60 of cells like '-w DefaultWidth,COL0:WIDTH0,COL1:WIDTH1,...'. COLn is the index starting from 0"`
	Header        uint   `flag:"h,the number of row-header"`
	Tsv           bool   `flag:"t,use TAB as field-separator"`
	Csv           bool   `flag:"c,use Comma as field-separator"`
	Semicolon     bool   `flag:"semicolon,use Semicolon as field-separator"`
	Iana          string `flag:"iana,IANA-registered-name to decode/encode NonUTF8 text(for example: Shift_JIS,EUC-JP... )"`
	NonUTF8       bool   `flag:"nonutf8,do not judge as utf8"`
	Help          bool   `flag:"help,this help"`
	Auto          string `flag:"auto,autopilot"`
	Utf16le       bool   `flag:"16le,Force read/write as UTF-16LE"`
	Utf16be       bool   `flag:"16be,Force read/write as UTF-16BE"`
	FixColumn     bool   `flag:"fixcol,Do not insert/delete a column"`
	ReadOnly      bool   `flag:"readonly,Read Only Mode"`
	ProtectHeader bool   `flag:"p,Protect the header line"`
	Title         string `flag:"title,Set title string"`
	ReverseVideo  bool   `flag:"rv,Enable reverse-video display (invert foreground and background colors)"`
	DebugBell     bool   `flag:"debug-bell,Enable Debug Bell"`
}

func NewFlag() *Flag {
	return &Flag{
		CellWidth: "14",
		Header:    1,
	}
}

func (this *Flag) Bind(fs *flag.FlagSet) *Flag {
	struct2flag.Bind(fs, this)
	return this
}

func (f *Flag) Run() error {
	if f.Help {
		flag.Usage()
		return nil
	}
	disable := colorable.EnableColorsStdout(nil)
	if disable != nil {
		defer disable()
	}
	if f.ReverseVideo || csvi.IsRevertVideoWithEnv() {
		csvi.RevertColor()
	} else if noColor := os.Getenv("NO_COLOR"); len(noColor) > 0 {
		csvi.MonoChrome()
	}
	if f.DebugBell {
		csvi.EnableDebugBell(os.Stderr)
	}
	var pilot csvi.Pilot
	if f.Auto != "" {
		pilot = &_AutoPilot{script: f.Auto}
		defer pilot.Close()
	}

	mode := &uncsv.Mode{}
	if f.Iana != "" {
		if err := mode.SetEncoding(f.Iana); err != nil {
			return fmt.Errorf("-iana %w", err)
		}
	}
	if f.NonUTF8 {
		mode.NonUTF8 = true
	}
	if f.Utf16le {
		mode.SetUTF16LE()
	}
	if f.Utf16be {
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
		if f.Tsv {
			mode.Comma = '\t'
		}
		if f.Csv {
			mode.Comma = ','
		}
		if f.Semicolon {
			mode.Comma = ';'
		}
		reader = multiFileReader(args...)
	}
	io.WriteString(out, ansi.CURSOR_OFF)
	defer io.WriteString(out, ansi.CURSOR_ON)

	cw := csvi.NewCellWidth()
	if err := cw.Parse(f.CellWidth); err != nil {
		return err
	}
	var titles []string
	if f.Title != "" {
		titles = []string{f.Title}
	}
	_, err := csvi.Config{
		Mode:          mode,
		Pilot:         pilot,
		CellWidth:     cw,
		HeaderLines:   int(f.Header),
		FixColumn:     f.FixColumn,
		ReadOnly:      f.ReadOnly,
		ProtectHeader: f.ProtectHeader,
		Titles:        titles,
	}.Edit(reader, out)

	return err
}

func Run() error {
	f := NewFlag().Bind(flag.CommandLine)
	flag.Parse()
	return f.Run()
}
