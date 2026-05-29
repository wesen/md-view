# md-view

<p align="center">
  <img src="docs/images/hero.jpg" alt="md-view — a markdown viewer that just works" width="400">
</p>

> A markdown viewer that just works. One command, beautiful rendering, browser opens automatically.

`md-view` is a lightweight daemon that renders Markdown files as GitHub-flavored HTML and opens them in your browser. It runs in the background, auto-starts when you need it, and watches your files for live reload.

## Install

```bash
go install github.com/go-go-golems/md-view/cmd/md-view@latest
```

## 30-Second Quick Start

```bash
md-view view ./README.md
```

That's it. Your browser opens on a rendered page. Edit the file, the page refreshes.

## Commands

| Command | What it does |
|---------|-------------|
| `md-view view <FILE>` | View a markdown file in your browser |
| `md-view serve` | Start the server in foreground (debugging) |
| `md-view status` | Show daemon PID, port, uptime |
| `md-view stop` | Stop the daemon |

## Key Features

- **One command** — `md-view view file.md` does everything
- **Background daemon** — auto-starts, stays running, no setup
- **GitHub-flavored rendering** — tables, task lists, fenced code blocks, strikethrough
- **Syntax highlighting** — 200+ languages, server-side via Chroma (no JS required)
- **Mermaid diagrams** — ` ```mermaid ` blocks rendered as SVG, works offline (mermaid.js embedded)
- **Dark theme** — toggle button, `--dark` flag, or `?theme=dark` URL param; code highlighting switches too
- **Live reload** — page refreshes when the file changes
- **Frontmatter support** — YAML frontmatter parsed and displayed as a collapsible key-value table; the `Title` field becomes the browser tab title
- **i3 / Sway ready** — windows titled `md-view: <filename>`, just add a floating rule to your config
- **Zero config** — random port, XDG state dir, defaults to Firefox `--new-window`

## Architecture

```
┌──────────────┐   Unix Socket    ┌──────────────────┐    HTTP     ┌─────────┐
│  md-view CLI │ ─── JSON cmd ──► │  md-view server  │ ─────────► │ Browser │
│  (ephemeral) │                  │  (daemon)        │            │         │
└──────────────┘                  │                  │  ◄── SSE ── │         │
                                  │  - Renders .md   │   reload   │         │
                                  │  - Serves HTML   │            └─────────┘
                                  │  - Watches files │
                                  └──────────────────┘
```

## Documentation

- **[Getting Started](docs/getting-started.md)** — install, first view, common workflows
- **[User Guide](docs/user-guide.md)** — all commands, flags, i3 integration, troubleshooting

## Build from Source

```bash
git clone https://github.com/go-go-golems/md-view.git
cd md-view
make build
```

## License

MIT
