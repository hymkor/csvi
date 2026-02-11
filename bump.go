//go:build run

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var rxVersion = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

func findVersion1(fname string) (string, error) {
	fd, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	sc := bufio.NewScanner(fd)
	for sc.Scan() {
		line := sc.Text()
		if rxVersion.MatchString(line) {
			return line, nil
		}
	}
	return "", sc.Err()
}

func findVersion(args []string) (string, error) {
	for _,arg := range args {
		notes, err := filepath.Glob(arg)
		if err != nil {
			notes = []string{arg}
		}
		for _, fname := range notes {
			version, err := findVersion1(fname)
			if err != nil {
				return "", err
			}
			if version != "" {
				return version, nil
			}
		}
	}
	return "", errors.New("not found")
}

var (
	flagGoSource = flag.String("gosource", "", "Specify package name; Output as golang source")
	flagSuffix   = flag.String("suffix", "", "Suffix for version")
)

func mains(args []string) error {
	version, err := findVersion(args)
	if err != nil {
		return err
	}
	if *flagSuffix != "" {
		version += *flagSuffix
	}
	if *flagGoSource != "" {
		fmt.Printf("package %s\n\nvar version = \"%s\"\n", *flagGoSource, version)
		return nil
	}
	fmt.Println(version)
	return nil
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
