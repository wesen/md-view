package main

import (
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// buildMenu constructs the application menu bar. Menu callbacks run in Go and
// CANNOT touch the DOM, so each one does its work in Go and emits a Wails event
// that the frontend listens for (the golden rule: file-error / file-opened /
// theme-changed / close-file). See the Wails article "Why both channels exist".
func buildMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	// --- File menu ---
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open…", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		html, err := app.OpenFile()
		if err != nil {
			if app.ctx != nil {
				wailsruntime.EventsEmit(app.ctx, "file-error", err.Error())
			}
			return
		}
		if html != "" && app.ctx != nil {
			wailsruntime.EventsEmit(app.ctx, "file-opened", map[string]string{
				"html":  html,
				"path":  app.currentFile,
				"title": app.currentFileTitle(),
			})
		}
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Close", keys.CmdOrCtrl("w"), func(_ *menu.CallbackData) {
		// Clear backend state (currentFile + watcher + window title) before
		// the frontend cleans up the DOM. Without this, GetCurrentFile() still
		// reports the closed file, the toolbar buttons still target it, and a
		// save would re-show it via ReopenCurrent().
		app.CloseFile()
		if app.ctx != nil {
			wailsruntime.EventsEmit(app.ctx, "close-file", nil)
		}
	})

	// --- View menu ---
	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("Toggle Theme", keys.Key("t"), func(_ *menu.CallbackData) {
		theme := app.ToggleTheme()
		if app.ctx != nil {
			wailsruntime.EventsEmit(app.ctx, "theme-changed", theme)
		}
	})

	return appMenu
}
