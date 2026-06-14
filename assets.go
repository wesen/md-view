package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// addAllowedDirTree registers a directory and all of its ancestor directories
// (up to, but NOT including, the filesystem root) as eligible for serving via
// ServeReferencedFile. Called with the opened file's directory when a file is
// opened, so that relative Markdown images such as ![](../assets/x.png) or
// ../../shared/logo.png — which renderer.rewriteImagePaths resolves to an
// absolute /file/... URL outside the file's own directory — still load.
//
// The root "/" is deliberately never registered. This keeps /etc/passwd and
// other system paths 403 even after a file deep under /home/... opens. This
// is intentionally tighter than the deleted pkg/server, which registered every
// ancestor including "/" and so effectively disabled the allow-list once any
// file was opened.
func (a *App) addAllowedDirTree(dir string) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for abs != string(filepath.Separator) {
		a.allowedDirs[abs] = struct{}{}
		parent := filepath.Dir(abs)
		if parent == abs {
			break // reached the root
		}
		abs = parent
	}
}

// isAllowed reports whether absPath is a regular file inside one of the
// allowed directories. The "+separator" prefix check prevents a directory
// named e.g. /tmp/foo from authorizing /tmp/foobar (an adjacent dir).
func (a *App) isAllowed(absPath string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for dir := range a.allowedDirs {
		if absPath == dir || strings.HasPrefix(absPath, dir+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// ServeReferencedFile is the AssetServer.Handler: Wails calls it for asset
// requests not satisfied by the embedded frontend (embed.FS). We answer
// /file/<abs-path> requests so that relative Markdown images (rewritten by
// renderer.rewriteImagePaths to /file/...) load from the open file's
// directory. Non-/file/ requests fall through with 404 (embedded assets are
// served by Wails before this handler is consulted).
//
// Security: only paths within an allowed directory are served; everything
// else is 403. This mirrors pkg/server.handleFileServing (server.go:418).
func (a *App) ServeReferencedFile(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/file/") {
		http.NotFound(w, r)
		return
	}
	// /file/<abs-path-without-leading-slash> -> /<abs-path>
	target := strings.TrimPrefix(r.URL.Path, "/file/")
	if target == "" {
		http.NotFound(w, r)
		return
	}
	if !strings.HasPrefix(target, "/") {
		target = "/" + target
	}

	absPath, err := filepath.Abs(target)
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	if !a.isAllowed(absPath) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Resolve symlinks before the allow check would be stricter; the daemon
	// model did not, and markdown image dirs are typically real paths, so we
	// keep parity. Documented as a follow-up if needed.
	info, err := os.Stat(absPath)
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}

	f, err := os.Open(absPath) // #nosec G304 -- path is allow-list checked above
	if err != nil {
		http.Error(w, "cannot open file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = f.Close() }()

	// http.ServeContent handles Range, Last-Modified, and Content-Type by
	// extension. We avoid http.ServeFile because it redirects to "clean" the
	// URL, which breaks absolute paths with a leading slash.
	http.ServeContent(w, r, filepath.Base(absPath), info.ModTime(), f)
}
