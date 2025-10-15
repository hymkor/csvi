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

func (f *Flag) mode() (*uncsv.Mode, error) {
	mode := &uncsv.Mode{}
	if f.Iana != "" {
		if err := mode.SetEncoding(f.Iana); err != nil {
			return nil, fmt.Errorf("-iana %w", err)
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
	}
	return mode, nil
}

func (f *Flag) setGlobalColor() {
	if f.ReverseVideo || csvi.IsRevertVideoWithEnv() {
		csvi.RevertColor()
	} else if noColor := os.Getenv("NO_COLOR"); len(noColor) > 0 {
		csvi.MonoChrome()
	}
}

func (f *Flag) dataSourceAndTtyOut() (io.Reader, io.Writer) {
	if len(f.flagSet.Args()) <= 0 {
		ttyOut := colorable.NewColorableStderr()
		if isatty.IsTerminal(os.Stdin.Fd()) {
			return nil, ttyOut
		}
		return os.Stdin, ttyOut
	}
	return multiFileReader(f.flagSet.Args()...),
		colorable.NewColorableStdout()
}

func (f *Flag) pilot() csvi.Pilot {
	if f.Auto == "" {
		return nil
	}
	return &autoPilot{script: f.Auto}
}

func (f *Flag) Run() error {
	if f.Help {
		f.flagSet.Usage()
		return nil
	}
	disable := colorable.EnableColorsStdout(nil)
	if disable != nil {
		defer disable()
	}
	f.setGlobalColor()

	dataSource, ttyOut := f.dataSourceAndTtyOut()

	io.WriteString(ttyOut, ansi.CURSOR_OFF)
	defer io.WriteString(ttyOut, ansi.CURSOR_ON)

	mode, err := f.mode()
	if err != nil {
		return err
	}

	cw := csvi.NewCellWidth()
	if err := cw.Parse(f.CellWidth); err != nil {
		return err
	}
	var titles []string
	if f.Title != "" {
		titles = []string{f.Title}
	}
	_, err = csvi.Config{
		Mode:          mode,
		Pilot:         f.pilot(),
		CellWidth:     cw,
		HeaderLines:   int(f.Header),
		FixColumn:     f.FixColumn,
		ReadOnly:      f.ReadOnly,
		ProtectHeader: f.ProtectHeader,
		Titles:        titles,
	}.Edit(dataSource, ttyOut)

	return err
}
