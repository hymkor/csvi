package startup

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"

	"github.com/hymkor/csvi"
	"github.com/hymkor/csvi/uncsv"

	"github.com/hymkor/csvi/internal/ansi"
)

func (f *Flag) Run() error {
	if f.Help {
		f.flagSet.Usage()
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
	if len(f.flagSet.Args()) <= 0 {
		out = colorable.NewColorableStderr()
	} else {
		out = colorable.NewColorableStdout()
	}
	if len(f.flagSet.Args()) <= 0 && isatty.IsTerminal(uintptr(os.Stdin.Fd())) {
		// Start with one empty line
		mode.Comma = '\t'
	} else {
		mode.Comma = ','
		args := f.flagSet.Args()
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
