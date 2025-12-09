package csvi

import (
	"errors"

	"github.com/hymkor/csvi/uncsv"
)

type pasteType int

const (
	pasteAfter pasteType = iota
	pasteBefore
	pasteOver
)

type pasteFunc func(head, dst **RowPtr, col *int, pt pasteType) error

func noPaste(head, dst **RowPtr, col *int, pt pasteType) error { return nil }

func (app *application) yankCurrentCell(src *RowPtr, col int) pasteFunc {
	dup := src.Cell[col].Clone()
	paste := func(head, dst **RowPtr, dcol *int, pt pasteType) error {
		if pt == pasteOver {
			if m := app.checkWriteProtect(*dst); m != "" {
				return errors.New(m)
			}
			(*dst).Cell[*dcol].SetSource(dup.Source(), app.Config.Mode)
		} else {
			if m := app.checkWriteProtectAndColumn(*dst); m != "" {
				return errors.New(m)
			}
			if pt == pasteAfter {
				(*dcol)++
			}
			(*dst).InsertCell(*dcol, dup, app.Config.Mode)
		}
		return nil
	}
	return paste
}

func (app *application) removeCurrentCell(src *RowPtr, col int) pasteFunc {
	paste := app.yankCurrentCell(src, col)
	if len(src.Cell) <= 1 {
		src.Replace(0, "", app.Config.Mode)
	} else {
		src.Delete(col)
	}
	return paste
}

func (app *application) makeRowPaster(dup *uncsv.Row) pasteFunc {
	paste := func(head, dst **RowPtr, _ *int, pt pasteType) error {
		if m := app.checkWriteProtect(*dst); m != "" {
			return errors.New(m)
		}
		if pt == pasteOver {
			for i := 0; i < len(dup.Cell); i++ {
				if i >= len((*dst).Cell) {
					(*dst).Cell = append((*dst).Cell, dup.Cell[i])
				} else {
					(*dst).Cell[i].SetSource(dup.Cell[i].Source(), app.Config.Mode)
				}
			}
		} else if pt == pasteAfter {
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
		return nil
	}
	return paste
}

func (app *application) yankCurrentRow(src *RowPtr) pasteFunc {
	dup := &uncsv.Row{Term: src.Term}
	for _, c := range src.Cell {
		dup.Cell = append(dup.Cell, c.Clone())
	}
	return app.makeRowPaster(dup)
}

func (app *application) removeCurrentRow(head, src **RowPtr) pasteFunc {
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

func (app *application) yankCurrentColumn(col int) pasteFunc {
	dup := []uncsv.Cell{}
	for p := app.Front(); p != nil; p = p.Next() {
		var newCell uncsv.Cell
		if col < len(p.Cell) {
			newCell = p.Cell[col]
		}
		dup = append(dup, newCell)
	}
	return func(head, dst **RowPtr, col *int, pt pasteType) error {
		var m string
		if pt == pasteOver {
			m = app.checkWriteProtect(*dst)
		} else {
			m = app.checkWriteProtectAndColumn(*dst)
		}
		if m != "" {
			return errors.New(m)
		}
		pos := *col
		if pt == pasteAfter {
			pos++
		}
		i := 0
		for p := app.Front(); p != nil; p = p.Next() {
			var newSrc []byte
			if i < len(dup) {
				newSrc = dup[i].Source()
			} else {
				newSrc = []byte{}
			}
			i++
			if pt == pasteOver {
				p.Cell[pos].SetSource(newSrc, app.Config.Mode)
			} else {
				var newCell uncsv.Cell
				newCell.SetSource(newSrc, app.Config.Mode)
				p.InsertCell(pos, newCell, app.Config.Mode)
			}
		}
		return nil
	}
}

func (app *application) removeCurrentColumn(col int) pasteFunc {
	paste := app.yankCurrentColumn(col)
	for p := app.Front(); p != nil; p = p.Next() {
		if len(p.Cell) > 1 && col < len(p.Cell) {
			copy(p.Cell[col:], p.Cell[col+1:])
			p.Cell = p.Cell[:len(p.Cell)-1]
		}
	}
	return paste
}
