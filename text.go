package tuingo

import (
	"strings"

	"github.com/rivo/uniseg"
)

// Width is the number of terminal columns in plain text, counting whole grapheme clusters.
// Wide characters occupy two columns; combining marks stay with their base character.
func Width(s string) int { return uniseg.StringWidth(s) }

// Wrap breaks text at word boundaries to fit w terminal columns. Words wider than w
// remain intact on their own line. Non-positive widths return no lines.
func Wrap(text string, w int) []string {
	if w <= 0 {
		return nil
	}
	var lines []string
	var cur string
	for _, word := range strings.Fields(text) {
		switch {
		case cur == "":
			cur = word
		case Width(cur)+1+Width(word) <= w:
			cur += " " + word
		default:
			lines = append(lines, cur)
			cur = word
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// Shorten keeps the leading words that fit in w columns, with an ellipsis when any were cut.
func Shorten(text string, w int) string {
	if w <= 0 {
		return ""
	}
	words := strings.Fields(text)
	if normalized := strings.Join(words, " "); Width(normalized) <= w {
		return normalized
	}
	out := ""
	for _, word := range words {
		next := word
		if out != "" {
			next = out + " " + word
		}
		if Width(next) > w-1 {
			break
		}
		out = next
	}
	if out == strings.Join(words, " ") {
		return out
	}
	return strings.TrimRight(out, " ") + "…"
}

// PadRight pads or truncates s to exactly n columns.
func PadRight(s string, n int) string {
	if n <= 0 {
		return ""
	}
	s = Truncate(s, n)
	return s + strings.Repeat(" ", n-Width(s))
}

// Truncate keeps the longest prefix fitting n terminal columns without splitting a
// grapheme cluster. Non-positive widths return an empty string.
func Truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	g := uniseg.NewGraphemes(s)
	end, width := 0, 0
	for g.Next() {
		width += g.Width()
		if width > n {
			break
		}
		_, end = g.Positions()
	}
	return s[:end]
}
