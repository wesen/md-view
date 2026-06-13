package main

import (
	"context"
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

// OpenFile shows a native file dialog and returns the rendered HTML.
// Phase 0 stub: returns empty (no renderer wired yet).
func (a *App) OpenFile() (string, error) {
	return "", nil
}

// OpenFileAtPath opens a specific Markdown file by absolute path.
func (a *App) OpenFileAtPath(_ string) (string, error) {
	return "", nil
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
