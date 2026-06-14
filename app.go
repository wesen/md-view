package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-go-golems/md-view/pkg/renderer"
	"github.com/go-go-golems/md-view/pkg/watcher"
	"github.com/wailsapp/wails/v2/pkg/options"
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
	allowedDirs map[string]struct{} // directories images may be served from (DR-5)

	// PendingOpen is the file requested on the command line before the WebView
	// was ready (first instance). OnDomReady opens it. Empty = open empty window.
	PendingOpen string
	PendingDark bool
}

// NewApp creates a new App instance with default values.
func NewApp() *App {
	return &App{
		theme:       "light",
		recentFiles: []string{},
		watched:     map[string]struct{}{},
		allowedDirs: map[string]struct{}{},
	}
}

// Startup is called when the Wails application starts. The context.Context
// is required for all runtime operations (dialogs, events, window control)
// and MUST be saved as a struct field. The file watcher is started here so a
// later-edited open file can trigger a `file-changed` event (live reload).
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.loadRecentFiles()
	// Subscribe to Wails drag-and-drop. DragAndDrop.EnableFileDrop in the app
	// options only arms the plumbing; the dropped paths are delivered as a
	// `wails:file-drop` event that must be subscribed to here, otherwise the
	// OnFileDrop handler below never fires and dropping a file does nothing.
	runtime.OnFileDrop(ctx, a.OnFileDrop)
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
	a.saveRecentFiles()
	if a.watcher != nil {
		_ = a.watcher.Close()
	}
}

// OnDomReady is called once the WebView DOM is ready. It opens the file
// requested on the command line (PendingOpen), which couldn't be rendered
// during Startup because the window didn't exist yet. This is what makes
// `md-view view README.md` actually display the file on first launch.
func (a *App) OnDomReady(_ context.Context) {
	if a.PendingOpen == "" {
		return
	}
	file := a.PendingOpen
	dark := a.PendingDark
	a.PendingOpen = ""
	a.PendingDark = false
	if dark {
		a.theme = "dark"
		runtime.EventsEmit(a.ctx, "theme-changed", a.theme)
	}
	html, err := a.openPath(file)
	if err != nil {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "file-error", err.Error())
		}
		return
	}
	if html != "" && a.ctx != nil {
		runtime.EventsEmit(a.ctx, "file-opened", map[string]string{
			"html":  html,
			"path":  a.currentFile,
			"title": a.currentFileTitle(),
		})
	}
}

// OnSecondInstanceLaunch is the SingleInstanceLock callback: Wails calls it
// in instance #1 when a second `md-view` starts, forwarding the 2nd process's
// os.Args. We parse them as `view` args and open the file in THIS window —
// this is the drop-in mechanism that replaces the daemon's "reuse running
// server" behavior with zero filesystem state.
func (a *App) OnSecondInstanceLaunch(data options.SecondInstanceData) {
	args := ParseViewArgs(data.Args)
	if args.File == "" {
		// No file requested — just bring the existing window forward.
		if a.ctx != nil {
			runtime.WindowShow(a.ctx)
		}
		return
	}
	// Resolve a relative path against the 2nd instance's working directory.
	file := args.File
	if data.WorkingDirectory != "" && !filepath.IsAbs(file) {
		file = filepath.Join(data.WorkingDirectory, file)
	}
	if args.Dark {
		a.theme = "dark"
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "theme-changed", a.theme)
		}
	}
	html, err := a.openPath(file)
	if err != nil {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "file-error", err.Error())
		}
		return
	}
	if html != "" && a.ctx != nil {
		runtime.WindowShow(a.ctx)
		runtime.EventsEmit(a.ctx, "file-opened", map[string]string{
			"html":  html,
			"path":  a.currentFile,
			"title": a.currentFileTitle(),
		})
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
	// Register the file's directory AND its ancestors (except the filesystem
	// root) so relative images like ![](../assets/x.png) or ../../shared/ that
	// resolve outside the file's own dir still load. The root "/" is never
	// registered, so /etc/passwd stays 403 even after a deep /home/... file
	// opens — tighter than the deleted pkg/server, which registered every
	// ancestor including "/" and thus disabled the allow-list once any file
	// was opened.
	a.addAllowedDirTree(filepath.Dir(abs))
	a.pushRecent(abs)
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

// currentFileTitle returns the base name of the current file ("" if none).
// Used by the menu's file-opened event payload.
func (a *App) currentFileTitle() string {
	if a.currentFile == "" {
		return ""
	}
	return filepath.Base(a.currentFile)
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

// UploadToRemarkable uploads the given markdown file to a reMarkable device via
// the `remarquee upload md` CLI. Returns the remarquee stdout (the upload
// result message). Mirrors pkg/server.handleUploadRemarkable (deleted in the
// cutover): validate path -> exec remarquee -> return output / wrapped error.
func (a *App) UploadToRemarkable(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("file not found: %s", abs)
	}
	cmd := exec.Command("remarquee", "upload", "md", abs, "--non-interactive") // #nosec G702 -- fixed args; path validated above
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		logger.Error().Str("file", abs).Str("error", errMsg).Msg("reMarkable upload failed")
		return "", errors.New(errMsg)
	}
	output := strings.TrimSpace(stdout.String())
	logger.Info().Str("file", abs).Str("output", output).Msg("reMarkable upload succeeded")
	return output, nil
}

// RawFile returns the raw bytes of a markdown file (used by the toolbar's
// "download markdown" button). Mirrors pkg/server.handleRaw (deleted).
func (a *App) RawFile(path string) ([]byte, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	return os.ReadFile(abs)
}

// DownloadMarkdown writes the given markdown file to a user-chosen location via
// a native save dialog, returning the chosen path ("" if cancelled). Used by
// the toolbar's "download" button as a friendlier alternative to RawFile.
func (a *App) DownloadMarkdown(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}
	dest, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Markdown",
		DefaultFilename: filepath.Base(abs),
	})
	if err != nil {
		return "", fmt.Errorf("save dialog error: %w", err)
	}
	if dest == "" {
		return "", nil // user cancelled
	}
	return dest, os.WriteFile(dest, data, 0o644)
}

// OnFileDrop is called when the user drops a file onto the window (requires
// DragAndDrop.EnableFileDrop in the Wails options). It picks the first
// Markdown-looking file, opens it, and emits `file-opened` so the frontend
// swaps the content — the same event path the menu uses.
func (a *App) OnFileDrop(_, _ int, paths []string) {
	for _, path := range paths {
		if !isMarkdownExt(path) {
			continue
		}
		html, err := a.openPath(path)
		if err != nil || html == "" || a.ctx == nil {
			if err != nil && a.ctx != nil {
				runtime.EventsEmit(a.ctx, "file-error", err.Error())
			}
			return
		}
		runtime.EventsEmit(a.ctx, "file-opened", map[string]string{
			"html":  html,
			"path":  a.currentFile,
			"title": a.currentFileTitle(),
		})
		return
	}
}
