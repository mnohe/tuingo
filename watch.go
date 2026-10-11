package tuingo

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watch watches root and every folder under it except .git, and calls fn once per burst of
// changes, after settle of quiet. It returns the function that stops the watch.
func Watch(root string, settle time.Duration, fn func()) (stop func(), err error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			_ = w.Add(path)
		}
		return nil
	})
	go func() {
		var timer *time.Timer
		for {
			select {
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				if strings.Contains(ev.Name, "/.git/") {
					continue
				}
				if ev.Has(fsnotify.Create) {
					if st, err := os.Stat(ev.Name); err == nil && st.IsDir() {
						_ = w.Add(ev.Name)
					}
				}
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(settle, fn)
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	return func() { _ = w.Close() }, nil
}
