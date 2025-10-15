package csvi

import (
	"testing"
)

func TestCwParse(t *testing.T) {
	cw := NewCellWidth()
	err := cw.Parse("18")
	if err != nil {
		t.Fatal(err.Error())
	}
	if cw.Default != 18 {
		t.Fatalf("cw.Default=%d", cw.Default)
	}
	if len(cw.Option) != 0 {
		t.Fatalf("len(cw.Option)=%d", len(cw.Option))
	}

	cw = NewCellWidth()
	err = cw.Parse("13,7:10")
	if err != nil {
		t.Fatal(err.Error())
	}
	if w := cw.Get(1); w != 13 {
		t.Fatalf("Get(1) = %d", w)
	}
	if w := cw.Get(7); w != 10 {
		t.Fatalf("Get(7) = %d", w)
	}
}
