package csvi_test

import (
	"bytes"
	"strings"
	"testing"
)

var sortSource = `name,score
banana,30
apple,10
cherry,2`

func TestSortAscendingByColumn(t *testing.T) {
	exp := `name,score
cherry,2
apple,10
banana,30
`
	testCase(t, sortSource, "l|s", exp)
}

func TestSortDescendingByColumn(t *testing.T) {
	exp := `name,score
banana,30
apple,10
cherry,2`
	testCase(t, sortSource, "l|S", exp)
}

func TestSortDefaultColumn(t *testing.T) {
	exp := `name,score
apple,10
banana,30
cherry,2`
	testCase(t, sortSource, "s", exp)
}

// The cursor should keep pointing at the same logical row after a sort,
// not at whatever row happens to land at the old screen position.
func TestSortCursorFollowsRow(t *testing.T) {
	src := "name,score\nbanana,30\napple,10\ncherry,2\n"
	exp := `name,score
apple,10
EDITED,30
cherry,2
`
	testCase(t, src, "j|s|r|EDITED", exp)
}

// "9" must sort before "10" numerically, not lexically.
func TestSortNumericAware(t *testing.T) {
	src := "name,score\na,9\nb,10\nc,2"
	exp := `name,score
c,2
a,9
b,10
`
	testCase(t, src, "l|s", exp)
}

// Sorting must not leave more than one row with an empty Term: only
// the row that ends up last may have no trailing newline.
func TestSortRelocatesMissingNewline(t *testing.T) {
	src := "name,score\nbanana,30\napple,10\ncherry,2" // no trailing newline
	exp := `name,score
cherry,2
apple,10
banana,30
`
	testCase(t, src, "l|s", exp)
}

func TestSortHeaderNotReordered(t *testing.T) {
	src := "z,score\nbanana,30\napple,10\ncherry,2\n"
	exp := `z,score
apple,10
banana,30
cherry,2
`
	// header cell "z" would sort after the data if it were included
	testCase(t, src, "s", exp)
}

func TestSortReadOnlyIsBlocked(t *testing.T) {
	exp := sortSource
	testCase(t, sortSource, "l|s", exp, "-readonly")
}

func TestSortNoHeader(t *testing.T) {
	src := "banana,30\napple,10\ncherry,2\n"
	exp := `apple,10
banana,30
cherry,2
`
	testCase(t, src, "s", exp, "-h", "0")
}

// Inf and exponential notation are still valid numbers and must sort
// numerically as long as the column has no NaN or non-numeric values.
func TestSortInfAndExponent(t *testing.T) {
	src := "name,val\na,10\ne,-Inf\nc,2\nf,1e2\nd,Inf\n"
	exp := `name,val
e,-Inf
c,2
a,10
f,1e2
d,Inf
`
	testCase(t, src, "l|s", exp)
}

// NaN compares false against everything under IEEE754, so it has no
// position in the numeric order at all. It is treated as a missing
// value (like a blank cell) instead: the rest of the column still
// sorts numerically and the NaN goes last.
func TestSortNaNSortsLastLikeBlank(t *testing.T) {
	src := "name,val\na,10\nb,NaN\nc,2\nd,Inf\ne,-Inf\nf,1e2\n"
	exp := `name,val
e,-Inf
c,2
a,10
f,1e2
d,Inf
b,NaN
`
	testCase(t, src, "l|s", exp)
}

// Descending too: a missing value sorts last regardless of direction,
// so NaN must not float to the top.
func TestSortNaNSortsLastDescending(t *testing.T) {
	src := "name,val\na,10\nb,NaN\nc,2\nd,30\n"
	exp := `name,val
d,30
a,10
c,2
b,NaN
`
	testCase(t, src, "l|S", exp)
}

// NaN and blank cells are both missing values: they end up together at
// the bottom, keeping their relative order (stable sort). "nan" is
// matched case-insensitively, as strconv.ParseFloat does.
func TestSortNaNAndBlankAreBothMissing(t *testing.T) {
	src := "name,val\na,10\nb,nan\nc,\nd,2\ne,NAN\n"
	exp := `name,val
d,2
a,10
b,nan
c,
e,NAN
`
	testCase(t, src, "l|s", exp)
}

// In a column that is not numeric, "NaN" is not a missing-number
// marker but ordinary text, and must be compared as such.
func TestSortNaNInTextColumnIsPlainText(t *testing.T) {
	src := "name,val\na,banana\nb,NaN\nc,apple\nd,\n"
	exp := `name,val
b,NaN
c,apple
a,banana
d,
`
	testCase(t, src, "l|s", exp)
}

// Mixing numeric and non-numeric values in one column can otherwise
// create a cycle under a per-pair comparator: "9" < "80" numerically,
// "80" < "85x" as strings, but "85x" < "9" as strings too. The column
// must fall back to a plain, consistent string sort as a whole.
func TestSortMixedNumericAndTextFallsBackToString(t *testing.T) {
	src := "name,val\na,9\nb,80\nc,85x\n"
	exp := `name,val
b,80
c,85x
a,9
`
	testCase(t, src, "l|s", exp)
}

// Blank cells must still sort last even when the rest of the column is
// text, not just when it is numeric: the string-comparison branch
// must not let "" (the lexically smallest value) sort first.
func TestSortBlankStaysLastInStringMode(t *testing.T) {
	src := "name,val\na,banana\nb,\nc,apple\n"
	exp := `name,val
c,apple
a,banana
b,
`
	testCase(t, src, "l|s", exp)
}

// Blank cells (missing values) are common in real CSVs and must not
// disable numeric sorting for the rest of the column. They sort after
// every non-blank value, regardless of direction.
func TestSortBlankStaysLastAscending(t *testing.T) {
	src := "name,score\na,30\nb,\nc,2\nd,10\n"
	exp := `name,score
c,2
d,10
a,30
b,
`
	testCase(t, src, "l|s", exp)
}

func TestSortBlankStaysLastDescending(t *testing.T) {
	src := "name,score\na,30\nb,\nc,2\nd,10\n"
	exp := `name,score
a,30
d,10
c,2
b,
`
	testCase(t, src, "l|S", exp)
}

// Multiple blanks keep their relative order (stable sort).
func TestSortMultipleBlanksStayInOrder(t *testing.T) {
	src := "name,score\na,30\nb,\nc,2\nd,\ne,10\n"
	exp := `name,score
c,2
e,10
a,30
b,
d,
`
	testCase(t, src, "l|s", exp)
}

// Rows with equal (non-blank) values must keep their original relative
// order after sorting.
func TestSortStableWithDuplicateValues(t *testing.T) {
	src := "name,score\na,10\nb,5\nc,10\nd,5\ne,10\n"
	exp := `name,score
b,5
d,5
a,10
c,10
e,10
`
	testCase(t, src, "l|s", exp)
}

// Sorting a column that is already in the requested order (a stable
// sort leaves every row in place) must not mark the file dirty. If it
// did, quitting would ask to save changes that do not actually exist;
// we detect that by checking whether the "Save changes" prompt was
// ever written, since a clean quit skips it entirely.
func TestSortNoOpDoesNotPromptToSaveOnQuit(t *testing.T) {
	src := "name,score\napple,10\nbanana,30\ncherry,40\n" // already ascending by name
	opt, err := newTestOptions("-auto", "s|q")
	if err != nil {
		t.Fatal(err.Error())
	}
	var out bytes.Buffer
	err = opt.RunInOut(strings.NewReader(src), &out)
	if err != nil {
		t.Fatalf("unexpected error (likely means the session got stuck on an unexpected prompt): %v", err)
	}
	if strings.Contains(out.String(), "Save changes") {
		t.Fatal("a no-op sort left the file marked dirty: quit prompted to save changes")
	}
}

// Columns are often padded, and padding is not part of the value: a
// padded numeric column must still sort numerically ("9" before "10"),
// and the cells must be written back exactly as they were read.
func TestSortIgnoresSurroundingSpaces(t *testing.T) {
	src := "name,val\na, 10\nb,9 \nc, 2 \n"
	exp := `name,val
c, 2 
b,9 
a, 10
`
	testCase(t, src, "l|s", exp)
}

// A cell holding nothing but padding carries no value, so it is
// missing like a blank one and must not turn the column into text.
func TestSortWhitespaceOnlyCellIsMissing(t *testing.T) {
	src := "name,val\na,30\nb,   \nc,2\nd,10\n"
	exp := `name,val
c,2
d,10
a,30
b,   
`
	testCase(t, src, "l|s", exp)
}

// The comparison mode is chosen from the data, not by the user, so the
// status line has to say which one was used: a column that sorts as
// text because of one stray non-number otherwise just looks wrongly
// sorted.
func TestSortReportsWhatItDid(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
		auto string
		want string
	}{
		{
			name: "numeric",
			src:  "name,score\nbanana,30\napple,10\ncherry,2\n",
			auto: "l|s|q|n",
			want: "Sorted by column 2 (numeric, ASC)",
		},
		{
			name: "text",
			src:  "name,score\nbanana,30\napple,x\ncherry,2\n",
			auto: "l|S|q|n",
			want: "Sorted by column 2 (text, DESC)",
		},
		{
			name: "already sorted",
			src:  "name,score\napple,10\nbanana,30\ncherry,40\n",
			auto: "s|q",
			want: "Already sorted by column 1 (text, ASC)",
		},
		{
			name: "single data row",
			src:  "name,score\napple,10\n",
			auto: "l|s|q",
			want: "Already sorted by column 2 (numeric, ASC)",
		},
		{
			name: "no data rows",
			src:  "name,score\n",
			auto: "s|q",
			want: "No rows to sort",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opt, err := newTestOptions("-auto", tt.auto)
			if err != nil {
				t.Fatal(err.Error())
			}
			var out bytes.Buffer
			if err := opt.RunInOut(strings.NewReader(tt.src), &out); err != nil {
				t.Fatalf("unexpected error (likely means the session got stuck on an unexpected prompt): %v", err)
			}
			if !strings.Contains(out.String(), tt.want) {
				t.Errorf("status line did not report %q", tt.want)
			}
		})
	}
}
