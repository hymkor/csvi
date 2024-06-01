package csvi

import (
	"flag"
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
			cursor = cursor.Next()
			return row
		}, w)
}

func cmdWrite(app *_Application) error {
	fname := "-"
	var err error
	if args := flag.Args(); len(args) >= 1 {
		fname, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}
	fname, err = app.GetFilename(app, "write to>", fname)
	if err != nil {
		return nil
	}
	if fname == "-" {
		dump(app, os.Stdout)
		return nil
	}
	fd, err := os.OpenFile(fname, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0666)
	if os.IsExist(err) {
		if _, ok := overWritten[fname]; ok {
			os.Remove(fname)
		} else {
			if !app.YesNo("Overwrite as \"" + fname + "\" [y/n] ?") {
				return nil
			}
			backupName := fname + "~"
			os.Remove(backupName)
			os.Rename(fname, backupName)
			overWritten[fname] = struct{}{}
		}
		fd, err = os.OpenFile(fname, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0666)
	}
	if err != nil {
		return err
	}
	dump(app, fd)
	if err := fd.Close(); err != nil {
		return err
	}
	return nil
}
