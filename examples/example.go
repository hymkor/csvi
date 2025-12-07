//go:build example

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-colorable"

	"github.com/hymkor/csvi"
	"github.com/hymkor/csvi/uncsv"
)

func main() {
	source := `A,B,C,D
"A1","B1","C1","D1"
"A2","B2","C2","D2"`

	cfg := &csvi.Config{
		Mode: &uncsv.Mode{Comma: ','},
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
