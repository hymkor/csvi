package csvi

import (
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/hymkor/csvi/candidate"
)

func cutStrInWidth(s string, cellwidth int) (string, int) {
	w := 0
	escape := false
	for n, c := range s {
		if escape {
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				escape = false
			}
			continue
		}
		if c == '\x1B' {
			escape = true
			continue
		}
		w1 := runewidth.RuneWidth(c)
		if w+w1 > cellwidth {
			var buffer strings.Builder
			buffer.WriteString(s[:n])
			for _, c := range s[n:] {
				if escape {
					buffer.WriteRune(c)
					if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
						escape = false
					}
					continue
				}
				if c == '\x1B' {
					buffer.WriteByte('\x1B')
					escape = true
				}
			}
			return buffer.String(), w
		}
		w += w1
	}
	return s, w
}

func makeCandidate(row, col int, cursor *RowPtr) candidate.Candidate {
	result := candidate.Candidate(make([]string, 0, 100))
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

func searchExactForward(cursor *RowPtr, c int, target string) (*RowPtr, int) {
	c++
	for cursor != nil {
		for c < len(cursor.Cell) {
			if strings.EqualFold(cursor.Cell[c].Text(), target) {
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

func searchExactBackward(cursor *RowPtr, c int, target string) (*RowPtr, int) {
	c--
	for {
		for c >= 0 {
			if strings.EqualFold(cursor.Cell[c].Text(), target) {
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
