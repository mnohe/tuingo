package tuingo

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var themed sync.Once

// RoundedTheme sets tview's package-wide look once: the terminal's own background, and
// rounded single borders whether or not a box has focus, for screens that show focus by
// color.
func RoundedTheme() {
	themed.Do(func() {
		tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
		tview.Borders.TopLeft, tview.Borders.TopRight = '╭', '╮'
		tview.Borders.BottomLeft, tview.Borders.BottomRight = '╰', '╯'
		tview.Borders.HorizontalFocus, tview.Borders.VerticalFocus = tview.Borders.Horizontal, tview.Borders.Vertical
		tview.Borders.TopLeftFocus, tview.Borders.TopRightFocus = '╭', '╮'
		tview.Borders.BottomLeftFocus, tview.Borders.BottomRightFocus = '╰', '╯'
	})
}
