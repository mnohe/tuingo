package tuingo

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestFooterColumnFollowsTheLayout(t *testing.T) {
	f := NewFooter()
	f.Set("←↑↓→", "move", "f", "filter", "/", "find")
	for key, want := range map[string]int{"←↑↓→": 0, "f": 11, "/": 21, "x": -1} {
		if got := f.Column(key); got != want {
			t.Errorf("Column(%q) = %d, want %d", key, got, want)
		}
	}
	// The column matches where the footer draws the key.
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(40, 1)
	f.SetRect(0, 0, 40, 1)
	f.Draw(s)
	if r, _, _, _ := s.GetContent(21, 0); r != '/' {
		t.Errorf("column 21 holds %q, want /", r)
	}
}

func TestAnchorStandsOnTheFooterAboveItsKey(t *testing.T) {
	f := NewFooter()
	f.Set("q", "quit", "f", "filter")
	f.SetRect(0, 9, 40, 1)
	for _, c := range []struct {
		w, x int
	}{
		{10, 8}, // above f
		{35, 5}, // moved left to fit
		{50, 0}, // never past the left edge
	} {
		b := tview.NewBox()
		a := Anchor(b, f, "f", func(int, int) (int, int) { return c.w, 4 })
		s := tcell.NewSimulationScreen("UTF-8")
		if err := s.Init(); err != nil {
			t.Fatal(err)
		}
		s.SetSize(40, 10)
		a.Draw(s)
		s.Fini()
		x, y, w, h := b.GetRect()
		if x != c.x || y != 5 || h != 4 || w != min(c.w, 40) {
			t.Errorf("width %d: rect %d,%d %dx%d", c.w, x, y, w, h)
		}
	}
}

func TestAnchorReadsTheColumnOnEveryDraw(t *testing.T) {
	f := NewFooter()
	f.Set("f", "filter")
	f.SetRect(0, 9, 40, 1)
	b := tview.NewBox()
	a := Anchor(b, f, "f", func(int, int) (int, int) { return 10, 4 })
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(40, 10)
	a.Draw(s)
	f.Set("q", "quit", "f", "filter")
	a.Draw(s)
	if x, _, _, _ := b.GetRect(); x != 8 {
		t.Errorf("x = %d after the footer changed, want 8", x)
	}
}
