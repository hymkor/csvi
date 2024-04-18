package main

import (
	"strings"

	"github.com/mattn/go-runewidth"
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

func makeCandidate(row, col int, cursor *_RowPtr) candidate {
	result := candidate(make([]string, 0, 100))
	set := make(map[string]struct{})
	for ; cursor != nil; cursor = cursor.Prev() {
		if col >= len(cursor.Cell) {
			break
		}
		value := cursor.Cell[col].Text()
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

func searchForward(cursor *_RowPtr, c int, target string) (*_RowPtr, int) {
	c++
	for cursor != nil {
		for c < len(cursor.Cell) {
			if strings.Contains(cursor.Cell[c].Text(), target) {
				return cursor, c
			}
			c++
		}
		cursor = cursor.Next()
		c = 0
	}
	return nil, c
}

func searchBackward(cursor *_RowPtr, c int, target string) (*_RowPtr, int) {
	c--
	for {
		for c >= 0 {
			if strings.Contains(cursor.Cell[c].Text(), target) {
				return cursor, c
			}
			c--
		}
		cursor = cursor.Prev()
		if cursor == nil {
			return nil, c
		}
		c = len(cursor.Cell) - 1
	}
}
