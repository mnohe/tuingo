package tuingo

import "github.com/gdamore/tcell/v2"

// heavyEdge and doubleEdge give, for each box-drawing glyph a border can be drawn with, the
// glyph that draws the same piece of line thick or doubled. Rounded corners have no thick
// form, so they turn square.
var (
	heavyEdge = map[rune]rune{
		'─': '━', '│': '┃', '═': '━', '║': '┃',
		'┌': '┏', '╭': '┏', '╔': '┏',
		'┐': '┓', '╮': '┓', '╗': '┓',
		'└': '┗', '╰': '┗', '╚': '┗',
		'┘': '┛', '╯': '┛', '╝': '┛',
	}
	doubleEdge = map[rune]rune{
		'─': '═', '│': '║', '━': '═', '┃': '║',
		'┌': '╔', '╭': '╔', '┏': '╔',
		'┐': '╗', '╮': '╗', '┓': '╗',
		'└': '╚', '╰': '╚', '┗': '╚',
		'┘': '╝', '╯': '╝', '┛': '╝',
	}
)

// thicken redraws the line glyphs on the edge of a rectangle already drawn, thick where the
// screen can show thick lines and doubled where it cannot, keeping each cell's style. Titles
// and other text on the edge stay as they are.
func thicken(screen tcell.Screen, x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	edge := heavyEdge
	if !screen.CanDisplay('┃', false) || !screen.CanDisplay('━', false) {
		edge = doubleEdge
	}
	swap := func(cx, cy int) {
		r, comb, style, _ := screen.GetContent(cx, cy)
		if t, ok := edge[r]; ok {
			screen.SetContent(cx, cy, t, comb, style)
		}
	}
	for cx := x; cx < x+w; cx++ {
		swap(cx, y)
		if h > 1 {
			swap(cx, y+h-1)
		}
	}
	for cy := y + 1; cy < y+h-1; cy++ {
		swap(x, cy)
		if w > 1 {
			swap(x+w-1, cy)
		}
	}
}
