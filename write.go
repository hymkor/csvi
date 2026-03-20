package csvi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/nyaosorg/go-inline-animation"

	"github.com/hymkor/go-safewrite"
	"github.com/hymkor/go-safewrite/perm"

	"github.com/hymkor/csvi/internal/ansi"
	"github.com/hymkor/csvi/uncsv"
)

func (app *Application) dump(ctx context.Context, w io.Writer) error {
	cursor := app.Front()
	return app.Config.Mode.DumpBy(
		ctx,
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

func (app *Application) dumpWithAnimationAndCancel(fd io.Writer) error {
	ctx, cancel := app.withSlowOperation("Saving...")
	defer cancel()

	return app.dump(ctx, fd)
}

func (app *Application) cmdWrite(fname string) (string, error) {
	if fname == "-" {
		err := app.dumpWithAnimationAndCancel(os.Stdout)
		return "Output to STDOUT", err
	}

	prompt := func(info *safewrite.Info) bool {
		if info.Status != safewrite.NONE {
			return true
		}
		if info.ReadOnly() {
			return app.yesNo("Overwrite READONLY file \"" + info.Name + "\" [y/n] ?")
		}
		return app.yesNo("Overwrite as \"" + info.Name + "\" [y/n] ?")
	}
	fd, err := safewrite.Open(fname, prompt)
	if err != nil {
		return "", err
	}

	if err := app.dumpWithAnimationAndCancel(fd); err != nil {
		return "", err
	}

	if err := fd.Close(); err != nil {
		var be *safewrite.BackupError
		if errors.As(err, &be) {
			return "",
				fmt.Errorf("failed to backup %q to %q (tmp: %q)",
					filepath.Base(be.Target),
					filepath.Base(be.Backup),
					filepath.Base(be.Tmp))
		}
		var re *safewrite.ReplaceError
		if errors.As(err, &re) {
			return "",
				fmt.Errorf("failed to replace %q to %q",
					filepath.Base(re.Tmp),
					filepath.Base(re.Target))
		}
		return "", err
	}
	perm.Track(fd)
	return fmt.Sprintf("Saved as \"%s\"", fname), nil
}

func (app *Application) cmdSave() (string, error) {
	var wg sync.WaitGroup

	ctx, cancel := app.ctrlC.NotifyContext(context.Background())

	if app.fetchFunc != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if ctx.Err() != nil {
					return
				}
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
		}()
	}
	fname, err := app.GetFilename(app, "write to>", app.getSavePath())
	if err != nil {
		return "", err
	}
	io.WriteString(app.out, ansi.YELLOW+"\rReading all data... "+ansi.ERASE_SCRN_AFTER)
	end := animation.Dots.Progress(app.out)
	wg.Wait()
	end()
	ctxErr := ctx.Err()
	cancel()
	if ctxErr != nil {
		return "", errors.New("Save interrupted")
	}
	message, err := app.cmdWrite(fname)
	if err == nil {
		app.resetDirty()
	}
	app.lastSavePath = fname
	return message, err
}
