---
Title: Fix image path resolution, add copy-to-clipboard on code blocks, verify mermaid diagrams
Ticket: MD-RENDER-ENHANCE
Status: active
Topics:
    - markdown
    - renderer
    - frontend
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/renderer/renderer.go
      Note: Core renderer — needs image path rewriting
    - Path: pkg/renderer/static/base.css
      Note: Light theme CSS — needs copy button styles
    - Path: pkg/renderer/static/dark.css
      Note: Dark theme CSS — needs dark copy button styles
    - Path: pkg/renderer/static/mermaid-init.js
      Note: Mermaid init — verify correct initialization
    - Path: pkg/server/server.go
      Note: HTTP server — needs /file/ handler for serving images from markdown file directories
ExternalSources: []
Summary: Three renderer enhancements for md-view: (1) fix image path resolution by serving the markdown file's parent directory and rewriting <img src>, (2) add copy-to-clipboard button on code blocks, (3) verify mermaid diagram rendering works end-to-end with theme toggle and live reload.
LastUpdated: 2026-05-28T17:29:23.984909704-04:00
WhatFor: ""
WhenToUse: ""
---


# Fix image path resolution, add copy-to-clipboard on code blocks, verify mermaid diagrams

## Overview

md-view renders markdown files at `/render?file=/abs/path/file.md`. Three rendering gaps:

1. **Images don't load** — relative `![](./img.png)` references resolve against `/render` instead of the file's directory. Fix: add a `/file/` static handler + post-process `<img>` src attributes.
2. **No copy button on code blocks** — add a clipboard icon to every `<pre><code>` block with `navigator.clipboard.writeText()`.
3. **Mermaid verification** — mermaid.js is already embedded; verify it works with theme toggle, live reload, and error handling.

**Design doc**: [design/01-renderer-enhancement-design.md](design/01-renderer-enhancement-design.md)

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- markdown
- renderer
- frontend

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
