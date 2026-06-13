# Agent Guidelines for go-go-golems/md-view

## Build Commands

md-view is a **Wails v2 desktop application** (single binary, since the MD-WAILS cutover). It MUST be built with `wails build`, not plain `go build` — Wails injects build tags a raw `go build` omits (the binary refuses to start otherwise).

- Run (dev, hot-reload): `wails dev -tags webkit2_41` or `make wails-dev`
- Build (production): `make build` (runs `make frontend-css` then `wails build -tags webkit2_41`) → `build/bin/md-view`
- View a file: `build/bin/md-view view README.md` (or `--dark`)
- Test: `go test -tags webkit2_41 ./...` or `make test`
- Run single test: `go test -tags webkit2_41 . -run TestParseViewArgs`
- Regenerate frontend CSS: `make frontend-css` (writes `frontend/dist/chroma.css` + `ui.css`)
- Lint: `make lint`
- Format: `go fmt ./...`

Linux requires `libwebkit2gtk-4.1-dev` + `libsoup-3.0-dev` (hence `-tags webkit2_41`).

IMPORTANT: To run the app and interact with it, use tmux so it's easy to kill. Use capture-pane to read output. When verifying bound Go methods, `wails dev` exposes a browser-accessible dev server (default http://localhost:34115) where `window.go.main.App.*` is callable — drive it with Playwright instead of simulating native dialogs.

## Project Structure

- `main.go`: Wails entry point — Cobra root + `view` command, `wails.Run`, `SingleInstanceLock`, embeds `frontend/dist`
- `app.go`: `App` struct (Wails-bound backend) — open/render/theme/recent/drop, `OnDomReady`, `OnSecondInstanceLaunch`
- `menu.go`, `events.go`, `assets.go`, `recent.go`, `cli.go`: menu, live-reload watcher, image handler, recent-files, CLI arg parsing
- `frontend/dist/`: embedded frontend (HTML/CSS/JS) — `index.html`, `app.js`, `augment.js`, `base.css`, `dark.css`, `chroma.css`, `ui.css`, mermaid, copy-button
- `pkg/renderer/`: Markdown → HTML (goldmark + chroma + frontmatter); `RenderBody` returns a chrome-free fragment
- `pkg/watcher/`: fsnotify wrapper for live reload (emits `file-changed`)
- `cmd/gen-chroma-css/`: generates `frontend/dist/{chroma,ui}.css` at build time
- `docs/`: User documentation
- `ttmp/`: Ticket workspace (MD-WAILS design + diary)

The old `pkg/{daemon,protocol,server,commands}` and `cmd/md-view` were deleted in the MD-WAILS cutover — the daemon/socket/HTTP/browser model is replaced by a single in-process Wails binary.

<runningProcessesGuidelines>
- When testing TUIs, use tmux and capture-pane to interact with the UI.
- When using tmux, try to batch as many commands as possible when using send-keys.
- When running long-running processes (servers, etc...), use tmux to more easily interact and kill them.
- Kill a process using port $PORT: `lsof-who -p $PORT -k`. When building a web server, ALWAYS use this command to kill the process.
</runningProcessesGuidelines>

<goGuidelines>
- When implementing go interfaces, use the var _ Interface = &Foo{} to make sure the interface is always implemented correctly.
- Always use a context argument when appropriate.
- Use cobra for command-line applications.
- Use the "defaults" package name, instead of "default" package name, as it's reserved in go.
- Use github.com/pkg/errors for wrapping errors.
- When starting goroutines, use errgroup.
- Only use the toplevel go.mod, don't create new ones.
- When writing a new experiment / app, add zerolog logging to help debug and figure out how it works, add --log-level flag to set the log level.
- When using go:embed, import embed as `_ "embed"`
- When using build tagged features, make sure the software compiles without the tag as well
</goGuidelines>

<debuggingGuidelines>
If me or you the LLM agent seem to go down too deep in a debugging/fixing rabbit hole in our conversations, remind me to take a breath and think about the bigger picture instead of hacking away. Say: "I think I'm stuck, let's TOUCH GRASS". IMPORTANT: Don't try to fix errors by yourself more than twice in a row. Then STOP. Don't do anything else.
</debuggingGuidelines>

<generalGuidelines>
Don't add backwards compatibility layers or adapters unless explicitly asked. If you think there is a need for a backwards compatibility or adapting to an existing interface, STOP AND ASK ME IF THAT IS NECESSARY. Usually, I don't need backwards compatibility.

If it looks like your edits aren't applied, stop immediately and say "STOPPING BECAUSE EDITING ISN'T WORKING".
</generalGuidelines>
