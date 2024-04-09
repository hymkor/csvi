package main

import (
	"container/list"
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/hymkor/csvi/uncsv"
)

func cutStrInWidth(s string, cellwidth int) (string, int) {
	w := 0
	for n, c := range s {
		w1 := runewidth.RuneWidth(c)
		if w+w1 > cellwidth {
			return s[:n], w
		}
		w += w1
	}
	return s, w
}

type candidate []string

func (c candidate) Len() int {
	return len(c)
}

func (c candidate) At(n int) string {
	return c[len(c)-n-1]
}

func (c candidate) Delimiters() string {
	return ""
}

func (c candidate) Enclosures() string {
	return ""
}

func (c candidate) List(field []string) (fullnames, basenames []string) {
	return c, c
}

func makeCandidate(row, col int, cursor *list.Element) candidate {
	result := candidate(make([]string, 0, 100))
	set := make(map[string]struct{})
	for ; cursor != nil; cursor = cursor.Prev() {
		row := cursor.Value.(*uncsv.Row)
		if col >= len(row.Cell) {
			break
		}
		value := row.Cell[col].Text()
		if value == "" {
			break
		}
		if _, ok := set[value]; !ok {
			result = append(result, value)
			set[value] = struct{}{}
			if len(set) > 100 {
				break
			}
		}
	}
	if len(result) <= 0 {
		result = append(result, "")
	}
	return result
}

func searchForward(cursor *list.Element, r, c int, target string) (*list.Element, int, int) {
	c++
	for cursor != nil {
		row := cursor.Value.(*uncsv.Row)
		for c < len(row.Cell) {
			if strings.Contains(row.Cell[c].Text(), target) {
				return cursor, r, c
			}
			c++
		}
		r++
		cursor = cursor.Next()
		c = 0
	}
	return nil, r, c
}

func searchBackward(cursor *list.Element, r, c int, target string) (*list.Element, int, int) {
	c--
	for {
		row := cursor.Value.(*uncsv.Row)
		for c >= 0 {
			if strings.Contains(row.Cell[c].Text(), target) {
				return cursor, r, c
			}
			c--
		}
		r--
		cursor = cursor.Prev()
		if r < 0 || cursor == nil {
			return nil, r, c
		}
		c = len(cursor.Value.(*uncsv.Row).Cell) - 1
	}
}
