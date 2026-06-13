package main

import (
	"embed"
	"log"
	"net/http"

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

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "md-view",
		Width:     1024,
		Height:    768,
		MinWidth:  480,
		MinHeight: 360,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: http.HandlerFunc(app.ServeReferencedFile), // serves /file/<abs> images (DR-5)
		},
		OnStartup:  app.Startup,
		OnShutdown: app.Shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
