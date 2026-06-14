package main

import (
	"github.com/go-go-golems/logcopter/pkg/logcopter"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// logger is the structured logger for the desktop backend (logcopter/zerolog).
// Named `logger` (not `log`) to avoid clashing with the stdlib "log" used by
// main.go for startup-fatal errors.
var logger = logcopter.Package("md-view.desktop")

// watchFile registers abs with the fsnotify watcher (once per file) and spawns
// a goroutine that translates write events into a Wails `file-changed` event.
// The frontend listens for it, calls ReopenCurrent(), and swaps the content —
// this replaces the daemon model's SSE /events endpoint.
//
// Safe to call on every openPath; duplicate watches for the same path are skipped.
func (a *App) watchFile(abs string) {
	if a.watcher == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.watched[abs]; ok {
		return
	}
	ch, err := a.watcher.Watch(abs)
	if err != nil {
		logger.Warn().Err(err).Str("file", abs).Msg("cannot watch file; live reload disabled for this file")
		return
	}
	a.watched[abs] = struct{}{}
	logger.Debug().Str("file", abs).Msg("watching file for live reload")

	go func(c <-chan struct{}, path string) {
		for range c {
			runtime.EventsEmit(a.ctx, "file-changed", map[string]string{"path": path})
		}
	}(ch, abs)
}

// unwatchFile stops watching abs: it removes the path from the App's watched
// set and calls watcher.Unwatch, which closes the subscriber channel and so
// lets the watchFile goroutine above exit (no more `file-changed` events for
// this path). Called by CloseFile. Safe to call for a path that isn't watched
// (no-op). The watcher.Unwatch call is made OUTSIDE a.mu to avoid nesting the
// App and watcher mutexes; the brief window before the goroutine exits is
// harmless because CloseFile clears currentFile first.
func (a *App) unwatchFile(abs string) {
	if a.watcher == nil {
		return
	}
	a.mu.Lock()
	_, watched := a.watched[abs]
	if watched {
		delete(a.watched, abs)
	}
	a.mu.Unlock()
	if watched {
		a.watcher.Unwatch(abs)
	}
}
