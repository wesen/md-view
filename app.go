package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-go-golems/md-view/pkg/renderer"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound application backend. Its public methods are
// automatically exposed to the JavaScript frontend as Promise-returning
// functions under window['go']['main']['App'].
//
// Phase 0 scaffold: bound methods are stubs. File rendering (Phase 1),
// live reload (Phase 3), image serving (Phase 4), menus/drag-drop/recent
// (Phase 5), single-instance CLI dispatch (Phase 6) are layered on later.
type App struct {
	ctx         context.Context // required for all Wails runtime calls
	currentFile string
	theme       string // "light" | "dark"
	recentFiles []string
}

// NewApp creates a new App instance with default values.
func NewApp() *App {
	return &App{
		theme:       "light",
		recentFiles: []string{},
	}
}

// Startup is called when the Wails application starts. The context.Context
// is required for all runtime operations (dialogs, events, window control)
// and MUST be saved as a struct field.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Shutdown is called when the application is about to quit.
func (a *App) Shutdown(_ context.Context) {
	// Phase 5: persist recent files here.
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
	runtime.WindowSetTitle(a.ctx, "md-view: "+body.Title)

	html := body.Body
	if body.Frontmatter != "" {
		html = body.Frontmatter + "\n" + html
	}
	return html, nil
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
