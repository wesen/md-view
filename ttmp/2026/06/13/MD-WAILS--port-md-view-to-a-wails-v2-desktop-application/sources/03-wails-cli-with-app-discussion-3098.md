---
title: "Wails discussion #3098: Include a CLI with app"
source: "https://github.com/wailsapp/wails/discussions/3098"
retrieved: 2026-06-13
ticket: MD-WAILS
why: "Confirms the 'one binary, both CLI and GUI' goal is a documented Wails use case (the exact requirement for a drop-in md-view replacement). Cites Docker Desktop as the model. Use this to justify DR: single-binary CLI+GUI coexistence."
---

# Wails discussion #3098: Include a CLI with app

> Verbatim capture of the opening post from the GitHub Discussions API. This is the use-case statement for "one binary that is both a CLI and a desktop app" — exactly the MD-WAILS requirement.

## Metadata

- **URL:** https://github.com/wailsapp/wails/discussions/3098
- **Title:** Include a CLI with app
- **Author:** manterfield
- **Created:** 2023-12-05
- **Retrieved:** 2026-06-13 (via `api.github.com/repos/wailsapp/wails/discussions/3098`)

## Body (verbatim)

> This might be a dumb question, so apologies if so. I couldn't find anything in docs and I'm not yet 100% comfortable with how Wails builds/deploys to figure it out from scratch.
>
> We're looking at moving our browser/CLI based app to Wails and distributing as a desktop app, but we'd still want a CLI interface too. Ideally users wouldn't have to download two separate things.
>
> Some apps I have installed on Mac also have a CLI. Docker desktop is an example.
>
> Is this something we can do today with Wails? If not, I would be grateful for any hacky workarounds people have used to achieve the same result.

## Relevance to MD-WAILS

- This is precisely the MD-WAILS requirement: replace md-view with a **single binary** that is both the CLI (`md-view view README.md`) and the desktop app, "ideally users wouldn't have to download two separate things."
- The Docker Desktop analogy is the right mental model: one installed application that also exposes a terminal command.
- Combined with the Cobra integration pattern (`sources/02-...`) and `SingleInstanceLock` (`sources/01-...`), this confirms the architecture is achievable with stock Wails v2 — no workarounds needed as of v2.7.0+.

## Provenance

- Captured via GitHub Discussions REST API: `curl https://api.github.com/repos/wailsapp/wails/discussions/3098`
- Reply comments (with concrete implementation suggestions) were not fetched. View the page directly for the community's recommended approaches.
