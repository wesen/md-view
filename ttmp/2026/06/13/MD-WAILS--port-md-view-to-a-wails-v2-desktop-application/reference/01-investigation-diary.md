---
title: "Investigation diary"
ticket: MD-WAILS
doc-type: reference
status: active
intent: long-term
topics: [markdown, go, architecture, wails, desktop, web]
---

# Investigation diary

## Goal

Capture the chronological investigation that produced the MD-WAILS design/implementation guide: what was read, what was learned, what the open questions are, and how a reviewer or the implementing intern should validate the analysis.

## Step 1: Gather evidence from both source repos and the Wails article

The task was to create a ticket that ports `md-view` to Wails v2, with an intern-grade design/implementation guide. Before writing any conclusions, I read the two provided sources end to end and mapped the existing md-view codebase.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to port md-view to use wails.io. Read ~/code/wesen/2026-06-13--wails-demo/ and also /home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/06/13/ARTICLE - Wails v2 Desktop Applications - Technical Deep Dive.md Create a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Produce a docmgr ticket (MD-WAILS) with a comprehensive, evidence-backed design+implementation guide aimed at a new intern, then deliver it to reMarkable. The guide must explain both md-view (current state) and Wails (target), map one to the other, and give a phased file-level plan.

**Inferred user intent:** A reusable onboarding + implementation document that lets an unfamiliar engineer execute the port without rediscovering the architecture or the Wails sharp edges.

### What I did

- Read the two provided sources in full:
  - `/home/manuel/code/wesen/2026-06-13--wails-demo` — `app.go`, `main.go`, `render.go`, `render_test.go`, `wails.json`, `go.mod`, `frontend/dist/{index.html,app.js}`, `frontend/wailsjs/go/main/App.js`.
  - The Wails v2 deep-dive article in the Obsidian vault.
- Read the current md-view codebase: `cmd/md-view/main.go`, `pkg/renderer/renderer.go`, `pkg/server/server.go`, `pkg/daemon/daemon.go`, `pkg/protocol/protocol.go`, `pkg/watcher/watcher.go`, `pkg/commands/{run,view,serve,status,stop}.go`, the build-tag files, the embedded `static/*` JS, `go.mod`, `Makefile`, `README.md`, `AGENT.md`.
- Captured precise line numbers for every key symbol (via `grep -n`) so the guide's references are verifiable.
- Added `wails`, `desktop`, `architecture` topics to the docmgr vocabulary (they were missing).

### What worked

- The two-source comparison made the gap analysis crisp: md-view's renderer is reusable, its process plumbing (daemon/socket/HTTP) is what Wails replaces.
- The Wails article already documents the "menu callbacks must emit events" and `context.Context` rules, which became core guidance in the design doc.
- `grep -n` line anchoring made file references exact and auditable.

### What didn't work

- Nothing failed. One thing to note: the demo's `render.go` uses `gomarkdown`+`bluemonday`, while md-view uses `goldmark`+`goldmark-highlighting`. I flagged this explicitly in the design doc (§5.3 and DR-3) so the intern does **not** copy the demo's renderer.

### What I learned

- md-view's `Render(filePath, opts)` is already a pure function — exactly the shape Wails needs. The only structural change is splitting it into body-fragment vs. full-page assembly (DR-3) so the WebView can own page chrome.
- Live reload maps cleanly: `fsnotify` watcher (`pkg/watcher`, already reusable) → goroutine → `runtime.EventsEmit("file-changed")`. SSE + `reload.js` are retired.
- Image serving is the most subtle port: `AssetServer.Handler` reuses the existing `/file/` URL scheme and allow-list (DR-5), avoiding a rewrite of `rewriteImagePaths`.

### What was tricky to build

- **Two communication channels.** Wails bound-method vs. event distinction is the #1 bug source. I made it a first-class section (§4.3) and a golden rule, and used events everywhere Go initiates DOM updates (menus, file drop, live reload).
- **Asset handler semantics.** Wails calls `AssetServer.Handler` only for requests not in the embedded FS; the design relies on this to route `/file/...` without colliding with real assets. Flagged as a risk to validate.
- **Coexistence vs. replacement.** Per `AGENT.md`'s "no backwards-compat shims without asking," I chose coexistence (DR-2): a new `wailsapp/` entry point, the CLI untouched — no adapters.

### What warrants a second pair of eyes

- DR-3's `RenderBody` refactor: confirm the CLI/server path still works after the split (the full-page assembler must remain callable by `pkg/server`).
- DR-5's path handling on Windows (drive letters, leading-slash stripping in `/file/...`).
- Whether a CSP is needed given md-view does not run Bluemonday today (open question OQ-3).

### What should be done in the future

- Implement the phases in order; Phase 0 (env + scaffolding) first.
- Add a CI `wails build` job on a pinned Linux image (webkit2gtk-4.1).
- Follow-up ticket for theme persistence and (optionally) a CSP.

### Code review instructions

- Start at the design doc: `design-impl-guide/01-wails-port-analysis-design-and-implementation-guide.md`.
- Verify the gap-analysis table (§6) against the cited files/lines.
- Spot-check 3 decision records (DR-1, DR-3, DR-5) against the actual code they reference.
- Confirm `pkg/watcher` and `pkg/renderer` are genuinely GUI-independent (they are).

### Technical details

- Docmgr ticket: `ttmp/2026/06/13/MD-WAILS--port-md-view-to-a-wails-v2-desktop-application/`.
- Vocabulary additions: `wails`, `desktop`, `architecture`.
- Validation command before upload: `docmgr doctor --ticket MD-WAILS --stale-after 30`.

## Step 2: Revise scope to a drop-in replacement (single binary) and capture sources

After the v1 guide (coexistence: two binaries) was written, the user clarified the real goal: **replace** md-view entirely with a single Wails binary that is **drop-in compatible at the CLI** (`md-view view README.md` still works). This is a fundamental architecture change, not a tweak.

### Prompt Context

**User prompt (verbatim):** "I actually want to replace the current md-view app with this, so that it's dropin compatible CLI wise. Update the document." then "store interesting resources into th esource/ folder of the ticket so that we can reference them later on"

**Assistant interpretation:** (1) Rewrite the design doc from "ship a second `md-view-desktop` binary alongside the CLI" to "one `md-view` binary that is both the CLI and the desktop app, with `view` preserved and daemon verbs removed". (2) Capture the Wails references I found into the ticket's `sources/` folder for later citation.

**Inferred user intent:** A single install/distribution artifact; existing `md-view view <file>` muscle memory and scripts keep working; no daemon/socket/PID state.

### What I did

- Researched the Wails multi-invocation mechanism. Found that Wails v2.7.0+ has a **built-in `SingleInstanceLock`** option (no plugin) — captured its exact API from `pkg/options/options.go` into `sources/01-wails-single-instance-lock-api.md`.
- Captured two GitHub Discussions via the REST API: Cobra+Wails coexistence (#1271) and the "one binary, CLI+GUI" use case (#3098) into `sources/02-…` and `sources/03-…`, plus a `00-sources-index.md`.
- Rewrote the design doc (v2): new §2.2 CLI-compatibility contract; gap-analysis rows changed (daemon/socket/PID → `SingleInstanceLock`; Cobra CLI → adapted; serve/stop/status → deleted); package layout changed from a separate `wailsapp/` to a repo-root Wails project + `internal/desktop/` + deletion of `pkg/{daemon,protocol,server,commands}`; DR-2 flipped from coexistence to replacement; added DR-7 (SingleInstanceLock) and DR-8 (flag trimming); new pseudocode for `main.go` (Cobra + wails.Run + SingleInstanceLock), `OnSecondInstanceLaunch`, and `OnDomReady`; phases restructured (Phase 6 = single-instance dispatch, Phase 7 = cutover/deletion).
- Updated tasks/changelog/index to match.

### What worked

- The `SingleInstanceLock` API is the clean, zero-state replacement for md-view's entire daemon+socket+PID subsystem — it forwards the 2nd process's `os.Args` directly to instance #1. This made the drop-in story concrete instead of hand-wavy.
- GitHub raw source + Discussions API fetched cleanly; only the Cloudflare-protected `wails.io` guide pages needed to be reconstructed from the source file (more authoritative anyway).

### What didn't work

- `wails.io` docs are behind a Cloudflare "Just a moment…" challenge and cannot be fetched headless with curl. Worked around by taking the authoritative type definitions from the `options.go` source.

### What I learned

- "Drop-in CLI compatible" for a GUI app really means two things: (a) the primary verb (`view <file> [--dark]`) still works, and (b) a second invocation reuses the running app. Wails `SingleInstanceLock` gives (b) natively; Cobra-in-main.go gives (a).
- The double-click gotcha (Discussion #1271): a bare binary launch must still open the GUI, so the Cobra root command itself must run the app — not require a subcommand.

### What was tricky to build

- **Entry-point tension.** go-go-golems convention is `cmd/<binary>/main.go`, but `wails build` expects `main.go` next to `wails.json`. Documented both paths (repo-root `main.go`, or `cmd/md-view/main.go` via `wails generate module` + manual `go build`) and flagged it as a Phase-0 resolution (DR-2).
- **Second-instance timing.** `OnSecondInstanceLaunch` can fire before the first instance's DOM is ready; the file must be buffered in `PendingOpen` and opened in `OnDomReady`. Captured in the pseudocode and risks.

### What warrants a second pair of eyes

- DR-2's entry-point choice and whether moving off `cmd/md-view/main.go` is acceptable to the project conventions.
- DR-8's flag-trimming: confirm removing `--port`/`--browser` won't break known user scripts (or keep `--no-reload` as a no-op).
- The `serve`/`stop`/`status` removal — verify nothing in CI or docs depends on them.

### What should be done in the future

- Execute Phases 0–8; the cutover (Phase 7) is the irreversible step.
- Add a CI `wails build` job on a pinned webkit2gtk-4.1 image.

### Code review instructions

- Diff the v2 design doc against v1: the changed sections are §1, §2, §6 rows, §7.1, DR-1/2/3, new DR-7/8, §9.1/9.2/9.3, §10 phases 6/7, §12, §13.
- Verify the `SingleInstanceLock` API in `sources/01-…` matches what the pseudocode uses.
- Confirm the deletion list (Phase 7) is complete: `pkg/daemon`, `pkg/protocol`, `pkg/server`, `pkg/commands`, `cmd/md-view/main.go`.

### Technical details

- Sources stored at `sources/{00-sources-index,01-wails-single-instance-lock-api,02-wails-cobra-integration-discussion-1271,03-wails-cli-with-app-discussion-3098}.md`.
- `SingleInstanceLock` added in Wails v2.7.0; demo pins v2.12.0 → satisfied.

## Step 3: Phase 0 — scaffold the Wails app and open a window

With the design locked as a drop-in replacement, work began on the actual codebase. Phase 0's goal was narrowly scoped: get a native window to open from `wails dev`, with the bound `App` struct in place, so every later phase has a running surface to build on. The renderer, menus, and single-instance lock all come later.

### Prompt Context

**User prompt (verbatim):** "Add detailed tasks to the ticket, then work on them one by one, keeping a detailed diary as you work, and committing at appropriate intervals. (see skills if needed)"

**Assistant interpretation:** Turn the design doc's phase plan into an atomic, checkable task list, then execute the phases in order — implementing, verifying, diary-updating, and committing at each checkpoint.

**Inferred user intent:** Move from design to working software, with an auditable trail (diary) and clean git history (per-phase commits).

### Environment check (before scaffolding)

- Wails CLI **v2.12.0** already installed at `/home/manuel/go/bin/wails` (≥ v2.7.0 → `SingleInstanceLock` available).
- `webkit2gtk-4.1` and `libsoup-3.0` **present** (pkg-config); webkit2gtk-4.0 absent (fine, we target 4.1).
- Go **1.26.3**.

So Phase 0's install steps were already satisfied; only scaffolding remained.

### What I did

- Wrote a detailed atomic `tasks.md` (Phase 0–8, each with sub-items and a commit checkpoint).
- **Layout decision (deviation from design doc §7.1):** the design proposed `internal/desktop/` for the bound `App`. Wails generates the JS binding namespace from the *package name*, and `internal/...` packages produce awkward/hard-to-predict namespaces. To keep the demo's proven `window['go']['main']['App']` namespace (so the copied frontend works unchanged), the bound `App` lives in **`package main` at the repo root**. Pure testable helpers (CLI arg parsing, path checks) will go in `internal/desktop` later. This is layering, not an adapter.
- Created the repo-root Wails project: `wails.json` (outputfilename `md-view`, empty frontend build/install — vanilla JS), `main.go` (`wails.Run` + `//go:embed all:frontend/dist`), `app.go` (stub `App` with `Startup`/`Shutdown` + stub bound methods: `OpenFile`, `OpenFileAtPath`, `GetCurrentFile`, `GetRecentFiles`, `ToggleTheme`, `GetTheme`, `OnFileDrop`).
- Copied the demo's `frontend/dist/{index.html,app.js,style.css,chroma.css}` as the shell.
- `go get github.com/wailsapp/wails/v2@v2.12.0` + `go mod tidy`. Note: tidy prunes Wails if nothing imports it — had to write `main.go` (which imports wails) *before* tidy, else wails gets removed.
- Fixed a `.gitignore` collision: the bare `dist/` rule (for GoReleaser output) was silently ignoring `frontend/dist/`. Re-anchored it to `/dist/` (root build output only).

### Why

- A stub-with-window first proves the toolchain end-to-end (embed, bindings, WebView) before any rendering logic is added. Failures here are cheap to find.
- Rooting the Wails project at the repo root (rather than a subdir) matches the demo exactly and keeps the final "single `md-view` binary" story simple (`go build .` / GoReleaser root).

### What worked

- `wails dev -tags webkit2_41` compiled, generated bindings, packaged, and opened a window; DevServer on `http://localhost:34115`; "Serving assets from disk: …/frontend/dist".
- The generated `frontend/wailsjs/go/main/App.js` confirms the namespace is `window['go']['main']['App']` with all 8 stub methods bound — the copied frontend's calls will resolve.
- Both binaries coexist during transition: `go build -tags webkit2_41 .` (new Wails app) and `go build ./cmd/md-view` (old CLI) both succeed — no package conflict from having two `main` packages.

### What didn't work

- **`go mod tidy` removed Wails on the first run** because no Go file imported it yet. Fix: author the importing `main.go` first, then tidy. Obvious in hindsight; recorded so it isn't repeated.
- **`frontend/dist/` was invisible to git** due to the `dist/` ignore rule. Found via `git check-ignore -v`.

### What I learned

- Wails dev regenerates `frontend/wailsjs/` on every run — correctly gitignored (`/frontend/wailsjs/`) to avoid churn.
- Two `main` packages (root + `cmd/md-view`) compile together cleanly; the cutover (Phase 7) simply deletes the old one.

### What was tricky to build

- The `.gitignore` `dist/` shadowing was non-obvious — `git status` just silently omitted `frontend/`, which would have meant committing a broken app (no frontend assets) if not caught. The root-anchor (`/dist/`) is the surgical fix that preserves the GoReleaser-output ignore.

### What warrants a second pair of eyes

- The decision to keep the bound `App` in `package main` at root (vs `internal/desktop` as the design proposed). It trades a namespace-predictability guarantee for slightly less tidy packaging; confirm this is acceptable, or move `App` into a non-`internal` subpackage (e.g. `pkg/desktop`) later — that still gives a clean `window.go.desktop.App` namespace and testability.
- The `.gitignore` `/dist/` change — verify GoReleaser/install still output to the right place.

### What should be done in the future

- Revisit package layout if `package main` root grows unwieldy.
- The stub `App` methods must be implemented in Phases 1, 3, 5, 6.

### Code review instructions

- Start at repo root: `main.go` (embed + `wails.Run`), `app.go` (stub `App`), `wails.json`.
- Validate: `wails dev -tags webkit2_41` opens a window; `frontend/wailsjs/go/main/App.js` lists the 8 bound methods.
- Confirm `.gitignore` line 18 is now `/dist/` and `frontend/dist/index.html` is tracked.

### Technical details

- Commit (this step): see `git log` for "feat(MD-WAILS): scaffold Wails v2 app (Phase 0)".
- `wails dev` build tag: `-tags webkit2_41` (webkit2gtk-4.1 present, 4.0 absent).
- Embed: `//go:embed all:frontend/dist` in `main.go`.

## Step 4: Phase 1 — RenderBody refactor + wire the renderer to the window

Phase 1's goal was to make the Wails window actually render a Markdown file, by reusing md-view's existing Goldmark renderer through a new fragment-returning `RenderBody`, and wiring it to a bound `App.openPath`.

### Prompt Context

**User prompt (verbatim):** (same as Step 3 — "Add detailed tasks … work on them one by one, keeping a detailed diary … committing at appropriate intervals.")

**Assistant interpretation:** Execute Phase 1 of the task list: refactor `pkg/renderer` to expose a chrome-free body fragment, keep the legacy full-page `Render` delegating to it (so `pkg/server` keeps compiling until the Phase 7 cutover), and make `App.OpenFile`/`OpenFileAtPath` render a real file.

**Inferred user intent:** Prove the renderer→Wails-bridge data path end to end (the core of the whole port) before adding styling, live reload, and the single-instance CLI dispatch.

### What I did

- **Added `BodyHTML` + `RenderBody`** to `pkg/renderer/renderer.go`. `RenderBody(filePath, opts) (*BodyHTML, error)` returns `{Frontmatter, Body, Title}` — the read→frontmatter→Goldmark→`rewriteImagePaths`→title core, with no `<html>`/`<head>`/CSS/`<script>`.
- **Refactored `Render`** to be a thin assembler over `RenderBody` (it calls it, then wraps the fragment with page chrome). This keeps `pkg/server` (the only `Render` consumer) compiling through the transition, per DR-3.
- **Added 2 `RenderBody` tests** (`TestRenderBody`, `TestRenderBodyWithFrontmatter`) covering title resolution, body content, the "no page chrome" invariant, and frontmatter formatting.
- **Wired `App.OpenFile`/`OpenFileAtPath`/`openPath`** in `app.go`: native file dialog (`runtime.OpenFileDialog`) → `renderer.RenderBody` → set `currentFile` + `runtime.WindowSetTitle` → return `frontmatter + "\n" + body` as the HTML string. Kept the return type as `string` (not the design's `FileResult` struct) to stay compatible with the copied demo frontend, which does `.then(html => showContent(html))`. `FileResult` can be introduced later if the frontend is upgraded.
- **Verified end to end** via `wails dev`'s browser mode + Playwright: navigated to `http://localhost:34115`, polled for `window.go.main.App` readiness, called `OpenFileAtPath('/tmp/md-view-test.md')`, and asserted the returned HTML contained `<h1>Hello md-view</h1>`, `<strong>bold</strong>`, `<table>`, and `chroma` classes. All passed.

### Why

- Splitting `Render` into `RenderBody` + assembler is the linchpin of DR-3: the WebView owns page chrome (stable `index.html`), and only the content fragment is swapped per file. This makes theme toggling, Mermaid re-init, and asset caching far simpler than full-page reloads.
- Keeping `Render` delegating (rather than deleting it) lets `pkg/server` keep compiling during the transition, so `go build ./...` and `go test ./...` stay green until the Phase 7 cutover deletes `pkg/server`.

### What worked

- All 8 renderer tests green after the refactor (6 existing `Render`-based + 2 new `RenderBody`); the refactor was behavior-preserving for `Render`.
- `gofmt` clean; `go build -tags webkit2_41 ./...` builds the new root main, the old `cmd/md-view`, and all packages together — the two-`main` coexistence still holds.
- Wails dev browser mode is an excellent verification tool: it exposes bound Go methods at `window.go.main.App.*` over HTTP, so I can drive the app from Playwright without simulating native dialog clicks.

### What didn't work

- **First refactor edit failed** ("Could not find edits[0]"): I mistyped the Goldmark options block's closing-paren nesting in the `oldText`. Fixed by re-reading the exact bytes with `sed -n | cat -A` (tabs show as `^I`) and reconstructing the `oldText` precisely. Lesson: for multi-line Go edits with nested calls, always source the exact whitespace from `cat -A` rather than transcribing by eye.
- Minor: `favicon.ico` 404 in the console (no favicon served). Harmless; deferred to a later cleanup.

### What I learned

- `rewriteImagePaths`'s `port` parameter is genuinely unused (it builds port-independent `/file/...` URLs) — confirmed by reading the body, so `RenderBody` passes `opts.Port` through purely for `Options` API compatibility. This matters for Phase 4 (image serving needs no port).
- The Wails binding contract for a `(string, error)` Go method is a JS `Promise<string>`; the demo frontend already consumes it as a plain string. Changing the return to a struct would require regenerating bindings and updating the frontend — deferred.

### What was tricky to build

- **Preserving `Render`'s exact output while extracting its core.** The title resolution and frontmatter formatting lived mid-function in `Render`; moving them into `RenderBody` required updating every downstream reference in `Render` (`title`, `fmHTML`, `renderedHTML` → `body.Title`, `body.Frontmatter`, `body.Body`). Verified by the unchanged `Render` tests passing.
- **The edit-tool whitespace sensitivity.** Documented above.

### What warrants a second pair of eyes

- The decision to return a plain `string` from `OpenFile`/`OpenFileAtPath` (vs the design's `FileResult{HTML,Path,Title}`). It works and matches the frontend today, but if we later want path/title without extra round-trips, `FileResult` is better. Acceptable to evolve.
- `Render` is now a thin assembler kept alive only for `pkg/server`. Confirm we still want it deleted at cutover (Phase 7) vs. kept as an `ExportHTML` helper.

### What should be done in the future

- Phase 2: swap the demo's `style.css`/`chroma.css` for md-view's `base.css` + generated dual-theme Chroma CSS so content looks like the real md-view.
- Add a favicon (or a 204 handler) to silence the console 404.

### Code review instructions

- `pkg/renderer/renderer.go`: `BodyHTML` + `RenderBody` (new), `Render` (now delegates).
- `pkg/renderer/renderer_test.go`: `TestRenderBody`, `TestRenderBodyWithFrontmatter` (new).
- `app.go`: `OpenFile`, `OpenFileAtPath`, `openPath` (implemented).
- Validate: `go test ./pkg/renderer -count=1`; `wails dev -tags webkit2_41` then call `window.go.main.App.OpenFileAtPath('/tmp/md-view-test.md')` in the browser.

### Technical details

- Commit (this step): see `git log` for "feat(MD-WAILS): RenderBody refactor + wire renderer to window (Phase 1)".
- Verification artifact: `.playwright-mcp/phase1-render.png` (gitignored).

## Step 5: Phase 2 — assets, dual-theme Chroma CSS, augmentation

Phase 2 made the rendered content look like the real md-view (GitHub styling, dual-theme syntax highlighting) and wired the Mermaid + copy-button augmentation to re-run after each content swap — a necessity in the Wails fragment model (only `#content` changes; scripts don't re-execute).

### Prompt Context

**User prompt (verbatim):** (same as Step 3 — work through the tasks, diary, commit.)

**Assistant interpretation:** Execute Phase 2: copy md-view's static assets into the frontend, generate the dual-theme Chroma CSS at build time, link everything, and make Mermaid + copy buttons work on dynamically-swapped content.

**Inferred user intent:** Feature parity with md-view's visual output (styling, dark mode, diagrams, copy buttons) inside the desktop window.

### What I did

- Copied `base.css`, `dark.css`, `mermaid.min.js`, `mermaid-init.js`, `copy-button.js` from `pkg/renderer/static/` into `frontend/dist/`.
- **DR-4 generator:** created `cmd/gen-chroma-css` that writes `frontend/dist/chroma.css` (dual-theme: github light + dracula dark prefixed with `[data-theme="dark"]`) AND `frontend/dist/ui.css` (md-view's in-page UI chrome — frontmatter block, copy/reMarkable/toolbar buttons — via a new public `renderer.UICSS()` wrapper around the existing `themeCSS`). Added `make frontend-css`.
- Added `renderer.UICSS()` (public) returning `themeCSS(false)` (which already includes light + dark overrides).
- **`augment.js`:** ported `copy-button.js` + `mermaid-init.js` into idempotent, re-runnable functions exposed as `window.MDSAugmentPage()` (copy buttons + mermaid render) and `window.MDSMermaidRerender(theme)` (re-render diagrams on theme change). Idempotency guards prevent double-wrapping.
- Updated `index.html` to link `style.css` (app chrome) + `base.css` + `chroma.css` + `ui.css` + `dark.css`, and load `mermaid.min.js` + `augment.js` + `app.js`.
- Updated `app.js`: `applyTheme` now sets `data-theme` on **both** `<html>` and `<body>` (md-view's CSS targets `[data-theme="dark"]` on an ancestor) and calls `MDSMermaidRerender`; `showContent` calls `MDSAugmentPage()` after every content swap.

### Why

- The daemon model ran augmentation IIFEs once per full page load. The Wails model keeps page chrome stable and swaps only `#content`, so augmentation must be callable repeatedly — hence `augment.js`.
- Generating CSS once (vs. per render) is DR-4: deterministic, cacheable, and it lets the renderer's `ChromaCSSBoth`/`UICSS` helpers produce static assets instead of being inlined on every `Render`.

### What worked

- Verified end-to-end via `wails dev` browser mode + Playwright against `/tmp/md-view-phase2.md` (frontmatter + go code + mermaid + table): frontmatter details block present, 1 copy button (after a guard to skip `language-mermaid`), mermaid rendered to **SVG**, all correct.
- **Theme toggle verified by computed style:** light keyword `rgb(207,34,46)` / code bg `#f6f8fa` vs dark keyword `rgb(255,121,198)` / bg `#161b22` — both color and background flip. A fresh-backend screenshot pair (light + dark via the real theme button) confirmed GitHub styling, Mermaid diagrams, and readable contrast in both themes.
- All renderer tests green; `go build -tags webkit2_41 ./...` clean; `make frontend-css` regenerates both CSS files.

### What didn't work

- **First `applyTheme` test showed "no color change":** I only set `data-theme` on `<html>`, but `<body>` retained the stale `dark` value from a prior toggle, so the dark CSS kept matching via the body. Fix: `applyTheme` sets **both** elements. This is a real invariant — document it: the data-theme attribute must be consistent on whatever element(s) the CSS targets as ancestors.
- **Vision QA flagged "copy button not visible":** that's by design — `.md-view-copy-btn { opacity: 0 }`, revealed on `.md-view-code-container:hover` (inherited from the original md-view). A static screenshot can't show a hover-revealed control; the DOM (count=1) and base.css confirm it's correct.
- **Screenshot showed dark when light was expected (once):** a `wails dev` hot reload re-ran `app.js` init, which called `GetTheme()` → the backend's persisted `dark` → reverted my manual light override. Restarting dev for a clean backend fixed it. Lesson: when testing theming through the dev server, remember backend `App` state persists across browser reloads.

### What I learned

- Chroma token classes for Go are subtypes (`kn` keyword-namespace, `kd` keyword-declaration, `nx`, `s`, …), not bare `.k` — a `.chroma .k` selector matches nothing; use `.kn`/`.kd` to probe. The dual-theme CSS itself is fine.
- `themeCSS(dark bool)` ignores its `dark` parameter — it always returns light + darkOverrides concatenated. So `UICSS() = themeCSS(false)` gives both themes; the param is vestigial.

### What was tricky to build

- **Augmentation ordering:** `initCopyButtons` ran before `initMermaid`, so the mermaid code block got a copy button before being converted to a diagram. Fixed by skipping `language-mermaid` code blocks in `initCopyButtons`. Without this, a stray copy button sat next to each diagram.
- **data-theme placement & consistency** (documented above).

### What warrants a second pair of eyes

- The `applyTheme` dual-attribute (`html` + `body`) set: confirm this is acceptable vs. picking one canonical element (md-view's full-page `Render` set it on `<html>` only). Both work as long as consistent; setting both is defensive.
- Whether `ui.css` (generated from `themeCSS`) is the right home for the button chrome, or whether those rules should migrate into `base.css`/`dark.css` and `UICSS`/`themeCSS` be retired.

### What should be done in the future

- Phase 3: live reload (watcher → `file-changed` event → `ReopenCurrent()` → re-augment). The `MDSAugmentPage` re-run hook is already in place.
- Consider adding a favicon handler to silence the console 404.

### Code review instructions

- `pkg/renderer/renderer.go`: new `UICSS()`.
- `cmd/gen-chroma-css/main.go`: writes `frontend/dist/{chroma,ui}.css`.
- `Makefile`: `frontend-css`, `wails-dev`, `wails-build` targets.
- `frontend/dist/{index.html,app.js,augment.js,base.css,dark.css,chroma.css,ui.css,mermaid.min.js,copy-button.js}`.
- Validate: `make frontend-css`; `wails dev -tags webkit2_41`; render `/tmp/md-view-phase2.md`; toggle theme.

### Technical details

- Commit (this step): see `git log` for "feat(MD-WAILS): assets, dual-theme Chroma CSS, augmentation (Phase 2)".
- Verification artifacts: `phase2-light-clean.png`, `phase2-dark-clean.png` (gitignored, in repo root).

## Step 6: Phase 3 — live reload via Wails events (replaces SSE)

Phase 3 restored the "edit the file, the view refreshes" behavior without any HTTP/SSE. The existing `pkg/watcher` (fsnotify) is reused; a goroutine translates each write into a Wails `file-changed` event; the frontend calls `ReopenCurrent()` and swaps the content.

### Prompt Context

**User prompt (verbatim):** (same as Step 3 — work through the tasks, diary, commit.)

**Assistant interpretation:** Execute Phase 3: wire `pkg/watcher` into `App.Startup`, emit `file-changed` per watched file, add a `ReopenCurrent` bound method, and have `app.js` listen and re-render.

**Inferred user intent:** Feature parity with the daemon model's SSE live reload, but in-process and event-driven.

### What I did

- Added `watcher *watcher.FileWatcher`, `mu sync.Mutex`, and `watched map[string]struct{}` to `App`. `Startup` creates + starts the watcher; `Shutdown` closes it. Duplicate watches are skipped.
- `openPath` now calls `a.watchFile(abs)` after setting `currentFile`.
- Added `events.go` with `watchFile(abs)`: registers the path with `watcher.Watch` (once) and spawns a goroutine that ranges over its `<-chan struct{}` and calls `runtime.EventsEmit(a.ctx, "file-changed", {path})`.
- Added the `ReopenCurrent() (string, error)` bound method: re-renders `currentFile` via `openPath` (returns `""` if none open).
- Added a `file-changed` listener in `app.js` that calls `ReopenCurrent()` and `showContent(html)` (which re-runs augmentation).
- Structured logging via `logcopter.Package("md-view.desktop")` as `logger` (named to avoid clashing with stdlib `log` used in `main.go` for fatal-on-startup).

### Why

- The daemon model pushed reloads over an SSE `/events` stream; the file watcher fanned events to all SSE clients. In a single-process Wails app there's no HTTP, so the watcher → Wails event → bound-method path is the direct equivalent, with no network hop.
- Reusing `pkg/watcher` unchanged keeps the cutover (Phase 7) smaller.

### What worked

- Verified end to end via `wails dev` browser mode + Playwright: opened `/tmp/md-view-reload.md` (H1 "Original Title", first line), appended `## Edited section\n\nEDITED_LIVE_RELOAD_MARKER …` to the file on disk, and `wait_for("EDITED_LIVE_RELOAD_MARKER")` succeeded. The post-edit DOM had the new `<h2>Edited section</h2>` and the marker, while the original H1 + first line were preserved (a content swap, not a full reload).
- gofmt clean; `go build -tags webkit2_41 ./...` ok; `go test ./...` green.

### What didn't work

- **`var log` collided with stdlib `log`** (imported by `main.go` for `log.Fatal` on startup error): `./events.go:9:5: log already declared through import of package log`. Fixed by naming the logcopter logger `logger` instead. A package can't have both a `var log` and an imported `"log"`; the convention here is `logger` for the zerolog instance, `log` for stdlib.
- The first batched edit to rename `log`→`logger` partially failed (edits targeting `app.go` were in the `events.go` call) — a copy/paste error in my edit targeting. Re-scoped to one file per concern and it applied cleanly.
- The Wails window was closed mid-session by the user; closing the window exits the app (tmux session gone) — expected desktop behavior. Verification had already completed before the close.

### What I learned

- Closing the main Wails window terminates the process (no tray-by-default). Relevant for Phase 6's single-instance story: when instance #1's window closes, the app exits, so a subsequent `md-view view` correctly starts fresh (no lingering instance) — `SingleInstanceLock` will simply not find a lock.
- `watcher.Watch` is idempotent-enough to call per openPath, but `fsnotify.Add` on an already-watched path is a no-op that still returns a *new* channel per `Watch` call — hence the `watched` set guard to avoid spawning duplicate goroutines.

### What was tricky to build

- **The `var log` vs stdlib `log` clash** is the kind of thing that only surfaces at compile across two files in the same package. Documented so Phase 5/6 additions use `logger` consistently.
- Ensuring the watcher goroutine doesn't leak: it exits when the channel closes (`watcher.Close` in `Shutdown`), and the `watched` set prevents duplicates.

### What warrants a second pair of eyes

- The `watched` map never has entries removed (files stay watched for the app's lifetime). Acceptable for a single-window viewer; if multi-file/multi-tab is added later, eviction on close would be needed.
- Live reload only fires for the *currently open* file's path (the frontend compares `data.path`); opening file B stops updates for file A until A is reopened. Confirm this matches expectations (it mirrors the daemon model, which watched the file in the active tab).

### What should be done in the future

- Phase 4: image serving (`AssetServer.Handler` + allow-list).
- Add a favicon handler to clear the recurring 404.

### Code review instructions

- `app.go`: new fields (`watcher`, `mu`, `watched`), `Startup`/`Shutdown` watcher lifecycle, `openPath` calls `watchFile`, new `ReopenCurrent` bound method.
- `events.go` (new): `watchFile` → goroutine → `EventsEmit("file-changed")`.
- `frontend/dist/app.js`: `file-changed` listener → `ReopenCurrent()` → `showContent`.
- Validate: `wails dev -tags webkit2_41`; open a file; append to it on disk; the window updates within ~1s.

### Technical details

- Commit (this step): see `git log` for "feat(MD-WAILS): live reload via Wails events (Phase 3)".
- Event name: `file-changed`; payload `{path string}`.

## Step 7: Phase 4 — image serving via AssetServer.Handler + allow-list (DR-5)

Phase 4 made relative Markdown images (`![](images/x.png)`) render in the window, by answering the `/file/<abs>` URLs that `renderer.rewriteImagePaths` already emits, through a Wails `AssetServer.Handler` backed by an allow-list.

### Prompt Context

**User prompt (verbatim):** (same as Step 3 — work through the tasks, diary, commit.)

**Assistant interpretation:** Execute Phase 4: implement `App.ServeReferencedFile` + an `allowedDirs` allow-list mirroring `pkg/server.handleFileServing` (server.go:418), and wire it as the Wails `AssetServer.Handler`.

**Inferred user intent:** Images in rendered Markdown load from the open file's directory, with the same path-traversal protection as the daemon model.

### What I did

- Added `allowedDirs map[string]struct{}` to `App`; `openPath` calls `addAllowedDir(filepath.Dir(abs))` on open (mirrors server.go:286-294).
- Created `assets.go` with `addAllowedDir`, `isAllowed` (prefix check with `+filepath.Separator` to avoid `/tmp/foo` authorizing `/tmp/foobar`), and `ServeReferencedFile` (the `AssetServer.Handler`): serves `/file/<abs>` requests via `http.ServeContent` after the allow check; non-`/file/` → 404; disallowed → 403.
- Wired `Handler: http.HandlerFunc(app.ServeReferencedFile)` into `main.go`'s `AssetServer.Options`.

### Why

- `renderer.rewriteImagePaths` already rewrites relative `<img src>` to `/file/<abs-path>` (port-independent — confirmed in Phase 1). Reusing that scheme means zero renderer change; only the *responder* for `/file/...` moves from `pkg/server` to `App`.
- The allow-list is the security boundary: only directories of opened files (and their ancestors, in the daemon model) may be read. Kept the same check to preserve the trust model.

### What worked

- Verified end to end via `wails dev` browser mode + Playwright against `/tmp/md-img-test/README.md` (with `images/diagram.png`): the renderer rewrote `src` to `/file/tmp/md-img-test/images/diagram.png`; `fetch(src)` returned **200**; a traversal `fetch('/file/etc/passwd')` returned **403**. Both the happy path and the security guard pass.
- gofmt clean; `go build -tags webkit2_41 ./...` ok; `go test ./...` green.

### What didn't work

- **First wire attempt: `Handler: app`** — `*App` doesn't implement `http.Handler` (no `ServeHTTP`). Fix: `Handler: http.HandlerFunc(app.ServeReferencedFile)`. Cleaner than bolting `ServeHTTP` onto `App`.
- **Import scoping:** I initially added `net/http`/`os`/`strings` to `app.go`, but those are used in `assets.go` (where the handler lives), leaving them unused in `app.go`. Reverted `app.go`'s imports to its actual usage; `assets.go` owns its own imports. One package, multiple files — imports are per-file in Go, which I momentarily forgot.

### What I learned

- Wails' `assetserver.Options.Handler` is an `http.Handler` consulted only for requests the embedded `Assets` (`embed.FS`) can't satisfy. So `/file/...` reaches our handler, while `/app.js`, `/style.css`, etc. are served from `embed.FS` first — no collision risk with frontend asset names (as long as none is named under `/file/`).
- `http.ServeFile` redirects to "clean" the URL, breaking absolute paths with a leading slash; `http.ServeContent` does not. Same choice the daemon made (server.go comment).

### What was tricky to build

- **The `+filepath.Separator` detail in `isAllowed`** is the difference between "safe" and "directory-confusion vulnerable": `strings.HasPrefix("/tmp/foobad", "/tmp/foo")` is true without the separator. The daemon had this right; preserved exactly.
- Per-file imports in a multi-file package (documented above).

### What warrants a second pair of eyes

- The allow-list grows monotonically (no eviction). For a single-window viewer this is fine; if multi-file is added, evict on close.
- Symlinks: `isAllowed` checks the *requested* path, not `filepath.EvalSymlinks`. The daemon model had the same limitation. A symlink inside an allowed dir pointing outside would currently be followed. Documented as a follow-up; acceptable parity for now.

### What should be done in the future

- Harden with `filepath.EvalSymlinks` before the allow check if the threat model demands it.
- Phase 5: menus, drag-and-drop, recent files, window title.

### Code review instructions

- `assets.go` (new): `addAllowedDir`, `isAllowed`, `ServeReferencedFile`.
- `app.go`: `allowedDirs` field, `openPath` calls `addAllowedDir`.
- `main.go`: `Handler: http.HandlerFunc(app.ServeReferencedFile)`.
- Validate: `wails dev -tags webkit2_41`; open a file with a relative image; the image renders; `fetch('/file/etc/passwd')` → 403.

### Technical details

- Commit (this step): see `git log` for "feat(MD-WAILS): image serving via AssetServer.Handler (Phase 4)".
- Handler type: `http.Handler`; adapter: `http.HandlerFunc`.

## Step 8: Phase 5 — menus, drag-and-drop, recent files, window title

Phase 5 added the desktop app's interaction surface: a native menu bar (File→Open/Close, View→Toggle Theme) that emits events, drag-and-drop file opening, and persistent recent-files. The window title was already set in `openPath` (Phase 1).

### Prompt Context

**User prompt (verbatim):** (same as Step 3 — work through the tasks, diary, commit.)

**Assistant interpretation:** Execute Phase 5: build the menu (event-emitting), implement `OnFileDrop`, add recent-files JSON persistence, and confirm `WindowSetTitle`.

**Inferred user intent:** A first-class desktop UX (menus, shortcuts, drag-drop, remembered files) matching the demo and replacing the browser-tab experience.

### What I did

- `menu.go` — `buildMenu(app)`: File→Open… (Ctrl/Cmd-O) calls `app.OpenFile()` and emits `file-opened` (or `file-error`); File→Close (Ctrl/Cmd-W) emits `close-file`; View→Toggle Theme (T) calls `app.ToggleTheme()` and emits `theme-changed`. Follows the "menu callbacks emit events" golden rule (Go can't touch the DOM).
- `recent.go` — recent-files persistence ported from the demo: `configPath` (`os.UserConfigDir()/md-view`), `loadRecentFiles`/`saveRecentFiles` (JSON array), `pushRecent` (prepend, dedup, cap 10), `isMarkdownExt`.
- `app.go` — `Startup` calls `loadRecentFiles()`; `Shutdown` calls `saveRecentFiles()` then closes the watcher; `openPath` calls `pushRecent(abs)`; added `currentFileTitle()`; implemented `OnFileDrop(_, _, paths)` (picks first `.md`, `openPath`, emits `file-opened`).
- `main.go` — wired `Menu: buildMenu(app)` and `DragAndDrop: &options.DragAndDrop{EnableFileDrop: true}`.

### Why

- Menus and drag-drop are the two ways users open files without typing a path; both must route through the same `openPath` → `file-opened` event path as the toolbar Open button, so the frontend has one DOM-update path.
- Recent-files persistence replaces the daemon's PID/port/socket state files with one tiny JSON under the XDG config dir — the same pattern the demo proved.

### What worked

- Verified recent-files end-to-end via `wails dev` browser mode + Playwright: opened two files via `OpenFileAtPath`; `GetRecentFiles()` returned both, most-recent-first (`/tmp/md-view-phase2.md` then `/tmp/md-view-reload.md`). Stopping the dev server (Ctrl-C → `Shutdown`) wrote `~/.config/md-view/recent.json` = `["/tmp/md-view-phase2.md","/tmp/md-view-reload.md"]`. Persistence round-trips (load on next Startup).
- gofmt clean; `go vet` clean; build + tests green.
- The frontend's existing `loadRecentFiles()` (calls `GetRecentFiles()`, renders the sidebar, re-runs on every `showContent`) already consumes the Go list — no frontend change needed for the sidebar.

### What didn't work

- Nothing failed. One note: `document.title` stays "md-view" in the **browser** dev mode — `runtime.WindowSetTitle` sets the *native window* title, not the browser tab, so it's only visible in the real desktop window (or `wails build`), not Playwright-on-browser. Not a bug; just a verification limitation.

### What I learned

- The menu, the toolbar, drag-drop, and recent-files clicks all funnel through `openPath` + the `file-opened` event. Keeping that single chokepoint is what made Phase 5 small — adding new entry points just means calling `openPath` and emitting the event.
- `os.UserConfigDir()` is cross-platform (XDG on Linux, `~/Library/Application Support` on macOS, `%AppData%` on Windows), so recent-files land in the right place per OS without `#ifdef`.

### What was tricky to build

- **Menu callback → DOM gap:** the natural mistake is to call `app.OpenFile()` in the menu callback and expect the UI to update. It won't — the result lives in Go. The demo's pattern (emit `file-opened` with `{html,path,title}`) is mandatory; reproduced exactly.
- Remembering that `saveRecentFiles` only runs on `Shutdown` — so the JSON isn't on disk until the app quits cleanly. Verified by stopping the dev server and reading the file afterward.

### What warrants a second pair of eyes

- `OnFileDrop` silently returns on a non-`.md` drop (no error event). Acceptable, but consider emitting a `file-error` ("not a markdown file") if UX wants feedback.
- Recent files are saved only on shutdown; a crash loses the session. Consider saving after each `pushRecent` if durability matters (cheap write).

### What should be done in the future

- Phase 6: the single-instance CLI dispatch (`SingleInstanceLock` + Cobra `view`) — the core of drop-in compatibility.
- Persist the theme preference alongside recent files (open question OQ-2).
- Add a favicon handler to clear the 404.

### Code review instructions

- `menu.go` (new): `buildMenu`.
- `recent.go` (new): config path + load/save/push + `isMarkdownExt`.
- `app.go`: `Startup`/`Shutdown` recent lifecycle, `openPath` → `pushRecent`, `currentFileTitle`, `OnFileDrop`.
- `main.go`: `Menu` + `DragAndDrop` options.
- Validate: open two files; `~/.config/md-view/recent.json` appears on quit and reloads next start.

### Technical details

- Commit (this step): see `git log` for "feat(MD-WAILS): menus, drag-drop, recent files (Phase 5)".
- Config dir: `os.UserConfigDir()/md-view`; file: `recent.json`.
