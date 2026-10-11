package tuingo

import (
	"strings"

	"github.com/rivo/tview"
)

// Tabs is a row of labels with one active, and a stamp at the right edge.
type Tabs struct {
	*tview.Flex
	labels *tview.TextView
	stamp  *tview.TextView
}

// NewTabs returns an empty row.
func NewTabs() *Tabs {
	t := &Tabs{
		Flex:   tview.NewFlex(),
		labels: tview.NewTextView().SetDynamicColors(true),
		stamp:  tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight),
	}
	t.AddItem(t.labels, 0, 1, false).AddItem(t.stamp, 16, 0, false)
	return t
}

// Set replaces the labels, marking the one at active.
func (t *Tabs) Set(labels []string, active int) {
	var parts []string
	for i, l := range labels {
		if i == active {
			parts = append(parts, "[::bu] "+tview.Escape(l)+" [::-]")
		} else {
			parts = append(parts, "[::d] "+tview.Escape(l)+" [::-]")
		}
	}
	t.labels.SetText(strings.Join(parts, ""))
}

// SetStamp sets the text at the right edge.
func (t *Tabs) SetStamp(text string) {
	t.stamp.SetText("[::d]" + tview.Escape(text) + "[::-]")
}
