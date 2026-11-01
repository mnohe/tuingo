package tuingo

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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

func TestTableNarrowLayoutPreservesCellsWithoutBlankRows(t *testing.T) {
	table := Table{Rows: [][]string{{"long", "text"}, {"界", "x"}}}
	lines := table.Render(1)
	var plain strings.Builder
	for _, line := range lines {
		if tview.TaggedStringWidth(line) == 0 {
			t.Fatalf("spurious blank row: %q", lines)
		}
		plain.WriteString(tview.NewTextView().SetDynamicColors(true).SetText(line).GetText(true))
	}
	if got := strings.ReplaceAll(plain.String(), " ", ""); got != "longtext�x" {
		t.Fatalf("lost content: %q", got)
	}
	for _, width := range []int{0, -1} {
		if got := table.Render(width); len(got) != 0 {
			t.Fatalf("width %d: %q", width, got)
		}
	}
}

func TestTableStyledCellsAndContinuation(t *testing.T) {
	v := NewTextBox(func(rune) {}, func() {})
	v.ShowBlocks("", []Block{Table{Rows: [][]string{
		{"[red::b]A", "[green::u]one two three"},
		{"B", "plain"},
	}}}, nil, "")
	s := drawn(t, v, 18) // Inner width 12, last column width 9.
	x, y, _, _ := v.GetInnerRect()
	_, _, style, _ := s.GetContent(x, y)
	if styleColor(style) != tcell.ColorRed || styleNotBold(style) {
		t.Fatalf("first cell style: %v", style)
	}
	_, _, style, _ = s.GetContent(x+3, y+1)
	if styleColor(style) != tcell.ColorGreen || style.GetUnderlineStyle() == tcell.UnderlineStyleNone {
		t.Fatalf("wrapped style: %v", style)
	}
	_, _, style, _ = s.GetContent(x, y+2)
	if styleColor(style) == tcell.ColorRed || style.GetUnderlineStyle() != tcell.UnderlineStyleNone {
		t.Fatalf("style leaked to next row: %v", style)
	}
}

func styleColor(style tcell.Style) tcell.Color { fg, _, _ := style.Decompose(); return fg }
func styleNotBold(style tcell.Style) bool      { _, _, a := style.Decompose(); return a&tcell.AttrBold == 0 }
