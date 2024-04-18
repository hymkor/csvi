package main

import (
	"container/list"

	"github.com/hymkor/csvi/uncsv"
)

type _RowPtr struct {
	*uncsv.Row
	element *list.Element
	lnum    int
	list    *list.List
}

func (r *_RowPtr) Next() *_RowPtr {
	next := r.element.Next()
	if next == nil {
		return nil
	}
	return &_RowPtr{Row: next.Value.(*uncsv.Row), element: next, lnum: r.lnum + 1, list: r.list}
}

func (r *_RowPtr) Prev() *_RowPtr {
	prev := r.element.Prev()
	if prev == nil {
		return nil
	}
	return &_RowPtr{Row: prev.Value.(*uncsv.Row), element: prev, lnum: r.lnum - 1, list: r.list}
}

func (r *_RowPtr) Remove() *uncsv.Row {
	return r.list.Remove(r.element).(*uncsv.Row)
}

func (r *_RowPtr) Clone() *_RowPtr {
	return &_RowPtr{Row: r.element.Value.(*uncsv.Row), element: r.element, lnum: r.lnum, list: r.list}
}

func frontPtr(L *list.List) *_RowPtr {
	front := L.Front()
	return &_RowPtr{Row: front.Value.(*uncsv.Row), element: front, lnum: 0, list: L}
}

func backPtr(L *list.List) *_RowPtr {
	back := L.Back()
	return &_RowPtr{Row: back.Value.(*uncsv.Row), element: back, lnum: L.Len() - 1, list: L}
}

func (r *_RowPtr) InsertAfter(val *uncsv.Row) *_RowPtr {
	next := r.list.InsertAfter(val, r.element)
	return &_RowPtr{Row: next.Value.(*uncsv.Row), element: next, lnum: r.lnum + 1, list: r.list}
}

func (r *_RowPtr) InsertBefore(val *uncsv.Row) *_RowPtr {
	next := r.list.InsertBefore(val, r.element)
	return &_RowPtr{Row: next.Value.(*uncsv.Row), element: next, lnum: r.lnum + 1, list: r.list}
}
