package csvi_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hymkor/csvi/startup"
)

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

func TestCase3(t *testing.T) {
	path1 := makeSource(t, "t1.csv", "first\n")
	path2 := makeSource(t, "t2.csv", "second\n")
	path3 := makeSource(t, "t3.csv", "third\n")
	outputPath := filepath.Join(t.TempDir(), "t4.csv")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	instance := startup.NewFlag().Bind(fs)
	err := fs.Parse([]string{"-auto", fmt.Sprintf("w|%s|q|y", outputPath), path1, path2, path3})
	if err != nil {
		t.Fatal(err.Error())
	}
	enable := disableStdout(t)
	err = instance.Run()
	enable()
	if err != nil {
		t.Fatal(err.Error())
	}
	checkResult(t, outputPath, "first\nsecond\nthird\n")
}
