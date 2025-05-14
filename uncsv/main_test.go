package uncsv

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

func try(t *testing.T, source string, expect ...string) {
	t.Helper()
	r := bufio.NewReader(strings.NewReader(source))
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

func upd(t *testing.T, source string, expect string, f func([]Row, *Mode)) {
	t.Helper()
	mode := &Mode{Comma: ','}
	rows, err := ReadAll(strings.NewReader(source), mode)
	if err != nil {
		t.Fatalf("error=%s", err.Error())
	}

	f(rows, mode)

	var buffer strings.Builder
	mode.Dump(rows, &buffer)
	result := buffer.String()
	if result != expect {
		t.Fatalf("expect `%v`, but `%v`", expect, result)
	}
}

func TestUpdateWithCRLF(t *testing.T) {
	upd(t, "abcdef,12345\r\n", "abcdef,\"123\r\n456\"\r\n", func(r []Row, m *Mode) {
		r[0].Replace(1, "123\r\n456", m)
	})
}

func TestUpdateWithLFOnly(t *testing.T) {
	upd(t, "abcdef,12345\n", "abcdef,777777\n", func(r []Row, m *Mode) {
		r[0].Replace(1, "777777", m)
	})
}

func TestUpdateWithBom(t *testing.T) {
	upd(t, "\uFEFF12345\r\n", "\uFEFF77777\r\n", func(r []Row, m *Mode) {
		r[0].Replace(0, "77777", m)
	})
}

func TestDeleteWithBom(t *testing.T) {
	upd(t, "\uFEFF123,456,789\r\nXYZ", "\uFEFF456,789\r\nXYZ", func(r []Row, m *Mode) {
		r[0].Delete(0)
	})
}

func TestSlicesInsert(t *testing.T) {
	source := []byte{1, 2, 3, 4}
	result := slicesInsert(source, 2, 7)
	expect := []byte{1, 2, 7, 3, 4}
	if !bytes.Equal(expect, result) {
		t.Fatal()
	}
}

func TestSlicesDelete(t *testing.T) {
	source := []byte{1, 2, 3, 4}
	result := slicesDelete(source, 1, 3)
	expect := []byte{1, 4}
	if !bytes.Equal(expect, result) {
		t.Fatal()
	}
}
