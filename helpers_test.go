package csvi_test

import (
	"io"
	"os"
	"testing"
	"flag"
	"errors"
	"path/filepath"
	"fmt"
	"strings"

	"github.com/mattn/go-colorable"

	"github.com/hymkor/csvi/csviapp"
)

func newTestOptions(args ...string) (*csviapp.Options, error) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	opt := csviapp.NewOptions().Bind(flagSet)
	return opt, flagSet.Parse(args)
}

func testRun(t *testing.T, dataSource io.Reader, args ...string) {
	t.Helper()
	instance, err := newTestOptions(args...)
	if err != nil {
		t.Fatal(err.Error())
	}
	var ttyOut io.Writer = io.Discard
	if testing.Verbose() {
		disable := colorable.EnableColorsStdout(nil)
		if disable != nil {
			defer disable()
		}
		ttyOut = colorable.NewColorableStderr()
	}
	err = instance.RunInOut(dataSource, ttyOut)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err.Error())
	}
}

func checkResult(t *testing.T, path, expect string) {
	t.Helper()
	bin, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err.Error())
	}
	result := string(bin)
	if expect != result {
		t.Fatalf("Expect %#v, but %#v", expect, result)
	}
}

func testCase(t *testing.T, source, process, result string, options ...string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.csv")
	args := make([]string, 0, len(options)+3)
	args = append(args, options...)
	args = append(args, "-auto")
	args = append(args, fmt.Sprintf("%s|w|%s|q|y", process, path))
	testRun(t, strings.NewReader(source), args...)
	checkResult(t, path, result)
}

func makeSource(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0666)
	if err != nil {
		t.Fatal(err.Error())
	}
	return path
}

func disableStdout(t *testing.T) func() {
	if testing.Verbose() {
		return func() {}
	}
	fd, err := os.Create(os.DevNull)
	if err != nil {
		t.Fatal(err.Error())
	}
	stdoutSave := os.Stdout
	os.Stdout = fd
	return func() {
		fd.Close()
		os.Stdout = stdoutSave
	}
}
