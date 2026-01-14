package csviapp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"

	"github.com/hymkor/csvi"
	"github.com/hymkor/csvi/uncsv"

	"github.com/hymkor/csvi/internal/ansi"
)

var errNotSingleChar = errors.New("delimiter must be a single character")

func (f *Options) mode() (*uncsv.Mode, error) {
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
		switch len(f.Delimiter) {
		case 0:
		case 1:
			mode.Comma = f.Delimiter[0]
		default:
			return nil, errNotSingleChar
		}
	}
	return mode, nil
}

func (f *Options) setGlobalColor() {
	if f.ReverseVideo || csvi.IsRevertVideoWithEnv() {
		csvi.RevertColor()
	} else if noColor := os.Getenv("NO_COLOR"); len(noColor) > 0 {
		csvi.MonoChrome()
	}
}

func (f *Options) dataSourceAndTtyOut() (io.Reader, io.Writer) {
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

func (f *Options) pilot() csvi.Pilot {
	if f.Auto == "" {
		return nil
	}
	return &autoPilot{script: f.Auto}
}

func (f *Options) Run() error {
	if f.Help {
		f.flagSet.Usage()
		return nil
	}
	disable := colorable.EnableColorsStdout(nil)
	if disable != nil {
		defer disable()
	}
	f.setGlobalColor()

	return f.RunInOut(f.dataSourceAndTtyOut())
}

func (f *Options) callExtEditor(text string, app *csvi.Application) (string, error) {
	fd, err := os.CreateTemp("", "csvi*.txt")
	if err != nil {
		return text, err
	}
	fname := fd.Name()
	_, err1 := io.WriteString(fd, text)
	err2 := fd.Close()
	if err1 != nil {
		return text, err1
	}
	if err2 != nil {
		return text, err2
	}
	defer func() {
		os.Remove(fname)
	}()
	fmt.Fprintf(app, "%s\r%v %v%s", ansi.YELLOW, f.ExtEditor, fname, ansi.ERASE_LINE)
	cmd := exec.Command(f.ExtEditor, fname)
	if err := cmd.Run(); err != nil {
		return text, err
	}
	output, err := os.ReadFile(fname)
	if err != nil {
		return text, err
	}
	return string(output), nil
}

func (f *Options) RunInOut(dataSource io.Reader, ttyOut io.Writer) error {
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

	var extEditor func(string, *csvi.Application) (string, error)
	if f.ExtEditor != "" {
		extEditor = f.callExtEditor
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
		OutputSep:     f.OutputSep,
		SavePath:      f.SavePath,
		ExtEditor:     extEditor,
	}.Edit(dataSource, ttyOut)

	return err
}
