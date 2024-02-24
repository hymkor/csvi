package main

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/mattn/go-runewidth"
)

func readCsvAll(in *csv.Reader) ([][]string, error) {
	csvlines := [][]string{}
	for {
		csv1, err := in.Read()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			return csvlines, nil
		}
		for i, c := range csv1 {
			csv1[i] = strings.ReplaceAll(c, emptyDummyCode, "")
		}
		csvlines = append(csvlines, csv1)
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

func makeCandidate(row, col int, csvlines [][]string) candidate {
	result := candidate(make([]string, 0, row))
	set := make(map[string]struct{})
	for ; row >= 0; row-- {
		if col >= len(csvlines[row]) {
			break
		}
		value := csvlines[row][col]
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

func searchForward(csvlines [][]string, r, c int, target string) (bool, int, int) {
	c++
	for r < len(csvlines) {
		for c < len(csvlines[r]) {
			if strings.Contains(csvlines[r][c], target) {
				return true, r, c
			}
			c++
		}
		r++
		c = 0
	}
	return false, r, c
}

func searchBackward(csvlines [][]string, r, c int, target string) (bool, int, int) {
	c--
	for {
		for c >= 0 {
			if strings.Contains(csvlines[r][c], target) {
				return true, r, c
			}
			c--
		}
		r--
		if r < 0 {
			return false, r, c
		}
		c = len(csvlines[r]) - 1
	}
}
