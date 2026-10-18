// Command docgen draws each primitive on a simulation screen and writes the cells as an SVG
// picture into the folder given as its argument, so the pictures in docs/ come from the
// same Draw methods a terminal runs.
package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/mnohe/tuingo"
)

// The terminal a picture imitates: its cell size in pixels, font, and the colors that
// stand in for the terminal's own foreground and background.
const (
	cellW, cellH = 9, 18
	fontSize     = 15
	font         = "Menlo, Consolas, 'DejaVu Sans Mono', monospace"
	screenFg     = "#d4d4d4"
	screenBg     = "#1e1e1e"
)

// scene is a picture: its file name, its size in cells, and what it draws.
type scene struct {
	name string
	w, h int
	draw func(s tcell.Screen)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: docgen <folder>")
		os.Exit(2)
	}
	tuingo.RoundedTheme()
	for _, sc := range scenes() {
		s := tcell.NewSimulationScreen("UTF-8")
		if err := s.Init(); err != nil {
			fail(err)
		}
		s.SetSize(sc.w, sc.h)
		sc.draw(s)
		err := os.WriteFile(filepath.Join(os.Args[1], sc.name+".svg"), []byte(svg(s)), 0o644)
		s.Fini()
		if err != nil {
			fail(err)
		}
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// at sets p's rectangle and draws it.
func at(s tcell.Screen, p tview.Primitive, x, y, w, h int) {
	p.SetRect(x, y, w, h)
	p.Draw(s)
}

func scenes() []scene {
	keys := []tuingo.Key{
		{Rune: 'o', Label: "Open worktree"},
		{Rune: 'y', Label: "Accept", Kind: tuingo.KeyCloses},
		{Rune: 'c', Label: "Request changes", Kind: tuingo.KeyCloses},
		{Rune: 'a', Label: "Answer", Kind: tuingo.KeyCloses},
	}
	text := []string{
		"[::b]Release notes[::-]  Draft the notes for the next version",
		"",
		"Collect the merged changes since the last tag, group them by",
		"area, and link each one to the change that introduced it.",
	}
	textBox := func(w int, plus bool) func(tcell.Screen) {
		return func(s tcell.Screen) {
			b := tuingo.NewTextBox(func(rune) {}, func() {})
			b.Show("Task", text, keys, "")
			at(s, b, 0, 0, w, 12)
			if plus {
				// The box takes + in its input capture, which tview runs before the handler.
				b.GetInputCapture()(tcell.NewEventKey(tcell.KeyRune, '+', 0))
				b.Draw(s)
			}
		}
	}
	return []scene{
		{"cardpane", 64, 12, func(s tcell.Screen) {
			p := tuingo.NewCardPane("board")
			colors := map[string][3]string{
				"plan":   {"#3b82f6", "#ffffff", "#e5e7eb"},
				"build":  {"#16a34a", "#ffffff", "#e5e7eb"},
				"review": {"#d97706", "#ffffff", "#e5e7eb"},
			}
			card := func(id string) tuingo.Card {
				c := colors[id]
				title, body := tuingo.CardStyles(c[0], c[1], c[2], id == "plan")
				return tuingo.Card{Label: id, Glyph: "●", Text: "the " + id + " stage of the task", Title: title, Body: body}
			}
			arrow := tcell.StyleDefault.Foreground(tcell.ColorGray)
			p.DrawTree(&tuingo.Node{ID: "plan", Children: []*tuingo.Node{{ID: "build"}, {ID: "review"}}}, 1, 1, 22, card, arrow)
			p.Select("plan")
			at(s, p, 0, 0, 64, 12)
		}},
		{"textbox", 72, 12, textBox(72, false)},
		{"textbox-folded", 40, 12, textBox(40, false)},
		{"textbox-panel", 40, 12, textBox(40, true)},
		{"keypanel", 24, 5, func(s tcell.Screen) {
			p := tuingo.NewKeyPanel()
			p.Set("c", "Request changes", "a", "Answer", "r", "Reassign")
			w, h := p.Size()
			at(s, p, 0, 0, w, h)
		}},
		{"confirm", 48, 9, func(s tcell.Screen) {
			at(s, tuingo.Confirm("Discard the draft?", func(bool) {}), 0, 0, 48, 9)
		}},
		{"tabs", 60, 1, func(s tcell.Screen) {
			t := tuingo.NewTabs()
			t.Set([]string{"Board", "Backlog", "Done"}, 0)
			t.SetStamp("updated 10:42")
			at(s, t, 0, 0, 60, 1)
		}},
		{"footer", 60, 1, func(s tcell.Screen) {
			f := tuingo.NewFooter()
			f.Set("←↑↓→", "move", "enter", "open", "/", "search", "q", "quit")
			at(s, f, 0, 0, 60, 1)
		}},
		{"notice", 60, 1, func(s tcell.Screen) {
			n := tuingo.NewNotice(0, func(func()) {})
			n.SetText("Release notes moved to review")
			at(s, n, 0, 0, 60, 1)
		}},
		{"divider", 40, 6, func(s tcell.Screen) {
			pinned := tview.NewTextView().SetText("pinned\npane")
			scroll := tview.NewTextView().SetText("panes that\nscroll past")
			at(s, pinned, 1, 1, 12, 4)
			at(s, tuingo.NewDivider().SetColor(tcell.ColorGray), 14, 0, 1, 6)
			at(s, scroll, 16, 1, 22, 4)
		}},
		{"center", 48, 12, func(s tcell.Screen) {
			b := tview.NewBox().SetBorder(true).SetTitle(" centered ")
			tuingo.Center(b, func(sw, sh int) (int, int) { return sw / 2, sh / 2 }).Draw(s)
		}},
	}
}

// svg writes the screen's cells as an SVG picture: a background rectangle per run of
// cells sharing a background, and a group per run sharing a style, holding a text element
// per character so each sits in its own column whatever the viewer's font.
func svg(s tcell.Screen) string {
	w, h := s.Size()
	var b strings.Builder
	fmt.Fprintf(&b, "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\">\n",
		w*cellW, h*cellH, w*cellW, h*cellH)
	fmt.Fprintf(&b, "<rect width=\"100%%\" height=\"100%%\" fill=\"%s\"/>\n", screenBg)
	fmt.Fprintf(&b, "<g font-family=\"%s\" font-size=\"%d\" xml:space=\"preserve\">\n", font, fontSize)
	for y := 0; y < h; y++ {
		for x := 0; x < w; {
			fg, bg, st := look(s, x, y)
			end := x + 1
			for end < w {
				f2, b2, s2 := look(s, end, y)
				if f2 != fg || b2 != bg || s2 != st {
					break
				}
				end++
			}
			if bg != screenBg {
				fmt.Fprintf(&b, "<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\"/>\n",
					x*cellW, y*cellH, (end-x)*cellW, cellH, bg)
			}
			var cells []string
			for c := x; c < end; c++ {
				if r, _, _, _ := s.GetContent(c, y); r != ' ' && r != 0 {
					cells = append(cells, fmt.Sprintf("<text x=\"%d\" y=\"%d\">%s</text>",
						c*cellW, y*cellH+cellH*3/4, html.EscapeString(string(r))))
				}
			}
			if len(cells) > 0 {
				fmt.Fprintf(&b, "<g fill=\"%s\"%s>%s</g>\n", fg, st, strings.Join(cells, ""))
			}
			x = end
		}
	}
	b.WriteString("</g>\n</svg>\n")
	return b.String()
}

// look is a cell's ink and fill as #rrggbb, with reverse applied, and its other attributes
// as SVG attributes.
func look(s tcell.Screen, x, y int) (fg, bg, attrs string) {
	_, _, style, _ := s.GetContent(x, y)
	f, b, a := style.Decompose()
	fg, bg = hex(f, screenFg), hex(b, screenBg)
	if a&tcell.AttrReverse != 0 {
		fg, bg = bg, fg
	}
	if a&tcell.AttrBold != 0 {
		attrs += " font-weight=\"bold\""
	}
	if a&tcell.AttrDim != 0 {
		attrs += " fill-opacity=\"0.55\""
	}
	if a&tcell.AttrUnderline != 0 {
		attrs += " text-decoration=\"underline\""
	}
	return fg, bg, attrs
}

// hex is a color as #rrggbb, or def for the terminal's default.
func hex(c tcell.Color, def string) string {
	if c == tcell.ColorDefault {
		return def
	}
	return fmt.Sprintf("#%06x", c.Hex())
}
