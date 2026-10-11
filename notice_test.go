package tuingo

import (
	"testing"
	"time"
)

func TestNoticeClearsOnlyTheLatest(t *testing.T) {
	n := NewNotice(30*time.Millisecond, func(f func()) { f() })
	n.Show("one")
	time.Sleep(15 * time.Millisecond)
	n.Show("two")
	time.Sleep(25 * time.Millisecond)
	if got := n.GetText(true); got != "two" {
		t.Errorf("the first message's expiry cleared the second: %q", got)
	}
	time.Sleep(30 * time.Millisecond)
	if got := n.GetText(true); got != "" {
		t.Errorf("the second message should have cleared: %q", got)
	}
}
