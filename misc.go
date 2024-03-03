package main

import (
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/mattn/go-tty"

	"github.com/hymkor/go-cursorposition"

	"github.com/hymkor/csview/uncsv"
)

func initAmbiguousWidth(tty1 *tty.TTY) error {
	// Measure how far the cursor moves while the `▽` is printed
	w, err := cursorposition.AmbiguousWidthGoTty(tty1, os.Stderr)
	if err != nil {
		return err
	}
	runewidth.DefaultCondition.EastAsianWidth = w >= 2
	return nil
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
	if row < 0 || row >= len(csvlines) {
		return candidate([]string{})
	}
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
