package csvi_test

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestCase3(t *testing.T) {
	path1 := makeSource(t, "t1.csv", "first\n")
	path2 := makeSource(t, "t2.csv", "second\n")
	path3 := makeSource(t, "t3.csv", "third\n")
	outputPath := filepath.Join(t.TempDir(), "t4.csv")

	instance, err := newTestOptions("-auto", fmt.Sprintf("w|%s|q|y", outputPath), path1, path2, path3)

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
