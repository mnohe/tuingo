package tuingo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Anchored draws a primitive standing on a footer, above the key that opened it.
type Anchored struct {
	tview.Primitive
	footer *Footer
	key    string
	size   func(screenW, screenH int) (w, h int)
}

// Anchor wraps p so it draws with its bottom edge on the row above footer and its left edge
// above key, moved left as far as it takes to fit on the screen, never past its left edge.
// The size comes from size, and the column from the footer, on every draw, so a footer that
// lists different keys in different views moves the dialog with its key. When the footer
// does not list key, the dialog stands at the footer's left edge.
func Anchor(p tview.Primitive, footer *Footer, key string, size func(screenW, screenH int) (w, h int)) *Anchored {
	return &Anchored{Primitive: p, footer: footer, key: key, size: size}
}

// Draw sizes and places the primitive, then draws it.
func (a *Anchored) Draw(screen tcell.Screen) {
	sw, sh := screen.Size()
	w, h := a.size(sw, sh)
	fx, fy, _, _ := a.footer.GetRect()
	x := fx + max(a.footer.Column(a.key), 0)
	x = max(min(x, sw-w), 0)
	h = min(h, fy)
	a.Primitive.SetRect(x, fy-h, min(w, sw), h)
	a.Primitive.Draw(screen)
}
