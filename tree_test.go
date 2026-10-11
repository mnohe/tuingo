package tuingo

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestDrawTreePlacesChildrenRightAndStacked(t *testing.T) {
	p := NewCardPane("T")
	root := &Node{ID: "a", Children: []*Node{{ID: "b"}, {ID: "c", Children: []*Node{{ID: "d"}}}}}
	card := func(id string) Card { return Card{Label: id, Text: "text " + id} }
	used := p.DrawTree(root, 1, 0, 10, card, tcell.StyleDefault)
	want := map[string][3]interface{}{"a": {1, 0, ""}, "b": {15, 0, "a"}, "c": {15, 4, "a"}, "d": {29, 4, "c"}}
	for _, r := range p.Rects() {
		w := want[r.ID]
		if r.X != w[0] || r.Y != w[1] || r.Parent != w[2] {
			t.Errorf("%s at (%d,%d) off %q, want (%v,%v) off %q", r.ID, r.X, r.Y, r.Parent, w[0], w[1], w[2])
		}
	}
	if used != 7 {
		t.Errorf("rows used: %d, want 7", used)
	}
	var row strings.Builder
	for x := 10; x < 16; x++ {
		row.WriteRune(p.At(x, 0))
	}
	if got := row.String(); got != " ─┬─▶ " {
		t.Errorf("connector on row 0: %q", got)
	}
	if p.At(12, 2) != '│' || p.At(12, 4) != '└' {
		t.Errorf("junction column: %q %q", p.At(12, 2), p.At(12, 4))
	}
}

func TestDrawLane(t *testing.T) {
	p := NewCardPane("T")
	card := func(id string) Card { return Card{Label: id} }
	if used := p.DrawLane([]string{"a", "b"}, 1, 0, 10, card); used != 7 {
		t.Errorf("rows used: %d, want 7", used)
	}
	if r := p.Rects()[1]; r.Y != 4 {
		t.Errorf("second card at row %d, want 4", r.Y)
	}
	if used := p.DrawLane(nil, 1, 0, 10, card); used != 0 {
		t.Errorf("empty lane uses %d rows", used)
	}
}
