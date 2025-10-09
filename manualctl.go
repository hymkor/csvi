package csvi

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/mattn/go-runewidth"
	"github.com/mattn/go-tty"

	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/completion"
	"github.com/nyaosorg/go-readline-ny/keys"
	"github.com/nyaosorg/go-readline-skk"

	"github.com/hymkor/go-cursorposition"
)

type _ManualCtl struct {
	*tty.TTY
}

func newManualCtl() (_ManualCtl, error) {
	var rc _ManualCtl
	var err error

	rc.TTY, err = tty.Open()
	return rc, err
}

func (m _ManualCtl) Calibrate() error {
	// Measure how far the cursor moves while the `▽` is printed
	w, err := cursorposition.AmbiguousWidthGoTty(m.TTY, os.Stderr)
	if err != nil {
		return err
	}
	runewidth.DefaultCondition.EastAsianWidth = w >= 2
	return nil
}

func (m _ManualCtl) Close() error {
	return m.TTY.Close()
}

func (m _ManualCtl) Size() (int, int, error) {
	return m.TTY.Size()
}

func (m _ManualCtl) GetKey() (string, error) {
	return readline.GetKey(m.TTY)
}

var predictColor = [...]string{"\x1B[3;22;34m", "\x1B[23;39m"}

var skkInitOnce sync.Once

func skkInit() {
	skkInitOnce.Do(func() {
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
}

func (m _ManualCtl) ReadLine(out io.Writer, prompt, defaultStr string, c Candidate) (string, error) {
	skkInit()
	editor := &readline.Editor{
		Writer:  out,
		Default: defaultStr,
		History: c,
		Cursor:  65535,
		PromptWriter: func(w io.Writer) (int, error) {
			return fmt.Fprintf(w, "\r%s%s%s", _ANSI_YELLOW, prompt, _ANSI_ERASE_LINE)
		},
		LineFeedWriter: func(readline.Result, io.Writer) (int, error) {
			return 0, nil
		},
		Coloring:     &skk.Coloring{},
		PredictColor: predictColor,
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

func (m _ManualCtl) GetFilename(out io.Writer, prompt, defaultStr string) (string, error) {
	skkInit()
	editor := &readline.Editor{
		Writer:  out,
		Default: defaultStr,
		Cursor:  len(defaultStr) - len(filepath.Ext(defaultStr)),
		PromptWriter: func(w io.Writer) (int, error) {
			return fmt.Fprintf(w, "\r\x1B[0;33;40;1m%s%s", prompt, _ANSI_ERASE_LINE)
		},
		LineFeedWriter: func(readline.Result, io.Writer) (int, error) {
			return 0, nil
		},
		Coloring:     &skk.Coloring{},
		PredictColor: predictColor,
	}
	editor.BindKey(keys.CtrlI, completion.CmdCompletionOrList{
		Completion: completion.File{},
	})

	defer io.WriteString(out, _ANSI_CURSOR_OFF)
	editor.BindKey(keys.Escape, readline.CmdInterrupt)
	return editor.ReadLine(context.Background())
}
