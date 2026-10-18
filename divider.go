package tuingo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Divider is a vertical double line down the height of its rectangle, one column wide, for
// marking an edge that does not move: a pinned pane beside panes that scroll past it.
type Divider struct {
	*tview.Box
	color tcell.Color
}

// NewDivider returns a divider in the terminal's default color.
func NewDivider() *Divider {
	return &Divider{Box: tview.NewBox(), color: tcell.ColorDefault}
}

// SetColor sets the line's color.
func (d *Divider) SetColor(c tcell.Color) *Divider {
	d.color = c
	return d
}

// Draw draws the line in the rectangle's leftmost column.
func (d *Divider) Draw(screen tcell.Screen) {
	d.Box.DrawForSubclass(screen, d)
	x, y, _, h := d.GetRect()
	style := tcell.StyleDefault.Foreground(d.color).Background(tview.Styles.PrimitiveBackgroundColor)
	for row := y; row < y+h; row++ {
		screen.SetContent(x, row, '║', nil, style)
	}
}
