package csvi

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hymkor/csvi/uncsv"
)

var overWritten = map[string]struct{}{}

func dump(app *_Application, w io.Writer) {
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

func cmdWrite(app *_Application) (string, error) {
	fname := "-"
	var err error
	if args := flag.Args(); len(args) >= 1 {
		fname, err = filepath.Abs(args[0])
		if err != nil {
			return "", err
		}
	}
	fname, err = app.GetFilename(app, "write to>", fname)
	if err != nil {
		return "", err
	}
	if fname == "-" {
		dump(app, os.Stdout)
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
	dump(app, fd)
	if err := fd.Close(); err != nil {
		return "", err
	}
	return fmt.Sprintf("Saved as \"%s\"", fname), nil
}
