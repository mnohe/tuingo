package tuingo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Row is a row of a Checklist: the key the owner knows it by and the label it shows. A
// header is drawn bold and owns the rows under it up to the next header, which are
// indented; a muted row is drawn dim, for a row whose header is off.
type Row struct {
	Key, Label string
	Header     bool
	Muted      bool
}

// Checklist is a bordered, titled list of rows drawn [ ] and [x], each toggling on and off
// the moment it is chosen. Space toggles the row under the cursor, and enter and esc close
// the dialog; what a toggle means stays with the owner.
type Checklist struct {
	*tview.List
	rows  []Row
	title string
}

// NewChecklist returns an empty dialog; toggle receives the key of the row space is pressed
// on, and close runs on enter or esc.
func NewChecklist(toggle func(key string), close func()) *Checklist {
	c := &Checklist{List: newDialogList()}
	c.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch {
		case ev.Key() == tcell.KeyRune && ev.Rune() == ' ':
			if i := c.GetCurrentItem(); i < len(c.rows) {
				toggle(c.rows[i].Key)
			}
		case ev.Key() == tcell.KeyEnter, ev.Key() == tcell.KeyEsc:
			close()
		default:
			return ev
		}
		return nil
	})
	return c
}

// Show lists the rows under title, ticking those whose key is in checked, with the cursor
// on the first row.
func (c *Checklist) Show(title string, rows []Row, checked map[string]bool) {
	c.title = title
	c.SetTitle(" " + title + " ")
	c.Clear()
	c.rows = nil
	c.Mark(rows, checked)
	c.SetCurrentItem(0)
}

// Mark redraws the rows and their ticks, which may have changed with the last toggle,
// leaving the cursor where it is, or on the last row when there are fewer rows now.
func (c *Checklist) Mark(rows []Row, checked map[string]bool) {
	at := c.GetCurrentItem()
	for c.GetItemCount() > len(rows) {
		c.RemoveItem(c.GetItemCount() - 1)
	}
	for c.GetItemCount() < len(rows) {
		c.AddItem("", "", 0, nil)
	}
	c.rows = rows
	owned := false
	for i, r := range rows {
		mark := "[ ] "
		if checked[r.Key] {
			mark = "[x] "
		}
		text := tview.Escape(mark + r.Label)
		switch {
		case r.Header:
			owned = true
			text = "[::b]" + text + "[::-]"
		case owned:
			text = "  " + text
		}
		if r.Muted {
			text = "[::d]" + text + "[::-]"
		}
		c.SetItemText(i, text, "")
	}
	if len(rows) > 0 {
		c.SetCurrentItem(min(at, len(rows)-1))
	}
}

// Size is the dialog's size on a screen: wide enough for the title and the widest label
// with its box, its indent and the border, one row per row, and never larger than the
// screen. It fits Center.
func (c *Checklist) Size(sw, sh int) (int, int) {
	labels := make([]string, len(c.rows))
	owned := false
	for i, r := range c.rows {
		owned = owned || r.Header
		labels[i] = r.Label
		if owned && !r.Header {
			labels[i] = "  " + r.Label
		}
	}
	return dialogSize(c.title, labels, sw, sh)
}
