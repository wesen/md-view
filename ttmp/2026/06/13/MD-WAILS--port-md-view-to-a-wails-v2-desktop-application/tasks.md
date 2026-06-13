# Tasks

Detailed, atomic task list for the MD-WAILS drop-in replacement. Check items off as they complete; each phase ends with a commit + diary update.

## Phase 0 — Scaffolding (env ready: wails v2.12.0, webkit2gtk-4.1, libsoup-3.0 present)

- [x] 0.1 Confirm wails CLI + webkit deps + go toolchain
- [x] 0.2 Layout decision: repo-root Wails project (`wails.json` + `main.go` + `frontend/` at root); old `cmd/md-view` coexists until Phase 7 cutover. **Deviation:** bound `App` in `package main` at root (not `internal/desktop`) for a predictable `window.go.main.App` binding namespace (see diary Step 3)
- [x] 0.3 Create `wails.json` (outputfilename `md-view`)
- [x] 0.4 Create `frontend/dist/{index.html, app.js, style.css}` shell (copied from demo)
- [x] 0.5 Create root `main.go` (minimal `wails.Run` + `//go:embed all:frontend/dist`)
- [x] 0.6 Create `app.go` (stub `App`: `Startup`/`Shutdown`, ctx field, 8 stub bound methods)
- [x] 0.7 Add wails/v2 dep (`go get`), tidy
- [x] 0.8 Add `.gitignore` entries (`build/`, `frontend/wailsjs/`, `frontend/node_modules/`); fix `dist/` → `/dist/` so `frontend/dist/` is tracked
- [x] 0.9 Verify: `wails dev -tags webkit2_41` opens a window (DevServer :34115, bindings generated)
- [x] 0.10 Commit

## Phase 1 — Reuse the renderer (RenderBody refactor, DR-3)

- [ ] 1.1 Add `BodyHTML` type + `RenderBody(filePath, opts) (*BodyHTML, error)` to `pkg/renderer` (frontmatter HTML + body HTML + title), reusing `extractFrontmatter`, goldmark convert, `rewriteImagePaths`
- [ ] 1.2 Decide fate of full-page `Render` (keep as `ExportHTML` or remove — since `pkg/server` is deleted in cutover)
- [ ] 1.3 Port `renderer_test.go` assertions to `RenderBody`; keep them green
- [ ] 1.4 Add `OpenFile`/`OpenFileAtPath`/`openPath` bound methods on `App`; call `renderer.RenderBody`
- [ ] 1.5 Verify: window opens a `.md` file and renders HTML
- [ ] 1.6 Commit

## Phase 2 — Assets, CSS, Chroma (DR-4)

- [ ] 2.1 Copy `base.css`, `dark.css`, `mermaid.min.js`, `mermaid-init.js`, `copy-button.js` from `pkg/renderer/static/` into `frontend/dist/`
- [ ] 2.2 Generate `frontend/dist/chroma.css` (both themes) via a tiny generator + `make chroma-css`
- [ ] 2.3 Link assets in `index.html`; wire theme toggle + mermaid init
- [ ] 2.4 Verify: dark toggle recolors code; a ```` ```mermaid ```` block renders
- [ ] 2.5 Commit

## Phase 3 — Live reload via events (replaces SSE)

- [ ] 3.1 `internal/desktop/events.go`: `Startup` starts `pkg/watcher`; per-file `Watch` → goroutine → `EventsEmit("file-changed", path)`
- [ ] 3.2 `app.js`: `EventsOn('file-changed')` → `ReopenCurrent()` → swap `#content.innerHTML` + re-run augmentations
- [ ] 3.3 Verify: edit open file on disk → window reloads within ~1s
- [ ] 3.4 Commit

## Phase 4 — Image serving (DR-5)

- [ ] 4.1 Implement `App.ServeReferencedFile` + `allowedDirs` (mirror `server.go:418`), wire into `AssetServer.Handler`
- [ ] 4.2 Confirm `rewriteImagePaths` emits `/file/<abs>` (port-independent)
- [ ] 4.3 Verify: relative image renders; traversal path → 403
- [ ] 4.4 Commit

## Phase 5 — Menus, drag-and-drop, recent files, window title

- [ ] 5.1 `BuildMenu(app)`: File (Open… Ctrl/Cmd-O, Close), View (Toggle Theme); callbacks emit events
- [ ] 5.2 `OnFileDrop(x,y,paths)` → `openPath` → `file-opened`
- [ ] 5.3 Recent-files JSON persistence (`UserConfigDir`) + `GetRecentFiles`/`RemoveRecentFile`
- [ ] 5.4 `WindowSetTitle` on open
- [ ] 5.5 Commit

## Phase 6 — Single-instance CLI dispatch + `view` command (DR-7, DR-8)

- [ ] 6.1 Add `SingleInstanceLock` to `wails.Run`; implement `OnSecondInstanceLaunch` → `ParseViewArgs` → `openPath`/`WindowShow`
- [ ] 6.2 Cobra `view` command + bare `md-view` root → `wails.Run` with `PendingOpen`; open in `OnDomReady`
- [ ] 6.3 Trim flags (keep `--dark`; remove/error `--browser`/`--no-browser`/`--port`)
- [ ] 6.4 Verify: `md-view view a.md` then `md-view view b.md --dark` reuses window
- [ ] 6.5 Commit

## Phase 7 — Cutover: delete old packages, single binary

- [ ] 7.1 Delete `pkg/daemon`, `pkg/protocol`, `pkg/server`, `pkg/commands`, `cmd/md-view/main.go`
- [ ] 7.2 `go build ./...` + `go test ./...` clean
- [ ] 7.3 Update `Makefile`, `.goreleaser.yaml`, `.github/workflows`
- [ ] 7.4 Update `docs/`, `README.md`, `AGENT.md`
- [ ] 7.5 Commit

## Phase 8 — reMarkable + toolbar buttons

- [ ] 8.1 Bound `UploadToRemarkable`, `RawFile`, `DownloadMarkdown`
- [ ] 8.2 Adapt `remarkable-button.js`, `toolbar-buttons.js` to bound methods
- [ ] 8.3 Verify + commit
