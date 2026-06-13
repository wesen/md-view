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
