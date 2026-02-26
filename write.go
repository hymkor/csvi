package csvi

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/nyaosorg/go-inline-animation"

	"github.com/hymkor/go-safewrite"

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
		end := animation.Dots.Progress(app.out)
		defer end()
		app.dump(os.Stdout)
		return "Output to STDOUT", nil
	}

	prompt := func(info *safewrite.Info) bool {
		if info.Status != safewrite.NONE {
			return true
		}
		if info.ReadOnly() {
			if app.yesNo("Overwrite READONLY file \"" + info.Name + "\" [y/n] ?") {
				return true
			}
			return false
		}
		return app.yesNo("Overwrite as \"" + info.Name + "\" [y/n] ?")
	}
	fd, err := safewrite.Open(fname, prompt)
	if err != nil {
		return "", err
	}
	end := animation.Dots.Progress(app.out)
	app.dump(fd)
	end()
	if err := fd.Close(); err != nil {
		var be *safewrite.BackupError
		if errors.As(err, &be) {
			return "",
				fmt.Errorf("Failed to backup %q to %q (tmp: %q)",
					filepath.Base(be.Target),
					filepath.Base(be.Backup),
					filepath.Base(be.Tmp))
		}
		var re *safewrite.ReplaceError
		if errors.As(err, &re) {
			return "",
				fmt.Errorf("Failed to replace %q to %q",
					filepath.Base(re.Tmp),
					filepath.Base(re.Target))
		}
		return "", err
	}
	app.registerOnClose(fname, func() {
		safewrite.RestorePerm(fd)
	})
	return fmt.Sprintf("Saved as \"%s\"", fname), nil
}

func (app *Application) cmdSave() (string, error) {
	var wg sync.WaitGroup
	chStop := make(chan struct{})
	defer close(chStop)

	if app.fetchFunc != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-chStop:
					return
				default:
					row, err := app.fetchFunc()
					if err != nil && !errors.Is(err, io.EOF) {
						app.fetchFunc = nil
						app.tryFetchFunc = nil
						return
					}
					if row != nil {
						app.push(row)
					}
					if errors.Is(err, io.EOF) {
						app.fetchFunc = nil
						app.tryFetchFunc = nil
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
	end := animation.Dots.Progress(app.out)
	wg.Wait()
	end()
	message, err := app.cmdWrite(fname)
	if err == nil {
		app.resetDirty()
	}
	app.lastSavePath = fname
	return message, err
}
