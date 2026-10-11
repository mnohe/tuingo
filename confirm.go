package tuingo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Confirm asks a yes-or-no question with buttons, answering the y, n and esc keys too.
func Confirm(text string, done func(yes bool)) *tview.Modal {
	m := tview.NewModal().SetText(text).AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(i int, _ string) { done(i == 0) })
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
