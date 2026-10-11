package tuingo

import (
	"testing"
)

func pane() *CardPane {
	p := NewCardPane("T")
	p.SetRect(0, 0, 22, 8) // inner 20x6
	p.Place("a", 0, 0, 10, 3, "")
	p.Place("b", 14, 0, 10, 3, "a")
	p.Place("c", 14, 4, 10, 3, "a")
	p.Place("d", 0, 8, 10, 3, "")
	p.Select("a")
	return p
}

func TestMoveFollowsArrowsAndColumns(t *testing.T) {
	p := pane()
	steps := []struct {
		dx, dy int
		want   string
	}{
		{1, 0, "b"},  // right: first child
		{0, 1, "c"},  // down: nearest below, same column
		{-1, 0, "a"}, // left: parent
		{0, 1, "d"},  // down from a: d is nearer than c by column weight
		{0, -1, "a"}, // up
		{-1, 0, "a"}, // no parent: stays
	}
	for _, s := range steps {
		p.Move(s.dx, s.dy)
		if got := p.Selected(); got != s.want {
			t.Errorf("move (%d,%d): got %s, want %s", s.dx, s.dy, got, s.want)
		}
	}
}

func TestRevealScrollsTheLeast(t *testing.T) {
	p := pane()
	p.Select("d")
	p.Reveal()
	if x, y := p.Offset(); x != 0 || y != 5 {
		t.Errorf("d (y 8..11) in a 6-high viewport: offset (%d,%d), want (0,5)", x, y)
	}
	p.Select("b")
	p.Reveal()
	if x, y := p.Offset(); x != 4 || y != 0 {
		t.Errorf("b (x 14..24) in a 20-wide viewport: offset (%d,%d), want (4,0)", x, y)
	}
	p.Select("a")
	p.Reveal()
	if x, y := p.Offset(); x != 0 || y != 0 {
		t.Errorf("back to a: offset (%d,%d), want (0,0)", x, y)
	}
}

func TestSelectFallsBackToFirst(t *testing.T) {
	p := pane()
	p.Select("gone")
	if got := p.Selected(); got != "a" {
		t.Errorf("unknown id selects the first rectangle, got %s", got)
	}
	p.Clear()
	p.Select("a")
	if got := p.Selected(); got != "" {
		t.Errorf("nothing placed selects nothing, got %q", got)
	}
}
