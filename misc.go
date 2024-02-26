package main

import (
	"io"
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/hymkor/csview/unbreakable-csv"
)

func readCsvAll(in io.ByteReader, mode *csv.Mode) ([]csv.Row, error) {
	csvlines := []csv.Row{}
	for {
		row, err := csv.ReadLine(in, mode)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			return csvlines, nil
		}
		csvlines = append(csvlines, *row)
	}
}

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

func makeCandidate(row, col int, csvlines []csv.Row) candidate {
	result := candidate(make([]string, 0, row))
	set := make(map[string]struct{})
	for ; row >= 0; row-- {
		if col >= len(csvlines[row].Cell) {
			break
		}
		value := csvlines[row].Cell[col].Text()
		if value == "" {
			break
		}
		if _, ok := set[value]; !ok {
			result = append(result, value)
			set[value] = struct{}{}
		}
	}
	return result
}

func searchForward(csvlines []csv.Row, r, c int, target string) (bool, int, int) {
	c++
	for r < len(csvlines) {
		for c < len(csvlines[r].Cell) {
			if strings.Contains(csvlines[r].Cell[c].Text(), target) {
				return true, r, c
			}
			c++
		}
		r++
		c = 0
	}
	return false, r, c
}

func searchBackward(csvlines []csv.Row, r, c int, target string) (bool, int, int) {
	c--
	for {
		for c >= 0 {
			if strings.Contains(csvlines[r].Cell[c].Text(), target) {
				return true, r, c
			}
			c--
		}
		r--
		if r < 0 {
			return false, r, c
		}
		c = len(csvlines[r].Cell) - 1
	}
}
