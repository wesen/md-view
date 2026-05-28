---
Title: Renderer Enhancement Design
Ticket: MD-RENDER-ENHANCE
Status: active
Topics:
    - markdown
    - renderer
    - frontend
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-28T17:30:46.579790506-04:00
WhatFor: ""
WhenToUse: ""
---
Summary: Design for three renderer enhancements: image path resolution, copy-to-clipboard on code blocks, and mermaid diagram verification.
WhatFor: Reference for implementing the three features in pkg/renderer and pkg/server.

# Renderer Enhancement Design

## Problem

`md-view` renders markdown files via an HTTP endpoint at `/render?file=/abs/path/to/file.md`. Three rendering gaps need to be addressed:

1. **Images don't load** — relative `![](./img.png)` references resolve against the URL `/render`, not the file's directory, so the browser 404s.
2. **No copy button on code blocks** — `<pre><code>` blocks have no way to copy content to clipboard.
3. **Mermaid diagrams need verification** — mermaid.js is embedded but needs to be validated end-to-end with theme toggle and live reload.

---

## 1. Image Path Resolution

### Root Cause

The browser resolves relative image URLs against the page URL:
```
http://localhost:PORT/render?file=/home/user/project/README.md
```

So `![](./images/diagram.png)` becomes `http://localhost:PORT/images/diagram.png` — which 404s because the server has no route for it.

### Design Options

#### Option A: Serve the file's parent directory as static assets (RECOMMENDED)

Add a new HTTP handler that serves files from the markdown file's parent directory:

```go
// In server.go, add route:
mux.HandleFunc("/file/", s.handleFileServing)

func (s *Server) handleFileServing(w http.ResponseWriter, r *http.Request) {
    // Extract the absolute path after /file/
    // e.g. /file/home/user/project/images/diagram.png
    filePath := strings.TrimPrefix(r.URL.Path, "/file/")
    
    // Security: only serve files, only under allowed directories
    // (must be the same dir as a currently-watched/rendered md file)
    
    absPath, err := filepath.Abs(filePath)
    // ... validate, serve file
    http.ServeFile(w, r, absPath)
}
```

Then in `renderer.go`, post-process the rendered HTML to rewrite relative `src` attributes:

```go
// After goldmark conversion, rewrite <img src="...">
// For relative paths, prefix with /file/<parent-dir>/
func rewriteImagePaths(html string, mdFilePath string) string {
    dir := filepath.Dir(mdFilePath)
    // regex or HTML parser: find <img src="X"> where X is not absolute URL
    // prepend /file/ + dir + "/" to relative paths
}
```

**Pros**: Simple, works with any file type (images, PDFs, etc.), no base tag hacks.
**Cons**: Need path validation to avoid directory traversal.

#### Option B: HTML `<base>` tag

```html
<base href="/file/home/user/project/">
```

**Pros**: Zero rewriting, browser handles all relative resolution.
**Cons**: Breaks `/render?file=...` links, SSE endpoint URLs, and mermaid script URLs. Too invasive.

#### Option C: Goldmark image transformer

Use goldmark's `transformer.Transformer` interface to rewrite AST nodes before rendering.

**Pros**: Cleanest integration, handles edge cases (title attributes, etc.).
**Cons**: More code, deeper goldmark API knowledge needed.

### Decision

**Option A** with a security check: the `/file/` handler only serves files that share a directory prefix with recently-rendered markdown files (tracked in a simple allowlist). The renderer rewrites `<img>` `src` attributes in post-processing.

### Implementation Sketch

1. `server.go`: Add `allowedDirs map[string]bool` field to `Server`, populated when `/render` is called.
2. `server.go`: Add `/file/` handler that validates the requested path is under an allowed dir, then uses `http.ServeFile`.
3. `renderer.go`: Add `rewriteImagePaths()` function called after `md.Convert()`.
4. Use a simple regex `<img\s[^>]*src="([^"]+)"` to find and rewrite relative paths.

---

## 2. Copy-to-Clipboard Button on Code Blocks

### Design

Add a small copy icon button to every `<pre><code>` block that copies the code text to the clipboard.

### Approach

Pure CSS + JS, no dependencies:

1. **CSS**: Style the copy button as a small icon in the top-right corner of the `<pre>` block.
2. **JS**: On page load, find all `<pre><code>` blocks, wrap them in a container `<div>`, inject a `<button>` with a clipboard SVG icon.
3. **JS**: On click, use `navigator.clipboard.writeText()` to copy `code.textContent`. Show brief "Copied!" feedback.

### Implementation

Embed a new `static/copy-button.js` file:

```javascript
(function() {
    document.querySelectorAll('pre code').forEach(function(codeBlock) {
        var pre = codeBlock.parentElement;
        var container = document.createElement('div');
        container.className = 'md-view-code-block';
        
        var button = document.createElement('button');
        button.className = 'md-view-copy-btn';
        button.title = 'Copy to clipboard';
        button.innerHTML = '<svg>...</svg>'; // clipboard icon
        
        pre.parentNode.insertBefore(container, pre);
        container.appendChild(pre);
        container.appendChild(button);
        
        button.addEventListener('click', function() {
            navigator.clipboard.writeText(codeBlock.textContent).then(function() {
                button.innerHTML = '<svg>...</svg>'; // checkmark icon
                button.title = 'Copied!';
                setTimeout(function() {
                    button.innerHTML = '<svg>...</svg>'; // back to clipboard
                    button.title = 'Copy to clipboard';
                }, 2000);
            });
        });
    });
})();
```

Add CSS for the button (positioned top-right, subtle, with hover state) to both `base.css` and `dark.css` (or in the themeCSS function with dark overrides).

### Files to Change

- `pkg/renderer/static/copy-button.js` (new)
- `pkg/renderer/static/base.css` (add copy button styles)
- `pkg/renderer/static/dark.css` (add dark theme copy button styles)
- `pkg/renderer/renderer.go` (embed + inject the script)

---

## 3. Mermaid Diagram Verification

### Current State

Mermaid support is already partially implemented:
- `pkg/renderer/static/mermaid.min.js` — embedded library
- `pkg/renderer/static/mermaid-init.js` — detects `code.language-mermaid` blocks, wraps them, initializes mermaid
- `renderer.go` embeds both and injects them into the HTML
- `server.go` serves `mermaid.min.js` at `/static/mermaid.min.js`
- Theme toggle re-renders diagrams via MutationObserver

### What to Verify

1. **Basic rendering**: ` ```mermaid ` blocks render as SVG
2. **Theme toggle**: switching light/dark re-renders with correct mermaid theme
3. **Live reload**: after file change, page refreshes and mermaid re-initializes
4. **Error handling**: invalid mermaid syntax shows a readable error, not a blank page
5. **Script loading order**: mermaid.js loads before init script runs

### Potential Issues

- The init script loads mermaid.js from `http://localhost:PORT/static/mermaid.min.js` — hardcoded port, works but fragile if port changes.
- After live reload, the full page refreshes so mermaid re-initializes — this should work but needs testing.
- Error blocks: if mermaid can't parse, `mermaid.run()` may throw — need a try/catch (already present in theme toggle but not in initial render).

---

## Scope & Priority

| # | Feature | Priority | Complexity | Risk |
|---|---------|----------|------------|------|
| 1 | Image path resolution | **High** | Medium | Security (path traversal) |
| 2 | Copy-to-clipboard button | Medium | Low | None |
| 3 | Mermaid verification | Medium | Low | None |

Recommended implementation order: 1 → 2 → 3
