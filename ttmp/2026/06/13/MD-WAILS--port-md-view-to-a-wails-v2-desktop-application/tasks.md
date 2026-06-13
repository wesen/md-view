# Tasks

## TODO

### Phase 0 — Environment & scaffolding
- [ ] Install `libwebkit2gtk-4.1-dev libsoup-3.0-dev build-essential`
- [ ] `go install github.com/wailsapp/wails/v2/cmd/wails@latest`; run `wails doctor`
- [ ] Add repo-root `wails.json` + `main.go` + `internal/desktop/{app,events,assets,cli}.go` + `frontend/dist/{index.html,app.js,style.css}`
- [ ] Resolve entry-point: `wails build` vs `cmd/md-view` manual build (DR-2)
- [ ] `wails dev -tags webkit2_41` opens the shell

### Phase 1 — Reuse the renderer
- [ ] Add `RenderBody` to `pkg/renderer` (DR-3)
- [ ] Port renderer tests to `RenderBody`
- [ ] Call `RenderBody` from `internal/desktop/app.go` `openPath`

### Phase 2 — Assets, CSS, Chroma
- [ ] Copy `base.css`, `dark.css`, `mermaid.min.js`, `mermaid-init.js`, `copy-button.js` into `frontend/dist`
- [ ] Add `chroma-css` generator (DR-4) + `make chroma-css`
- [ ] Link assets in `index.html`; verify theme toggle + mermaid

### Phase 3 — Live reload via events
- [ ] `internal/desktop/events.go`: watcher → `EventsEmit("file-changed")`
- [ ] `app.js`: `EventsOn('file-changed')` → `ReopenCurrent()` → swap DOM

### Phase 4 — Image serving
- [ ] Implement `App.ServeReferencedFile` + `allowedDirs` (DR-5)

### Phase 5 — Menus, drag-and-drop, recent files, title
- [ ] `BuildMenu(app)` (callbacks emit events)
- [ ] `OnFileDrop` → `openPath` → `file-opened`
- [ ] Recent-files JSON persistence + `WindowSetTitle`

### Phase 6 — Single-instance CLI dispatch + `view` command (NEW)
- [ ] Add `SingleInstanceLock` (DR-7) + `OnSecondInstanceLaunch` → `ParseViewArgs` → `openPath`
- [ ] Wire Cobra `view` command + bare `md-view` root → `wails.Run` with `PendingOpen`; open in `OnDomReady`
- [ ] Trim flags per DR-8 (`--dark` kept; `--browser/--no-browser/--port` removed)
- [ ] Verify: `md-view view a.md` then `md-view view b.md --dark` reuses window

### Phase 7 — Cutover: delete old packages, single binary (NEW)
- [ ] Delete `pkg/daemon`, `pkg/protocol`, `pkg/server`, `pkg/commands`, `cmd/md-view/main.go`
- [ ] `go build ./...` + `go test ./...` clean with only renderer/watcher/desktop/root main
- [ ] Update `Makefile`, `.goreleaser.yaml`, `.github/workflows` (single `md-view` binary)
- [ ] Update `docs/user-guide.md`, `README.md`, `AGENT.md`

### Phase 8 — reMarkable + toolbar buttons
- [ ] Bound `UploadToRemarkable`, `RawFile`, `DownloadMarkdown`
- [ ] Adapt `remarkable-button.js` and `toolbar-buttons.js`

## DONE

- [x] Create ticket MD-WAILS and design/implementation guide (v1: coexistence)
- [x] Evidence-gather md-view codebase + Wails demo + article
- [x] Capture Wails sources (SingleInstanceLock API, Cobra discussion #1271, CLI-with-app #3098) into sources/
- [x] Revise guide to DROP-IN REPLACEMENT scope (single `md-view` binary, SingleInstanceLock replaces daemon)
