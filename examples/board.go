// Run with: go run ./examples
package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/mnohe/tuingo"
)

func main() {
	tuingo.RoundedTheme()
	app := tview.NewApplication()
	pane := tuingo.NewCardPane("release board")
	tree := &tuingo.Node{ID: "plan", Children: []*tuingo.Node{
		{ID: "build", Children: []*tuingo.Node{{ID: "test"}}},
		{ID: "review"},
	}}
	summaries := map[string]string{
		"plan":   "Choose what ships in 1.0.0",
		"build":  "Keep the API small and useful",
		"test":   "Check Unicode: 界 é 👩‍💻",
		"review": "Review the release notes",
	}
	redraw := func() {
		selected := pane.Selected()
		if selected == "" {
			selected = "plan"
		}
		pane.Clear()
		pane.DrawTree(tree, 0, 0, 24, func(id string) tuingo.Card {
			title, body := tuingo.CardStyles("#3b82f6", "#ffffff", "#e5e7eb", id == selected)
			return tuingo.Card{Label: id, Glyph: "●", Text: summaries[id], Title: title, Body: body}
		}, tcell.StyleDefault.Foreground(tcell.ColorGray))
		pane.Select(selected)
	}
	pane.SetChangedFunc(redraw)
	redraw()
	footer := tuingo.NewFooter()
	footer.Set("←↑↓→ / hjkl", "move", "q / Ctrl-C", "quit")
	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(pane, 0, 1, true).
		AddItem(footer, 1, 0, false)
	app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyCtrlC || ev.Key() == tcell.KeyRune && ev.Rune() == 'q' {
			app.Stop()
			return nil
		}
		return ev
	})
	if err := app.SetRoot(root, true).Run(); err != nil {
		panic(err)
	}
}
