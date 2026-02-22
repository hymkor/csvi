package safewrite

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	overwritten          = make(map[string]struct{})
	ErrOverWriteRejected = errors.New("overwrite rejected")
)

func Open(
	name string,
	confirmOverwrite func() bool,
) (io.WriteCloser, error) {

	info, err := os.Stat(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fd, err := os.Create(name)
			if err != nil {
				err = fmt.Errorf("create %q: %w", name, err)
			}
			return fd, err
		}
		err = fmt.Errorf("stat %q: %w", name, err)
		return nil, err
	}

	if info.Mode()&os.ModeDevice != 0 {
		fd, err := os.OpenFile(name, os.O_WRONLY, 0666)
		if err != nil {
			err = fmt.Errorf("OpenFile %q: %w", name, err)
		}
		return fd, err
	}

	if !confirmOverwrite() {
		return nil, ErrOverWriteRejected
	}

	dir := filepath.Dir(name)
	base := filepath.Base(name)

	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return nil, err
	}

	return &replaceWriter{
		File:   tmp,
		target: name,
		tmp:    tmp.Name(),
	}, nil
}

type replaceWriter struct {
	*os.File
	target string
	tmp    string
}

func (w *replaceWriter) Close() error {
	if err := w.File.Close(); err != nil {
		return err
	}
	backup := w.target + "~"
	if _, ok := overwritten[w.target]; !ok {
		overwritten[w.target] = struct{}{}
		_ = os.Remove(backup)
		if err := os.Rename(w.target, backup); err != nil {
			return err
		}
	}
	return os.Rename(w.tmp, w.target)
}
