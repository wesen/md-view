package watcher

import (
	"os"
	"path/filepath"
	"testing"
)

// TestUnwatchClosesChannel verifies that Unwatch closes the subscriber
// channel for a watched file, so the App-level `for range ch` goroutine exits
// and no further `file-changed` events are emitted for a closed file.
func TestUnwatchClosesChannel(t *testing.T) {
	fw, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = fw.Close() }()
	fw.Start()

	dir := t.TempDir()
	f := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(f, []byte("# hi\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	ch, err := fw.Watch(f)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	fw.Unwatch(f)

	// Unwatch closes the channel under fw.mu; receiving must yield ok==false.
	_, ok := <-ch
	if ok {
		t.Fatal("expected subscriber channel to be closed after Unwatch")
	}
}

// TestUnwatchUnknownPathIsNoOp verifies Unwatch is safe for a path that was
// never watched (no panic, no error).
func TestUnwatchUnknownPathIsNoOp(t *testing.T) {
	fw, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = fw.Close() }()
	fw.Start()

	// Must not panic.
	fw.Unwatch("/does/not/exist.md")
}

// TestUnwatchRemovesFromInternalMap verifies that after Unwatch the path is
// gone from the internal paths map (so Start() won't dispatch to it).
func TestUnwatchRemovesFromInternalMap(t *testing.T) {
	fw, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = fw.Close() }()
	fw.Start()

	dir := t.TempDir()
	f := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(f, []byte("a"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := fw.Watch(f); err != nil {
		t.Fatalf("Watch: %v", err)
	}

	fw.mu.Lock()
	_, before := fw.paths[f]
	fw.mu.Unlock()
	if !before {
		t.Fatal("expected path to be registered before Unwatch")
	}

	fw.Unwatch(f)

	fw.mu.Lock()
	_, after := fw.paths[f]
	fw.mu.Unlock()
	if after {
		t.Fatal("expected path to be removed from internal map after Unwatch")
	}
}
