package csvi

import (
	"github.com/mattn/go-runewidth"
)

const (
	minimumWidth = 3
	maximumWidth = 40
)

func (app *Application) narrow(cellWidth *CellWidth) {
	if w := cellWidth.Get(app.cursorCol) - 1; w > minimumWidth {
		cellWidth.Set(app.cursorCol, w)
	}
	app.clearCache()
}

func (app *Application) widen(cellWidth *CellWidth) {
	if w := cellWidth.Get(app.cursorCol); w < maximumWidth {
		cellWidth.Set(app.cursorCol, w+1)
	}
	app.clearCache()
}

func (app *Application) fitWidthCurrent(cellWidth *CellWidth) {
	maxW := 0
	for p := app.startRow; p != nil; p = p.Next() {
		w := runewidth.StringWidth(p.Row.Cell[app.cursorCol].Text())
		if w > maxW {
			maxW = w
		}
	}
	maxW++
	if maxW < minimumWidth {
		maxW = minimumWidth
	} else if maxW > maximumWidth {
		maxW = maximumWidth
	}
	cellWidth.Set(app.cursorCol, maxW)
	app.clearCache()
}

func (app *Application) fitWidthAll(cellWidth *CellWidth) {
	maxW := []int{}
	for p := app.startRow; p != nil; p = p.Next() {
		for i, c := range p.Row.Cell {
			if i >= len(maxW) {
				maxW = append(maxW, 0)
			}
			w := runewidth.StringWidth(c.Text())
			if w > maxW[i] {
				maxW[i] = w
			}
		}
	}
	for i, w := range maxW {
		w++
		if w < minimumWidth {
			w = minimumWidth
		} else if w > maximumWidth {
			w = maximumWidth
		}
		cellWidth.Set(i, w)
	}
	app.clearCache()
}
