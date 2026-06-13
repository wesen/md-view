package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// assets holds the embedded frontend (HTML/CSS/JS). Wails serves these to
// the WebView. In `wails dev` they load from disk for hot reload; in
// `wails build` they are baked into this binary.
//
//go:embed all:frontend/dist
var assets embed.FS

// singleInstanceID is the UniqueId for Wails' SingleInstanceLock. A second
// `md-view` process with this id is caught by the lock; its os.Args are
// forwarded to instance #1 via OnSecondInstanceLaunch. This replaces the
// daemon's "reuse running server over a Unix socket" with zero filesystem state.
const singleInstanceID = "github.com/go-go-golems/md-view"

// viewFlags holds the parsed `md-view view` flags. The compatibility surface
// is `view <file> [--dark]`; browser/port flags from the old CLI are removed
// (DR-8) and unknown flags are tolerated by ParseViewArgs.
type viewFlags struct {
	dark bool
}

var (
	viewDark bool
)

func main() {
	// `md-view view [file] [--dark]` — the primary, drop-in command.
	viewCmd := &cobra.Command{
		Use:   "view [file]",
		Short: "View a markdown file in the md-view window",
		Long: `View a markdown file rendered as HTML in the md-view desktop window.

If the app is already running, the file opens in the existing window
(via the single-instance lock); otherwise a new window opens.

Examples:
  md-view view ./README.md
  md-view view --dark ./notes.md
  md-view view ./doc.md         # while running: reuses the window`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := ""
			if len(args) == 1 {
				file = args[0]
			}
			return runDesktop(file, viewDark)
		},
	}
	viewCmd.Flags().BoolVar(&viewDark, "dark", false, "Use the dark theme")

	// Bare `md-view` (no subcommand) opens an empty window — this is also what
	// happens when the binary is double-clicked. (Wails single-instance: a 2nd
	// bare launch just focuses the existing window.)
	rootCmd := &cobra.Command{
		Use:   "md-view",
		Short: "A markdown viewer desktop application",
		RunE: func(cmd *cobra.Command, args []string) error {
			file := ""
			if len(args) > 0 {
				file = args[0]
			}
			return runDesktop(file, false)
		},
	}
	rootCmd.AddCommand(viewCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// runDesktop starts the Wails app, opening `file` (if non-empty) with the
// given initial theme. It blocks until the window closes.
func runDesktop(file string, dark bool) error {
	app := NewApp()
	app.PendingOpen = file
	app.PendingDark = dark
	if dark {
		app.theme = "dark"
	}

	return wails.Run(&options.App{
		Title:     "md-view",
		Width:     1024,
		Height:    768,
		MinWidth:  480,
		MinHeight: 360,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: http.HandlerFunc(app.ServeReferencedFile), // serves /file/<abs> images (DR-5)
		},
		Menu:       buildMenu(app),
		OnStartup:  app.Startup,
		OnDomReady: app.OnDomReady,
		OnShutdown: app.Shutdown,
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               singleInstanceID,
			OnSecondInstanceLaunch: app.OnSecondInstanceLaunch,
		},
	})
}
