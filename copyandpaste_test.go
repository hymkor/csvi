package csvi_test

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hymkor/csvi/startup"
)

func testRun(t *testing.T, args ...string) {
	t.Helper()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	instance := startup.NewFlag().Bind(fs)
	err := fs.Parse(args)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = instance.Run()
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err.Error())
	}
}

func makeSource(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.csv")
	err := os.WriteFile(path, []byte(text), 0666)
	if err != nil {
		t.Fatal(err.Error())
	}
	return path
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

func testCase(t *testing.T, source, process, result string) {
	t.Helper()
	path := makeSource(t, source)
	testRun(t, "-auto", fmt.Sprintf("%s|w|%s|y|q|y", process, path), path)
	checkResult(t, path, result)
}

func TestDeleteCell(t *testing.T) { // `x`
	testCase(t,
		"あ,い,う,え,お\nか,き,く,け,こ",
		"<|x",
		"い,う,え,お\nか,き,く,け,こ")
}

func TestDeleteRow(t *testing.T) { // `D`
	testCase(t,
		"あ,い,う,え,お\nか,き,く,け,こ",
		"<|D",
		"か,き,く,け,こ")
}
func TestDeleteColumn(t *testing.T) { // `dc`
	testCase(t,
		"あ,い,う,え,お\nか,き,く,け,こ",
		"<|d|c",
		"い,う,え,お\nき,く,け,こ")
}

func TestCopyPasteCell(t *testing.T) { // `yl` and `p`
	testCase(t,
		"あ,い,う,え,お\nか,き,く,け,こ",
		"<|y|l|l|p",
		"あ,い,あ,う,え,お\nか,き,く,け,こ")
}

func TestCopyPasteCellB(t *testing.T) { // `yl` and `P`
	testCase(t,
		"あ,い,う,え,お\nか,き,く,け,こ",
		"<|y|l|$|P",
		"あ,い,う,え,あ,お\nか,き,く,け,こ")
}

func TestCopyPasteRow(t *testing.T) { // `yy` and `p`
	testCase(t,
		"あ,い,う,え,お\nか,き,く,け,こ",
		"<|y|y|>|p",
		"あ,い,う,え,お\nか,き,く,け,こ\nあ,い,う,え,お\n")
}

func TestCopyPasteRowB(t *testing.T) { // `yy` and `P`
	testCase(t,
		"あ,い,う,え,お\r\nか,き,く,け,こ",
		"<|y|y|P",
		"あ,い,う,え,お\r\nあ,い,う,え,お\r\nか,き,く,け,こ")
}

func TestCopyPasteColumn(t *testing.T) { // `yc` and `p`
	testCase(t,
		"あ,い,う,え,お\nか,き,く,け,こ",
		"<|y|c|$|p",
		"あ,い,う,え,お,あ\nか,き,く,け,こ,か")
}

func TestCopyPasteColumnB(t *testing.T) { // `yc` and `P`
	testCase(t,
		"あ,い,う,え,お\nか,き,く,け,こ",
		"<|y|c|$|P",
		"あ,い,う,え,あ,お\nか,き,く,け,か,こ")
}
