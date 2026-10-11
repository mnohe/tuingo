package tuingo

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Darker scales a #rrggbb color toward black by factor.
func Darker(hex string, factor float64) string {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	round := func(c int) int { return int(float64(c)*factor + 0.5) }
	return fmt.Sprintf("#%02x%02x%02x", round(r), round(g), round(b))
}

// CardStyles returns a card's title and body styles from its fill and inks: the title row
// bold on the fill, the body on a darker shade. A selected card's title row is inverted to
// near-white so it reads whatever the fill.
func CardStyles(fill, ink, bodyInk string, selected bool) (title, body tcell.Style) {
	title = tcell.StyleDefault.Bold(true).Foreground(tcell.GetColor(ink)).Background(tcell.GetColor(fill))
	if selected {
		title = tcell.StyleDefault.Bold(true).Foreground(tcell.GetColor("#000000")).Background(tcell.GetColor("#f2f2f2"))
	}
	body = tcell.StyleDefault.Foreground(tcell.GetColor(bodyInk)).Background(tcell.GetColor(Darker(fill, 0.55)))
	return title, body
}
