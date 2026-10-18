package tuingo

import (
	"strings"

	"github.com/rivo/tview"
)

// KeyPanel is a bordered column of keys, one key and label per line in the legend's style.
// It shows keys and handles none; the screen that opens it runs them.
type KeyPanel struct {
	*tview.TextView
	widest, lines int
}

// NewKeyPanel returns an empty panel.
func NewKeyPanel() *KeyPanel {
	p := &KeyPanel{TextView: tview.NewTextView().SetDynamicColors(true)}
	p.SetBorder(true).SetBorderPadding(0, 0, 1, 1)
	return p
}

// Set replaces the panel's lines with the key and label pairs, in order.
func (p *KeyPanel) Set(pairs ...string) {
	var rows []string
	p.widest = 0
	for i := 0; i+1 < len(pairs); i += 2 {
		rows = append(rows, Legend(pairs[i], pairs[i+1]))
		if w := Width(pairs[i]) + 1 + Width(pairs[i+1]); w > p.widest {
			p.widest = w
		}
	}
	p.lines = len(rows)
	p.SetText(strings.Join(rows, "\n"))
}

// Size is the panel's size around its lines: the widest line and one row per line, plus the
// border and a column of padding either side.
func (p *KeyPanel) Size() (int, int) {
	return p.widest + 4, p.lines + 2
}
