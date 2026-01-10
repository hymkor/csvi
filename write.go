package csvi

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"

	"github.com/hymkor/csvi/internal/ansi"
	"github.com/hymkor/csvi/uncsv"
)

var overWritten = map[string]struct{}{}

func (app *Application) dump(w io.Writer) {
	cursor := app.Front()
	app.Config.Mode.DumpBy(
		func() *uncsv.Row {
			if cursor == nil {
				return nil
			}
			row := cursor.Row
			row.MarkAsSave()
			cursor = cursor.Next()
			return row
		}, w)
}

var errCanceled = errors.New("canceled")

func (app *Application) cmdWrite(fname string) (string, error) {
	if fname == "-" {
		app.dump(os.Stdout)
		return "Output to STDOUT", nil
	}

	perm := os.FileMode(0666)
	openflag := os.O_WRONLY | os.O_EXCL | os.O_CREATE

	fd, err := os.OpenFile(fname, openflag, perm)
	if err == nil {
		app.dump(fd)
		if err := fd.Close(); err != nil {
			return "", err
		}
		return fmt.Sprintf("Saved as \"%s\"", fname), nil
	}
	if !errors.Is(err, os.ErrExist) {
		return "", err
	}
	stat, err := os.Stat(fname)
	if err != nil {
		return "", err
	}
	perm = stat.Mode().Perm()
	if _, done := overWritten[fname]; done || !stat.Mode().IsRegular() {
		openflag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	} else {
		if !app.YesNo("Overwrite as \"" + fname + "\" [y/n] ?") {
			return "", errCanceled
		}
		backup := fname + "~"
		if err := os.Remove(backup); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		if err := os.Rename(fname, backup); err != nil {
			return "", err
		}
		overWritten[fname] = struct{}{}
	}
	fd, err = os.OpenFile(fname, openflag, perm)
	if err != nil {
		return "", err
	}
	app.dump(fd)
	if err := fd.Close(); err != nil {
		return "", err
	}
	return fmt.Sprintf("Saved as \"%s\"", fname), nil
}

func (app *Application) cmdSave() (string, error) {
	var wg sync.WaitGroup
	chStop := make(chan struct{})
	defer close(chStop)

	if app.fetch != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-chStop:
					return
				default:
					row, err := app.fetch()
					if err != nil && !errors.Is(err, io.EOF) {
						return
					}
					if row != nil {
						app.Push(row)
					}
					if errors.Is(err, io.EOF) {
						return
					}
				}
			}
		}()
	}
	fname, err := app.GetFilename(app, "write to>", app.getSavePath())
	if err != nil {
		return "", err
	}
	io.WriteString(app.out, ansi.YELLOW+"\rw: Wait a moment for reading all data..."+ansi.ERASE_LINE)
	wg.Wait()
	message, err := app.cmdWrite(fname)
	if err == nil {
		app.ResetDirty()
	}
	app.lastSavePath = fname
	return message, err
}
