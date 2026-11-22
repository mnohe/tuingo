package tuingo

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestConfirmAnswersAreClickableOnTheBorder(t *testing.T) {
	var answers []bool
	m := Confirm("Discard the draft?", func(yes bool) { answers = append(answers, yes) })
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(60, 12)
	m.Draw(s)
	_, y, _, h := m.GetRect()
	bottom := y + h - 1
	line := row(s, bottom)
	for _, chip := range []string{"y Yes", "n No"} {
		i := strings.Index(line, chip)
		if i < 0 {
			t.Fatalf("%q absent from bottom border %q", chip, line)
		}
		x := Width(line[:i])
		_, _, style, _ := s.GetContent(x, bottom)
		if _, _, attr := style.Decompose(); attr&tcell.AttrReverse == 0 {
			t.Errorf("%q is missing chip styling", chip)
		}
		m.MouseHandler()(tview.MouseLeftClick, tcell.NewEventMouse(x, bottom, tcell.Button1, 0), func(tview.Primitive) {})
	}
	if len(answers) != 2 || !answers[0] || answers[1] {
		t.Fatalf("answers %v, want [true false]", answers)
	}
}
