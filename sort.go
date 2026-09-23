package csvi

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/hymkor/csvi/uncsv"
)

// cmdSort sorts the rows below the header by the col-th column and
// returns a message for the status line.
func (app *Application) cmdSort(cfg *Config, ascending bool) string {
	if cfg.ReadOnly {
		return msgReadOnly
	}
	ctx, cancel := app.withSlowOperation("Reading all data...")
	drainErr := app.drainAllData(ctx)
	interrupted := ctx.Err() != nil
	cancel()
	if interrupted {
		return "Sort interrupted"
	}
	if drainErr != nil {
		return drainErr.Error()
	}
	msg, changed := app.sortByColumn(app.cursorCol, ascending)
	if changed {
		app.repaint()
		app.clearCache()
		app.setHardDirty()
	}
	return msg
}

func cellTextAt(row *uncsv.Row, col int) string {
	if col < 0 || col >= len(row.Cell) {
		return ""
	}
	return row.Cell[col].Text()
}

// numericCell pairs a row with the already-parsed value of the cell
// being sorted on, so the comparator never calls strconv.
type numericCell struct {
	row *uncsv.Row
	num float64
}

// splitNumericColumn scans the col-th column and splits the rows into
// ones with a number and ones without (blank or NaN, both treated as
// missing). It reports numeric=false as soon as it meets a value that
// isn't a number at all, since the column then has to be compared as
// text: deciding numeric vs. text pair by pair is not a consistent
// order (e.g. "9"<"80" numerically but "80"<"85x"<"9" as strings).
func splitNumericColumn(rows []*uncsv.Row, col int) (values []numericCell, missing []*uncsv.Row, numeric bool) {
	for i, r := range rows {
		text := strings.TrimSpace(cellTextAt(r, col))
		if text == "" {
			missing = append(missing, r)
			continue
		}
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, nil, false
		}
		if math.IsNaN(f) {
			missing = append(missing, r)
			continue
		}
		if values == nil {
			values = make([]numericCell, 0, len(rows)-i)
		}
		values = append(values, numericCell{row: r, num: f})
	}
	return values, missing, true
}

// sortedRows returns rows reordered by the col-th column, leaving the
// rows slice itself untouched so the caller can tell whether anything
// actually moved, and reports whether the column was compared as
// numbers. The sort is stable, and rows with no value (blank, or NaN
// in a numeric column) sort last regardless of direction.
func sortedRows(rows []*uncsv.Row, col int, ascending bool) ([]*uncsv.Row, bool) {
	if values, missing, numeric := splitNumericColumn(rows, col); numeric {
		sort.SliceStable(values, func(i, j int) bool {
			if ascending {
				return values[i].num < values[j].num
			}
			return values[j].num < values[i].num
		})
		sorted := make([]*uncsv.Row, 0, len(rows))
		for i := range values {
			sorted = append(sorted, values[i].row)
		}
		return append(sorted, missing...), true
	}

	// Unlike the numeric path, this does not trim: padding around text
	// is part of the text, not formatting around a value.
	sorted := make([]*uncsv.Row, 0, len(rows))
	var blank []*uncsv.Row
	for _, r := range rows {
		if cellTextAt(r, col) == "" {
			blank = append(blank, r)
		} else {
			sorted = append(sorted, r)
		}
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		ti := cellTextAt(sorted[i], col)
		tj := cellTextAt(sorted[j], col)
		if ascending {
			return ti < tj
		}
		return tj < ti
	})
	return append(sorted, blank...), false
}

// sortByColumn sorts the data rows (below app.HeaderLines) by the
// col-th cell, leaving the header in place, and moves the cursor to
// follow the row it was on. It returns a message for the status line
// and reports whether it actually reordered anything.
func (app *Application) sortByColumn(col int, ascending bool) (string, bool) {
	header := app.HeaderLines
	if header < 0 {
		header = 0
	}
	first := app.Front()
	for i := 0; i < header && first != nil; i++ {
		first = first.Next()
	}
	if first == nil {
		return "No rows to sort", false
	}

	ptrs := make([]*RowPtr, 0, app.Len()-header)
	for p := first; p != nil; p = p.Next() {
		ptrs = append(ptrs, p)
	}
	rows := make([]*uncsv.Row, len(ptrs))
	for i, p := range ptrs {
		rows[i] = p.Row
	}

	sorted, numeric := sortedRows(rows, col, ascending)
	report := sortReport(col, ascending, numeric)

	// A stable sort of an already-sorted (or all-equal) column leaves
	// every row in place; report no change instead of dirtying the
	// file over a no-op.
	changed := false
	for i := range rows {
		if rows[i] != sorted[i] {
			changed = true
			break
		}
	}
	if !changed {
		return "Already sorted " + report, false
	}

	cursorRow := app.cursorRow.Row
	for _, p := range ptrs {
		app.csvLines.Remove(p.element)
	}

	// At most one row can have an empty Term (the one that used to sit
	// at the end of the file). If sorting moved it away from the last
	// position, it needs a real terminator or it merges with the next.
	last := len(sorted) - 1
	for i, r := range sorted {
		if i == last || r.Term != "" {
			continue
		}
		r.Term = app.Config.Mode.DefaultTerm
		if r.Term == "" {
			r.Term = uncsv.OsNewline
		}
	}

	for _, r := range sorted {
		app.csvLines.PushBack(r)
	}

	app.startRow = app.Front()
	app.cursorRow = app.Front()
	for p := app.cursorRow; p != nil; p = p.Next() {
		if p.Row == cursorRow {
			app.cursorRow = p
			break
		}
	}
	return "Sorted " + report, true
}

// sortReport describes a sort as "by column 2 (numeric, DESC)".
func sortReport(col int, ascending, numeric bool) string {
	kind := "text"
	if numeric {
		kind = "numeric"
	}
	order := "ASC"
	if !ascending {
		order = "DESC"
	}
	return fmt.Sprintf("by column %d (%s, %s)", col+1, kind, order)
}
