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

// CardTitle supplies a card's title label, with an optional lead and trailing marks.
type CardTitle struct {
	Lead  string
	Label string
	Marks []string
}

// DrawCard draws a card with its top-left at (x, y), w columns wide, and places its
// rectangle under id, hanging off parent.
func (p *CardPane) DrawCard(id string, x, y, w int, c Card, parent string) {
	p.DrawCardWithTitle(id, x, y, w, c, CardTitle{Label: c.Label}, parent)
}

// DrawCardWithTitle draws a card using title in place of c.Label. Marks drop
// from the right before the label is truncated. The lead and marks use the
// title colors without bold.
func (p *CardPane) DrawCardWithTitle(id string, x, y, w int, c Card, title CardTitle, parent string) {
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
	available := inner - Width(glyph)
	lead := Truncate(title.Lead, available)
	leadWidth := Width(lead)
	labelWidth := Width(title.Label)
	gap := 0
	if lead != "" && title.Label != "" {
		gap = 1
	}
	marks := ""
	marksWidth := 0
	for _, mark := range title.Marks {
		if mark == "" {
			continue
		}
		separator := ""
		if lead != "" || title.Label != "" || marks != "" {
			separator = " "
		}
		width := len(separator) + Width(mark)
		if leadWidth+gap+labelWidth+marksWidth+width > available {
			break
		}
		marks += separator + mark
		marksWidth += width
	}
	label := Truncate(title.Label, available-leadWidth-gap-marksWidth)
	if label != "" && gap > 0 {
		label = " " + label
	}
	p.Put(x, y, strings.Repeat(" ", w), c.Title)
	plain := c.Title.Bold(false)
	p.Put(x+1, y, lead, plain)
	p.Put(x+1+leadWidth, y, label, c.Title)
	p.Put(x+1+leadWidth+Width(label), y, marks, plain)
	p.Put(x+1+available, y, glyph, c.Title)
	for i, l := range lines {
		p.Put(x, y+1+i, PadRight(" "+PadRight(l, inner), w), c.Body)
	}
	p.Place(id, x, y, w, CardHeight, parent)
}
