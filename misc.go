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
