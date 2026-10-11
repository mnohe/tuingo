package tuingo

import (
	"time"

	"github.com/rivo/tview"
)

// Notice is a one-line message that clears itself after a while unless replaced.
type Notice struct {
	*tview.TextView
	ttl     time.Duration
	id      int
	redraw  func(func())
	pending *time.Timer
}

// NewNotice returns an empty notice. redraw runs a change on the UI loop and redraws; ttl
// is how long a message stays.
func NewNotice(ttl time.Duration, redraw func(func())) *Notice {
	return &Notice{TextView: tview.NewTextView().SetDynamicColors(true), ttl: ttl, redraw: redraw}
}

// Show replaces the message. An earlier message's expiry no longer clears the line.
func (n *Notice) Show(text string) {
	n.id++
	id := n.id
	n.SetText(text)
	if n.pending != nil {
		n.pending.Stop()
	}
	n.pending = time.AfterFunc(n.ttl, func() {
		n.redraw(func() { n.clear(id) })
	})
}

func (n *Notice) clear(id int) {
	if id == n.id {
		n.SetText("")
	}
}
