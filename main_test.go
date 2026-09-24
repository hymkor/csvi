package csvi

import (
	"container/list"
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/hymkor/csvi/uncsv"
)

// drainAllData must react to context cancellation promptly even when
// the underlying source never delivers another row and the blocking
// fetch would hang forever. It does this by polling tryFetchFunc
// (which has its own short timeout) instead of calling the blocking
// fetchFunc directly.
func TestDrainAllDataRespectsCancellation(t *testing.T) {
	app := &Application{}
	blocked := make(chan struct{})
	app.fetchFunc = func() (*uncsv.Row, error) {
		<-blocked // would hang forever in this test
		return nil, nil
	}
	app.tryFetchFunc = func() (*uncsv.Row, error) {
		select {
		case <-blocked:
			return nil, nil
		case <-time.After(20 * time.Millisecond):
			return nil, os.ErrDeadlineExceeded
		}
	}
	defer close(blocked)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		app.drainAllData(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("drainAllData did not return after context cancellation; it is likely blocked on fetchFunc instead of polling tryFetchFunc")
	}
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("test setup error: context was not actually canceled: %v", ctx.Err())
	}
}

// A genuine (non-EOF, non-timeout) fetch error must be reported back
// to the caller, not swallowed as if the drain had simply finished.
// Otherwise the caller (an uncanceled context, no error) cannot tell
// a partial read apart from a complete one, and an operation like
// sort would silently act on incomplete data.
func TestDrainAllDataReturnsHardError(t *testing.T) {
	app := &Application{csvLines: list.New()}
	wantErr := errors.New("boom")
	calls := 0
	app.tryFetchFunc = func() (*uncsv.Row, error) {
		calls++
		if calls == 1 {
			row := uncsv.NewRow(&uncsv.Mode{})
			return &row, nil
		}
		return nil, wantErr
	}
	app.fetchFunc = app.tryFetchFunc

	err := app.drainAllData(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected drainAllData to return %v, got %v", wantErr, err)
	}
	if app.fetchFunc != nil || app.tryFetchFunc != nil {
		t.Fatal("fetchFunc/tryFetchFunc should be nil-ed out after a hard error")
	}
	if app.Len() != 1 {
		t.Fatalf("expected the one row fetched before the error to still be pushed, got %d rows", app.Len())
	}
}

// A source that ends right at a trailing newline reports a final,
// non-nil but zero-valued row together with io.EOF (the same thing
// happens after the last line during ordinary background loading).
// drainAllData must filter that out like the other loading paths do,
// instead of leaving a spurious empty row in app.csvLines.
func TestDrainAllDataSkipsZeroRowAtEOF(t *testing.T) {
	app := &Application{csvLines: list.New()}
	calls := 0
	app.tryFetchFunc = func() (*uncsv.Row, error) {
		calls++
		if calls == 1 {
			row := uncsv.NewRow(&uncsv.Mode{})
			return &row, nil
		}
		return &uncsv.Row{Cell: []uncsv.Cell{{}}}, io.EOF
	}
	app.fetchFunc = app.tryFetchFunc

	if err := app.drainAllData(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Len() != 1 {
		t.Fatalf("expected the phantom zero row at EOF to be filtered out, got %d rows", app.Len())
	}
}
