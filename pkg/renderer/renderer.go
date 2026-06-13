// Package renderer converts Markdown files to styled HTML pages.
package renderer

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	chroma_html "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed static/base.css
var defaultCSS []byte

//go:embed static/dark.css
var darkCSS []byte

//go:embed static/reload.js
var reloadJS []byte

//go:embed static/mermaid-init.js
var mermaidInitJS []byte

//go:embed static/mermaid.min.js
var mermaidJS []byte

//go:embed static/copy-button.js
var copyButtonJS []byte

//go:embed static/remarkable-button.js
var remarkableButtonJS []byte

//go:embed static/toolbar-buttons.js
var toolbarButtonsJS []byte

// CSS returns the embedded GitHub-flavored CSS (light theme).
func CSS() []byte {
	return defaultCSS
}

// DarkCSS returns the embedded dark theme CSS.
func DarkCSS() []byte {
	return darkCSS
}

// ReloadJS returns the embedded live-reload script.
func ReloadJS() []byte {
	return reloadJS
}

// MermaidJS returns the embedded mermaid.js library.
func MermaidJS() []byte {
	return mermaidJS
}

// CopyButtonJS returns the embedded copy-to-clipboard script.
func CopyButtonJS() []byte {
	return copyButtonJS
}

// RemarkableButtonJS returns the embedded reMarkable upload button script.
func RemarkableButtonJS() []byte {
	return remarkableButtonJS
}

// ToolbarButtonsJS returns the embedded toolbar buttons script.
func ToolbarButtonsJS() []byte {
	return toolbarButtonsJS
}

// ChromaCSS returns the CSS for syntax highlighting.
// chromaStyle is the Chroma style name (e.g. "github" for light, "dracula" for dark).
func ChromaCSS(chromaStyle string) (string, error) {
	formatter := chroma_html.New(chroma_html.WithClasses(true))
	style := styles.Get(chromaStyle)
	if style == nil {
		style = styles.Fallback
	}
	buf := &bytes.Buffer{}
	if err := formatter.WriteCSS(buf, style); err != nil {
		return "", fmt.Errorf("cannot generate chroma CSS: %w", err)
	}
	return buf.String(), nil
}

// ChromaCSSBoth returns CSS for both light and dark themes, wrapped
// in [data-theme] selectors so the toggle works for code highlighting.
func ChromaCSSBoth() (string, error) {
	lightCSS, err := ChromaCSS("github")
	if err != nil {
		return "", err
	}
	darkCSS, err := ChromaCSS("dracula")
	if err != nil {
		return "", err
	}

	var buf strings.Builder

	// Light theme is the default
	buf.WriteString(lightCSS)

	// Dark theme: prefix each .chroma rule with [data-theme="dark"]
	buf.WriteString("\n")
	darkLines := strings.Split(darkCSS, "\n")
	for _, line := range darkLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Lines that contain .chroma selectors need the dark prefix
		// Chroma outputs lines like: /* Comment */ .chroma .k { color: #ff79c6 }
		if strings.Contains(trimmed, ".chroma") || strings.Contains(trimmed, ".bg") {
			buf.WriteString("[data-theme=\"dark\"] ")
		}
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// UICSS returns the CSS for md-view's in-page UI chrome: the frontmatter
// <details> block, the fixed-position buttons (copy/reMarkable/toolbar),
// the copy-to-clipboard code-block button, and their dark-theme overrides.
// The desktop frontend links this as a static asset (ui.css) instead of
// inlining it per render. It includes both light and [data-theme="dark"] rules.
func UICSS() string {
	return themeCSS(false)
}

// Options for rendering.
type Options struct {
	// NoReload disables SSE live reload injection.
	NoReload bool
	// File is the absolute path of the markdown file (used for SSE endpoint).
	File string
	// Title is the page title (defaults to filename if empty).
	Title string
	// Port is the HTTP port (used for SSE endpoint URL).
	Port int
	// Dark enables the dark theme.
	Dark bool
}

// frontmatterData holds parsed YAML frontmatter key-value pairs.
type frontmatterData struct {
	Title   string
	Entries []fmEntry
}

type fmEntry struct {
	Key   string
	Value string
}

// extractFrontmatter splits input into YAML frontmatter and body.
func extractFrontmatter(data []byte) (*frontmatterData, []byte, bool) {
	content := string(data)
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return nil, data, false
	}

	end := strings.Index(content[3:], "\n---")
	if end == -1 {
		return nil, data, false
	}

	rawFM := content[3 : 3+end]
	body := content[3+end+4:]
	body = strings.TrimPrefix(body, "\n")
	body = strings.TrimPrefix(body, "\r\n")

	fm := parseFrontmatter(rawFM)
	return fm, []byte(body), true
}

// parseFrontmatter parses simple YAML key: value pairs from frontmatter.
func parseFrontmatter(raw string) *frontmatterData {
	data := &frontmatterData{}
	lines := strings.Split(raw, "\n")

	i := 0
	for i < len(lines) {
		line := lines[i]

		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			colonIdx := strings.Index(line, ":")
			if colonIdx == -1 {
				i++
				continue
			}

			key := strings.TrimSpace(line[:colonIdx])
			value := strings.TrimSpace(line[colonIdx+1:])

			if value == "" {
				var nested []string
				i++
				for i < len(lines) && (strings.HasPrefix(lines[i], "  ") || strings.HasPrefix(lines[i], "\t") || strings.TrimSpace(lines[i]) == "") {
					nested = append(nested, lines[i])
					i++
				}
				value = strings.Join(nested, "\n")
			} else {
				value = stripQuotes(value)
				i++
			}

			if strings.EqualFold(key, "Title") && data.Title == "" {
				data.Title = stripQuotes(value)
			}

			data.Entries = append(data.Entries, fmEntry{Key: key, Value: value})
		} else {
			i++
		}
	}

	return data
}

var reQuoted = regexp.MustCompile(`^"(.*)"$`)

func stripQuotes(s string) string {
	if m := reQuoted.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

// formatFrontmatterHTML renders YAML frontmatter as a collapsible <details> block.
func formatFrontmatterHTML(fm *frontmatterData) string {
	var buf strings.Builder

	buf.WriteString(`<details class="md-view-frontmatter">
<summary>Frontmatter</summary>
<div class="md-view-fm-table">
`)

	for _, entry := range fm.Entries {
		key := htmlEscape(entry.Key)
		value := htmlEscape(strings.TrimSpace(entry.Value))

		if strings.Contains(value, "\n") {
			fmt.Fprintf(&buf, `<div class="md-view-fm-row">
<span class="md-view-fm-key">%s</span>
<pre class="md-view-fm-value">%s</pre>
</div>
`, key, value)
		} else {
			fmt.Fprintf(&buf, `<div class="md-view-fm-row">
<span class="md-view-fm-key">%s</span>
<span class="md-view-fm-value">%s</span>
</div>
`, key, value)
		}
	}

	buf.WriteString(`</div>
</details>`)
	return buf.String()
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// themeCSS returns the CSS for the frontmatter section, adjusted for theme.
func themeCSS(dark bool) string {
	light := `
.md-view-frontmatter {
    margin-bottom: 24px;
    border: 1px solid #d0d7de;
    border-radius: 6px;
    background: #f6f8fa;
    padding: 0;
}
.md-view-frontmatter > summary {
    padding: 8px 12px;
    cursor: pointer;
    font-size: 13px;
    color: #656d76;
    user-select: none;
    list-style: none;
    display: flex;
    align-items: center;
    gap: 6px;
}
.md-view-frontmatter > summary::before {
    content: "▶";
    font-size: 10px;
    transition: transform 0.15s;
}
.md-view-frontmatter[open] > summary::before {
    transform: rotate(90deg);
}
.md-view-frontmatter > summary:hover {
    background: #eaeef2;
}
.md-view-fm-table {
    border-top: 1px solid #d0d7de;
    font-size: 13px;
    line-height: 1.5;
}
.md-view-fm-row {
    display: flex;
    border-bottom: 1px solid #eaeef2;
    align-items: baseline;
}
.md-view-fm-row:last-child {
    border-bottom: none;
}
.md-view-fm-key {
    min-width: 120px;
    padding: 6px 12px;
    font-weight: 600;
    color: #24292e;
    background: #f0f2f4;
    flex-shrink: 0;
}
.md-view-fm-value {
    padding: 6px 12px;
    color: #656d76;
    word-break: break-word;
}
.md-view-fm-value pre {
    margin: 0;
    padding: 0;
    font-size: 12px;
    line-height: 1.4;
    background: transparent;
    white-space: pre-wrap;
}
.md-view-theme-toggle {
    position: fixed;
    top: 12px;
    right: 12px;
    z-index: 100;
    background: #f6f8fa;
    border: 1px solid #d0d7de;
    border-radius: 6px;
    padding: 4px 10px;
    cursor: pointer;
    font-size: 13px;
    color: #24292e;
    opacity: 0.7;
    transition: opacity 0.15s;
}
.md-view-theme-toggle:hover {
    opacity: 1;
}
/* reMarkable upload button */
.md-view-remarkable-btn {
    position: fixed;
    top: 12px;
    right: 80px;
    z-index: 100;
    background: #f6f8fa;
    border: 1px solid #d0d7de;
    border-radius: 6px;
    padding: 4px 8px;
    cursor: pointer;
    color: #656d76;
    opacity: 0.7;
    transition: opacity 0.15s, color 0.15s;
    display: flex;
    align-items: center;
    justify-content: center;
    line-height: 1;
}
.md-view-remarkable-btn:hover {
    opacity: 1;
    color: #24292e;
}
/* Toolbar buttons (copy path, download) */
.md-view-toolbar-btn {
    position: fixed;
    top: 12px;
    z-index: 100;
    background: #f6f8fa;
    border: 1px solid #d0d7de;
    border-radius: 6px;
    padding: 4px 8px;
    cursor: pointer;
    color: #656d76;
    opacity: 0.7;
    transition: opacity 0.15s, color 0.15s;
    display: flex;
    align-items: center;
    justify-content: center;
    line-height: 1;
}
.md-view-toolbar-btn:hover {
    opacity: 1;
    color: #24292e;
}
.md-view-toolbar-btn-success {
    color: #1a7f37 !important;
    opacity: 1 !important;
}
.md-view-copy-path-btn { right: 160px; }
.md-view-download-btn { right: 120px; }
.md-view-remarkable-btn:disabled {
    cursor: wait;
    opacity: 0.9 !important;
}
.md-view-remarkable-btn-loading svg {
    animation: md-view-spin 0.8s linear infinite;
}
.md-view-remarkable-btn-success {
    color: #1a7f37 !important;
    opacity: 1 !important;
}
.md-view-remarkable-btn-error {
    color: #cf222e !important;
    opacity: 1 !important;
}
@keyframes md-view-spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}
/* reMarkable upload status toast */
.md-view-remarkable-toast {
    position: fixed;
    top: 50px;
    right: 12px;
    z-index: 101;
    background: #f6f8fa;
    border: 1px solid #d0d7de;
    border-radius: 6px;
    padding: 8px 14px;
    font-size: 13px;
    color: #24292e;
    max-width: 400px;
    word-break: break-word;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}
.md-view-remarkable-toast-success {
    border-color: #1a7f37;
    color: #1a7f37;
}
.md-view-remarkable-toast-error {
    border-color: #cf222e;
    color: #cf222e;
}
.md-view-remarkable-toast-loading {
    color: #656d76;
}
.md-view-remarkable-toast-loading svg {
    animation: md-view-spin 0.8s linear infinite;
}
`

	darkOverrides := `
[data-theme="dark"] .md-view-frontmatter {
    border-color: #30363d;
    background: #161b22;
}
[data-theme="dark"] .md-view-frontmatter > summary {
    color: #8b949e;
}
[data-theme="dark"] .md-view-frontmatter > summary:hover {
    background: #21262d;
}
[data-theme="dark"] .md-view-fm-table {
    border-top-color: #30363d;
}
[data-theme="dark"] .md-view-fm-row {
    border-bottom-color: #21262d;
}
[data-theme="dark"] .md-view-fm-key {
    color: #c9d1d9;
    background: #161b22;
}
[data-theme="dark"] .md-view-fm-value {
    color: #8b949e;
}
[data-theme="dark"] .md-view-theme-toggle {
    background: #21262d;
    border-color: #30363d;
    color: #c9d1d9;
}
[data-theme="dark"] .md-view-remarkable-btn {
    background: #21262d;
    border-color: #30363d;
    color: #8b949e;
}
[data-theme="dark"] .md-view-remarkable-btn:hover {
    color: #c9d1d9;
}
[data-theme="dark"] .md-view-remarkable-btn-success {
    color: #3fb950 !important;
}
[data-theme="dark"] .md-view-remarkable-btn-error {
    color: #f85149 !important;
}
[data-theme="dark"] .md-view-toolbar-btn {
    background: #21262d;
    border-color: #30363d;
    color: #8b949e;
}
[data-theme="dark"] .md-view-toolbar-btn:hover {
    color: #c9d1d9;
}
[data-theme="dark"] .md-view-toolbar-btn-success {
    color: #3fb950 !important;
}
[data-theme="dark"] .md-view-remarkable-toast {
    background: #21262d;
    border-color: #30363d;
    color: #c9d1d9;
}
[data-theme="dark"] .md-view-remarkable-toast-success {
    border-color: #3fb950;
    color: #3fb950;
}
[data-theme="dark"] .md-view-remarkable-toast-error {
    border-color: #f85149;
    color: #f85149;
}
`

	return light + darkOverrides
}

// rewriteImagePaths rewrites relative <img src="..."> paths to absolute /file/ URLs
// so the browser can fetch them from the server's /file/ handler.
// mdFilePath is the absolute path of the markdown file being rendered.
// port is the server's HTTP port.
var reImgSrc = regexp.MustCompile(`<img\s[^>]*src="([^"]+)"`)

func rewriteImagePaths(htmlContent string, mdFilePath string, port int) string {
	fileDir := filepath.Dir(mdFilePath)

	return reImgSrc.ReplaceAllStringFunc(htmlContent, func(imgTag string) string {
		submatch := reImgSrc.FindStringSubmatch(imgTag)
		if len(submatch) < 2 {
			return imgTag
		}
		src := submatch[1]

		// Skip absolute URLs, data URIs, anchors, and scheme-only
		if strings.HasPrefix(src, "http://") ||
			strings.HasPrefix(src, "https://") ||
			strings.HasPrefix(src, "data:") ||
			strings.HasPrefix(src, "#") ||
			strings.HasPrefix(src, "//") {
			return imgTag
		}

		// Skip paths that are already /file/ URLs
		if strings.HasPrefix(src, "/file/") {
			return imgTag
		}

		// Resolve relative path against the markdown file's directory
		resolved := filepath.Join(fileDir, src)
		resolved = filepath.Clean(resolved)

		// Build the new src: /file/<absolute-path-without-leading-slash>
		// This avoids // in the URL which triggers ServeMux redirects.
		// The handler re-adds the leading /.
		pathForURL := strings.TrimPrefix(resolved, "/")
		newSrc := "/file/" + pathForURL

		return strings.Replace(imgTag, `src="`+src+`"`, `src="`+newSrc+`"`, 1)
	})
}

// BodyHTML is the rendered fragment produced by RenderBody: the frontmatter
// block, the rendered Markdown body, and the resolved page title. It contains
// no page chrome (no <html>/<head>, no CSS, no <script>) — the caller assembles
// those. This is what the Wails frontend swaps into #content.innerHTML.
type BodyHTML struct {
	// Frontmatter is the HTML for the collapsible frontmatter <details> block,
	// or "" when the file has no frontmatter.
	Frontmatter string
	// Body is the rendered Markdown HTML, with relative image paths rewritten
	// to /file/<abs-path> URLs.
	Body string
	// Title is the resolved page title: opts.Title, else the frontmatter Title,
	// else the file's base name. It is NOT prefixed with "md-view: ".
	Title string
}

// RenderBody reads a markdown file and renders it to a body fragment
// (frontmatter HTML + body HTML + title), without any page chrome. Both the
// desktop app (via App.openPath) and the legacy full-page Render use it.
func RenderBody(filePath string, opts Options) (*BodyHTML, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", filePath, err)
	}

	fm, body, hasFM := extractFrontmatter(data)

	// Always use "github" style for goldmark highlighting — both light and dark
	// chroma CSS are included so the theme toggle works.
	chromaStyle := "github"

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle(chromaStyle),
				highlighting.WithFormatOptions(
					chroma_html.WithClasses(true),
				),
			),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(body, &buf); err != nil {
		return nil, fmt.Errorf("cannot convert markdown: %w", err)
	}

	// Rewrite relative image paths to /file/<abs-path>. The Port argument is
	// unused by rewriteImagePaths (it builds port-independent /file/ URLs) but
	// is kept for Options API compatibility.
	renderedHTML := rewriteImagePaths(buf.String(), filePath, opts.Port)

	// Resolve the page title.
	title := opts.Title
	if title == "" && hasFM && fm.Title != "" {
		title = fm.Title
	}
	if title == "" {
		title = filepath.Base(filePath)
	}

	// Frontmatter section (empty string if none).
	fmHTML := ""
	if hasFM {
		fmHTML = formatFrontmatterHTML(fm)
	}

	return &BodyHTML{
		Frontmatter: fmHTML,
		Body:        renderedHTML,
		Title:       title,
	}, nil
}

// Render reads a markdown file and returns a full standalone HTML document.
// It is a thin assembler over RenderBody: it takes the body fragment and wraps
// it with page chrome (CSS, scripts, mermaid, reload). Used by the legacy HTTP
// server (pkg/server); the desktop app calls RenderBody directly.
func Render(filePath string, opts Options) (string, error) {
	body, err := RenderBody(filePath, opts)
	if err != nil {
		return "", err
	}

	chromaCSS, err := ChromaCSSBoth()
	if err != nil {
		return "", err
	}

	// Reload script
	reloadScript := ""
	if !opts.NoReload && opts.File != "" {
		encodedPath := strings.ReplaceAll(opts.File, " ", "%20")
		reloadScript = fmt.Sprintf(
			`<script>
%s
new MDSReloader("http://localhost:%d/events?file=%s");
</script>`,
			string(reloadJS),
			opts.Port,
			encodedPath,
		)
	}

	// Mermaid init script (detects ```mermaid blocks and renders them)
	// mermaid.js is served from the daemon at /static/mermaid.min.js
	mermaidScript := fmt.Sprintf(`<script src="http://localhost:%d/static/mermaid.min.js"></script>
<script>
%s
</script>`, opts.Port, string(mermaidInitJS))

	// Copy-to-clipboard script
	copyButtonScript := fmt.Sprintf(`<script>
%s
</script>`, string(copyButtonJS))

	// reMarkable upload button script
	remarkableButtonScript := fmt.Sprintf(`<script>
%s
</script>`, string(remarkableButtonJS))

	// Toolbar buttons (copy path, download)
	toolbarButtonsScript := fmt.Sprintf(`<script>
%s
</script>`, string(toolbarButtonsJS))

	// Theme toggle script
	themeToggleScript := `<script>
(function() {
    var toggle = document.querySelector('.md-view-theme-toggle');
    if (!toggle) return;
    toggle.addEventListener('click', function() {
        var html = document.documentElement;
        var current = html.getAttribute('data-theme');
        var next = current === 'dark' ? 'light' : 'dark';
        html.setAttribute('data-theme', next);
        toggle.textContent = next === 'dark' ? '☀ Light' : '🌙 Dark';
        try { localStorage.setItem('md-view-theme', next); } catch(e) {}
    });
    // Restore saved theme
    try {
        var saved = localStorage.getItem('md-view-theme');
        if (saved) {
            document.documentElement.setAttribute('data-theme', saved);
            toggle.textContent = saved === 'dark' ? '☀ Light' : '🌙 Dark';
        }
    } catch(e) {}
})();
</script>`

	// Page title (RenderBody resolved the title; add the app prefix).
	title := "md-view: " + body.Title

	// Frontmatter section (RenderBody already formatted it).
	fmHTML := body.Frontmatter

	// Dark CSS (always included — activated by data-theme="dark")
	darkStyle := fmt.Sprintf(`<style>
%s
</style>`, string(darkCSS))

	// Theme attribute on <html>
	htmlThemeAttr := ""
	if opts.Dark {
		htmlThemeAttr = ` data-theme="dark"`
	}

	// Theme toggle button
	themeToggleBtn := `<button class="md-view-theme-toggle">🌙 Dark</button>`

	htmlPage := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en"%s>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<style>
%s
</style>
<style>
%s
</style>
<style>
%s
</style>
%s
</head>
<body class="markdown-body">
%s
%s
%s
%s
%s
</body>
</html>`,
		htmlThemeAttr,
		title,
		string(defaultCSS),
		chromaCSS,
		themeCSS(opts.Dark),
		darkStyle,
		themeToggleBtn,
		fmHTML,
		body.Body,
		mermaidScript,
		reloadScript+themeToggleScript+copyButtonScript+remarkableButtonScript+toolbarButtonsScript,
	)

	return htmlPage, nil
}
