package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/mattn/go-tty"
	"github.com/nyaosorg/go-windows-mbcs"

	"github.com/hymkor/csview/unbreakable-csv"
)

type WriteNopCloser struct {
	io.Writer
}

func (*WriteNopCloser) Close() error {
	return nil
}

func writeCsvTo(csvlines []csv.Row, mode *csv.Mode, codeFlag _CodeFlag, fd io.Writer) {
	if (codeFlag & isAnsi) != 0 {
		pipeIn, pipeOut := io.Pipe()
		go func() {
			mode.Dump(csvlines, pipeOut)
			pipeOut.Close()
		}()
		sc := bufio.NewScanner(pipeIn)
		bw := bufio.NewWriter(fd)
		for sc.Scan() {
			bytes, _ := mbcs.Utf8ToAnsi(sc.Text(), mbcs.ACP)
			bw.Write(bytes)
			bw.WriteByte('\r')
			bw.WriteByte('\n')
		}
		bw.Flush()
	} else {
		if (codeFlag & hasBom) != 0 {
			io.WriteString(fd, "\uFEFF")
		}
		mode.Dump(csvlines, fd)
	}
}

func cmdWrite(csvlines []csv.Row, mode *csv.Mode, codeFlag _CodeFlag, tty1 *tty.TTY, out io.Writer) (message string) {
	fname := "-"
	var err error
	if args := flag.Args(); len(args) >= 1 {
		fname, err = filepath.Abs(args[0])
		if err != nil {
			return err.Error()
		}
	}
	fname, err = getline(out, "write to>", fname, nil)
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

	writeCsvTo(csvlines, mode, codeFlag, fd)
	fd.Close()
	return
}
