package tuingo

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Confirm asks a yes-or-no question with answer chips on its bottom border.
// The y key answers Yes, n and Escape answer No, and Enter leaves it open.
func Confirm(text string, done func(yes bool)) *tview.Modal {
	m := tview.NewModal().SetText(text).SetDoneFunc(func(i int, _ string) { done(i == 0) })
	var targets []target
	m.SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
		chips := []chip{{"y", "Yes"}, {"n", "No"}}
		if rowWidth(chips) > w-3 {
			chips = []chip{{key: "y"}, {key: "n"}}
		}
		targets = drawChipRow(screen, x+w-3, y+h-1, w-3, chips, borderChipStyle(m.GetBorderColor()), targets[:0], func(i int) func() {
			return func() { done(i == 0) }
		})
		return x + 1, y + 1, w - 2, h - 2
	})
	m.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick {
			if hit, ok := targetAt(targets, event); ok {
				hit()
				return action, nil
			}
		}
		return action, event
	})
	m.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch {
		case ev.Rune() == 'y':
			done(true)
		case ev.Rune() == 'n', ev.Key() == tcell.KeyEsc:
			done(false)
		case ev.Key() == tcell.KeyEnter:
		default:
			return ev
		}
		return nil
	})
	return m
}
