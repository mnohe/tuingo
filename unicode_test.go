package tuingo

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestPlaneDrawsWholeGraphemesAndClipsWideEdges(t *testing.T) {
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(20, 6)
	p := NewCardPane("T")
	p.SetBorder(false).SetRect(0, 0, 8, 3)
	p.Put(0, 0, "界e\u0301👩‍💻x", tcell.StyleDefault)
	p.Draw(s)
	for _, tc := range []struct {
		x     int
		main  rune
		comb  string
		width int
	}{
		{0, '界', "", 2}, {2, 'e', "\u0301", 1}, {3, '👩', "‍💻", 2}, {5, 'x', "", 1},
	} {
		r, c, _, w := s.GetContent(tc.x, 0)
		if r != tc.main || string(c) != tc.comb || w != tc.width {
			t.Errorf("column %d: %q + %q width %d", tc.x, r, string(c), w)
		}
	}
	p.ScrollTo(1, 0)
	s.Clear()
	p.Draw(s)
	if r, _, _, _ := s.GetContent(0, 0); r != ' ' {
		t.Errorf("partial wide glyph at left edge: %q", r)
	}
	if r, c, _, _ := s.GetContent(1, 0); r != 'e' || string(c) != "\u0301" {
		t.Error("combining character lost after scrolling")
	}
	p.ScrollTo(0, 0)
	p.SetRect(0, 0, 4, 3)
	s.Clear()
	p.Draw(s)
	if r, _, _, _ := s.GetContent(3, 0); r != ' ' {
		t.Errorf("partial wide glyph at right edge: %q", r)
	}
	p.Put(1, 0, "a", tcell.StyleDefault)
	if p.At(0, 0) != ' ' || p.At(1, 0) != 'a' {
		t.Error("overwriting trailing column left a broken wide glyph")
	}
	p.Put(3, 0, "b", tcell.StyleDefault)
	if p.At(4, 0) != ' ' {
		t.Error("overwriting wide glyph left a trailing cell")
	}
}

func TestWideCardAndChipLayout(t *testing.T) {
	p := NewCardPane("T")
	p.DrawCard("id", 0, 0, 10, Card{Label: "界e\u0301", Glyph: "👩‍💻", Text: "界 界"}, "")
	if p.At(6, 0) != ' ' || p.At(7, 0) != '👩' || p.At(9, 0) != ' ' {
		t.Error("wide title glyph is not right aligned")
	}
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(30, 3)
	end := drawChip(s, 0, 0, chip{key: "界", label: "e\u0301👩‍💻"}, tcell.StyleDefault)
	if end != 8 {
		t.Errorf("chip ends at %d, want 8", end)
	}
	if r, c, _, _ := s.GetContent(4, 0); r != 'e' || string(c) != "\u0301" {
		t.Error("chip lost combining mark")
	}
	if r, _, _, _ := s.GetContent(5, 0); r != '👩' {
		t.Error("chip emoji in wrong column")
	}
	keys := []Key{{Rune: 'a', Label: "界界", Kind: KeyCloses}, {Rune: 'b', Label: "e\u0301", Kind: KeyCloses}}
	if shown, folded := fold(keys, 15); folded || len(shown) != 2 {
		t.Error("chips that fit folded")
	}
	if _, folded := fold(keys, 14); !folded {
		t.Error("wide chips overflowed without folding")
	}
	panel := NewKeyPanel()
	panel.Set("a", "界界")
	if w, _ := panel.Size(); w != 10 {
		t.Errorf("panel width %d, want 10", w)
	}
}

func TestChipKeyStylingUsesDisplayColumns(t *testing.T) {
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(30, 3)
	for y, c := range []chip{escChip, {key: "界e\u0301", label: "👩‍💻"}} {
		end := drawChip(s, 0, y, c, tcell.StyleDefault)
		if want := Width(c.text()) + 2; end != want {
			t.Errorf("chip %q ends at %d, want %d", c.text(), end, want)
		}
		for x := 1; x < end-1; {
			_, _, style, width := s.GetContent(x, y)
			_, _, attrs := style.Decompose()
			if bold, want := attrs&tcell.AttrBold != 0, x <= Width(c.key); bold != want {
				t.Errorf("chip %q column %d: bold %v, want %v", c.text(), x, bold, want)
			}
			x += max(1, width)
		}
	}
}
