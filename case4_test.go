package csvi_test

import (
	"os"
	"testing"
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
	op := "<|r|あ'|w|" + os.DevNull + "|r|あ''|u"
	exp := `あ',い,う,え,お
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
