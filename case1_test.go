package csvi_test

import (
	"testing"
)

var aiueo = `あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,ほ
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん`

func TestCaseSmallI(t *testing.T) {
	op := "<|i|new"
	exp := `new,あ,い,う,え,お
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

func TestCaseSmallA(t *testing.T) {
	op := "<|$|a|new"
	exp := `あ,い,う,え,お,new
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

func TestTailSmallX(t *testing.T) {
	op := "<|$|d| "
	exp := `あ,い,う,え
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

func TestHeadSmallX(t *testing.T) {
	op := "<|d| "
	exp := `い,う,え,お
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

func TestLastSmallO(t *testing.T) {
	op := ">|o|ががが"
	exp := `あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,ほ
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん
ががが`
	testCase(t, aiueo, op, exp)
}

func TestLastLargeO(t *testing.T) {
	op := ">|O|ががが"
	exp := `あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,ほ
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
ががが
わ,を,ん`
	testCase(t, aiueo, op, exp)
}

func TestHeadLargeO(t *testing.T) {
	op := "<|O|ががが"
	exp := `ががが
あ,い,う,え,お
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

func TestHeadSmallO(t *testing.T) {
	op := "<|o|ががが"
	exp := `あ,い,う,え,お
ががが
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

func TestTailD(t *testing.T) {
	op := ">|D"
	exp := `あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,ほ
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ`
	testCase(t, aiueo, op, exp)
}

func TestHeadLargeD(t *testing.T) {
	op := "<|D"
	exp := `か,き,く,け,こ
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

func TestHeadSmallR(t *testing.T) {
	op := "<|r|ぎ\nゃあ"
	exp := `"ぎ
ゃあ",い,う,え,お
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

func TestSearchDownInsert(t *testing.T) {
	op := "/|ま|$|j|i|foo"
	exp := `あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,ほ
ま,み,む,め,も
や,ゆ,foo,よ
ら,り,る,れ,ろ
わ,を,ん`
	testCase(t, aiueo, op, exp)
}

func TestSearchTailReplace(t *testing.T) {
	op := "/|ま|$|j|r|foo"
	exp := `あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,ほ
ま,み,む,め,も
や,ゆ,foo
ら,り,る,れ,ろ
わ,を,ん`
	testCase(t, aiueo, op, exp)
}

func TestSearchTailDownAppend(t *testing.T) {
	op := "/|ま|$|j|a|foo"
	exp := `あ,い,う,え,お
か,き,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,ほ
ま,み,む,め,も
や,ゆ,よ,foo
ら,り,る,れ,ろ
わ,を,ん`
	testCase(t, aiueo, op, exp)
}
