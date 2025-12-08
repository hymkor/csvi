package csvi

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/hymkor/csvi/internal/ansi"
	"github.com/hymkor/csvi/uncsv"
)

var overWritten = map[string]struct{}{}

func (app *_Application) dump(w io.Writer) {
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

func (app *_Application) getFname() (string, error) {
	fname := "-"
	if args := flag.Args(); len(args) >= 1 {
		var err error
		fname, err = filepath.Abs(args[0])
		if err != nil {
			return "", err
		}
	}
	return app.GetFilename(app, "write to>", fname)
}

func (app *_Application) cmdWrite(fname string) (string, error) {
	if fname == "-" {
		app.dump(os.Stdout)
		return "Output to STDOUT", nil
	}
	fd, err := os.OpenFile(fname, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0666)
	if os.IsExist(err) {
		if _, ok := overWritten[fname]; ok {
			os.Remove(fname)
		} else {
			if !app.YesNo("Overwrite as \"" + fname + "\" [y/n] ?") {
				return "", errCanceled
			}
			backupName := fname + "~"
			os.Remove(backupName)
			os.Rename(fname, backupName)
			overWritten[fname] = struct{}{}
		}
		fd, err = os.OpenFile(fname, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0666)
	}
	if err != nil {
		return "", err
	}
	app.dump(fd)
	if err := fd.Close(); err != nil {
		return "", err
	}
	return fmt.Sprintf("Saved as \"%s\"", fname), nil
}

func (app *_Application) trySave(fetch func() (bool, *uncsv.Row, error)) (string, error) {
	var wg sync.WaitGroup
	chStop := make(chan struct{})
	defer close(chStop)

	if fetch != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-chStop:
					return
				default:
					ok, row, err := fetch()
					if !ok {
						return
					}
					if err != nil && !errors.Is(err, io.EOF) {
						return
					}
					app.Push(row)
					if errors.Is(err, io.EOF) {
						return
					}
				}
			}
		}()
	}
	fname, err := app.getFname()
	if err != nil {
		return "", err
	}
	io.WriteString(app.out, ansi.YELLOW+"\rw: Wait a moment for reading all data..."+ansi.ERASE_LINE)
	wg.Wait()
	message, err := app.cmdWrite(fname)
	if err == nil {
		app.ResetDirty()
	}
	return message, err
}
