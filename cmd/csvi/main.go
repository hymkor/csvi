package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/hymkor/csvi/startup"
)

var version string

func main() {
	fmt.Fprintf(os.Stderr, "%s %s-%s-%s built with %s\n",
		os.Args[0], version, runtime.GOOS, runtime.GOARCH, runtime.Version())
	if err := startup.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
