package tuingo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Option is a row of a Choice: the key the owner knows it by and the label it shows.
type Option struct {
	Key, Label string
}

// Choice is a bordered, titled radio list in which exactly one row is current, drawn ( )
// and (•). Space picks the row under the cursor, and enter and esc close the dialog; what a
// pick means stays with the owner.
type Choice struct {
	*tview.List
	options []Option
	title   string
}

// NewChoice returns an empty dialog; pick receives the key of the row space is pressed on,
// and close runs on enter or esc.
func NewChoice(pick func(key string), close func()) *Choice {
	c := &Choice{List: newDialogList()}
	c.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch {
		case ev.Key() == tcell.KeyRune && ev.Rune() == ' ':
			if i := c.GetCurrentItem(); i < len(c.options) {
				pick(c.options[i].Key)
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

// Show lists the options under title, with the filled button and the cursor on the row
// whose key is current.
func (c *Choice) Show(title string, options []Option, current string) {
	c.title, c.options = title, options
	c.SetTitle(" " + title + " ")
	c.Clear()
	at := 0
	for i, o := range options {
		if o.Key == current {
			at = i
		}
		c.AddItem("", "", 0, nil)
	}
	c.Mark(current)
	c.SetCurrentItem(at)
}

// Mark moves the filled button to the row whose key is current, leaving the cursor where it
// is.
func (c *Choice) Mark(current string) {
	for i, o := range c.options {
		mark := "( ) "
		if o.Key == current {
			mark = "(•) "
		}
		c.SetItemText(i, mark+tview.Escape(o.Label), "")
	}
}

// Size is the dialog's size on a screen: wide enough for the title and the widest label
// with its button and the border, one row per option, and never larger than the screen. It
// fits Center.
func (c *Choice) Size(sw, sh int) (int, int) {
	labels := make([]string, len(c.options))
	for i, o := range c.options {
		labels[i] = o.Label
	}
	return dialogSize(c.title, labels, sw, sh)
}

// newDialogList returns the list Choice and Checklist are built on: bordered, titled on the
// left, one line per row, the cursor's row highlighted across its width.
func newDialogList() *tview.List {
	l := tview.NewList().ShowSecondaryText(false).SetHighlightFullLine(true)
	l.SetBorder(true).SetTitleAlign(tview.AlignLeft).SetBorderPadding(0, 0, 1, 1)
	return l
}

// dialogSize is the size of a dialog list on a screen: the title with a space either side
// and a column of border either side, or the widest label with its four-column mark and
// the border and padding, whichever is wider; one row per label and the border;
// and never larger than the screen.
func dialogSize(title string, labels []string, sw, sh int) (int, int) {
	w := Width(title) + 4
	for _, l := range labels {
		w = max(w, Width(l)+8)
	}
	return min(w, sw), min(len(labels)+2, sh)
}
