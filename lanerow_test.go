package tuingo

import (
	"slices"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// lane is a lane of a box, for reading back where the row put it.
func lane(key string, w int) Lane { return Lane{Key: key, Item: tview.NewBox(), Width: w} }

// laid draws the row at width w and returns each shown lane's x and width, keyed, for the
// lanes drawn.
func laid(t *testing.T, r *LaneRow, w int, lanes ...Lane) map[string][2]int {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(w, 3)
	for _, l := range lanes {
		l.Item.SetRect(-1, -1, 0, 0)
	}
	r.SetRect(0, 0, w, 3)
	r.Draw(s)
	got := map[string][2]int{}
	for _, l := range lanes {
		if x, _, lw, _ := l.Item.GetRect(); x >= 0 {
			got[l.Key] = [2]int{x, lw}
		}
	}
	return got
}

func keys(r *LaneRow) []string {
	var ks []string
	for _, l := range r.lanes {
		ks = append(ks, l.Key)
	}
	return ks
}

func TestLaneRowFillsWhatTheOthersLeave(t *testing.T) {
	a, b, c := lane("a", 10), lane("b", 0), lane("c", 12)
	var filled int
	r := NewLaneRow(8).SetFillFunc(func(w int) { filled = w })
	r.SetLanes(a, b, c)
	got := laid(t, r, 40, a, b, c)
	if got["a"] != [2]int{0, 10} || got["b"] != [2]int{10, 18} || got["c"] != [2]int{28, 12} || filled != 18 {
		t.Errorf("laid %v, filled %d", got, filled)
	}
	got = laid(t, r, 24, a, b, c)
	if got["b"] != [2]int{10, 8} || got["c"] != [2]int{18, 6} {
		t.Errorf("narrow: laid %v, want the fill at its minimum and c cut off", got)
	}
}

func TestLaneRowTellsTheOwnerItsWidthBeforeDrawing(t *testing.T) {
	a, b := lane("a", 10), lane("b", 10)
	r := NewLaneRow(0)
	r.SetLayFunc(func(w int) {
		if w < 20 {
			r.SetLanes(a)
		} else {
			r.SetLanes(a, b)
		}
	})
	if got := laid(t, r, 15, a, b); len(got) != 1 {
		t.Errorf("laid %v at 15 columns", got)
	}
}

func TestLaneRowKeepsPinnedLanesLeftmost(t *testing.T) {
	r := NewLaneRow(0, "p", "q")
	r.SetLanes(lane("a", 5), lane("q", 5), lane("b", 5), lane("p", 5))
	if got := keys(r); !slices.Equal(got, []string{"p", "q", "a", "b"}) {
		t.Errorf("order %v", got)
	}
	if r.Move("p", 1) || r.Move("a", -1) {
		t.Error("a pinned lane moved, or a lane moved past one")
	}
	if !r.Move("a", 1) || !slices.Equal(keys(r), []string{"p", "q", "b", "a"}) {
		t.Errorf("after move: %v", keys(r))
	}
}

func TestLaneRowKeepsItsOrderAcrossSetLanes(t *testing.T) {
	r := NewLaneRow(0)
	r.SetLanes(lane("a", 5), lane("b", 5), lane("c", 5))
	r.Move("a", 1)
	// b and a swapped; d is new and follows c, its predecessor in the call.
	r.SetLanes(lane("a", 5), lane("c", 5), lane("d", 5))
	if got := keys(r); !slices.Equal(got, []string{"a", "c", "d"}) {
		t.Errorf("shown %v", got)
	}
	if got := r.Order(); !slices.Equal(got, []string{"b", "a", "c", "d"}) {
		t.Errorf("order %v", got)
	}
	r.SetOrder([]string{"d", "c", "b", "a"})
	if got := keys(r); !slices.Equal(got, []string{"d", "c", "a"}) {
		t.Errorf("after SetOrder: %v", got)
	}
}

func TestLaneRowScrollsPastThePinnedLaneAndStopsAtTheEnd(t *testing.T) {
	p, a, b, c := lane("p", 10), lane("a", 10), lane("b", 10), lane("c", 10)
	r := NewLaneRow(0, "p")
	r.SetLanes(p, a, b, c)
	laid(t, r, 30, p, a, b, c)
	r.ScrollBy(1)
	got := laid(t, r, 30, p, a, b, c)
	if got["p"] != [2]int{0, 10} || got["b"] != [2]int{10, 10} || got["c"] != [2]int{20, 10} {
		t.Errorf("scrolled: %v", got)
	}
	if _, ok := got["a"]; ok {
		t.Error("a scrolled off but still drew")
	}
	r.ScrollBy(1)
	if r.Offset() != 1 {
		t.Errorf("offset %d past the last lane in view", r.Offset())
	}
	r.ScrollBy(-5)
	if r.Offset() != 0 {
		t.Errorf("offset %d before the first lane", r.Offset())
	}
}

func TestLaneRowRevealsTheLeastItTakes(t *testing.T) {
	p, a, b, c, d := lane("p", 10), lane("a", 10), lane("b", 10), lane("c", 10), lane("d", 10)
	r := NewLaneRow(0, "p")
	r.SetLanes(p, a, b, c, d)
	laid(t, r, 30, p, a, b, c, d)
	r.Reveal("d")
	if r.Offset() != 2 {
		t.Errorf("reveal d: offset %d, want 2", r.Offset())
	}
	r.Reveal("c")
	if r.Offset() != 2 {
		t.Errorf("reveal c, already in view: offset %d", r.Offset())
	}
	r.Reveal("a")
	if r.Offset() != 0 {
		t.Errorf("reveal a: offset %d, want 0", r.Offset())
	}
}

func TestLaneRowSendsFocusAndKeysToTheFillingLane(t *testing.T) {
	a := lane("a", 10)
	var got []rune
	fill := tview.NewInputField()
	fill.SetChangedFunc(func(text string) { got = []rune(text) })
	r := NewLaneRow(0).SetLanes(a, Lane{Key: "b", Item: fill})
	var focused tview.Primitive
	r.Focus(func(p tview.Primitive) { focused = p; p.Focus(nil) })
	if focused != fill {
		t.Fatalf("focus went to %T", focused)
	}
	r.InputHandler()(tcell.NewEventKey(tcell.KeyRune, 'x', 0), func(tview.Primitive) {})
	if string(got) != "x" {
		t.Errorf("filling lane got %q", string(got))
	}
}

func TestLaneRowScrollsBeforeItsFirstDraw(t *testing.T) {
	p, a, b, c := lane("p", 10), lane("a", 10), lane("b", 10), lane("c", 10)
	r := NewLaneRow(0, "p")
	r.SetLanes(p, a, b, c)
	r.ScrollBy(5)
	laid(t, r, 30, p, a, b, c)
	if r.Offset() != 1 {
		t.Errorf("offset %d after the first draw, want 1", r.Offset())
	}
}

// noHeavy is a screen that cannot show thick lines.
type noHeavy struct{ tcell.Screen }

func (noHeavy) CanDisplay(r rune, _ bool) bool { return r != '┃' && r != '━' }

func TestLaneRowDrawsPinnedBordersThick(t *testing.T) {
	boxed := func(key string) Lane {
		b := tview.NewBox().SetBorder(true).SetTitle("T").SetTitleAlign(tview.AlignLeft)
		return Lane{Key: key, Item: b, Width: 6}
	}
	for _, tc := range []struct {
		name   string
		wrap   func(tcell.Screen) tcell.Screen
		corner rune
		side   rune
	}{
		{"thick", func(s tcell.Screen) tcell.Screen { return s }, '┏', '┃'},
		{"doubled without thick lines", func(s tcell.Screen) tcell.Screen { return noHeavy{s} }, '╔', '║'},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim := tcell.NewSimulationScreen("UTF-8")
			if err := sim.Init(); err != nil {
				t.Fatal(err)
			}
			defer sim.Fini()
			sim.SetSize(12, 4)
			r := NewLaneRow(0, "p").SetLanes(boxed("p"), boxed("a"))
			r.SetRect(0, 0, 12, 4)
			r.Draw(tc.wrap(sim))
			at := func(x, y int) rune { c, _, _, _ := sim.GetContent(x, y); return c }
			if at(0, 0) != tc.corner || at(0, 1) != tc.side || at(5, 2) != tc.side {
				t.Errorf("pinned lane drawn %q %q %q, want %q and %q", at(0, 0), at(0, 1), at(5, 2), tc.corner, tc.side)
			}
			if at(1, 0) != 'T' {
				t.Errorf("pinned title became %q", at(1, 0))
			}
			if c := at(6, 1); c == tc.side {
				t.Errorf("scrolling lane's side drawn %q, want it left as drawn", c)
			}
		})
	}
}
