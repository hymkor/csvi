//go:build example

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"

	"github.com/hymkor/csvi"
	"github.com/hymkor/csvi/uncsv"
)

func main() {
	source := `01,02,03,04
"11","21","31","41"
"42","42","42","42"`

	cfg := &csvi.Config{
		Mode:          &uncsv.Mode{Comma: ','},
		HeaderLines:   1,
		ProtectHeader: true,
		OnCellValidated: func(e *csvi.CellValidatedEvent) (string, error) {
			// All cells must contain only numbers
			_, err := strconv.ParseFloat(e.Text, 64)
			if err != nil {
				return "", err
			}
			return e.Text, nil
		},
	}

	result, err := cfg.Edit(strings.NewReader(source), colorable.NewColorableStdout())

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// // env GOEXPERIMENT=rangefunc go run example
	// for row := range result.Each {
	//     os.Stdout.Write(row.Rebuild(cfg.Mode))
	// }
	result.Each(func(row *uncsv.Row) bool {
		os.Stdout.Write(row.Rebuild(cfg.Mode))
		return true
	})
}
