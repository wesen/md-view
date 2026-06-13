---
title: "Sources — captured Wails reference material"
ticket: MD-WAILS
doc-type: reference
status: active
intent: long-term
topics: [wails, go, architecture]
---

# Sources

Captured reference material for the MD-WAILS port. Each file is self-contained with a provenance header so it remains useful even if the upstream URLs change or go offline.

| # | File | What it proves | Upstream |
|---|------|----------------|----------|
| 01 | `01-wails-single-instance-lock-api.md` | Built-in `SingleInstanceLock`/`SecondInstanceData` API (Wails v2.7.0+) that replaces md-view's daemon+socket+PID subsystem for drop-in multi-invocation CLI | `pkg/options/options.go` (GitHub raw) + wails.io guide |
| 02 | `02-wails-cobra-integration-discussion-1271.md` | Cobra and `wails.Run` coexist in one `main.go`, with the double-click-vs-CLI gotcha | GitHub Discussion #1271 |
| 03 | `03-wails-cli-with-app-discussion-3098.md` | "One binary, CLI + GUI" is a documented Wails use case (Docker Desktop model) | GitHub Discussion #3098 |

## How these feed the design doc

- **01** is the foundation for the drop-in CLI behavior: a 2nd `md-view view b.md` forwards `os.Args` to instance #1 via `OnSecondInstanceLaunch` → `EventsEmit("open-from-cli", args)`. Retires `pkg/daemon` and `pkg/protocol`.
- **02** justifies keeping md-view's existing Cobra+Glazed CLI and routing the GUI through a command's `Run` → `wails.Run`.
- **03** is the use-case validation for the single-binary decision (vs. shipping two binaries).

## Retrieval notes

- GitHub raw source (`options.go`) and the GitHub Discussions REST API fetched cleanly with `curl`.
- The `wails.io` guide pages (`/docs/guides/single-instance-lock/`, `/docs/guides/file-association/`) are behind a Cloudflare "Just a moment..." challenge and could not be fetched headless. The authoritative type definitions were taken from the `options.go` source instead, which is more precise than the guide prose.
