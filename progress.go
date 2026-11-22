package tuingo

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rivo/uniseg"
)

const progressFrame = 100 * time.Millisecond

// Progress is a bordered display of the processes started for one operation.
// All of its methods, including those on Process, run on the UI loop.
type Progress struct {
	*tview.Box
	processes []*Process
	running   int
	idle      func()
	app       *tview.Application
	timer     *time.Timer
	frame     uint
}

// Process is one bar in a Progress dialog. Start returns its handle.
type Process struct {
	progress *Progress
	name     string
	done     int
	total    int
	finished bool
}

// NewProgress creates an empty dialog. idle runs when the last running process
// finishes, so it can close the dialog. Open Center(p, p.Size) through a Host.
func NewProgress(idle func()) *Progress {
	return &Progress{
		Box:  tview.NewBox().SetBorder(true).SetTitleAlign(tview.AlignLeft),
		idle: idle,
	}
}

// SetTitle names the operation shared by the bars.
func (p *Progress) SetTitle(title string) *Progress {
	if title != "" {
		title = " " + title + " "
	}
	p.Box.SetTitle(tview.Escape(title))
	return p
}

// Start appends a running bar. A non-positive total means indeterminate work.
func (p *Progress) Start(name string, total int) *Process {
	process := &Process{progress: p, name: name, total: total}
	p.processes = append(p.processes, process)
	p.running++
	p.refreshAnimation()
	return process
}

// Advance updates the absolute completed count and total of a running bar.
func (p *Process) Advance(done, total int) {
	if p.finished {
		return
	}
	p.done, p.total = max(0, done), total
	p.progress.refreshAnimation()
}

// Finish completes the bar. It stays visible until the dialog is closed.
func (p *Process) Finish() {
	if p.finished {
		return
	}
	p.finished = true
	if p.total > 0 {
		p.done = p.total
	}
	dialog := p.progress
	dialog.running--
	dialog.refreshAnimation()
	if dialog.running == 0 && dialog.idle != nil {
		dialog.idle()
	}
}

// Size fits the title and bars, subject to the available screen dimensions.
func (p *Progress) Size(sw, sh int) (int, int) {
	width := max(44, tview.TaggedStringWidth(p.GetTitle())+4)
	for _, process := range p.processes {
		width = max(width, Width(process.label())+4)
	}
	return min(sw, width), min(sh, len(p.processes)+2)
}

func (p *Process) label() string {
	if p.total <= 0 {
		return p.name
	}
	return fmt.Sprintf("%s %d/%d", p.name, p.done, p.total)
}

// Draw shows the newest bars with an omission row if the older ones don't fit.
func (p *Progress) Draw(screen tcell.Screen) {
	p.DrawForSubclass(screen, p)
	x, y, w, h := p.GetInnerRect()
	if w <= 0 || h <= 0 {
		return
	}
	margin := min(1, (w-1)/2)
	x, w = x+margin, w-2*margin
	processes := p.processes
	if len(processes) > h {
		tview.Print(screen, "...", x, y, w, tview.AlignLeft, tview.Styles.SecondaryTextColor)
		y++
		processes = processes[len(processes)-h+1:]
	}
	for i, process := range processes {
		process.draw(screen, x, y+i, w, p.frame, p.GetBackgroundColor())
	}
}

func (p *Process) draw(screen tcell.Screen, x, y, w int, frame uint, background tcell.Color) {
	start, end := 0, 0
	style := tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(background)
	switch {
	case p.finished:
		style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorDarkGray)
	case p.total > 0:
		end = min(w, int(float64(w)*(float64(min(p.done, p.total))/float64(p.total))))
	default:
		marker := max(1, w/4)
		start = int(frame % uint(w-marker+1))
		end = start + marker
	}
	at := func(col int) tcell.Style {
		return style.Reverse(col >= start && col < end)
	}
	for col := 0; col < w; col++ {
		screen.SetContent(x+col, y, ' ', nil, at(col))
	}
	graphemes := uniseg.NewGraphemes(p.label())
	for col := 0; graphemes.Next(); {
		width := graphemes.Width()
		if col+width > w {
			break
		}
		if width > 0 {
			runes := graphemes.Runes()
			screen.SetContent(x+col, y, runes[0], runes[1:], at(col))
			col += width
		}
	}
}

func (p *Progress) setApplication(app *tview.Application) {
	p.app = app
	p.refreshAnimation()
}

func (p *Progress) refreshAnimation() {
	if p.app != nil {
		for _, process := range p.processes {
			if !process.finished && process.total <= 0 {
				if p.timer == nil {
					p.scheduleFrame(p.app)
				}
				return
			}
		}
	}
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
}

func (p *Progress) scheduleFrame(app *tview.Application) {
	// The timer's identity distinguishes a queued frame from a previous run.
	var timer *time.Timer
	timer = time.AfterFunc(progressFrame, func() {
		app.QueueUpdateDraw(func() {
			if p.timer == timer {
				p.timer = nil
				p.frame++
				p.refreshAnimation()
			}
		})
	})
	p.timer = timer
}
