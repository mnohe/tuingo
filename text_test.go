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
		t.Errorf("pad counts terminal columns: %q", got)
	}
	if got := Truncate("◆ abcdef", 3); got != "◆ a" {
		t.Errorf("truncate counts terminal columns: %q", got)
	}
}

func TestDisplayWidthHelpers(t *testing.T) {
	for _, tc := range []struct {
		text  string
		width int
	}{
		{"界", 2}, {"e\u0301", 1}, {"👩‍💻", 2}, {"🇪🇸", 2}, {"a界e\u0301👩‍💻", 6},
	} {
		if got := Width(tc.text); got != tc.width {
			t.Errorf("Width(%q) = %d, want %d", tc.text, got, tc.width)
		}
	}
	for _, tc := range []struct {
		text string
		n    int
		want string
	}{
		{"a界b", 2, "a"}, {"a界b", 3, "a界"}, {"e\u0301x", 1, "e\u0301"},
		{"👩‍💻x", 2, "👩‍💻"}, {"🇪🇸x", 1, ""}, {"abc", -1, ""},
	} {
		if got := Truncate(tc.text, tc.n); got != tc.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tc.text, tc.n, got, tc.want)
		}
	}
	if got := PadRight("a界", 2); got != "a " || Width(got) != 2 {
		t.Errorf("padding a clipped wide character: %q", got)
	}
	if got := strings.Join(Wrap("界 界 e\u0301", 4), "|"); got != "界|界 e\u0301" {
		t.Errorf("wide wrap: %q", got)
	}
	if got := Shorten("界 a", 4); got != "界 a" {
		t.Errorf("exact fit should not add ellipsis: %q", got)
	}
	if got := Shorten("界 界", 4); got != "界…" {
		t.Errorf("wide shorten: %q", got)
	}
	if Shorten("abc", 0) != "" || PadRight("abc", -1) != "" || len(Wrap("abc", 0)) != 0 {
		t.Error("non-positive widths must return empty text")
	}
}
