package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/go-go-golems/md-view/pkg/renderer"
	"github.com/go-go-golems/md-view/pkg/watcher"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound application backend. Its public methods are
// automatically exposed to the JavaScript frontend as Promise-returning
// functions under window['go']['main']['App'].
type App struct {
	ctx         context.Context // required for all Wails runtime calls
	currentFile string
	theme       string // "light" | "dark"
	recentFiles []string
	watcher     *watcher.FileWatcher // fsnotify wrapper for live reload
	mu          sync.Mutex
	watched     map[string]struct{} // files already watched (avoid duplicate watches)
}

// NewApp creates a new App instance with default values.
func NewApp() *App {
	return &App{
		theme:       "light",
		recentFiles: []string{},
		watched:     map[string]struct{}{},
	}
}

// Startup is called when the Wails application starts. The context.Context
// is required for all runtime operations (dialogs, events, window control)
// and MUST be saved as a struct field. The file watcher is started here so a
// later-edited open file can trigger a `file-changed` event (live reload).
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	fw, err := watcher.New()
	if err != nil {
		logger.Warn().Err(err).Msg("cannot create file watcher; live reload disabled")
		return
	}
	a.watcher = fw
	fw.Start()
}

// Shutdown is called when the application is about to quit.
func (a *App) Shutdown(_ context.Context) {
	if a.watcher != nil {
		_ = a.watcher.Close()
	}
}

// --- Bound methods (stubs for Phase 0; implemented in later phases) ---

// OpenFile shows a native file dialog and returns the rendered HTML fragment
// (frontmatter block + body). Returns "" if the user cancels.
func (a *App) OpenFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Markdown File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Markdown Files", Pattern: "*.md;*.markdown;*.mdown;*.mkd"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("file dialog error: %w", err)
	}
	if path == "" {
		return "", nil // user cancelled
	}
	return a.openPath(path)
}

// OpenFileAtPath opens a specific Markdown file by absolute path and returns
// the rendered HTML fragment. Used by recent-files clicks and drag-and-drop.
func (a *App) OpenFileAtPath(path string) (string, error) {
	return a.openPath(path)
}

// openPath is the shared implementation: resolve → render → set state → set title.
// It returns the HTML fragment the frontend swaps into #content.innerHTML.
// Phase 4 will add the image-serving allow-list here; Phase 3 adds file watching.
func (a *App) openPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}

	body, err := renderer.RenderBody(abs, renderer.Options{})
	if err != nil {
		return "", err
	}

	a.currentFile = abs
	a.watchFile(abs)
	runtime.WindowSetTitle(a.ctx, "md-view: "+body.Title)

	html := body.Body
	if body.Frontmatter != "" {
		html = body.Frontmatter + "\n" + html
	}
	return html, nil
}

// ReopenCurrent re-renders the currently open file and returns the new HTML
// fragment. The frontend calls it in response to the `file-changed` event
// (live reload): it swaps #content and re-augments. Returns "" if no file is open.
func (a *App) ReopenCurrent() (string, error) {
	if a.currentFile == "" {
		return "", nil
	}
	return a.openPath(a.currentFile)
}

// GetCurrentFile returns the absolute path of the currently open file.
func (a *App) GetCurrentFile() string {
	return a.currentFile
}

// GetRecentFiles returns the list of recently opened file paths.
func (a *App) GetRecentFiles() []string {
	return a.recentFiles
}

// ToggleTheme switches between light and dark themes and returns the new theme.
func (a *App) ToggleTheme() string {
	if a.theme == "light" {
		a.theme = "dark"
	} else {
		a.theme = "light"
	}
	return a.theme
}

// GetTheme returns the current theme name.
func (a *App) GetTheme() string {
	return a.theme
}

// OnFileDrop is called when the user drops a file onto the window
// (Wails v2 DragAndDrop must be enabled in options).
func (a *App) OnFileDrop(_, _ int, _ []string) {}
