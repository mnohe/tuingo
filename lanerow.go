package tuingo

import (
	"slices"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Lane is an item of a LaneRow: the key the row orders it by, the primitive it draws, and
// its width in columns. A lane of width 0 fills the columns the others leave.
type Lane struct {
	Key   string
	Item  tview.Primitive
	Width int
}

// LaneRow lays lanes out left to right at their own widths and keeps the order they are
// shown in and how far they are scrolled. Pinned lanes stay leftmost in the order they were
// pinned, never move and never scroll, and their borders are drawn thick, or doubled on a
// terminal without thick lines; the others scroll past them a lane at a time, and
// lanes past the right edge are cut off there. The owner decides which lanes are shown and
// at what width, with SetLanes; the row decides nothing about what a lane holds.
//
// Focus goes to the filling lane, or the first lane when none fills, and keys go to
// whichever lane has focus.
type LaneRow struct {
	*tview.Box
	minFill int
	pinned  []string
	order   []string
	lanes   []Lane
	offset  int
	width   int
	lay     func(width int)
	fill    func(width int)
}

// NewLaneRow returns an empty row whose filling lane is never narrower than minFill, with
// the lanes whose keys are given pinned.
func NewLaneRow(minFill int, pinned ...string) *LaneRow {
	return &LaneRow{Box: tview.NewBox(), minFill: minFill, pinned: pinned}
}

// SetLayFunc sets a function that receives the row's width before each draw, so the owner
// can choose which lanes fit and call SetLanes.
func (r *LaneRow) SetLayFunc(f func(width int)) *LaneRow {
	r.lay = f
	return r
}

// SetFillFunc sets a function that receives the filling lane's width on each draw.
func (r *LaneRow) SetFillFunc(f func(width int)) *LaneRow {
	r.fill = f
	return r
}

// SetLanes replaces the lanes shown. They are drawn in the row's order: pinned lanes first,
// then lanes the row has ordered before in that order, and lanes it has not seen join the
// order in the sequence given, after the lane given before them.
func (r *LaneRow) SetLanes(lanes ...Lane) *LaneRow {
	for i, l := range lanes {
		if slices.Contains(r.order, l.Key) {
			continue
		}
		at := len(r.order)
		for j := i - 1; j >= 0; j-- {
			if k := slices.Index(r.order, lanes[j].Key); k >= 0 {
				at = k + 1
				break
			}
		}
		r.order = slices.Insert(r.order, at, l.Key)
	}
	r.lanes = slices.Clone(lanes)
	slices.SortStableFunc(r.lanes, func(a, b Lane) int { return r.rank(a.Key) - r.rank(b.Key) })
	return r
}

// rank is where a key sorts: pinned keys first in the order they were pinned, then the rest
// in the row's order.
func (r *LaneRow) rank(key string) int {
	if i := slices.Index(r.pinned, key); i >= 0 {
		return i - len(r.pinned)
	}
	return slices.Index(r.order, key)
}

// Order returns every key the row has ordered, shown or not and pinned ones aside, so an
// owner can keep it between runs.
func (r *LaneRow) Order() []string {
	return slices.DeleteFunc(slices.Clone(r.order), r.isPinned)
}

// SetOrder replaces the row's order with keys, as Order returned them; the lanes shown are
// sorted again.
func (r *LaneRow) SetOrder(keys []string) *LaneRow {
	r.order = slices.Clone(keys)
	return r.SetLanes(r.lanes...)
}

func (r *LaneRow) isPinned(key string) bool { return slices.Contains(r.pinned, key) }

// scrolling returns the index of the first lane that scrolls.
func (r *LaneRow) scrolling() int {
	n := 0
	for n < len(r.lanes) && r.isPinned(r.lanes[n].Key) {
		n++
	}
	return n
}

// Move swaps the lane with key with its neighbour among the shown lanes that scroll, the
// one after it for a positive by and before it for a negative one, and reports whether it
// moved. A pinned lane, or one at the end it would move past, does not move.
func (r *LaneRow) Move(key string, by int) bool {
	first := r.scrolling()
	i := slices.IndexFunc(r.lanes, func(l Lane) bool { return l.Key == key })
	if i < first || by == 0 {
		return false
	}
	j := i + 1
	if by < 0 {
		j = i - 1
	}
	if j < first || j >= len(r.lanes) {
		return false
	}
	a, b := slices.Index(r.order, r.lanes[i].Key), slices.Index(r.order, r.lanes[j].Key)
	r.order[a], r.order[b] = r.order[b], r.order[a]
	r.lanes[i], r.lanes[j] = r.lanes[j], r.lanes[i]
	return true
}

// ScrollBy shifts the scrolling lanes by lanes, left for a positive count, stopping at the
// first lane and once the last lane is in view. Before the row has drawn, it knows no width,
// so the next draw stops it.
func (r *LaneRow) ScrollBy(lanes int) {
	r.offset = max(0, r.offset+lanes)
	if r.width > 0 {
		r.offset = r.clamp(r.offset)
	}
}

// Offset is how many scrolling lanes are scrolled out of view to the left.
func (r *LaneRow) Offset() int { return r.offset }

// Reveal scrolls the least it takes to bring the lane with key whole into view, or as much
// of it as fits when it is wider than the room the pinned lanes leave.
func (r *LaneRow) Reveal(key string) {
	first := r.scrolling()
	i := slices.IndexFunc(r.lanes, func(l Lane) bool { return l.Key == key })
	if i < first {
		return
	}
	if n := i - first; n < r.offset {
		r.offset = n
		return
	}
	widths, room := r.widths(r.width)
	for {
		end := 0
		for _, w := range widths[first+r.offset : i+1] {
			end += w
		}
		if end <= room || first+r.offset == i {
			break
		}
		r.offset++
	}
}

// clamp keeps an offset between the first scrolling lane and the least that brings the
// last lane into view.
func (r *LaneRow) clamp(offset int) int {
	widths, room := r.widths(r.width)
	first := r.scrolling()
	most, used := 0, 0
	for i := len(widths) - 1; i >= first; i-- {
		used += widths[i]
		if used > room {
			most = i - first + 1
			break
		}
	}
	return max(0, min(offset, most))
}

// widths returns each shown lane's width at a row width, the filling lane's computed, and
// the columns left for the lanes that scroll.
func (r *LaneRow) widths(width int) ([]int, int) {
	ws := make([]int, len(r.lanes))
	fill, taken := -1, 0
	for i, l := range r.lanes {
		if l.Width == 0 && fill < 0 {
			fill = i
			continue
		}
		ws[i] = l.Width
		taken += l.Width
	}
	if fill >= 0 {
		ws[fill] = max(r.minFill, width-taken)
	}
	room := width
	for i := range r.scrolling() {
		room -= ws[i]
	}
	return ws, room
}

// Draw tells the owner the row's width, lays the lanes out and draws them: the pinned ones
// at the left edge, their borders drawn thick to mark that they stay put, then the scrolling
// ones from the offset, cut off at the right edge.
func (r *LaneRow) Draw(screen tcell.Screen) {
	r.DrawForSubclass(screen, r)
	x, y, w, h := r.GetInnerRect()
	r.width = w
	if r.lay != nil {
		r.lay(w)
	}
	r.offset = r.clamp(r.offset)
	widths, _ := r.widths(w)
	first := r.scrolling()
	cx, right := x, x+w
	for i, l := range r.lanes {
		if i >= first && i < first+r.offset {
			continue
		}
		if l.Width == 0 && r.fill != nil {
			r.fill(widths[i])
		}
		if cx >= right {
			continue
		}
		lw := min(widths[i], right-cx)
		l.Item.SetRect(cx, y, lw, h)
		l.Item.Draw(screen)
		if i < first {
			thicken(screen, cx, y, lw, h)
		}
		cx += widths[i]
	}
}

// filling is the lane focus goes to: the one that fills, or the first.
func (r *LaneRow) filling() tview.Primitive {
	for _, l := range r.lanes {
		if l.Width == 0 {
			return l.Item
		}
	}
	if len(r.lanes) > 0 {
		return r.lanes[0].Item
	}
	return nil
}

// Focus hands focus to the filling lane.
func (r *LaneRow) Focus(delegate func(p tview.Primitive)) {
	if p := r.filling(); p != nil {
		delegate(p)
		return
	}
	r.Box.Focus(delegate)
}

// HasFocus reports whether the row or one of its lanes has focus.
func (r *LaneRow) HasFocus() bool {
	for _, l := range r.lanes {
		if l.Item.HasFocus() {
			return true
		}
	}
	return r.Box.HasFocus()
}

// InputHandler sends keys to the lane that has focus.
func (r *LaneRow) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		for _, l := range r.lanes {
			if l.Item.HasFocus() {
				if handler := l.Item.InputHandler(); handler != nil {
					handler(event, setFocus)
				}
				return
			}
		}
	})
}

// MouseHandler sends mouse events to the lane under the pointer.
func (r *LaneRow) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return r.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !r.InRect(event.Position()) {
			return false, nil
		}
		for _, l := range r.lanes {
			if handler := l.Item.MouseHandler(); handler != nil {
				if consumed, capture = handler(action, event, setFocus); consumed {
					return consumed, capture
				}
			}
		}
		return false, nil
	})
}
