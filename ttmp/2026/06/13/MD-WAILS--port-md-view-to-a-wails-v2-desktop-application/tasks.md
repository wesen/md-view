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

- [x] 5.1 `menu.go` `buildMenu(app)`: File (Open… Ctrl/Cmd-O → emits file-opened/file-error; Close Ctrl/Cmd-W → close-file), View (Toggle Theme T → theme-changed). Callbacks emit events
- [x] 5.2 `OnFileDrop(_,_,paths)`: first `.md` → `openPath` → `file-opened`
- [x] 5.3 Recent-files JSON persistence (`recent.go`: `os.UserConfigDir()/md-view/recent.json`, load on Startup, save on Shutdown, pushRecent dedup+cap10) + `GetRecentFiles`/`currentFileTitle`
- [x] 5.4 `WindowSetTitle` on open (already in openPath since Phase 1)
- [x] 5.5 Commit

## Phase 6 — Single-instance CLI dispatch + `view` command (DR-7, DR-8)

- [x] 6.1 `SingleInstanceLock` wired + `OnSecondInstanceLaunch` → `ParseViewArgs` → `openPath`/`WindowShow` (code correct per Wails API; **known limitation**: did not dedupe on this Linux/D-Bus setup — 2nd invocation opens a new window; user accepts multiple windows)
- [x] 6.2 Cobra `view` command + bare `md-view` root → `wails.Run` with `PendingOpen`; opened in `OnDomReady`
- [x] 6.3 Flags trimmed (keep `--dark`; `view` uses `cobra.MaximumNArgs(1)`)
- [x] 6.4 Verify: `wails build` then `build/bin/md-view view README.md` opens window titled `md-view: README.md`; `--dark` works; `ParseViewArgs` 8 unit tests green
- [x] 6.5 Commit
- [ ] **FINDING (Phase 7 blocker):** production binary MUST be built via `wails build`, NOT plain `go build` (missing Wails build tags → refuses to start). Repoint Makefile/GoReleaser/CI in Phase 7.

## Phase 7 — Cutover: delete old packages, single binary

- [x] 7.1 Delete `pkg/daemon`, `pkg/protocol`, `pkg/server`, `pkg/commands`, `cmd/md-view/main.go` (graph verified: no external importers)
- [x] 7.2 `go build ./...` + `go test ./...` clean (renderer/watcher/root main/gen-chroma-css only); `go mod tidy` dropped glazed
- [x] 7.3 Update `Makefile` (`build`→`wails build`+`frontend-css`, `run`/`test`/`clean`/`install`; removed glazed-lint/bump-glazed/serve), `.goreleaser.yaml` (`main: .`, build tags, descriptions), `.github/workflows/{push,lint}.yml` (webkit deps + `-tags webkit2_41`)
- [x] 7.4 Update `AGENT.md` (Wails build commands + structure)
- [x] 7.5 Commit (`make build` → `build/bin/md-view view README.md` opens `md-view: README.md`; lint 0 issues; tests green)

## Phase 8 — reMarkable + toolbar buttons

- [x] 8.1 Bound `UploadToRemarkable`, `RawFile`, `DownloadMarkdown` (app.go:249/276/287)
- [x] 8.2 `frontend/dist/buttons.js` (`MDSInitButtons`) calls bound methods instead of `fetch`; loaded in `index.html`, called from `app.js:showContent`
- [x] 8.3 Verify + commit (reMarkable upload landed on device at `/ai/2026/06/13`)

## Phase 9 — Documentation cutover (match the Wails single-binary model)

The implementation review (design-doc/01-implementation-review-and-lessons-learned.md) flagged the user-facing docs as the biggest non-code weakness: `README.md`, `docs/getting-started.md`, and `docs/user-guide.md` still describe the **deleted** daemon/browser model after the Phase 7 cutover. This phase rewrites them to match the actual shipped system. AGENT.md / Makefile / `.goreleaser.yaml` / `wails.json` are already correct and only need a verification pass.

- [x] 9.1 Audit stale references: confirm the only stale docs are `README.md`, `docs/getting-started.md`, `docs/user-guide.md`; confirm `AGENT.md` references are historical-only (not stale). (Done — see diary Step 13.)
- [ ] 9.2 Rewrite **`README.md`**: drop "lightweight daemon"/browser/`serve`/`status`/`stop`; replace install (build-from-source via `make build` + GoReleaser-produced native packages — NOT `go install .../cmd/md-view`, that path is deleted and CGO/Wails won't `go install`); new architecture diagram (single Wails process: Cobra + Wails runtime + WebView + `pkg/renderer`/`pkg/watcher`); correct command table (`md-view view [file] [--dark]`, bare `md-view`); features (native window, live reload via events, `/file/` image serving, Mermaid, dual-theme, frontmatter, copy/reMarkable/download, drag-drop, recent files); note Linux `libwebkit2gtk-4.1-dev` + `libsoup-3.0-dev` requirement and the known Linux multi-window limitation.
- [ ] 9.3 Rewrite **`docs/getting-started.md`**: install (build from source + native packages), first view (native window, not browser tab), live reload (event-driven — drop `--no-reload`), dark theme (toggle + `--dark` — drop URL param/localStorage), Mermaid, multiple files (each `view` opens a window — note multi-window behavior), i3/Sway (native window titles `md-view: <filename>`). Remove `serve`/`status`/`stop`/browser-selection/URL-param sections.
- [ ] 9.4 Rewrite **`docs/user-guide.md`** (major surgery): REMOVE `serve`/`status`/`stop` command sections, HTTP API (render/raw/static/SSE), Unix Socket Protocol, Daemon Management (state files, stale PID), Browser Integration (browser selection), and `--browser`/`--no-browser`/`--no-reload`/`--port` flags. KEEP + update Markdown Features, Syntax Highlighting ("server-side" → "in-process"), Mermaid Diagrams, YAML Frontmatter, Page Titles (native window title). REWRITE the `view` command (only `--dark`), Dark Theme (in-memory, no URL param/localStorage persistence yet), Live Reload (Wails events, no SSE), i3/Sway (native window), Security (`/file/` allow-list, no socket/HTTP), Troubleshooting (web-kit dev libs, "will not build without the correct build tags" error, multi-window note — remove daemon/port/browser-conflict items).
- [ ] 9.5 Verify **`AGENT.md`** has no stale operational references (its "daemon" mentions are historical context explaining the cutover — keep). No rewrite expected.
- [ ] 9.6 Update ticket bookkeeping: diary steps, `index.md` summary, `changelog.md` entry, file relations. Commit docs at each milestone.
- [ ] 9.7 Final validation: `grep -rn` the repo (excluding `ttmp/`) for `daemon`, `md-view serve`, `md-view stop`, `md-view status`, `--browser`, `--no-reload`, `Unix Socket`, `SSE`, `/render`, `/events`, `cmd/md-view` and confirm zero live operational references remain.
