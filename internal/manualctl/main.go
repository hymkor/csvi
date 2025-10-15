package manualctl

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/mattn/go-tty"

	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/completion"
	"github.com/nyaosorg/go-readline-ny/keys"
	"github.com/nyaosorg/go-readline-skk"

	"github.com/hymkor/csvi/candidate"

	"github.com/hymkor/csvi/internal/ansi"
)

type ManualCtl struct {
	*tty.TTY
}

func New() (ManualCtl, error) {
	var rc ManualCtl
	var err error

	rc.TTY, err = tty.Open()
	return rc, err
}

func (m ManualCtl) GetKey() (string, error) {
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

func (m ManualCtl) ReadLine(out io.Writer, prompt, defaultStr string, c candidate.Candidate) (string, error) {
	skkInit()
	editor := &readline.Editor{
		Writer:  out,
		Default: defaultStr,
		History: c,
		Cursor:  65535,
		PromptWriter: func(w io.Writer) (int, error) {
			return fmt.Fprintf(w, "\r%s%s%s", ansi.YELLOW, prompt, ansi.ERASE_LINE)
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

	defer io.WriteString(out, ansi.CURSOR_OFF)
	editor.BindKey(keys.Escape, readline.CmdInterrupt)
	return editor.ReadLine(context.Background())
}

func (m ManualCtl) GetFilename(out io.Writer, prompt, defaultStr string) (string, error) {
	skkInit()
	editor := &readline.Editor{
		Writer:  out,
		Default: defaultStr,
		Cursor:  len(defaultStr) - len(filepath.Ext(defaultStr)),
		PromptWriter: func(w io.Writer) (int, error) {
			return fmt.Fprintf(w, "\r\x1B[0;33;40;1m%s%s", prompt, ansi.ERASE_LINE)
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

	defer io.WriteString(out, ansi.CURSOR_OFF)
	editor.BindKey(keys.Escape, readline.CmdInterrupt)
	return editor.ReadLine(context.Background())
}
