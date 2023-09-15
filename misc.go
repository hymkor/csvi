package main

import (
	"encoding/csv"
	"io"
	"strings"
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
