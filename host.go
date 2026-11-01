package tuingo

import (
	"slices"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// rootPage is the page of a Host that holds the application's own screen.
const rootPage = "tuingo.root"

// confirmPage is the page a Host opens a Confirm on.
const confirmPage = "tuingo.confirm"

// Host sits at an application's root, wrapping its screen, and owns the dialogs the
// application opens, so a dialog stays modal whatever the application does with focus.
// While a dialog is open the host hands every key and mouse event to the topmost dialog,
// taking focus back for it first when the application sent focus elsewhere. Closing a
// dialog restores focus to the primitive that had it when the dialog opened.
type Host struct {
	*tview.Pages
	app  *tview.Application
	open []hosted
}

// hosted is an open dialog: its page name, the primitive it draws, and the primitive that
// had focus when it opened.
type hosted struct {
	name       string
	item, prev tview.Primitive
}

// NewHost returns a host over root, the application's screen, for app. Pass the host to
// app.SetRoot in root's place.
func NewHost(app *tview.Application, root tview.Primitive) *Host {
	h := &Host{Pages: tview.NewPages(), app: app}
	h.AddPage(rootPage, root, true, true)
	return h
}

// Open shows p as a dialog over everything open so far, under name, and gives it focus.
// Opening a name that is already open replaces that dialog and keeps the focus it will
// restore. p draws itself where it likes; Center and Anchor place it.
func (h *Host) Open(name string, p tview.Primitive) {
	prev := h.app.GetFocus()
	if i := h.index(name); i >= 0 {
		prev = h.forget(i).prev
	}
	h.open = append(h.open, hosted{name: name, item: p, prev: prev})
	h.AddPage(name, p, true, true)
	h.app.SetFocus(p)
}

// Close removes the dialog open under name. When it was the topmost, focus goes back to
// the primitive that had it when the dialog opened.
func (h *Host) Close(name string) {
	i := h.index(name)
	if i < 0 {
		return
	}
	d := h.forget(i)
	h.RemovePage(name)
	if i < len(h.open) {
		return
	}
	if d.prev != nil {
		h.app.SetFocus(d.prev)
	} else {
		h.app.SetFocus(h.Pages)
	}
}

// forget drops the dialog at i and returns it. A dialog that would restore focus to it
// restores focus to where it would have instead.
func (h *Host) forget(i int) hosted {
	d := h.open[i]
	h.open = slices.Delete(h.open, i, i+1)
	for j := range h.open {
		if h.open[j].prev == d.item {
			h.open[j].prev = d.prev
		}
	}
	return d
}

// Confirm opens a Confirm asking text through the host, so no focus change skips it. The
// host closes it before calling done with the answer.
func (h *Host) Confirm(text string, done func(yes bool)) {
	h.Open(confirmPage, Confirm(text, func(yes bool) {
		h.Close(confirmPage)
		done(yes)
	}))
}

// Modal reports whether a dialog is open.
func (h *Host) Modal() bool {
	h.prune()
	return len(h.open) > 0
}

// index is the position of the dialog open under name, or -1.
func (h *Host) index(name string) int {
	return slices.IndexFunc(h.open, func(d hosted) bool { return d.name == name })
}

// prune forgets the dialogs whose pages the application removed itself.
func (h *Host) prune() {
	for i := len(h.open) - 1; i >= 0; i-- {
		if d := h.open[i]; !h.HasPage(d.name) || h.GetPage(d.name) != d.item {
			h.forget(i)
		}
	}
}

// top is the topmost open dialog.
func (h *Host) top() (hosted, bool) {
	h.prune()
	if len(h.open) == 0 {
		return hosted{}, false
	}
	return h.open[len(h.open)-1], true
}

// HasFocus reports true while a dialog is open, so the application hands the host every
// key wherever focus went, and otherwise whether the screen has focus.
func (h *Host) HasFocus() bool {
	if h.Modal() {
		return true
	}
	return h.Pages.HasFocus()
}

// InputHandler hands keys to the topmost dialog while one is open, giving it focus first
// when it lost it, and otherwise to the screen.
func (h *Host) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	d, ok := h.top()
	if !ok {
		return h.Pages.InputHandler()
	}
	return h.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if !d.item.HasFocus() {
			setFocus(d.item)
		}
		if handler := d.item.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}

// MouseHandler hands mouse events to the topmost dialog while one is open, and swallows
// those outside it, and otherwise to the screen.
func (h *Host) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	d, ok := h.top()
	if !ok {
		return h.Pages.MouseHandler()
	}
	return h.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if handler := d.item.MouseHandler(); handler != nil {
			if consumed, capture = handler(action, event, setFocus); consumed {
				return consumed, capture
			}
		}
		return true, nil
	})
}
