package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/hymkor/csvi/csviapp"
)

func main() {
	fmt.Fprintf(os.Stderr, "Csvi %s-%s-%s <https://hymkor.github.io/csvi>\n",
		version, runtime.GOOS, runtime.GOARCH)
	if err := csviapp.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
