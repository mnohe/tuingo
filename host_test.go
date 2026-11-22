package tuingo

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// key sends ev to the host the way the application does: only when the root has focus,
// with the application's SetFocus.
func key(app *tview.Application, h *Host, ev *tcell.EventKey) {
	if h.HasFocus() {
		h.InputHandler()(ev, func(p tview.Primitive) { app.SetFocus(p) })
	}
}

func esc() *tcell.EventKey     { return tcell.NewEventKey(tcell.KeyEsc, 0, 0) }
func r(c rune) *tcell.EventKey { return tcell.NewEventKey(tcell.KeyRune, c, 0) }

// screen is a lane and a details box that closes on esc, as on a board.
func screen(t *testing.T) (*tview.Application, *Host, *tview.Box, *TextBox, *bool) {
	app := tview.NewApplication()
	lane := tview.NewBox()
	closed := false
	h := (*Host)(nil)
	details := NewTextBox(func(rune) {}, func() { closed = true; h.Close("details") })
	h = NewHost(app, lane)
	app.SetRoot(h, true)
	app.SetFocus(lane)
	h.Open("details", details)
	return app, h, lane, details, &closed
}

func TestConfirmKeepsTheKeyboardWhenFocusMoves(t *testing.T) {
	app, h, _, details, closed := screen(t)
	var answers []bool
	h.Confirm("Accept?", func(yes bool) { answers = append(answers, yes) })
	// The board reloads and reopens the details while the question is open.
	app.SetFocus(details)
	key(app, h, esc())
	if *closed {
		t.Fatal("esc reached the details under the dialog")
	}
	if len(answers) != 1 || answers[0] {
		t.Fatalf("answers %v, want [false]", answers)
	}
	if app.GetFocus() != details {
		t.Errorf("focus on %T after the answer, want the details", app.GetFocus())
	}
	if h.HasPage(confirmPage) || !h.Modal() {
		t.Error("the confirm is still open, or the details closed with it")
	}
}

func TestConfirmAnswersYesAfterFocusLeftForALane(t *testing.T) {
	app, h, lane, _, _ := screen(t)
	var answers []bool
	h.Confirm("Accept?", func(yes bool) { answers = append(answers, yes) })
	app.SetFocus(lane)
	key(app, h, r('y'))
	if len(answers) != 1 || !answers[0] {
		t.Fatalf("answers %v, want [true]", answers)
	}
}

func TestConfirmEnterLeavesTheQuestionOpen(t *testing.T) {
	app, h, lane, _, _ := screen(t)
	var answers []bool
	h.Confirm("Accept?", func(yes bool) { answers = append(answers, yes) })
	app.SetFocus(lane)
	key(app, h, tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if len(answers) != 0 || !h.HasPage(confirmPage) {
		t.Fatalf("enter answered the question: answers %v", answers)
	}
}

func TestPromptKeepsTypingAfterFocusMovesAndTrimsTheAnswer(t *testing.T) {
	app, h, lane, details, _ := screen(t)
	var answer string
	var submitted bool
	h.Prompt("Why?", func(text string, ok bool) { answer, submitted = text, ok })
	app.SetFocus(lane)
	key(app, h, r(' '))
	key(app, h, tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if !h.HasPage(promptPage) {
		t.Fatal("blank answer closed the prompt")
	}
	for _, ch := range "  ready  " {
		key(app, h, r(ch))
	}
	key(app, h, tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if !submitted || answer != "ready" || app.GetFocus() != details {
		t.Fatalf("answer %q, submitted %v, focus %T", answer, submitted, app.GetFocus())
	}
	h.Prompt("Why?", func(text string, ok bool) { answer, submitted = text, ok })
	key(app, h, r('x'))
	key(app, h, esc())
	if submitted || answer != "" {
		t.Fatalf("cancel reported %q, %v", answer, submitted)
	}
}

func TestClosingRestoresFocusThroughStackedDialogs(t *testing.T) {
	app, h, lane, _, _ := screen(t)
	help := tview.NewBox()
	h.Open("help", help)
	h.Close("details")
	if app.GetFocus() != help {
		t.Fatalf("closing a lower dialog moved focus to %T", app.GetFocus())
	}
	h.Close("help")
	if app.GetFocus() != lane {
		t.Errorf("focus on %T, want the lane the details opened over", app.GetFocus())
	}
}

func TestHostForgetsADialogTheApplicationRemoved(t *testing.T) {
	app, h, lane, _, _ := screen(t)
	h.RemovePage("details")
	app.SetFocus(lane)
	if h.Modal() {
		t.Error("host still holds a removed dialog")
	}
	got := false
	lane.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey { got = true; return nil })
	key(app, h, r('x'))
	if !got {
		t.Error("key did not reach the screen with no dialog open")
	}
}
