---
Title: Add reMarkable upload button to md-view rendered pages
Ticket: MD-REMARKABLE
Status: active
Topics:
    - markdown
    - renderer
    - frontend
    - remarkable
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/renderer/renderer.go
      Note: Needs to embed and inject remarkable-button.js
    - Path: pkg/renderer/static/base.css
      Note: Needs button styles
    - Path: pkg/renderer/static/dark.css
      Note: Needs dark button styles
    - Path: pkg/server/server.go
      Note: Needs POST /upload-remarkable handler
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-29T08:19:12.152625618-04:00
WhatFor: ""
WhenToUse: ""
---


# Add reMarkable upload button to md-view rendered pages

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- markdown
- renderer
- frontend
- remarkable

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
