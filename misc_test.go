package csvi

import (
	"testing"
)

func TestCutStrInWidth(t *testing.T) {
	list := []struct {
		source string
		w      int
		expect string
	}{
		{source: "あいうえお", w: 80, expect: "あいうえお"},
		{source: "あいうえお", w: 3, expect: "あ"},
		{source: "\x1B[7mあい\x1B[0mうえお", w: 3, expect: "\x1B[7mあ\x1B[0m"},
	}

	for _, p := range list {
		result, _ := cutStrInWidth(p.source, p.w)
		if result != p.expect {
			t.Fatalf("source: %s & %d, expect: %s, but result:%s\x1B[0m",
				p.source, p.w, p.expect, result)
		}
	}
}
