package csvi

import (
	"container/list"
	"io"

	"github.com/hymkor/csvi/uncsv"
)

type RowPtr struct {
	*uncsv.Row
	element *list.Element
	lnum    int
	list    *list.List
}

func (r *RowPtr) Next() *RowPtr {
	next := r.element.Next()
	if next == nil {
		return nil
	}
	return &RowPtr{Row: next.Value.(*uncsv.Row), element: next, lnum: r.lnum + 1, list: r.list}
}

func (r *RowPtr) Prev() *RowPtr {
	prev := r.element.Prev()
	if prev == nil {
		return nil
	}
	return &RowPtr{Row: prev.Value.(*uncsv.Row), element: prev, lnum: r.lnum - 1, list: r.list}
}

func (r *RowPtr) Remove() *uncsv.Row {
	return r.list.Remove(r.element).(*uncsv.Row)
}

func (r *RowPtr) Clone() *RowPtr {
	return &RowPtr{Row: r.element.Value.(*uncsv.Row), element: r.element, lnum: r.lnum, list: r.list}
}

func frontPtr(L *list.List) *RowPtr {
	front := L.Front()
	if front == nil {
		return nil
	}
	return &RowPtr{Row: front.Value.(*uncsv.Row), element: front, lnum: 0, list: L}
}

func backPtr(L *list.List) *RowPtr {
	back := L.Back()
	return &RowPtr{Row: back.Value.(*uncsv.Row), element: back, lnum: L.Len() - 1, list: L}
}

func (r *RowPtr) InsertAfter(val *uncsv.Row) *RowPtr {
	next := r.list.InsertAfter(val, r.element)
	return &RowPtr{Row: next.Value.(*uncsv.Row), element: next, lnum: r.lnum + 1, list: r.list}
}

func (r *RowPtr) InsertBefore(val *uncsv.Row) *RowPtr {
	next := r.list.InsertBefore(val, r.element)
	r.lnum++
	return &RowPtr{Row: next.Value.(*uncsv.Row), element: next, lnum: r.lnum - 1, list: r.list}
}

func (r *RowPtr) Index() int {
	return r.lnum
}

type _Application struct {
	csvLines    *list.List
	removedRows []*uncsv.Row
	out         io.Writer
	*Config
}

type Result struct {
	*_Application
}

func (app *_Application) Write(data []byte) (int, error) {
	return app.out.Write(data)
}

func (app *_Application) Front() *RowPtr {
	return frontPtr(app.csvLines)
}

func (app *_Application) Back() *RowPtr {
	return backPtr(app.csvLines)
}

func (app *_Application) Len() int {
	return app.csvLines.Len()
}

func (app *_Application) Push(row *uncsv.Row) {
	app.csvLines.PushBack(row)
}

func (app *_Application) Each(callback func(*uncsv.Row) bool) {
	for p := app.Front(); p != nil; p = p.Next() {
		if !callback(p.Row) {
			break
		}
	}
}

func (app *_Application) RemovedRows(callback func(*uncsv.Row) bool) {
	for _, p := range app.removedRows {
		if !callback(p) {
			break
		}
	}
}
