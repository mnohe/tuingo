package tuingo

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchDeliversOncePerBurst(t *testing.T) {
	dir := t.TempDir()
	calls := make(chan struct{}, 10)
	stop, err := Watch(dir, 100*time.Millisecond, func() { calls <- struct{}{} })
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	for i := 0; i < 3; i++ {
		if err := os.WriteFile(filepath.Join(dir, "f"+string(rune('a'+i))), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	select {
	case <-calls:
	case <-time.After(2 * time.Second):
		t.Fatal("no call after a burst of writes")
	}
	select {
	case <-calls:
		t.Error("a burst should deliver one call")
	case <-time.After(300 * time.Millisecond):
	}
}
