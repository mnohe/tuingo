package tuingo

import (
	"strings"
	"unicode/utf8"
)

// Width is the number of columns a string takes, counting one per rune.
func Width(s string) int { return utf8.RuneCountInString(s) }

// Wrap breaks text into lines of at most w columns at word boundaries.
func Wrap(text string, w int) []string {
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
	words := strings.Fields(text)
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
	if w := Width(s); w < n {
		return s + strings.Repeat(" ", n-w)
	}
	return Truncate(s, n)
}

// Truncate keeps the first n columns of s.
func Truncate(s string, n int) string {
	if Width(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}
