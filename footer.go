package tuingo

import (
	"strings"

	"github.com/rivo/tview"
)

// Footer is a one-line key legend: each key in bold, its label faint.
type Footer struct {
	*tview.TextView
}

// NewFooter returns an empty legend.
func NewFooter() *Footer {
	return &Footer{tview.NewTextView().SetDynamicColors(true)}
}

// Set replaces the legend with the key and label pairs, in order.
func (f *Footer) Set(pairs ...string) {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, "[::b]"+tview.Escape(pairs[i])+"[::-] [::d]"+tview.Escape(pairs[i+1])+"[::-]")
	}
	f.SetText(strings.Join(parts, "  "))
}

// Legend formats key and label pairs the way the footer does, for use inside other text.
func Legend(pairs ...string) string {
	f := NewFooter()
	f.Set(pairs...)
	return f.GetText(false)
}
