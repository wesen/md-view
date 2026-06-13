.PHONY: all build run test lint lintmax golangci-lint-install gosec govulncheck goreleaser clean install dev tag-major tag-minor tag-patch release gifs frontend-css wails-dev

# md-view is now a single Wails v2 desktop binary (MD-WAILS cutover).
# It is built with `wails build` (NOT plain `go build` — Wails injects build
# tags that a raw go build omits, causing a "will not build without the correct
# build tags" runtime error).
all: build

BINARY := md-view
VERSION := v0.2.0
WAILS_BUILD_TAGS ?= webkit2_41

GORELEASER_ARGS ?= --skip=sign --snapshot --clean
GORELEASER_TARGET ?= --single-target
GOLANGCI_LINT_VERSION ?= $(shell cat .golangci-lint-version)
GOLANGCI_LINT_BIN ?= $(CURDIR)/.bin/golangci-lint
GOLANGCI_LINT_ARGS ?= --timeout=5m . ./cmd/... ./pkg/...
LINT_DIRS := $(shell git ls-files '*.go' | grep -vE '(^|/)ttmp/|(^|/)testdata/' | xargs -r -n1 dirname | sed 's#^#./#' | sort -u)
GOSEC_EXCLUDE_DIRS := -exclude-dir=.history -exclude-dir=testdata -exclude-dir=ttmp

# Build the frontend CSS (chroma.css + ui.css) before building the app.
# wails build embeds frontend/dist, so the generated CSS must be present.
build: frontend-css
	wails build -tags $(WAILS_BUILD_TAGS) -s

# Frontend assets: regenerate the static CSS the Wails frontend links (MD-WAILS DR-4).
# Produces frontend/dist/chroma.css (dual-theme code highlighting) and ui.css
# (frontmatter + button chrome, both themes).
frontend-css:
	GOWORK=off go run -tags $(WAILS_BUILD_TAGS) ./cmd/gen-chroma-css

# Development: hot-reload dev server (frontend + Go changes reload live).
wails-dev:
	wails dev -tags $(WAILS_BUILD_TAGS)

run: build
	./build/bin/$(BINARY) view $(FILE)

test:
	GOWORK=off go test -tags $(WAILS_BUILD_TAGS) ./...

golangci-lint-install:
	mkdir -p $(dir $(GOLANGCI_LINT_BIN))
	GOBIN=$(dir $(GOLANGCI_LINT_BIN)) GOWORK=off go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint: golangci-lint-install
	GOWORK=off $(GOLANGCI_LINT_BIN) config verify
	GOWORK=off $(GOLANGCI_LINT_BIN) run -v $(GOLANGCI_LINT_ARGS)

lintmax: golangci-lint-install
	GOWORK=off $(GOLANGCI_LINT_BIN) config verify
	GOWORK=off $(GOLANGCI_LINT_BIN) run -v --max-same-issues=100 $(GOLANGCI_LINT_ARGS)

gosec:
	GOWORK=off go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -exclude-generated -exclude=G101,G304,G301,G306 $(GOSEC_EXCLUDE_DIRS) $(LINT_DIRS)

govulncheck:
	GOWORK=off go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

goreleaser:
	GOWORK=off goreleaser release $(GORELEASER_ARGS) $(GORELEASER_TARGET)

tag-major:
	git tag $(shell svu major)

tag-minor:
	git tag $(shell svu minor)

tag-patch:
	git tag $(shell svu patch)

release:
	git push origin --tags
	GOWORK=off GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/md-view@$(shell svu current)

install: build
	cp ./build/bin/$(BINARY) $(shell which md-view 2>/dev/null || echo /usr/local/bin/md-view)

clean:
	rm -f $(BINARY)
	rm -rf build/bin
