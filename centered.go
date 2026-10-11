package tuingo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Centered draws a primitive centered on the screen at a size computed from the screen's.
type Centered struct {
	tview.Primitive
	size func(screenW, screenH int) (w, h int)
}

// Center wraps p so it draws centered, sized by size.
func Center(p tview.Primitive, size func(screenW, screenH int) (w, h int)) *Centered {
	return &Centered{Primitive: p, size: size}
}

// Draw sizes and places the primitive, then draws it.
func (c *Centered) Draw(screen tcell.Screen) {
	sw, sh := screen.Size()
	w, h := c.size(sw, sh)
	c.Primitive.SetRect((sw-w)/2, (sh-h)/2, w, h)
	c.Primitive.Draw(screen)
}
