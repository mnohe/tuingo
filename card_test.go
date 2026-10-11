package tuingo

import "testing"

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
