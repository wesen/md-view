package watcher

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches a file for changes and calls the callback.
type FileWatcher struct {
	watcher *fsnotify.Watcher
	mu      sync.Mutex
	paths   map[string][]notifyTarget
	done    chan struct{}
}

type notifyTarget struct {
	ch  chan struct{}
	sub <-chan struct{} // read-only view returned to callers
}

// New creates a new FileWatcher.
func New() (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("cannot create fsnotify watcher: %w", err)
	}
	return &FileWatcher{
		watcher: w,
		paths:   make(map[string][]notifyTarget),
		done:    make(chan struct{}),
	}, nil
}

// Watch registers a file for watching. Returns a channel that receives
// a signal when the file changes.
func (fw *FileWatcher) Watch(filePath string) (<-chan struct{}, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	ch := make(chan struct{}, 1)

	if _, exists := fw.paths[filePath]; !exists {
		if err := fw.watcher.Add(filePath); err != nil {
			return nil, fmt.Errorf("cannot watch file %s: %w", filePath, err)
		}
	}

	fw.paths[filePath] = append(fw.paths[filePath], notifyTarget{ch: ch, sub: ch})
	return ch, nil
}

// Start begins dispatching file change events.
func (fw *FileWatcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-fw.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fw.mu.Lock()
					if targets, ok := fw.paths[event.Name]; ok {
						for _, t := range targets {
							select {
							case t.ch <- struct{}{}:
							default:
							}
						}
					}
					fw.mu.Unlock()
				}
			case <-fw.done:
				return
			}
		}
	}()
}

// Close stops the watcher.
func (fw *FileWatcher) Close() error {
	close(fw.done)
	return fw.watcher.Close()
}

// Unwatch stops watching filePath: it removes the path from the fsnotify
// watcher and closes every subscriber channel returned by Watch for that
// path, which causes each caller's `for range ch` goroutine to exit.
//
// Both the event-dispatch send in Start() and this close run under fw.mu, so
// a send can never race with the close (no "send on closed channel" panic).
// Safe to call for a path that isn't watched (no-op).
func (fw *FileWatcher) Unwatch(filePath string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	targets, ok := fw.paths[filePath]
	if !ok {
		return
	}
	delete(fw.paths, filePath)
	_ = fw.watcher.Remove(filePath)
	for _, t := range targets {
		close(t.ch)
	}
}
