# Changelog

## 2026-05-29

- Initial workspace created


## 2026-05-29

Bumped glazed v1.3.6 (logcopter). Replaced stdlib log with logcopter package loggers. Added glazed-lint Makefile targets, lefthook, CI. Fixed errcheck + staticcheck.

### Related Files

- /home/manuel/code/wesen/2026-05-07--md-server/.github/workflows/lint.yml — Glazed-lint in CI
- /home/manuel/code/wesen/2026-05-07--md-server/Makefile — Glazed-lint targets
- /home/manuel/code/wesen/2026-05-07--md-server/lefthook.yml — Glazed-lint in pre-commit
- /home/manuel/code/wesen/2026-05-07--md-server/pkg/commands/run.go — Logcopter logging
- /home/manuel/code/wesen/2026-05-07--md-server/pkg/server/server.go — Logcopter logging + errcheck fixes

