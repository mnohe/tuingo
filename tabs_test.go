package tuingo

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestTabsUnderlineOnlyTheActiveLabel(t *testing.T) {
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(40, 1)
	tabs := NewTabs()
	tabs.Set([]string{"One", "Two", "Three"}, 0)
	tabs.SetRect(0, 0, 40, 1)
	tabs.Draw(s)
	s.Show()
	for x := 0; x < 18; x++ {
		_, _, st, _ := s.GetContent(x, 0)
		underlined := st.GetUnderlineStyle() != tcell.UnderlineStyleNone
		// " One " is columns 0-4.
		if want := x < 5; underlined != want {
			t.Errorf("column %d underlined = %v, want %v", x, underlined, want)
		}
	}
}
