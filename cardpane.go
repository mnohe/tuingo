// Package tuingo holds terminal primitives on tview and tcell for screens built from cards,
// panes and notices: a card pane, a debounced file watcher, a self-clearing notice, tabs, a
// key legend, a centered overlay, and column-counted text helpers.
package tuingo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Rect is a placed rectangle with an id and the id of the rectangle it hangs off.
type Rect struct {
	ID     string
	X, Y   int
	W, H   int
	Parent string
}

type cell struct {
	r     rune
	style tcell.Style
}

// CardPane is a bordered, titled viewport over a plane of cells larger than itself, with
// rectangles placed on the plane that a cursor moves between. The plane is drawn with Put,
// the rectangles declared with Place, and the pane scrolls so the selection stays visible.
type CardPane struct {
	*tview.Box
	rows     [][]cell
	rects    []Rect
	selected string
	ox, oy   int
	idle     tcell.Color
	focused  tcell.Color
	onSelect func(id string)
	onChange func()
}

// NewCardPane returns an empty pane with the given title in its border.
func NewCardPane(title string) *CardPane {
	p := &CardPane{Box: tview.NewBox(), idle: tcell.ColorGray, focused: tcell.ColorBlue}
	p.SetBorder(true).SetTitle(" " + title + " ").SetTitleAlign(tview.AlignLeft)
	return p
}

// SetFocusColors sets the border and title color when idle and when focused.
func (p *CardPane) SetFocusColors(idle, focused tcell.Color) *CardPane {
	p.idle, p.focused = idle, focused
	return p
}

// SetSelectedFunc sets the function called with the selected id on enter.
func (p *CardPane) SetSelectedFunc(fn func(id string)) *CardPane {
	p.onSelect = fn
	return p
}

// SetChangedFunc sets the function called after the cursor moves, for the owner to redraw
// the plane with the new selection.
func (p *CardPane) SetChangedFunc(fn func()) *CardPane {
	p.onChange = fn
	return p
}

// Clear empties the plane and the rectangles; the selection and scroll position stay, so a
// redraw of the same content keeps the cursor where it was.
func (p *CardPane) Clear() {
	p.rows, p.rects = nil, nil
}

// Put writes text on the plane at (x, y), one rune per column.
func (p *CardPane) Put(x, y int, text string, style tcell.Style) {
	for len(p.rows) <= y {
		p.rows = append(p.rows, nil)
	}
	row := p.rows[y]
	for _, r := range text {
		for len(row) <= x {
			row = append(row, cell{' ', tcell.StyleDefault})
		}
		row[x] = cell{r, style}
		x++
	}
	p.rows[y] = row
}

// At returns the rune on the plane at (x, y), a space where nothing was put.
func (p *CardPane) At(x, y int) rune {
	if y < len(p.rows) && x < len(p.rows[y]) {
		return p.rows[y][x].r
	}
	return ' '
}

// Place declares a rectangle the cursor can land on. parent is the id of the rectangle an
// arrow leads back to, or empty.
func (p *CardPane) Place(id string, x, y, w, h int, parent string) {
	p.rects = append(p.rects, Rect{id, x, y, w, h, parent})
}

// Rects returns the rectangles placed since the last Clear, in placement order.
func (p *CardPane) Rects() []Rect { return p.rects }

// Select moves the selection to id, or to the first rectangle when id is not placed.
func (p *CardPane) Select(id string) {
	for _, r := range p.rects {
		if r.ID == id {
			p.selected = id
			return
		}
	}
	p.selected = ""
	if len(p.rects) > 0 {
		p.selected = p.rects[0].ID
	}
}

// Selected returns the selected id, empty when nothing is placed.
func (p *CardPane) Selected() string { return p.selected }

func (p *CardPane) current() *Rect {
	for i := range p.rects {
		if p.rects[i].ID == p.selected {
			return &p.rects[i]
		}
	}
	return nil
}

// Move moves the selection: left to the parent, right to the first rectangle hanging off
// the selection, up and down to the nearest rectangle in that direction, preferring the same
// column. It reports whether the selection changed.
func (p *CardPane) Move(dx, dy int) bool {
	cur := p.current()
	if cur == nil {
		return false
	}
	best := ""
	switch {
	case dx < 0:
		best = cur.Parent
	case dx > 0:
		for _, r := range p.rects {
			if r.Parent == p.selected {
				best = r.ID
				break
			}
		}
	default:
		bestScore := -1.0
		for _, r := range p.rects {
			if (r.Y-cur.Y)*dy <= 0 {
				continue
			}
			score := float64(abs(r.Y-cur.Y)) + 0.5*float64(abs(r.X-cur.X))
			if bestScore < 0 || score < bestScore {
				best, bestScore = r.ID, score
			}
		}
	}
	if best == "" {
		return false
	}
	p.selected = best
	return true
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Reveal scrolls the least that brings the selected rectangle inside the viewport.
func (p *CardPane) Reveal() {
	_, _, w, h := p.GetInnerRect()
	if r := p.current(); r != nil {
		if r.Y < p.oy {
			p.oy = r.Y
		}
		if r.Y+r.H > p.oy+h {
			p.oy = r.Y + r.H - h
		}
		if r.X < p.ox {
			p.ox = r.X
		}
		if r.X+r.W > p.ox+w {
			p.ox = r.X + r.W - w
		}
	}
	if p.oy < 0 {
		p.oy = 0
	}
	if p.ox < 0 {
		p.ox = 0
	}
}

// ScrollTo sets the plane coordinate shown at the viewport's top-left corner.
func (p *CardPane) ScrollTo(x, y int) { p.ox, p.oy = x, y }

// Offset returns the plane coordinate shown at the viewport's top-left corner.
func (p *CardPane) Offset() (x, y int) { return p.ox, p.oy }

// Draw draws the border, then the window of the plane the viewport shows.
func (p *CardPane) Draw(screen tcell.Screen) {
	color := p.idle
	if p.HasFocus() {
		color = p.focused
	}
	p.SetBorderColor(color).SetTitleColor(color)
	p.DrawForSubclass(screen, p)
	x0, y0, w, h := p.GetInnerRect()
	for y := 0; y < h; y++ {
		py := y + p.oy
		if py >= len(p.rows) {
			break
		}
		row := p.rows[py]
		for x := 0; x < w; x++ {
			px := x + p.ox
			if px >= len(row) {
				break
			}
			screen.SetContent(x0+x, y0+y, row[px].r, nil, row[px].style)
		}
	}
}

// InputHandler moves the cursor on the arrow and vi keys and reports enter.
func (p *CardPane) InputHandler() func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
	return p.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		dx, dy := 0, 0
		switch {
		case event.Key() == tcell.KeyUp || event.Rune() == 'k':
			dy = -1
		case event.Key() == tcell.KeyDown || event.Rune() == 'j':
			dy = 1
		case event.Key() == tcell.KeyLeft || event.Rune() == 'h':
			dx = -1
		case event.Key() == tcell.KeyRight || event.Rune() == 'l':
			dx = 1
		case event.Key() == tcell.KeyEnter:
			if p.onSelect != nil && p.selected != "" {
				p.onSelect(p.selected)
			}
			return
		default:
			return
		}
		if p.Move(dx, dy) {
			if p.onChange != nil {
				p.onChange()
			}
			p.Reveal()
		}
	})
}
