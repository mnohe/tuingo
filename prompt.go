package tuingo

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// LinePrompt is a bordered question with a single-line input field.
type LinePrompt struct {
	*tview.Box
	question string
	field    *tview.InputField
}

// Prompt asks for a line of text. Enter submits the trimmed answer when it is
// nonempty. Escape cancels with an empty answer and false.
// Use Center to place it or Host.Prompt to manage it as a modal.
func Prompt(text string, done func(answer string, ok bool)) *LinePrompt {
	p := &LinePrompt{
		Box:      tview.NewBox().SetBorder(true),
		question: text,
		field:    tview.NewInputField(),
	}
	p.field.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEsc:
			done("", false)
		case tcell.KeyEnter:
			if answer := strings.TrimSpace(p.field.GetText()); answer != "" {
				done(answer, true)
			}
		}
	})
	return p
}

// Size fits the question and reserves space to type, within the screen.
func (p *LinePrompt) Size(sw, sh int) (int, int) {
	width := 28
	for _, line := range strings.Split(p.question, "\n") {
		width = max(width, Width(line)+4)
	}
	width = min(width, sw)
	return width, min(sh, len(p.lines(width))+4)
}

func (p *LinePrompt) lines(width int) []string {
	return tview.WordWrap(tview.Escape(p.question), max(1, width-4))
}

// Draw draws the question and input field, keeping the field visible first.
func (p *LinePrompt) Draw(s tcell.Screen) {
	p.DrawForSubclass(s, p)
	x, y, w, h := p.GetInnerRect()
	p.field.SetRect(x, y, 0, 0)
	if w <= 0 || h <= 0 {
		return
	}
	margin := min(1, (w-1)/2)
	x, w = x+margin, w-2*margin
	lines := p.lines(w + 4)
	shown := min(len(lines), h-1)
	for i, line := range lines[:shown] {
		tview.Print(s, line, x, y+i, w, tview.AlignLeft, tview.Styles.PrimaryTextColor)
	}
	fieldY := y + shown
	if shown > 0 && fieldY < y+h-1 {
		fieldY++
	}
	p.field.SetRect(x, fieldY, w, 1)
	p.field.Draw(s)
}

// Focus gives the input field keyboard focus.
func (p *LinePrompt) Focus(delegate func(tview.Primitive)) { delegate(p.field) }

// HasFocus reports whether the field has focus.
func (p *LinePrompt) HasFocus() bool { return p.field.HasFocus() }

// InputHandler forwards typing and editing to the input field.
func (p *LinePrompt) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return p.WrapInputHandler(p.field.InputHandler())
}

// PasteHandler forwards single-line paste to the input field.
func (p *LinePrompt) PasteHandler() func(string, func(tview.Primitive)) {
	return p.WrapPasteHandler(func(text string, focus func(tview.Primitive)) {
		text = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ").Replace(text)
		p.field.PasteHandler()(text, focus)
	})
}

// MouseHandler lets the input field position its cursor on a click.
func (p *LinePrompt) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return p.WrapMouseHandler(p.field.MouseHandler())
}
