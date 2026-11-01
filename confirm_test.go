package tuingo

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// buttonBackground draws m and returns the background behind the first letter of label on
// the button row.
func buttonBackground(t *testing.T, m *tview.Modal, label string) tcell.Color {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer s.Fini()
	s.SetSize(48, 9)
	m.SetRect(0, 0, 48, 9)
	m.Draw(s)
	for y := 0; y < 9; y++ {
		line := row(s, y)
		if strings.Contains(line, "Yes") && strings.Contains(line, "No") {
			x := len([]rune(line[:strings.Index(line, label)]))
			_, _, st, _ := s.GetContent(x, y)
			_, bg, _ := st.Decompose()
			return bg
		}
	}
	t.Fatal("no button row")
	return 0
}

func TestOnlyTheSelectedButtonHasABackground(t *testing.T) {
	m := Confirm("Discard the draft?", func(bool) {})
	app := tview.NewApplication()
	app.SetFocus(m)
	dialog, selected := tview.Styles.ContrastBackgroundColor, tview.Styles.PrimaryTextColor
	if bg := buttonBackground(t, m, "Yes"); bg != selected {
		t.Errorf("selected Yes on %v, want %v", bg, selected)
	}
	if bg := buttonBackground(t, m, "No"); bg != dialog {
		t.Errorf("unselected No on %v, want the dialog's %v", bg, dialog)
	}
	m.InputHandler()(tcell.NewEventKey(tcell.KeyTab, 0, 0), func(p tview.Primitive) { app.SetFocus(p) })
	if yes, no := buttonBackground(t, m, "Yes"), buttonBackground(t, m, "No"); yes != dialog || no != selected {
		t.Errorf("after tab: Yes on %v, No on %v", yes, no)
	}
}
