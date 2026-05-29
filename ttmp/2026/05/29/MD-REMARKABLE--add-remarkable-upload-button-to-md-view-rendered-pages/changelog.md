# Changelog

## 2026-05-29

- Initial workspace created


## 2026-05-29

Implemented reMarkable upload button (POST /upload-remarkable + remarkable-button.js), copy-path button, download-markdown button. All three appear in top bar with light/dark CSS.

### Related Files

- /home/manuel/code/wesen/2026-05-07--md-server/pkg/renderer/renderer.go — Embedded and injected both new scripts + CSS
- /home/manuel/code/wesen/2026-05-07--md-server/pkg/renderer/static/remarkable-button.js — New reMarkable upload button script
- /home/manuel/code/wesen/2026-05-07--md-server/pkg/renderer/static/toolbar-buttons.js — New copy-path and download buttons
- /home/manuel/code/wesen/2026-05-07--md-server/pkg/server/server.go — Added POST /upload-remarkable handler (commit 41b614b)

