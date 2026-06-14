# Changelog

## 2026-06-13

- Initial workspace created
- Added vocabulary topics: `wails`, `desktop`, `architecture`
- Created design/implementation guide `design-impl-guide/01-...md` — intern-grade analysis of md-view (current) + Wails (target), gap analysis, proposed architecture, 6 decision records (DR-1..DR-6), pseudocode/flows, 8-phase file-level plan, test strategy, risks/open questions
- Created investigation diary `reference/01-investigation-diary.md`
- Related 8 key files (md-view renderer/server/watcher/daemon/protocol + wails-demo main/app + Wails article) to the design doc
- Key decisions recorded: DR-1 retire daemon/protocol/server, keep renderer; DR-2 coexistence (new `wailsapp/` entry, CLI untouched); DR-3 split `Render` into `RenderBody`/page; DR-4 pre-generate Chroma CSS; DR-5 image serving via `AssetServer.Handler`; DR-6 vanilla-JS frontend

## 2026-06-13

Uploaded bundle (design guide + diary + index) to reMarkable /ai/2026/06/13/MD-WAILS

### Related Files

- /home/manuel/code/wesen/2026-05-07--md-server/ttmp/2026/06/13/MD-WAILS--port-md-view-to-a-wails-v2-desktop-application/design-impl-guide/01-wails-port-analysis-design-and-implementation-guide.md — Primary deliverable uploaded to reMarkable


## 2026-06-13

Added sources/ folder: captured Wails SingleInstanceLock API (options.go source), Cobra+Wails discussion #1271, CLI-with-app discussion #3098 — all self-contained with provenance

### Related Files

- /home/manuel/code/wesen/2026-05-07--md-server/ttmp/2026/06/13/MD-WAILS--port-md-view-to-a-wails-v2-desktop-application/sources/README.md — Index of captured reference materials


## 2026-06-13

SCOPE REVISION (v2): changed from coexistence (two binaries) to DROP-IN REPLACEMENT — single md-view binary, CLI compatible (view + --dark preserved; serve/stop/status removed). SingleInstanceLock replaces daemon/socket/PID. Added DR-7 (SingleInstanceLock) and DR-8 (flag trimming). Restructured phases (Phase 6 single-instance dispatch, Phase 7 cutover/deletion).

### Related Files

- /home/manuel/code/wesen/2026-05-07--md-server/ttmp/2026/06/13/MD-WAILS--port-md-view-to-a-wails-v2-desktop-application/design-impl-guide/01-wails-port-analysis-design-and-implementation-guide.md — Rewritten to replacement scope


## 2026-06-13

Added sources/01..03 (SingleInstanceLock API, Cobra discussion #1271, CLI-with-app #3098) + 00-sources-index; related to design doc

### Related Files

- /home/manuel/code/wesen/2026-05-07--md-server/ttmp/2026/06/13/MD-WAILS--port-md-view-to-a-wails-v2-desktop-application/sources/01-wails-single-instance-lock-api.md — Core drop-in mechanism


## 2026-06-13

Added implementation review / lessons-learned document for a new maintainer or intern; captures subsystem map, review scorecard, strengths, weaknesses, stale-docs finding, build/release lessons, and recommended next learning/follow-up work.

### Related Files

- /home/manuel/code/wesen/2026-05-07--md-server/ttmp/2026/06/13/MD-WAILS--port-md-view-to-a-wails-v2-desktop-application/design-doc/01-implementation-review-and-lessons-learned.md — Post-implementation technical review deliverable


## 2026-06-14

Phase 9 (documentation cutover): rewrote README.md, docs/getting-started.md, docs/user-guide.md to match the single-binary Wails model. Removed docs for deleted subsystems (HTTP API, Unix Socket Protocol, Daemon Management, browser selection, serve/status/stop, --browser/--no-reload/--port flags). Install now leads with make build (NOT go install .../cmd/md-view). Verified AGENT.md is clean (historical refs only). grep validation (task 9.7) passes: zero live operational references to the old model. Also marked Phase 8 done in tasks.md (was implemented but unchecked).

### Related Files

- /home/manuel/code/wesen/2026-05-07--md-server/AGENT.md — Verified clean — only historical daemon/socket refs explaining the cutover (no rewrite needed)
- /home/manuel/code/wesen/2026-05-07--md-server/README.md — Rewritten for the Wails single-binary model (install via make build
- /home/manuel/code/wesen/2026-05-07--md-server/docs/getting-started.md — Rewritten — native window first view
- /home/manuel/code/wesen/2026-05-07--md-server/docs/user-guide.md — Major surgery — removed HTTP API/Unix Socket/Daemon Management/browser-selection/serve-status-stop; rewrote view/Dark Theme/Live Reload/Security/Troubleshooting for Wails

