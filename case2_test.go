package csvi_test

import (
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
)

func testEnc(t *testing.T, enc encoding.Encoding) {
	t.Helper()
	encoder := enc.NewEncoder()
	src, err := encoder.String(aiueo)
	if err != nil {
		t.Fatal(err.Error())
	}
	op := "j|l|r|foo"
	exp, err := encoder.String(`あ,い,う,え,お
か,foo,く,け,こ
さ,し,す,せ,そ
た,ち,つ,て,と
な,に,ぬ,ね,の
は,ひ,ふ,へ,ほ
ま,み,む,め,も
や,ゆ,よ
ら,り,る,れ,ろ
わ,を,ん`)
	if err != nil {
		t.Fatal(err.Error())
	}
	testCase(t, src, op, exp)
}

func TestUTF16LE(t *testing.T) {
	testEnc(t, unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM))
}

func TestUTF16BE(t *testing.T) {
	testEnc(t, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM))
}

func TestShiftJIS(t *testing.T) {
	testEnc(t, japanese.ShiftJIS)
}
