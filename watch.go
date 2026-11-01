package tuingo

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watch watches root and every folder under it except .git, and calls fn once per burst of
// changes, after settle of quiet. root must be a directory; setup errors are returned.
// Newly created directory trees are watched on a best-effort basis. Runtime filesystem
// errors cannot be reported through this API.
//
// Callbacks run serially on a background goroutine; UI changes must go through the UI loop.
// stop cancels a pending callback and waits for an active callback to finish. It is safe
// to call stop repeatedly, but fn must not call stop or wait for a goroutine calling stop.
func Watch(root string, settle time.Duration, fn func()) (stop func(), err error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("watch %q: not a directory", root)
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	addTree := func(root string) error {
		return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				return nil
			}
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return w.Add(path)
		})
	}
	if err := addTree(root); err != nil {
		_ = w.Close()
		return nil, err
	}
	done, finished := make(chan struct{}), make(chan struct{})
	var once sync.Once
	go func() {
		defer close(finished)
		defer w.Close()
		var timer *time.Timer
		var expiry <-chan time.Time
		defer func() {
			if timer != nil {
				timer.Stop()
			}
		}()
		for {
			select {
			case <-done:
				return
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				if filepath.Base(ev.Name) == ".git" {
					continue
				}
				if ev.Has(fsnotify.Create) {
					if st, err := os.Stat(ev.Name); err == nil && st.IsDir() {
						_ = addTree(ev.Name)
					}
				}
				if timer != nil {
					timer.Stop()
				}
				timer = time.NewTimer(settle)
				expiry = timer.C
			case <-expiry:
				expiry = nil
				select {
				case <-done:
					return
				default:
					fn()
				}
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	return func() {
		once.Do(func() { close(done) })
		<-finished
	}, nil
}
