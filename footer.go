package tuingo

import (
	"strings"

	"github.com/rivo/tview"
)

// legendGap is the columns between two pairs of a legend row.
const legendGap = 2

// Footer is a one-line key legend: each key in bold, its label faint.
type Footer struct {
	*tview.TextView
	pairs []string
}

// NewFooter returns an empty legend.
func NewFooter() *Footer {
	return &Footer{TextView: tview.NewTextView().SetDynamicColors(true)}
}

// Set replaces the legend with the key and label pairs, in order.
func (f *Footer) Set(pairs ...string) {
	f.pairs = append(f.pairs[:0], pairs...)
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, "[::b]"+tview.Escape(pairs[i])+"[::-] [::d]"+tview.Escape(pairs[i+1])+"[::-]")
	}
	f.SetText(strings.Join(parts, strings.Repeat(" ", legendGap)))
}

// Column is the column, counted from the footer's left edge, at which the footer draws key,
// read from the pairs last passed to Set, or -1 when it lists no such key.
func (f *Footer) Column(key string) int {
	col := 0
	for i := 0; i+1 < len(f.pairs); i += 2 {
		if f.pairs[i] == key {
			return col
		}
		col += columns(f.pairs[i]) + 1 + columns(f.pairs[i+1]) + legendGap
	}
	return -1
}

// columns is the columns s takes on screen, as the footer draws it.
func columns(s string) int { return tview.TaggedStringWidth(tview.Escape(s)) }

// Legend formats key and label pairs the way the footer does, for use inside other text.
func Legend(pairs ...string) string {
	f := NewFooter()
	f.Set(pairs...)
	return f.GetText(false)
}
