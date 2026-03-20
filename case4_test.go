package csvi_test

import (
	"path/filepath"
	"testing"

	"github.com/hymkor/csvi/uncsv"
)

func TestUndo1(t *testing.T) {
	op := "<|r|あ'|u"
	exp := `あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,ほ
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん`
	testCase(t, aiueo, op, exp)
}

func TestUndo2(t *testing.T) {
	dummy := filepath.Join(t.TempDir(), "test.csv")
	op := "<|r|あ+|w|" + dummy + "|r|あ++|u"
	//println(op)
	exp := `あ+,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,ほ
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん`
	testCase(t, aiueo, op, exp)
}

func TestZeroLines(t *testing.T) {
	exp := "1,2,3" + uncsv.OsNewline + "4,5" + uncsv.OsNewline
	testCase(t, "", "i|1|a|2|a|3|o|4|a|5", exp)
}
