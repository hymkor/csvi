package csvi

import (
	"github.com/hymkor/csvi/uncsv"
)

type pasteFunc func(head, dst **RowPtr, col *int, offset bool)

func noPaste(head, dst **RowPtr, col *int, offset bool) {}

func (app *_Application) yankCurrentCell(src *RowPtr, col int) pasteFunc {
	dup := src.Cell[col].Clone()
	paste := func(head, dst **RowPtr, dcol *int, offset bool) {
		if offset {
			(*dcol)++
		}
		(*dst).InsertCell(*dcol, dup, app.Config.Mode)
	}
	return paste
}

func (app *_Application) removeCurrentCell(src *RowPtr, col int) pasteFunc {
	paste := app.yankCurrentCell(src, col)
	if len(src.Cell) <= 1 {
		src.Replace(0, "", app.Config.Mode)
	} else {
		src.Delete(col)
	}
	return paste
}

func (app *_Application) makeRowPaster(dup *uncsv.Row) pasteFunc {
	paste := func(head, dst **RowPtr, _ *int, offset bool) {
		if offset {
			(*dst).InsertAfter(dup)
			if (*dst).Term == "" {
				(*dst).Term = app.Config.Mode.DefaultTerm
			}
		} else {
			if (*head).lnum == (*dst).lnum {
				defer func() {
					*head = (*dst).Clone()
				}()
			}
			*dst = (*dst).InsertBefore(dup)
		}
	}
	return paste
}

func (app *_Application) yankCurrentRow(src *RowPtr) pasteFunc {
	dup := &uncsv.Row{Term: src.Term}
	for _, c := range src.Cell {
		dup.Cell = append(dup.Cell, c.Clone())
	}
	return app.makeRowPaster(dup)
}

func (app *_Application) removeCurrentRow(head, src **RowPtr) pasteFunc {
	if app.Len() <= 1 {
		return noPaste
	}
	newLnum := (*src).lnum
	paste := app.makeRowPaster((*src).Row)

	headPrev := (*head).Prev()
	prevP := (*src).Prev()
	removedRow := (*src).Remove()
	app.removedRows = append(app.removedRows, removedRow)
	if prevP == nil {
		(*src) = app.Front()
	} else if next := prevP.Next(); next != nil {
		(*src) = next
		(*src).lnum = newLnum
	} else {
		(*src) = prevP
		(*src).Term = removedRow.Term
	}
	if headPrev == nil {
		(*head) = app.Front()
	} else {
		(*head) = headPrev.Next()
	}
	return paste
}
