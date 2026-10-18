package tuingo

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// KeyKind is what a key's action does to the text box, which decides where the box shows it.
type KeyKind int

const (
	// KeyStays is a key whose action leaves the box open, shown on the legend row inside it.
	KeyStays KeyKind = iota
	// KeyCloses is a key whose action closes the box, shown as a chip on its bottom border.
	KeyCloses
)

// Key is a key the text box offers, with the label its legend shows and what it does to the
// box.
type Key struct {
	Rune  rune
	Label string
	Kind  KeyKind
}

// more is the chip that stands in for the closing keys folded off the border.
const more = "+ More…"

// chipInk is the ink of a border chip, dark so it reads on the border color.
var chipInk = tcell.GetColor("#000000")

// TextBox is a bordered, titled, scrolling box of dynamic-color text with a key legend
// along its bottom edge, and the keys that close it as chips on its bottom border. Chips
// that do not fit fold into a "+ More…" chip, and + opens a KeyPanel listing them.
// Scrolling keys are the text view's; esc and q close it; any other key it offers is
// reported, folded or not.
type TextBox struct {
	*tview.Box
	text    *tview.TextView
	hints   *tview.TextView
	panel   *KeyPanel
	keys    map[rune]bool
	closing []Key
	folded  []Key
	open    bool
	onKey   func(r rune)
	close   func()
}

// NewTextBox returns an empty box; onKey receives an offered key, close runs on esc or q.
func NewTextBox(onKey func(r rune), close func()) *TextBox {
	v := &TextBox{
		Box:   tview.NewBox(),
		text:  tview.NewTextView().SetDynamicColors(true).SetScrollable(true),
		hints: tview.NewTextView().SetDynamicColors(true),
		panel: NewKeyPanel(),
		onKey: onKey,
		close: close,
	}
	v.SetBorder(true).SetTitleAlign(tview.AlignLeft).SetBorderPadding(1, 1, 2, 2)
	v.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch {
		case v.open && (ev.Key() == tcell.KeyEsc || ev.Rune() == 'q'):
			v.open = false
		case ev.Key() == tcell.KeyEsc, ev.Rune() == 'q':
			v.close()
		case ev.Key() == tcell.KeyRune && v.keys[ev.Rune()]:
			v.open = false
			v.onKey(ev.Rune())
		case ev.Rune() == '+' && len(v.folded) > 0:
			v.open = true
		default:
			return ev
		}
		return nil
	})
	return v
}

// Show fills the box: the title in its border, the lines, the keys that leave it open in
// the legend with a faint note after them when given, and the keys that close it on the
// bottom border, the first the last to fold.
func (v *TextBox) Show(title string, lines []string, keys []Key, note string) {
	v.SetTitle(" " + title + " ")
	v.text.SetText(strings.Join(lines, "\n")).ScrollToBeginning()
	v.keys = map[rune]bool{}
	v.closing, v.folded, v.open = nil, nil, false
	var pairs []string
	for _, k := range keys {
		v.keys[k.Rune] = true
		if k.Kind == KeyCloses {
			v.closing = append(v.closing, k)
			continue
		}
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

// Draw draws the frame, the text, the legend on the two bottom rows, the closing keys on
// the bottom border, and the key panel above them when open.
func (v *TextBox) Draw(screen tcell.Screen) {
	v.DrawForSubclass(screen, v)
	x, y, w, h := v.GetInnerRect()
	v.text.SetRect(x, y, w, h-2)
	v.text.Draw(screen)
	v.hints.SetRect(x, y+h-2, w, 2)
	v.hints.Draw(screen)
	v.drawChips(screen)
}

// drawChips draws the closing keys right-aligned on the bottom border, each followed by a
// ─, folding from the right into a "+ More…" chip until the row fits, and the key panel
// listing the folded keys, its bottom on the row above the border and its right edge on
// the chip's.
func (v *TextBox) drawChips(screen tcell.Screen) {
	if len(v.closing) == 0 {
		return
	}
	bx, by, bw, bh := v.GetRect()
	shown, folded := fold(v.closing, bw-4)
	v.folded = v.closing[len(shown):]
	labels := make([]string, 0, len(shown)+1)
	for _, k := range shown {
		labels = append(labels, string(k.Rune)+" "+k.Label)
	}
	if folded {
		labels = append(labels, more)
	}
	right := bx + bw - 2
	style := tcell.StyleDefault.Foreground(chipInk).Background(v.GetBorderColor())
	if row := rowWidth(labels); row <= bw-4 {
		cx := right - row + 1
		for _, l := range labels {
			cx = drawChip(screen, cx, by+bh-1, l, style) + 1
		}
	}
	if !folded || !v.open {
		return
	}
	var pairs []string
	for _, k := range v.folded {
		pairs = append(pairs, string(k.Rune), k.Label)
	}
	v.panel.Set(pairs...)
	pw, ph := v.panel.Size()
	px, py := max(right-1-pw+1, bx+1), max(by+bh-1-ph, by+1)
	v.panel.SetRect(px, py, min(pw, bx+bw-1-px), min(ph, by+bh-1-py))
	v.panel.Draw(screen)
}

// fold returns the leading closing keys that fit in avail columns, and whether any were
// folded behind a "+ More…" chip, whose width it counts. When none fit, none are shown.
func fold(keys []Key, avail int) ([]Key, bool) {
	labels := make([]string, len(keys))
	for i, k := range keys {
		labels[i] = string(k.Rune) + " " + k.Label
	}
	if rowWidth(labels) <= avail {
		return keys, false
	}
	for n := len(keys) - 1; n > 0; n-- {
		if rowWidth(append(labels[:n:n], more)) <= avail {
			return keys[:n], true
		}
	}
	return nil, true
}

// rowWidth is the columns a row of chips takes: each label with a space either side and a
// ─ after it.
func rowWidth(labels []string) int {
	w := 0
	for _, l := range labels {
		w += Width(l) + 3
	}
	return w
}

// drawChip draws a label padded by a space either side at (x, y), its first rune bold,
// and returns the column after it.
func drawChip(screen tcell.Screen, x, y int, label string, style tcell.Style) int {
	screen.SetContent(x, y, ' ', nil, style)
	x++
	for i, r := range label {
		s := style
		if i == 0 {
			s = s.Bold(true)
		}
		screen.SetContent(x, y, r, nil, s)
		x++
	}
	screen.SetContent(x, y, ' ', nil, style)
	return x + 1
}

// InputHandler scrolls the text; the capture above takes the offered keys.
func (v *TextBox) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return v.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		v.text.InputHandler()(event, setFocus)
	})
}
