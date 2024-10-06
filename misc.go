package csvi

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

type Candidate []string

func (c Candidate) Len() int {
	return len(c)
}

func (c Candidate) At(n int) string {
	return c[len(c)-n-1]
}

func (c Candidate) Delimiters() string {
	return ""
}

func (c Candidate) Enclosures() string {
	return ""
}

func (c Candidate) List(field []string) (fullnames, basenames []string) {
	return c, c
}

func makeCandidate(row, col int, cursor *RowPtr) Candidate {
	result := Candidate(make([]string, 0, 100))
	set := make(map[string]struct{})
	count := 0
	for ; cursor != nil; cursor = cursor.Prev() {
		count++
		if col >= len(cursor.Cell) {
			if count == 1 {
				continue
			} else {
				break
			}
		}
		value := cursor.Cell[col].Text()
		if value == "" {
			if count == 1 {
				continue
			} else {
				break
			}
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

func searchForward(cursor *RowPtr, c int, target string) (*RowPtr, int) {
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

func searchBackward(cursor *RowPtr, c int, target string) (*RowPtr, int) {
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
