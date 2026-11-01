package tuingo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Confirm asks a yes-or-no question with buttons, answering the y, n and esc keys too. Only
// the selected button has a background of its own; the other sits on the dialog's, so the
// selection is the one thing that stands out.
func Confirm(text string, done func(yes bool)) *tview.Modal {
	m := tview.NewModal().SetText(text).AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(i int, _ string) { done(i == 0) }).
		SetButtonStyle(tcell.StyleDefault.
			Background(tview.Styles.ContrastBackgroundColor).
			Foreground(tview.Styles.PrimaryTextColor)).
		SetButtonActivatedStyle(tcell.StyleDefault.
			Background(tview.Styles.PrimaryTextColor).
			Foreground(tview.Styles.InverseTextColor))
	m.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch {
		case ev.Rune() == 'y':
			done(true)
		case ev.Rune() == 'n', ev.Key() == tcell.KeyEsc:
			done(false)
		default:
			return ev
		}
		return nil
	})
	return m
}
