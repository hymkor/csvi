package csvi

import (
	"errors"
	"io"

	"github.com/nyaosorg/go-readline-ny"
)

func cmdEditCellWith(editor func(string, *Application) (string, error), app *Application) string {
	if m := app.checkWriteProtect(app.cursorRow); m != "" {
		return m
	}
	cursor := &app.cursorRow.Cell[app.cursorCol]
	modifiedBefore := cursor.Modified()
	q := cursor.IsQuoted()
	app.clearCache()
	if text, err := editor(cursor.Text(), app); err == nil {
		app.cursorRow.Replace(app.cursorCol, text, app.Mode)
		if q {
			*cursor = cursor.Quote(app.Mode)
		}
	} else {
		return err.Error()
	}
	modifiedAfter := cursor.Modified()
	app.updateSoftDirty(modifiedBefore, modifiedAfter)
	return ""
}

func cmdEditCell(app *Application) string {
	return cmdEditCellWith(func(text string, app *Application) (string, error) {
		text, err := app.readlineAndValidate("replace cell>", text, app.cursorRow, app.cursorCol)
		if errors.Is(err, io.EOF) || errors.Is(err, readline.CtrlC) {
			err = errCanceled
		}
		return text, err
	}, app)
}

func cmdEditCellExtEditor(app *Application) string {
	if app.ExtEditor == nil {
		return cmdEditCell(app)
	}
	return cmdEditCellWith(app.ExtEditor, app)
}
