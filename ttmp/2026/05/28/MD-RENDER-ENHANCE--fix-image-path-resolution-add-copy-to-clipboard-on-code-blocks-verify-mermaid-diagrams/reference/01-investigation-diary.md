---
Title: Investigation Diary
Ticket: MD-RENDER-ENHANCE
Status: active
Topics:
    - markdown
    - renderer
    - frontend
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Chronological diary of investigation and implementation for the three renderer enhancements.
LastUpdated: 2026-05-28T17:30:00.000000000-04:00
WhatFor: ""
WhenToUse: ""
---

# Investigation Diary

## 2026-05-28 — Ticket Created

### What was done
- Created ticket MD-RENDER-ENHANCE for three renderer enhancements.
- Analyzed current codebase: `pkg/renderer/renderer.go`, `pkg/server/server.go`, static assets.
- Wrote design doc with three options for image path resolution (chose Option A: `/file/` handler + src rewriting).
- Sketched implementation for copy-to-clipboard (pure CSS + JS, no dependencies).
- Noted mermaid support is already partially implemented; needs end-to-end verification.

### Key findings
- **Image paths**: The page URL is `/render?file=/abs/path.md`, so relative images resolve against `/render` → 404. Need `/file/` handler + HTML rewriting.
- **Copy button**: Straightforward — embed `copy-button.js`, inject after rendering, add CSS to both themes.
- **Mermaid**: Already embedded (`mermaid.min.js` + `mermaid-init.js`). Init script wraps `code.language-mermaid` → `div.mermaid`. Theme toggle MutationObserver exists. Needs testing for: initial render, error handling, live reload re-init.

### Next steps
- [x] Implement `/file/` handler in `server.go` with directory allowlist
- [x] Implement `rewriteImagePaths()` in `renderer.go`
- [x] Create `copy-button.js` and inject into rendered HTML
- [x] Add copy button CSS to both themes
- [x] Test mermaid end-to-end

## 2026-05-28 — Implementation Complete

### What was done
- Implemented all three features across 3 commits.

### Feature 1: Image path resolution
- Added `/file/` handler in `server.go` that serves files from allowed directories.
- Directories are registered when `/render` is called (the md file's parent dir).
- Added `rewriteImagePaths()` in `renderer.go` that rewrites relative `<img src>` to `/file/<abs-path-no-leading-slash>`.
- **Bug found & fixed**: Go's `http.ServeMux` redirects `//` in URLs. Fixed by stripping the leading `/` from absolute paths in URLs and re-adding it in the handler.
- **Bug found & fixed**: `http.ServeFile` also redirects. Switched to `http.ServeContent` with manual file open.

### Feature 2: Copy-to-clipboard button
- Created `static/copy-button.js` with SVG clipboard/check icons.
- Wraps each `<pre><code>` in a container with a hover-reveal copy button.
- Uses `navigator.clipboard.writeText()` with fallback to text selection.
- CSS added to both `base.css` and `dark.css`.
- Script injected inline in rendered HTML.

### Feature 3: Mermaid verification
- Mermaid was already correctly implemented.
- Verified: init script wraps `code.language-mermaid` → `div.mermaid`.
- Verified: theme toggle MutationObserver re-renders diagrams.
- Verified: script order is correct (mermaid first, copy button after — so mermaid blocks get replaced before copy button runs).
- Added tests for mermaid presence and copy button injection.

### Test results
- All existing tests pass.
- New tests: `TestRewriteImagePaths`, `TestRenderWithImages`, `TestRenderWithCodeBlockHasCopyButton`, `TestRenderWithMermaidBlock`.
- Manual smoke test: all 3 features work against live server.

### Bugs found during real-world testing
- **Double-slash 307**: Go's `http.ServeMux` redirects `//` in URLs → fixed by stripping leading `/`.
- **ServeFile redirect**: `http.ServeFile` also redirects → switched to `http.ServeContent`.
- **../ paths 403**: `../artifacts/` images resolve outside the markdown file's parent dir → fixed by registering all ancestor directories as allowed, not just the immediate parent.
- **Stale binary**: The installed binary at `~/.local/bin/md-view` was not being updated on rebuild. Always `go build -o ~/.local/bin/md-view` after changes.
