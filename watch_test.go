package tuingo

import (
	"os"
	"path/filepath"
	"sync"
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

func TestWatchRejectsInvalidRoots(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, root := range []string{filepath.Join(dir, "missing"), file} {
		stop, err := Watch(root, time.Millisecond, func() {})
		if stop != nil {
			stop()
		}
		if err == nil {
			t.Errorf("Watch(%q) succeeded", root)
		}
	}
}

func TestWatchStopCancelsPendingCallback(t *testing.T) {
	dir := t.TempDir()
	calls := make(chan struct{}, 10)
	stop, err := Watch(dir, 300*time.Millisecond, func() { calls <- struct{}{} })
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if err := os.WriteFile(filepath.Join(dir, "file"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	// Allow the event to arm the debounce timer before stopping.
	time.Sleep(50 * time.Millisecond)
	stop()
	stop()
	select {
	case <-calls:
		t.Fatal("callback ran despite cancellation")
	case <-time.After(400 * time.Millisecond):
	}
}

func TestWatchStopWaitsForActiveCallback(t *testing.T) {
	dir := t.TempDir()
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	stop, err := Watch(dir, time.Millisecond, func() { close(started); <-release })
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	defer unblock()
	if err := os.WriteFile(filepath.Join(dir, "file"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("callback never started")
	}
	stopped := make(chan struct{})
	go func() { stop(); close(stopped) }()
	select {
	case <-stopped:
		t.Fatal("stop returned during callback")
	case <-time.After(20 * time.Millisecond):
	}
	unblock()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("stop did not return")
	}
}

func TestWatchAddsMovedDirectoryTreesAndSkipsGit(t *testing.T) {
	dir, staging := t.TempDir(), t.TempDir()
	nested := filepath.Join(staging, "tree", "nested")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0o700); err != nil {
		t.Fatal(err)
	}
	calls := make(chan struct{}, 10)
	stop, err := Watch(dir, 30*time.Millisecond, func() { calls <- struct{}{} })
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if err := os.WriteFile(filepath.Join(gitDir, "index"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	select {
	case <-calls:
		t.Fatal(".git write triggered callback")
	case <-time.After(100 * time.Millisecond):
	}
	if err := os.Rename(filepath.Join(staging, "tree"), filepath.Join(dir, "tree")); err != nil {
		t.Fatal(err)
	}
	wait := func() {
		t.Helper()
		select {
		case <-calls:
		case <-time.After(2 * time.Second):
			t.Fatal("missing callback")
		}
	}
	wait()
	if err := os.WriteFile(filepath.Join(dir, "tree", "nested", "file"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	wait()
}
