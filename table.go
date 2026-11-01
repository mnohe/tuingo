package tuingo

import (
	"github.com/rivo/tview"
	"strings"
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
// are empty. On narrow screens fixed columns are clipped to leave one last column.
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
	// Keep every line bounded even when the short columns alone fill the screen.
	used := 0
	for col := range widths {
		widths[col] = min(widths[col], max(0, width-1-used-2))
		used += widths[col] + 2
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
			clipped := tview.WordWrap(cell, w)
			cell = ""
			if len(clipped) > 0 {
				cell = clipped[0]
			}
			prefix.WriteString(cell + "[::BUILDSR]" + strings.Repeat(" ", max(0, w-tview.TaggedStringWidth(cell))) + "  ")
		}
		last := ""
		if n-1 < len(row) {
			last = row[n-1]
		}
		lines := tview.WordWrap(last, width-used)
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
		out[len(out)-1] += "[-:-:BUILDSR]"
	}
	return out
}

func (t Table) renderNarrow(width int) []string {
	var out []string
	for _, row := range t.Rows {
		out = append(out, tview.WordWrap(strings.Join(row, "  "), width)...)
		out = append(out, "[-:-:BUILDSR]")
	}
	return out
}
