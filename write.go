package main

import (
	"container/list"
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/hymkor/csvi/uncsv"
)

var overWritten = map[string]struct{}{}

func dump(csvlines *list.List, mode *uncsv.Mode, w io.Writer) {
	cursor := frontPtr(csvlines)
	mode.DumpBy(
		func() *uncsv.Row {
			if cursor == nil {
				return nil
			}
			row := cursor.Row
			cursor = cursor.Next()
			return row
		}, w)
}

func cmdWrite(pilot Pilot, csvlines *list.List, mode *uncsv.Mode, out io.Writer) error {
	fname := "-"
	var err error
	if args := flag.Args(); len(args) >= 1 {
		fname, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}
	fname, err = pilot.GetFilename(out, "write to>", fname)
	if err != nil {
		return nil
	}
	if fname == "-" {
		dump(csvlines, mode, os.Stdout)
		return nil
	}
	fd, err := os.OpenFile(fname, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0666)
	if os.IsExist(err) {
		if _, ok := overWritten[fname]; ok {
			os.Remove(fname)
		} else {
			if !yesNo(pilot, out, "Overwrite as \""+fname+"\" [y/n] ?") {
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
	dump(csvlines, mode, fd)
	if err := fd.Close(); err != nil {
		return err
	}
	return nil
}
