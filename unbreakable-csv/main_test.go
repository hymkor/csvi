package csv

import (
	"io"
	"strings"
	"testing"
)

func try(t *testing.T, source string, expect ...string) {
	t.Helper()
	r := strings.NewReader(source)
	mode := &Mode{Comma: ','}
	row, err := ReadLine(r, mode)
	if err != nil {
		if err != io.EOF || strings.HasSuffix(source, "\n") {
			t.Fatalf("error=%s", err.Error())
		}
	}
	if len(row.Cell) != len(expect) {
		t.Fatalf("size differs: %v", source)
	}
	for i, expect1 := range expect {
		if text := row.Cell[i].Text(); text != expect1 {
			t.Fatalf("[%d] expect %v but %v\n", i, expect1, text)
		}
	}
	if j := string(row.Rebuild(mode)); j != source {
		t.Fatalf("joined string expects %v but %v", source, j)
	}
}

func TestReadLine(t *testing.T) {
	try(t, "abcdef,12345\n", "abcdef", "12345")
	try(t, "\"abcdef,12345\",44444\n", "abcdef,12345", "44444")
	try(t, "\"abcdef\n12\"\"345\",44444", "abcdef\n12\"345", "44444")
}

func tryUpdate(t *testing.T, source string, n int, newText, expect string) {
	mode := &Mode{Comma: ','}
	row, err := ReadLine(strings.NewReader(source), mode)
	if err != nil && err != io.EOF {
		t.Fatalf("error=%s", err.Error())
	}
	row.Replace(n, newText, mode)
	result := string(row.Rebuild(mode))
	if result != expect {
		t.Fatalf("expect %v, but %v", expect, result)
	}
}

func TestRebuild(t *testing.T) {
	tryUpdate(t, "abcdef,12345\r\n", 1, "123\n456", "abcdef,\"123\n456\"\r\n")
	tryUpdate(t, "abcdef,12345\n", 1, "777777", "abcdef,777777\n")
}
