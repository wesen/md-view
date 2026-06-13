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

- [x] 1.1 Add `BodyHTML` type + `RenderBody(filePath, opts) (*BodyHTML, error)` to `pkg/renderer` (frontmatter HTML + body HTML + title), reusing `extractFrontmatter`, goldmark convert, `rewriteImagePaths`
- [x] 1.2 `Render` refactored to a thin assembler over `RenderBody` (kept so `pkg/server` compiles until Phase 7 cutover)
- [x] 1.3 Added `TestRenderBody` + `TestRenderBodyWithFrontmatter`; all 8 renderer tests green
- [x] 1.4 `OpenFile`/`OpenFileAtPath`/`openPath` bound methods on `App` call `renderer.RenderBody` (return `string` HTML fragment; `FileResult` deferred)
- [x] 1.5 Verify: window renders a `.md` file (Playwright via wails dev browser mode: h1/bold/table/chroma all present)
- [x] 1.6 Commit

## Phase 2 — Assets, CSS, Chroma (DR-4)

- [x] 2.1 Copy `base.css`, `dark.css`, `mermaid.min.js`, `mermaid-init.js`, `copy-button.js` from `pkg/renderer/static/` into `frontend/dist/`
- [x] 2.2 Generate `frontend/dist/chroma.css` (both themes) + `frontend/dist/ui.css` via `cmd/gen-chroma-css` + `make frontend-css`; added `renderer.UICSS()`
- [x] 2.3 Link assets in `index.html`; `app.js` applyTheme sets data-theme on html+body + re-renders mermaid; `augment.js` re-runs copy/mermaid after each swap
- [x] 2.4 Verify: dark toggle recolors code (computed styles flip: keyword rgb(207,34,46)→rgb(255,121,198)); a ```mermaid block renders to SVG; copy button present (hover-reveal by design)
- [x] 2.5 Commit

## Phase 3 — Live reload via events (replaces SSE)

- [x] 3.1 `app.go`: watcher/mu/watched fields; `Startup` creates+starts `pkg/watcher`; `Shutdown` closes it; `openPath` calls `watchFile`. `events.go`: per-file goroutine → `EventsEmit("file-changed", path)`
- [x] 3.2 `app.js`: `EventsOn('file-changed')` → `ReopenCurrent()` → `showContent` (re-augments). Added `ReopenCurrent` bound method
- [x] 3.3 Verify: append to open file on disk → window updates within ~1s (Playwright `wait_for` marker succeeded; H2+marker appeared, original content preserved)
- [x] 3.4 Commit

## Phase 4 — Image serving (DR-5)

- [x] 4.1 Implement `App.ServeReferencedFile` + `allowedDirs` (mirror `server.go:418`), wire into `AssetServer.Handler` via `http.HandlerFunc`
- [x] 4.2 Confirmed `rewriteImagePaths` emits port-independent `/file/<abs>` (renderer unchanged)
- [x] 4.3 Verify: relative image renders (fetch → 200); traversal `/file/etc/passwd` → 403
- [x] 4.4 Commit

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
