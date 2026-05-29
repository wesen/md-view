# Changelog

## 2026-05-28

- Initial workspace created


## 2026-05-28

Ticket created. Analyzed codebase, wrote design doc for three features: image path resolution (Option A: /file/ handler + src rewrite), copy-to-clipboard button (pure CSS+JS), mermaid verification.

### Related Files

- /home/manuel/code/wesen/2026-05-07--md-server/pkg/renderer/renderer.go — Analyzed current rendering pipeline
- /home/manuel/code/wesen/2026-05-07--md-server/pkg/server/server.go — Analyzed HTTP handlers


## 2026-05-28

All three features implemented and tested. Image paths fixed (with double-slash ServeMux bugfix). Copy-to-clipboard button added. Mermaid verified working.


## 2026-05-28

Fixed ../ image path resolution: register all ancestor directories as allowed for /file/ handler, not just the markdown file's immediate parent.

