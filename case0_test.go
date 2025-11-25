package csvi_test

import (
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hymkor/csvi/startup"
)

func TestFileIO(t *testing.T) {
	src := ""
	op := "i|ahaha"
	exp := "ahaha"
	testCase(t, src, op, exp)
}

func TestPipelineIO(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	instance := startup.NewFlag().Bind(fs)
	err := fs.Parse([]string{"-auto", "w|-|q|y"})
	if err != nil {
		t.Fatal(err.Error())
	}

	path := filepath.Join(t.TempDir(), "test.csv")
	fd, err := os.Create(path)
	if err != nil {
		t.Fatal(err.Error())
	}

	stdoutSave := os.Stdout
	os.Stdout = fd

	err = instance.RunInOut(strings.NewReader("ihihi\r\n"), io.Discard)

	os.Stdout = stdoutSave
	fd.Close()

	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err.Error())
	}

	checkResult(t, path, "ihihi\r\n")
}
