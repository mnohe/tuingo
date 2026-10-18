package tuingo

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestDividerDrawsADoubleLineDownItsHeight(t *testing.T) {
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Fini)
	s.SetSize(5, 6)
	d := NewDivider().SetColor(tcell.ColorRed)
	d.SetRect(2, 1, 1, 4)
	d.Draw(s)
	for y := 0; y < 6; y++ {
		r, _, style, _ := s.GetContent(2, y)
		inside := y >= 1 && y < 5
		if inside && r != '║' {
			t.Errorf("row %d: got %q, want the line", y, r)
		}
		if !inside && r == '║' {
			t.Errorf("row %d: line drawn outside the rectangle", y)
		}
		if fg, _, _ := style.Decompose(); inside && fg != tcell.ColorRed {
			t.Errorf("row %d: color %v, want red", y, fg)
		}
	}
	if r, _, _, _ := s.GetContent(3, 2); r == '║' {
		t.Error("line drawn past the first column")
	}
}
