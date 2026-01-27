package csvi_test

import (
	"testing"
)

func TestDeleteCell(t *testing.T) { // `x`
	src := "あ,い,う,え,お\nか,き,く,け,こ"
	op := "<|d| "
	exp := "い,う,え,お\nか,き,く,け,こ"

	testCase(t, src, op, exp)
	testCase(t, src, op, src, "-fixcol")   // can not update
	testCase(t, src, op, src, "-readonly") // can not update
}

func TestDeleteRow(t *testing.T) { // `dd`
	src := "あ,い,う,え,お\nか,き,く,け,こ"
	op := "<|d|d"
	exp := "か,き,く,け,こ"

	testCase(t, src, op, exp)
	testCase(t, src, op, exp, "-fixcol")
	testCase(t, src, op, src, "-readonly") // can not update
}
func TestDeleteColumn(t *testing.T) { // `dc`
	src := "あ,い,う,え,お\nか,き,く,け,こ"
	op := "<|d|c"
	exp := "い,う,え,お\nき,く,け,こ"

	testCase(t, src, op, exp)
	testCase(t, src, op, src, "-fixcol")   // can not update
	testCase(t, src, op, src, "-readonly") // can not update
}

func TestCopyPasteCell(t *testing.T) { // `yl` and `p`
	src := "あ,い,う,え,お\nか,き,く,け,こ"
	op := "<|y|l|l|p"
	exp := "あ,い,あ,う,え,お\nか,き,く,け,こ"

	testCase(t, src, op, exp)
	testCase(t, src, op, src, "-fixcol")   // can not update
	testCase(t, src, op, src, "-readonly") // can not update
}

func TestCopyPasteCellB(t *testing.T) { // `yl` and `P`
	src := "あ,い,う,え,お\nか,き,く,け,こ"
	op := "<|y|l|$|P"
	exp := "あ,い,う,え,あ,お\nか,き,く,け,こ"

	testCase(t, src, op, exp)
	testCase(t, src, op, src, "-fixcol")   // can not update
	testCase(t, src, op, src, "-readonly") // can not update
}

func TestCopyPasteCellOver(t *testing.T) { // `yl` and `ALT`|`ESC`+`p`
	src := "あ,い,う,え,お\nか,き,く,け,こ"
	op := "<|y|l|$|\x1B|p" // ESC-p
	exp := "あ,い,う,え,あ\nか,き,く,け,こ"

	testCase(t, src, op, exp)
	testCase(t, src, op, exp, "-fixcol")
	testCase(t, src, op, src, "-readonly") // can not update

	op = "<|y|l|$|\x1Bp" // ALT-p
	testCase(t, src, op, exp)
}

func TestCopyPasteRow(t *testing.T) { // `yy` and `p`
	src := "あ,い,う,え,お\nか,き,く,け,こ"
	op := "<|y|y|>|p"
	exp := "あ,い,う,え,お\nか,き,く,け,こ\nあ,い,う,え,お\n"

	testCase(t, src, op, exp)
	testCase(t, src, op, exp, "-fixcol")
	testCase(t, src, op, src, "-readonly") // can not update
}

func TestCopyPasteRowB(t *testing.T) { // `yy` and `P`
	src := "あ,い,う,え,お\r\nか,き,く,け,こ"
	op := "<|y|y|P"
	exp := "あ,い,う,え,お\r\nあ,い,う,え,お\r\nか,き,く,け,こ"

	testCase(t, src, op, exp)
	testCase(t, src, op, exp, "-fixcol")
	testCase(t, src, op, src, "-readonly") // can not update
}

func TestCopyPasteRowOver(t *testing.T) { // `yy` and `ALT`/`ESC`+`p`
	src := "あ,い,う,え,お\r\nか,き,く,け,こ"
	op := "<|y|y|\r|\x1B|p" // ESC-p
	exp := "あ,い,う,え,お\r\nあ,い,う,え,お"

	testCase(t, src, op, exp)
	testCase(t, src, op, exp, "-fixcol")
	testCase(t, src, op, src, "-readonly") // can not update

	op = "<|y|y|\r|\x1Bp" // ALT-p
	testCase(t, src, op, exp)
}

func TestCopyPasteColumn(t *testing.T) { // `yc` and `p`
	src := "あ,い,う,え,お\nか,き,く,け,こ"
	op := "<|y|c|$|p"
	exp := "あ,い,う,え,お,あ\nか,き,く,け,こ,か"

	testCase(t, src, op, exp)
	testCase(t, src, op, src, "-fixcol")   // can not update
	testCase(t, src, op, src, "-readonly") // can not update
}

func TestCopyPasteColumnB(t *testing.T) { // `yc` and `P`
	src := "あ,い,う,え,お\nか,き,く,け,こ"
	op := "<|y|c|$|P"
	exp := "あ,い,う,え,あ,お\nか,き,く,け,か,こ"

	testCase(t, src, op, exp)
	testCase(t, src, op, src, "-fixcol")   // can not update
	testCase(t, src, op, src, "-readonly") // can not update
}

func TestCopyPasteColumnOver(t *testing.T) { // `yc` and `ALT`/`Esc`+`p`
	src := "あ,い,う,え,お\nか,き,く,け,こ"
	op := "<|y|c|$|\x1B|p" // Esc-p
	exp := "あ,い,う,え,あ\nか,き,く,け,か"

	testCase(t, src, op, exp)
	testCase(t, src, op, exp, "-fixcol")
	testCase(t, src, op, src, "-readonly") // can not update

	op = "<|y|c|$|\x1Bp" // Alt-P
	testCase(t, src, op, exp)
}
