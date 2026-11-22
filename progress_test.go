package tuingo

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestProgressDisplaysNewestBarsAndCallsIdleOnce(t *testing.T) {
	calls := 0
	p := NewProgress(func() { calls++ }).SetTitle("Loading")
	p.Start("Oldest", 0).Finish()
	finished := p.Start("Checked", 0)
	running := p.Start("Loading tasks", 10)
	finished.Finish()
	running.Advance(4, 8)
	if calls != 1 {
		t.Fatalf("idle called %d times while the latest batch is running", calls)
	}
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(30, 7)
	Center(p, p.Size).Draw(s)
	if got := row(s, 4); !strings.Contains(got, "Loading tasks 4/8") {
		t.Errorf("determinate row = %q", got)
	}
	_, _, filled, _ := s.GetContent(2, 4)
	_, _, empty, _ := s.GetContent(27, 4)
	_, _, filledAttrs := filled.Decompose()
	_, _, emptyAttrs := empty.Decompose()
	if filledAttrs&tcell.AttrReverse == 0 || emptyAttrs&tcell.AttrReverse != 0 {
		t.Errorf("half-filled bar styles: %v and %v", filledAttrs, emptyAttrs)
	}
	s.SetSize(30, 5)
	p.Start("Older than these bars", 0).Finish()
	pending := p.Start("Pending", 0)
	Center(p, p.Size).Draw(s)
	for i, want := range []string{"Loading", "...", "Older than these bars", "Pending"} {
		if got := row(s, i); !strings.Contains(got, want) {
			t.Errorf("row %d = %q, want %q", i, got, want)
		}
	}
	_, _, finishedStyle, _ := s.GetContent(2, 2)
	if fg, bg, _ := finishedStyle.Decompose(); fg != tcell.ColorBlack || bg != tcell.ColorDarkGray {
		t.Errorf("finished style foreground %v, background %v", fg, bg)
	}
	running.Finish()
	running.Finish()
	if calls != 1 {
		t.Fatalf("idle called %d times with a process still pending", calls)
	}
	pending.Finish()
	if calls != 2 {
		t.Fatalf("idle called %d times after the last finish", calls)
	}
}
