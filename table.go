package tuingo

import (
	"strings"

	"github.com/rivo/tview"
	"github.com/rivo/uniseg"
)

// Block is text which lays itself out at a supplied number of terminal columns.
type Block interface{ Render(width int) []string }

// Lines is a block of ordinary dynamic-color text.
type Lines []string

// Render returns the lines; TextBox wraps ordinary prose as usual.
func (l Lines) Render(width int) []string { return append([]string(nil), l...) }

// Table lays out cells separated by two spaces. All columns but the last take
// their widest printed cell; the last wraps under itself in the remaining space.
// Cells accept tview style tags. Rows may have different lengths; missing cells
// are empty. On screens too narrow for the fixed columns, each row wraps from the left.
// A grapheme wider than the available line is shown as a replacement character.
type Table struct{ Rows [][]string }

// Render lays out the table at width, counting terminal cells rather than tags.
func (t Table) Render(width int) []string {
	if width <= 0 {
		return nil
	}
	n := 0
	for _, row := range t.Rows {
		n = max(n, len(row))
	}
	if n == 0 {
		return nil
	}
	widths := make([]int, n-1)
	for _, row := range t.Rows {
		for col := range widths {
			if col < len(row) {
				widths[col] = max(widths[col], tview.TaggedStringWidth(row[col]))
			}
		}
	}
	used := 0
	for _, w := range widths {
		used += w + 2
	}
	if used >= width {
		return t.renderNarrow(width)
	}
	var out []string
	for _, row := range t.Rows {
		var prefix strings.Builder
		for col, w := range widths {
			cell := ""
			if col < len(row) {
				cell = row[col]
			}
			prefix.WriteString(cell + tableReset + strings.Repeat(" ", max(0, w-tview.TaggedStringWidth(cell))) + "  ")
		}
		last := ""
		if n-1 < len(row) {
			last = row[n-1]
		}
		lines := tableWrap(last, width-used)
		if len(lines) == 0 {
			lines = []string{""}
		}
		for i, line := range lines {
			lead := prefix.String()
			if i > 0 {
				lead = strings.Repeat(" ", used)
			}
			out = append(out, lead+line)
		}
		out[len(out)-1] += tableReset
	}
	return out
}

func (t Table) renderNarrow(width int) []string {
	var out []string
	for _, row := range t.Rows {
		lines := tableWrap(strings.Join(row, tableReset+"  "), width)
		if len(lines) == 0 {
			lines = []string{""}
		}
		lines[len(lines)-1] += tableReset
		out = append(out, lines...)
	}
	return out
}

// Explicitly clear underline as well as the other attributes: tview v0.42.0's
// attribute reset leaves underline set.
const tableReset = "[-:-:BUILDSR]"

// tableWrap uses tview's tag-aware word wrapping. A terminal cannot display a
// two-column grapheme in one column, so replace just that grapheme, retaining
// its surrounding style tags and all other text.
func tableWrap(text string, width int) []string {
	if width <= 0 {
		return nil
	}
	if width == 1 {
		var b strings.Builder
		graphemes := uniseg.NewGraphemes(text)
		for graphemes.Next() {
			if graphemes.Width() > width {
				b.WriteRune('�')
			} else {
				b.WriteString(graphemes.Str())
			}
		}
		text = b.String()
	}
	return tview.WordWrap(text, width)
}
