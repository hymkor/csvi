package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/mattn/go-tty"

	"github.com/hymkor/csview/unbreakable-csv"
)

type WriteNopCloser struct {
	io.Writer
}

func (*WriteNopCloser) Close() error {
	return nil
}

var overWritten = map[string]struct{}{}

func cmdWrite(csvlines []csv.Row, mode *csv.Mode, tty1 *tty.TTY, out io.Writer) (message string) {
	fname := "-"
	var err error
	if args := flag.Args(); len(args) >= 1 {
		fname, err = filepath.Abs(args[0])
		if err != nil {
			return err.Error()
		}
	}
	fname, err = getfilename(out, "write to>", fname)
	if err != nil {
		return ""
	}
	var fd io.WriteCloser
	if fname == "-" {
		fd = &WriteNopCloser{Writer: os.Stdout}
	} else {
		fd, err = os.OpenFile(fname, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0666)
		if os.IsExist(err) {
			if _, ok := overWritten[fname]; ok {
				os.Remove(fname)
			} else {
				if !yesNo(tty1, out, "Overwrite as \""+fname+"\" [y/n] ?") {
					return ""
				}
				backupName := fname + "~"
				os.Remove(backupName)
				os.Rename(fname, backupName)
				overWritten[fname] = struct{}{}
			}
			fd, err = os.OpenFile(fname, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0666)
		}
		if err != nil {
			message = err.Error()
		}
	}

	mode.Dump(csvlines, fd)
	fd.Close()
	return
}
