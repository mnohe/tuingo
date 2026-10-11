package tuingo

import (
	"strings"
	"testing"
)

func TestWrapAndShorten(t *testing.T) {
	if got := strings.Join(Wrap("one two three four", 9), "|"); got != "one two|three|four" {
		t.Errorf("wrap: %q", got)
	}
	if got := Shorten("one two three four", 10); got != "one two…" {
		t.Errorf("shorten: %q", got)
	}
	if got := Shorten("one two", 10); got != "one two" {
		t.Errorf("shorten keeps text that fits: %q", got)
	}
	if got := PadRight("◆ ab", 6); got != "◆ ab  " || Width(got) != 6 {
		t.Errorf("pad counts runes: %q", got)
	}
	if got := Truncate("◆ abcdef", 3); got != "◆ a" {
		t.Errorf("truncate counts runes: %q", got)
	}
}
