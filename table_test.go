package tuingo

import (
	"github.com/rivo/tview"
	"strings"
	"testing"
)

func TestTableHangingIndentAndWidths(t *testing.T) {
	table := Table{Rows: [][]string{{"[::b]界[::B]", "one two three four"}, {"a", "done"}}}
	lines := table.Render(12)
	if len(lines) < 3 || !strings.HasPrefix(lines[1], "    ") {
		t.Fatalf("no hanging indent: %q", lines)
	}
	for _, line := range lines {
		if tview.TaggedStringWidth(line) > 12 {
			t.Fatalf("line too wide: %q", line)
		}
	}
	if len(table.Render(24)) >= len(lines) {
		t.Fatal("table did not rewrap")
	}
	for w := 1; w < 12; w++ {
		for _, line := range table.Render(w) {
			if tview.TaggedStringWidth(line) > w {
				t.Fatalf("width %d: %q", w, line)
			}
		}
	}
}
func TestTextBoxBlocksRerenderOnResize(t *testing.T) {
	v := NewTextBox(func(rune) {}, func() {})
	v.ShowBlocks("History", []Block{Lines{"Heading"}, Table{Rows: [][]string{{"time", "a comment long enough to wrap at narrow widths"}}}}, nil, "")
	drawn(t, v, 30)
	narrow := v.text.GetText(false)
	drawn(t, v, 70)
	if narrow == v.text.GetText(false) {
		t.Fatal("resize did not render content")
	}
}
