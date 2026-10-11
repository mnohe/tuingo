package tuingo

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Key is a key the text box offers, with the label its legend shows.
type Key struct {
	Rune  rune
	Label string
}

// TextBox is a bordered, titled, scrolling box of dynamic-color text with a key legend
// along its bottom edge. Scrolling keys are the text view's; esc and q close it; any other
// key in its legend is reported.
type TextBox struct {
	*tview.Box
	text  *tview.TextView
	hints *tview.TextView
	keys  map[rune]bool
	onKey func(r rune)
	close func()
}

// NewTextBox returns an empty box; onKey receives a legend key, close runs on esc or q.
func NewTextBox(onKey func(r rune), close func()) *TextBox {
	v := &TextBox{
		Box:   tview.NewBox(),
		text:  tview.NewTextView().SetDynamicColors(true).SetScrollable(true),
		hints: tview.NewTextView().SetDynamicColors(true),
		onKey: onKey,
		close: close,
	}
	v.SetBorder(true).SetTitleAlign(tview.AlignLeft).SetBorderPadding(1, 1, 2, 2)
	v.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch {
		case ev.Key() == tcell.KeyEsc, ev.Rune() == 'q':
			v.close()
		case ev.Key() == tcell.KeyRune && v.keys[ev.Rune()]:
			v.onKey(ev.Rune())
		default:
			return ev
		}
		return nil
	})
	return v
}

// Show fills the box: the title in its border, the lines, the keys in the legend, and a
// faint note after them when given.
func (v *TextBox) Show(title string, lines []string, keys []Key, note string) {
	v.SetTitle(" " + title + " ")
	v.text.SetText(strings.Join(lines, "\n")).ScrollToBeginning()
	v.keys = map[rune]bool{}
	var pairs []string
	for _, k := range keys {
		v.keys[k.Rune] = true
		pairs = append(pairs, string(k.Rune), k.Label)
	}
	hints := Legend(pairs...)
	if note != "" {
		hints += "  [::d]" + tview.Escape(note) + "[::-]"
	}
	if hints != "" {
		hints += "  "
	}
	v.hints.SetText("\n" + hints + Legend("esc", "close"))
}

// Size is the box's size on a screen: nine tenths of the width up to 120 columns, and 85
// percent of the height. It fits Center.
func (v *TextBox) Size(sw, sh int) (int, int) {
	w := sw * 9 / 10
	if w > 120 {
		w = 120
	}
	return w, sh * 85 / 100
}

// Draw draws the frame, the text, and the legend on the two bottom rows.
func (v *TextBox) Draw(screen tcell.Screen) {
	v.DrawForSubclass(screen, v)
	x, y, w, h := v.GetInnerRect()
	v.text.SetRect(x, y, w, h-2)
	v.text.Draw(screen)
	v.hints.SetRect(x, y+h-2, w, 2)
	v.hints.Draw(screen)
}

// InputHandler scrolls the text; the capture above takes the legend keys.
func (v *TextBox) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return v.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		v.text.InputHandler()(event, setFocus)
	})
}
