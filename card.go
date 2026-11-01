package tuingo

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// CardHeight is the rows a card takes: a title row and two body lines.
const CardHeight = 3

// Card is what a card shows: a label and a glyph on the title row, text wrapped to two
// body lines, in a style for each.
type Card struct {
	Label, Glyph string
	Text         string
	Title, Body  tcell.Style
}

// DrawCard draws a card with its top-left at (x, y), w columns wide, and places its
// rectangle under id, hanging off parent.
func (p *CardPane) DrawCard(id string, x, y, w int, c Card, parent string) {
	if w <= 0 {
		return
	}
	inner := max(0, w-2)
	lines := Wrap(c.Text, inner)
	if len(lines) == 0 {
		lines = []string{""}
	}
	if len(lines) > 2 {
		lines = []string{lines[0], Shorten(strings.Join(lines[1:], " "), inner)}
	}
	for len(lines) < 2 {
		lines = append(lines, "")
	}
	glyph := Truncate(c.Glyph, inner)
	label := PadRight(c.Label, inner-Width(glyph)) + glyph
	p.Put(x, y, PadRight(" "+label, w), c.Title)
	for i, l := range lines {
		p.Put(x, y+1+i, PadRight(" "+PadRight(l, inner), w), c.Body)
	}
	p.Place(id, x, y, w, CardHeight, parent)
}
