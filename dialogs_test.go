package tuingo

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// listed draws p at w by h on a simulation screen and returns its rows as text.
func listed(t *testing.T, p interface {
	SetRect(x, y, w, h int)
	Draw(tcell.Screen)
}, w, h int) []string {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(w, h)
	p.SetRect(0, 0, w, h)
	p.Draw(s)
	rows := make([]string, h)
	for y := range rows {
		rows[y] = row(s, y)
	}
	return rows
}

var space = tcell.NewEventKey(tcell.KeyRune, ' ', 0)

func TestChoicePicksTheRowUnderTheCursor(t *testing.T) {
	var picked []string
	closed := 0
	c := NewChoice(func(k string) { picked = append(picked, k) }, func() { closed++ })
	c.Show("Sort", []Option{{"new", "Newest"}, {"old", "Oldest"}, {"az", "By name"}}, "old")
	rows := listed(t, c, 20, 5)
	if !strings.Contains(rows[1], "( ) Newest") || !strings.Contains(rows[2], "(•) Oldest") {
		t.Errorf("rows: %q", rows)
	}
	if c.GetCurrentItem() != 1 {
		t.Errorf("cursor on %d, want 1", c.GetCurrentItem())
	}
	c.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, 0), nil)
	c.InputHandler()(space, nil)
	c.Mark("az")
	if c.GetCurrentItem() != 2 {
		t.Errorf("Mark moved the cursor to %d", c.GetCurrentItem())
	}
	if rows := listed(t, c, 20, 5); !strings.Contains(rows[3], "(•) By name") || !strings.Contains(rows[2], "( ) Oldest") {
		t.Errorf("after Mark: %q", rows)
	}
	c.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, 0), nil)
	c.InputHandler()(tcell.NewEventKey(tcell.KeyEsc, 0, 0), nil)
	if strings.Join(picked, ",") != "az" || closed != 2 {
		t.Errorf("picked %q, closed %d times", picked, closed)
	}
}

func TestChecklistTogglesAndKeepsTheCursor(t *testing.T) {
	var toggled []string
	c := NewChecklist(func(k string) { toggled = append(toggled, k) }, func() {})
	rows := []Row{{Key: "flag", Label: "Flagged"}, {Key: "team", Label: "Team", Header: true}, {Key: "ann", Label: "Ann", Muted: true}}
	c.Show("Show", rows, map[string]bool{"flag": true})
	got := listed(t, c, 20, 5)
	for i, want := range []string{"[x] Flagged", "[ ] Team", "  [ ] Ann"} {
		if !strings.Contains(got[i+1], want) {
			t.Errorf("row %d: %q, want %q", i, got[i+1], want)
		}
	}
	c.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, 0), nil)
	c.InputHandler()(space, nil)
	c.Mark(rows, map[string]bool{"flag": true, "team": true})
	if c.GetCurrentItem() != 1 || strings.Join(toggled, ",") != "team" {
		t.Errorf("cursor %d, toggled %q", c.GetCurrentItem(), toggled)
	}
	if !strings.Contains(listed(t, c, 20, 5)[2], "[x] Team") {
		t.Error("Mark did not tick the header")
	}
}

func TestChecklistDrawsHeadersBoldAndMutedRowsDim(t *testing.T) {
	c := NewChecklist(func(string) {}, func() {})
	c.Show("Show", []Row{{Key: "flag", Label: "Flagged"}, {Key: "team", Label: "Team", Header: true}, {Key: "ann", Label: "Ann", Muted: true}}, nil)
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(20, 5)
	c.SetRect(0, 0, 20, 5)
	c.Draw(s)
	attr := func(x, y int) tcell.AttrMask { _, _, st, _ := s.GetContent(x, y); _, _, a := st.Decompose(); return a }
	if attr(6, 2)&tcell.AttrBold == 0 {
		t.Error("header is not bold")
	}
	if attr(8, 3)&tcell.AttrDim == 0 {
		t.Error("muted row is not dim")
	}
}

func TestDialogSizeFitsTitleLabelsAndScreen(t *testing.T) {
	c := NewChoice(func(string) {}, func() {})
	c.Show("Sort", []Option{{"a", "Newest first"}, {"b", "Old"}}, "a")
	if w, h := c.Size(80, 24); w != 20 || h != 4 {
		t.Errorf("size %dx%d, want 20x4", w, h)
	}
	if w, h := c.Size(10, 3); w != 10 || h != 3 {
		t.Errorf("size %dx%d on a small screen", w, h)
	}
	k := NewChecklist(func(string) {}, func() {})
	k.Show("A very long title here", []Row{{Key: "a", Label: "x"}}, nil)
	if w, _ := k.Size(80, 24); w != 26 {
		t.Errorf("checklist width %d, want 26", w)
	}
}
