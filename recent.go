package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// appName is the per-user config directory name for persistent app state
// (recent files, and later: theme preference).
const appName = "md-view"

// configPath returns the app's persistent config directory:
//
//	Linux:   $XDG_CONFIG_HOME/md-view  (~/.config/md-view)
//	macOS:   ~/Library/Application Support/md-view
//	Windows: %AppData%/md-view
func (a *App) configPath() string {
	if configDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(configDir, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "."+appName)
}

func (a *App) recentFilesPath() string {
	return filepath.Join(a.configPath(), "recent.json")
}

// loadRecentFiles reads the recent-files list from disk. Missing/corrupt file
// is non-fatal — the list stays empty.
func (a *App) loadRecentFiles() {
	data, err := os.ReadFile(a.recentFilesPath())
	if err != nil {
		return
	}
	var files []string
	if err := json.Unmarshal(data, &files); err != nil {
		return
	}
	a.recentFiles = files
}

// saveRecentFiles persists the recent-files list to disk.
func (a *App) saveRecentFiles() {
	_ = os.MkdirAll(a.configPath(), 0o755)
	data, err := json.Marshal(a.recentFiles)
	if err != nil {
		return
	}
	_ = os.WriteFile(a.recentFilesPath(), data, 0o644)
}

// pushRecent prepends path (most-recent-first), removes duplicates, and caps
// the list at 10 entries. Call saveRecentFiles on shutdown (or here) to persist.
func (a *App) pushRecent(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return
	}
	for i, f := range a.recentFiles {
		if f == abs {
			a.recentFiles = append(a.recentFiles[:i], a.recentFiles[i+1:]...)
			break
		}
	}
	a.recentFiles = append([]string{abs}, a.recentFiles...)
	if len(a.recentFiles) > 10 {
		a.recentFiles = a.recentFiles[:10]
	}
}

// isMarkdownExt reports whether path has a Markdown extension.
func isMarkdownExt(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown", ".mdown", ".mkd":
		return true
	}
	return false
}
