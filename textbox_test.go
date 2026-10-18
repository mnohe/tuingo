package tuingo

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

var closing = []Key{
	{Rune: 'o', Label: "Open worktree"},
	{Rune: 'y', Label: "Accept", Kind: KeyCloses},
	{Rune: 'c', Label: "Request changes", Kind: KeyCloses},
}

// drawn draws a box of the given width on a simulation screen and returns the screen.
func drawn(t *testing.T, v *TextBox, w int) tcell.SimulationScreen {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Fini)
	s.SetSize(w, 12)
	v.SetRect(0, 0, w, 12)
	v.Draw(s)
	return s
}

// row reads one screen row as text.
func row(s tcell.SimulationScreen, y int) string {
	w, _ := s.Size()
	var b strings.Builder
	for x := 0; x < w; x++ {
		r, _, _, _ := s.GetContent(x, y)
		b.WriteRune(r)
	}
	return b.String()
}

// edge is a border row without its right corner.
func edge(border string) string {
	r := []rune(border)
	return string(r[:len(r)-1])
}

func box(keys []Key) (*TextBox, *[]rune) {
	var got []rune
	v := NewTextBox(func(r rune) { got = append(got, r) }, func() {})
	v.Show("T", []string{"body"}, keys, "")
	return v, &got
}

func press(v *TextBox, ev *tcell.EventKey) {
	v.InputHandler()(ev, nil)
}

func TestClosingKeysGoOnTheBorder(t *testing.T) {
	v, _ := box(closing)
	s := drawn(t, v, 60)
	legend, border := row(s, 9), row(s, 11)
	if !strings.Contains(legend, "o Open worktree") || !strings.Contains(legend, "esc close") {
		t.Errorf("legend row lacks the key that stays: %q", legend)
	}
	if strings.Contains(legend, "Accept") || strings.Contains(legend, "Request changes") {
		t.Errorf("legend row shows a key that closes: %q", legend)
	}
	if !strings.HasSuffix(edge(border), " y Accept ─ c Request changes ─") {
		t.Errorf("border row: %q", border)
	}
}

func TestBoxWithoutClosingKeysDrawsAsBefore(t *testing.T) {
	v, _ := box(closing[:1])
	s := drawn(t, v, 60)
	if strings.ContainsAny(row(s, 11), "oO") {
		t.Errorf("border row has a chip: %q", row(s, 11))
	}
}

func TestChipBackgroundFollowsTheBorderColor(t *testing.T) {
	v, _ := box(closing)
	v.SetBorderColor(tcell.ColorRed)
	s := drawn(t, v, 60)
	x := strings.IndexRune(edge(row(s, 11)), 'y')
	x = len([]rune(row(s, 11)[:x]))
	_, _, style, _ := s.GetContent(x, 11)
	if fg, bg, _ := style.Decompose(); bg != tcell.ColorRed || fg != chipInk {
		t.Errorf("chip colors: fg %v bg %v", fg, bg)
	}
}

func TestChipsFoldFromTheRight(t *testing.T) {
	for _, c := range []struct {
		w    int
		want string
	}{
		{40, " y Accept ─ c Request changes ─"},
		{30, " y Accept ─ + More… ─"},
		{15, " + More… ─"},
	} {
		v, _ := box(closing)
		border := row(drawn(t, v, c.w), 11)
		if !strings.HasSuffix(edge(border), c.want) {
			t.Errorf("width %d: got %q, want suffix %q", c.w, border, c.want)
		}
	}
}

func TestFoldedKeyStillReachesOnKey(t *testing.T) {
	v, got := box(closing)
	drawn(t, v, 15)
	press(v, tcell.NewEventKey(tcell.KeyRune, 'c', 0))
	if string(*got) != "c" {
		t.Errorf("onKey got %q", string(*got))
	}
}

func TestPlusOpensThePanelAndEscClosesOnlyIt(t *testing.T) {
	v, _ := box(closing)
	closed := false
	v.close = func() { closed = true }
	drawn(t, v, 30)
	press(v, tcell.NewEventKey(tcell.KeyRune, '+', 0))
	s := drawn(t, v, 30)
	if !strings.Contains(row(s, 9), "c Request changes") {
		t.Errorf("panel does not list the folded key: %q", row(s, 9))
	}
	press(v, tcell.NewEventKey(tcell.KeyEsc, 0, 0))
	if v.open || closed {
		t.Errorf("esc: panel open %v, box closed %v", v.open, closed)
	}
}

func TestPlusDoesNothingWhenEveryChipShows(t *testing.T) {
	v, _ := box(closing)
	drawn(t, v, 60)
	press(v, tcell.NewEventKey(tcell.KeyRune, '+', 0))
	if v.open {
		t.Error("+ opened the panel with no key folded")
	}
}
