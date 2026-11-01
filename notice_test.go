package tuingo

import (
	"testing"
	"time"
)

func TestNoticeClearsOnlyTheLatest(t *testing.T) {
	// Queue timer updates, as the application UI loop does, and run them on this goroutine.
	updates := make(chan func(), 2)
	n := NewNotice(time.Millisecond, func(f func()) { updates <- f })
	next := func() func() {
		t.Helper()
		select {
		case f := <-updates:
			return f
		case <-time.After(2 * time.Second):
			t.Fatal("notice did not queue its expiry")
			return nil
		}
	}
	n.Show("one")
	first := next()
	n.Show("two")
	first()
	if got := n.GetText(true); got != "two" {
		t.Errorf("the first message's expiry cleared the second: %q", got)
	}
	next()()
	if got := n.GetText(true); got != "" {
		t.Errorf("the second message should have cleared: %q", got)
	}
}
