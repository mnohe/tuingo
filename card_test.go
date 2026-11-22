package tuingo

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestCardTitleStylesAndDropsWholeMarks(t *testing.T) {
	p := NewCardPane("")
	title := CardTitle{Lead: "◆", Label: "PROJ-9", Marks: []string{"#41", "+2"}}
	card := Card{Glyph: "*", Title: tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true)}
	for _, test := range []struct {
		width int
		want  string
	}{
		{18, "◆ PROJ-9 #41 +2"},
		{15, "◆ PROJ-9 #41"},
		{13, "◆ PROJ-9"},
		{9, "◆ PROJ"},
	} {
		p.Clear()
		p.DrawCardWithTitle("card", 0, 0, test.width, card, title, "")
		var line strings.Builder
		for x := 1; x < test.width-2; x++ {
			line.WriteRune(p.At(x, 0))
		}
		if got := strings.TrimSpace(line.String()); got != test.want || p.At(test.width-2, 0) != '*' {
			t.Fatalf("width %d: title %q, glyph %q", test.width, got, p.At(test.width-2, 0))
		}
		for x := 1; x <= Width(test.want); x++ {
			fg, _, attr := p.rows[0][x].style.Decompose()
			wantBold := x >= 3 && x <= 8
			if p.At(x, 0) != ' ' && (fg != tcell.ColorBlue || (attr&tcell.AttrBold != 0) != wantBold) {
				t.Errorf("width %d, column %d: foreground %v, attributes %v", test.width, x, fg, attr)
			}
		}
	}
}

func TestDrawCardTitleRowAndTwoLines(t *testing.T) {
	p := NewCardPane("T")
	p.DrawCard("x", 0, 0, 12, Card{Label: "ID-1", Glyph: "?", Text: "one two three four five six"}, "")
	row := func(y int) string {
		s := ""
		for x := 0; x < 12; x++ {
			s += string(p.At(x, y))
		}
		return s
	}
	if got := row(0); got != " ID-1     ? " {
		t.Errorf("title row: %q", got)
	}
	if got := row(1); got != " one two    " {
		t.Errorf("first line: %q", got)
	}
	if got := row(2); got != " three fo…  " && got != " three…     " {
		t.Errorf("second line shortened: %q", got)
	}
	if r := p.Rects()[0]; r.ID != "x" || r.W != 12 || r.H != CardHeight {
		t.Errorf("placed: %+v", r)
	}
}
