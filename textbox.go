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

// chip is a key and its label as a chip on the border shows them, the key in bold.
type chip struct{ key, label string }

// text is the chip's label as drawn, without its padding.
func (c chip) text() string { return c.key + " " + c.label }

// moreChip stands in for the closing keys folded off the border, and escChip, on the top
// border, cancels the box.
var (
	moreChip = chip{"+", "More…"}
	escChip  = chip{"esc", "╳"}
)

// TextBox is a bordered, titled, scrolling box of dynamic-color text with a right-aligned
// key legend along its bottom edge, and the keys that close it as chips on its bottom
// border. Chips that do not fit fold into a "+ More…" chip, and + opens a KeyPanel listing
// them. Scrolling keys are the text view's; any other key it offers is reported, folded or
// not. While it is dismissible, as it is by default, esc and q cancel it, and an "esc ╳"
// chip at the top right of its border says so.
type TextBox struct {
	*tview.Box
	text     *tview.TextView
	hints    *tview.TextView
	panel    *KeyPanel
	keys     map[rune]bool
	closing  []Key
	folded   []Key
	open     bool
	onKey    func(r rune)
	close    func()
	stuck    bool
	blocks   []Block
	rendered string
	targets  []target
}

// target is a place on screen a click runs an action from: a chip, or a row of the key
// panel.
type target struct {
	x, y, w int
	act     func()
}

// NewTextBox returns an empty, dismissible box; onKey receives an offered key, close runs on
// esc or q.
func NewTextBox(onKey func(r rune), close func()) *TextBox {
	v := &TextBox{
		Box:   tview.NewBox(),
		text:  tview.NewTextView().SetDynamicColors(true).SetScrollable(true),
		hints: tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight),
		panel: NewKeyPanel(),
		onKey: onKey,
		close: close,
	}
	v.SetBorder(true).SetTitleAlign(tview.AlignLeft).SetBorderPadding(1, 1, 2, 2)
	v.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch {
		case v.open && (ev.Key() == tcell.KeyEsc || ev.Rune() == 'q'):
			v.open = false
		case !v.stuck && (ev.Key() == tcell.KeyEsc || ev.Rune() == 'q'):
			v.close()
		case ev.Key() == tcell.KeyRune && v.keys[ev.Rune()]:
			v.run(ev.Rune())
		case ev.Rune() == '+' && len(v.folded) > 0:
			v.open = true
		default:
			return ev
		}
		return nil
	})
	return v
}

// run closes the panel and reports an offered key.
func (v *TextBox) run(r rune) {
	v.open = false
	v.onKey(r)
}

// SetDismissible sets whether esc and q cancel the box by calling close, shown by an
// "esc ╳" chip at the top right of the border. A box that is not dismissible ignores them
// and shows no chip, so the reader closes it with one of its closing keys; the panel of
// folded keys still closes on esc and q.
func (v *TextBox) SetDismissible(on bool) *TextBox {
	v.stuck = !on
	return v
}

// Show fills the box: the title in its border, the lines, the keys that leave it open in
// the right-aligned legend with a faint note left of them when given, and the keys that
// close it on the bottom border, the first the last to fold.
func (v *TextBox) Show(title string, lines []string, keys []Key, note string) {
	v.blocks = nil
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
		if hints != "" {
			hints = "  " + hints
		}
		hints = "[::d]" + tview.Escape(note) + "[::-]" + hints
	}
	v.hints.SetText("\n" + hints)
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

// Draw draws the frame, the text, the legend on the two bottom rows, the esc chip on the
// top border, the closing keys on the bottom border, and the key panel above them when
// open.
func (v *TextBox) Draw(screen tcell.Screen) {
	v.DrawForSubclass(screen, v)
	v.targets = v.targets[:0]
	x, y, w, h := v.GetInnerRect()
	if v.blocks != nil {
		var lines []string
		for _, block := range v.blocks {
			if block != nil {
				lines = append(lines, block.Render(w)...)
			}
		}
		text := strings.Join(lines, "\n")
		if text != v.rendered {
			v.text.SetText(text)
			v.rendered = text
		}
	}
	v.text.SetRect(x, y, w, max(0, h-2))
	v.text.Draw(screen)
	v.hints.SetRect(x, y+h-2, w, 2)
	v.hints.Draw(screen)
	v.drawEsc(screen)
	v.drawChips(screen)
}

// chipStyle is a chip's look: reversed, it fills with the border color and its ink is the
// screen's background, whatever the terminal's theme.
func (v *TextBox) chipStyle() tcell.Style {
	return tcell.StyleDefault.Foreground(v.GetBorderColor()).Reverse(true)
}

// drawEsc draws the "esc ╳" chip right-aligned on the top border, followed by a ─, when
// the box is dismissible and the chip fits beside the title with a ─ between them.
func (v *TextBox) drawEsc(screen tcell.Screen) {
	if v.stuck {
		return
	}
	bx, by, bw, _ := v.GetRect()
	row := rowWidth([]chip{escChip})
	if 1+tview.TaggedStringWidth(v.GetTitle())+1+row+1 > bw {
		return
	}
	x := bx + bw - 1 - row
	drawChip(screen, x, by, escChip, v.chipStyle())
	v.targets = append(v.targets, target{x, by, row - 1, v.close})
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
	avail := bw - 4
	shown, folded := fold(v.closing, avail)
	v.folded = v.closing[len(shown):]
	chips := make([]chip, 0, len(shown)+2)
	for _, k := range shown {
		chips = append(chips, chip{string(k.Rune), k.Label})
	}
	if folded {
		chips = append(chips, moreChip)
	}
	right := bx + bw - 2
	style := v.chipStyle()
	// The panel's right edge sits on the "+ More…" chip's, or on the border's when no chip
	// fits.
	edge := right - 1
	if row := rowWidth(chips); row <= avail {
		cx := right - row + 1
		for i, c := range chips {
			act := func() { v.open = !v.open }
			if i < len(shown) {
				r := shown[i].Rune
				act = func() { v.run(r) }
			}
			start := cx
			cx = drawChip(screen, cx, by+bh-1, c, style) + 1
			v.targets = append(v.targets, target{start, by + bh - 1, cx - 1 - start, act})
			if c == moreChip {
				edge = cx - 2
			}
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
	px, py := max(edge-pw+1, bx+1), max(by+bh-1-ph, by+1)
	pw, ph = min(pw, bx+bw-1-px), min(ph, by+bh-1-py)
	v.panel.SetRect(px, py, pw, ph)
	v.panel.Draw(screen)
	for i, k := range v.folded {
		if row := py + 1 + i; row < py+ph-1 {
			r := k.Rune
			v.targets = append(v.targets, target{px + 1, row, pw - 2, func() { v.run(r) }})
		}
	}
}

// fold returns the leading closing keys that fit in avail columns, and whether any were
// folded behind a "+ More…" chip, whose width it counts. When none fit, none are shown.
func fold(keys []Key, avail int) ([]Key, bool) {
	chips := make([]chip, len(keys))
	for i, k := range keys {
		chips[i] = chip{string(k.Rune), k.Label}
	}
	if rowWidth(chips) <= avail {
		return keys, false
	}
	for n := len(keys) - 1; n > 0; n-- {
		if rowWidth(append(chips[:n:n], moreChip)) <= avail {
			return keys[:n], true
		}
	}
	return nil, true
}

// rowWidth is the columns a row of chips takes: each label with a space either side and a
// ─ after it.
func rowWidth(chips []chip) int {
	w := 0
	for _, c := range chips {
		w += Width(c.text()) + 3
	}
	return w
}

// drawChip draws a chip padded by a space either side at (x, y), its key bold, and returns
// the column after it.
func drawChip(screen tcell.Screen, x, y int, c chip, style tcell.Style) int {
	screen.SetContent(x, y, ' ', nil, style)
	x++
	bold := Width(c.key)
	for i, r := range []rune(c.text()) {
		s := style
		if i < bold {
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

// MouseHandler runs a chip's action, or a key panel row's, when it is clicked, as its key
// would, and closes the panel on a click anywhere else. Other mouse events go to the text,
// so the wheel scrolls it, and a click inside the box gives the box focus.
func (v *TextBox) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return v.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		x, y := event.Position()
		if !v.InRect(x, y) {
			return false, nil
		}
		if action == tview.MouseLeftClick {
			for _, t := range v.targets {
				if y == t.y && x >= t.x && x < t.x+t.w {
					setFocus(v)
					t.act()
					return true, nil
				}
			}
			if v.open {
				v.open = false
				return true, nil
			}
		}
		// The text would take focus from the box, whose capture handles the keys.
		v.text.MouseHandler()(action, event, func(tview.Primitive) { setFocus(v) })
		if action == tview.MouseLeftDown {
			setFocus(v)
		}
		return true, nil
	})
}

// ShowBlocks fills the box with content rendered at the inner width on every draw.
// Plain lines and tables can be mixed; Show remains available for finished lines.
func (v *TextBox) ShowBlocks(title string, blocks []Block, keys []Key, note string) {
	v.Show(title, nil, keys, note)
	v.blocks = append([]Block{}, blocks...)
	v.rendered = ""
}
