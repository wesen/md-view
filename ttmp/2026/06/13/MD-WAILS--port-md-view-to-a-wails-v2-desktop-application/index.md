---
Title: Port md-view to a Wails v2 desktop application
Ticket: MD-WAILS
Status: active
Topics:
    - markdown
    - go
    - architecture
    - wails
    - desktop
    - web
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "REPLACE md-view with a single Wails v2 binary (still named md-view) that is drop-in CLI compatible: `md-view view <file> [--dark]` opens a native window; a second invocation reuses it via SingleInstanceLock. Daemon/socket/PID/HTTP packages are deleted. Implementation (Phases 0-8) and user-facing docs cutover (Phase 9) are complete."
LastUpdated: 2026-06-14T00:00:00-04:00
WhatFor: "Intern-grade analysis, design, implementation guide, and post-implementation review for REPLACING md-view with a drop-in Wails v2 desktop binary. All phases (0-9) complete: code shipped, user docs rewritten to match."
WhenToUse: "Read before starting any Wails port work; the phased plan in the design doc is the execution roadmap."
---

# Port md-view to a Wails v2 desktop application

## Overview

**REPLACE** `md-view` (currently a CLI → Unix-socket daemon → HTTP server → browser toolchain) with a **single Wails v2 binary** (still named `md-view`) that embeds a platform-native WebView and runs all business logic in Go, with a thin vanilla-JS frontend. The replacement is **drop-in at the CLI**: `md-view view <file> [--dark]` still works (now opening a native window), and a second `md-view view` reuses the running app via Wails' built-in `SingleInstanceLock`. The existing Markdown renderer (`pkg/renderer`, Goldmark + Chroma + Mermaid + frontmatter) is reused; the daemon/socket/HTTP/PID packages are **deleted** and replaced by Wails bound methods, events, and `SingleInstanceLock`.

A working proof-of-concept exists at `/home/manuel/code/wesen/2026-06-13--wails-demo` and the Wails internals are documented in the vault article "Wails v2 Desktop Applications - Technical Deep Dive".

**Start here:** [design-impl-guide/01-wails-port-analysis-design-and-implementation-guide.md](./design-impl-guide/01-wails-port-analysis-design-and-implementation-guide.md)

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active — implementation (Phases 0–8) and user-facing documentation cutover (Phase 9) are complete.** The code shipped a single Wails binary and the README/getting-started/user-guide now match it. Known limitation: the Linux `SingleInstanceLock` multi-window behavior (accepted). Open follow-ups: theme persistence (OQ-2), a real GoReleaser release dry run, and focused unit tests for `assets.go`/`recent.go`.

## Topics

- markdown
- go
- architecture
- wails
- desktop
- web

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
