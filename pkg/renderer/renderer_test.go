package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTitle string
		wantBody  string
		wantHasFM bool
	}{
		{
			name:      "no frontmatter",
			input:     "# Hello\n\nWorld",
			wantBody:  "# Hello\n\nWorld",
			wantHasFM: false,
		},
		{
			name:      "simple frontmatter",
			input:     "---\nTitle: Test Doc\n---\n# Hello\n\nWorld",
			wantTitle: "Test Doc",
			wantBody:  "# Hello\n\nWorld",
			wantHasFM: true,
		},
		{
			name:      "frontmatter with nested values",
			input:     "---\nTitle: My Page\nStatus: active\nTopics:\n  - go\n  - cli\n---\n# Content",
			wantTitle: "My Page",
			wantBody:  "# Content",
			wantHasFM: true,
		},
		{
			name:      "unclosed frontmatter",
			input:     "---\nTitle: Broken\n# Not frontmatter",
			wantBody:  "---\nTitle: Broken\n# Not frontmatter",
			wantHasFM: false,
		},
		{
			name:      "quoted title",
			input:     "---\nTitle: \"Quoted Title\"\n---\nBody",
			wantTitle: "Quoted Title",
			wantBody:  "Body",
			wantHasFM: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, hasFM := extractFrontmatter([]byte(tt.input))
			if hasFM != tt.wantHasFM {
				t.Fatalf("hasFrontmatter = %v, want %v", hasFM, tt.wantHasFM)
			}
			if !hasFM {
				if string(body) != tt.wantBody {
					t.Errorf("body = %q, want %q", string(body), tt.wantBody)
				}
				return
			}
			if fm.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", fm.Title, tt.wantTitle)
			}
			if string(body) != tt.wantBody {
				t.Errorf("body = %q, want %q", string(body), tt.wantBody)
			}
		})
	}
}

func TestRender(t *testing.T) {
	// Create a temp markdown file
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := "# Hello\n\nThis is **bold** text.\n"
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	html, err := Render(mdFile, Options{NoReload: true})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Check key HTML elements
	checks := []string{
		"md-view: test.md",
		"<h1>Hello</h1>",
		"<strong>bold</strong>",
		"markdown-body",
	}
	for _, check := range checks {
		if !contains(html, check) {
			t.Errorf("Render() missing %q in output", check)
		}
	}

	// Should NOT have reload script when NoReload is true
	if contains(html, "MDSReloader") {
		t.Error("Render() should not include reload script when NoReload=true")
	}
}

func TestRenderWithFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "fm.md")
	content := "---\nTitle: My Document\nStatus: draft\n---\n# Content\n"
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	html, err := Render(mdFile, Options{NoReload: true})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Title should come from frontmatter
	if !contains(html, "md-view: My Document") {
		t.Error("Render() should use frontmatter Title as page title")
	}

	// Frontmatter should be in a <details> block
	if !contains(html, "md-view-frontmatter") {
		t.Error("Render() should include frontmatter details block")
	}

	// Body should not contain the frontmatter YAML
	if contains(html, "---") {
		t.Error("Render() should strip frontmatter delimiters from body")
	}
}

func TestRenderWithFrontmatterLowercase(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "fm2.md")
	content := "---\ntitle: Lowercase Title\nstatus: draft\n---\n# Content\n"
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	html, err := Render(mdFile, Options{NoReload: true})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !contains(html, "md-view: Lowercase Title") {
		t.Error("Render() should use frontmatter title (lowercase) as page title")
	}
}

func TestRewriteImagePaths(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		mdFile  string
		wantSrc string
	}{
		{
			name:    "relative image",
			input:   `<img src="./images/diagram.png" alt="diagram">`,
			mdFile:  "/home/user/project/README.md",
			wantSrc: `/file/home/user/project/images/diagram.png`,
		},
		{
			name:    "relative image without dot",
			input:   `<img src="images/photo.jpg" alt="photo">`,
			mdFile:  "/home/user/project/README.md",
			wantSrc: `/file/home/user/project/images/photo.jpg`,
		},
		{
			name:    "absolute URL unchanged",
			input:   `<img src="https://example.com/img.png" alt="ext">`,
			mdFile:  "/home/user/project/README.md",
			wantSrc: "https://example.com/img.png",
		},
		{
			name:    "data URI unchanged",
			input:   `<img src="data:image/png;base64,abc123" alt="inline">`,
			mdFile:  "/home/user/project/README.md",
			wantSrc: "data:image/png;base64,abc123",
		},
		{
			name:    "no images",
			input:   `<p>Hello world</p>`,
			mdFile:  "/home/user/project/README.md",
			wantSrc: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriteImagePaths(tt.input, tt.mdFile, 8080)
			if tt.wantSrc == "" {
				if contains(result, "/file/") {
					t.Errorf("unexpected /file/ rewrite in %q", result)
				}
				return
			}
			if !contains(result, `src="`+tt.wantSrc+`"`) {
				t.Errorf("rewriteImagePaths() = %q, want src=%q", result, tt.wantSrc)
			}
		})
	}
}

func TestRenderWithImages(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := "# Hello\n\n![diagram](./images/diagram.png)\n"
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	html, err := Render(mdFile, Options{NoReload: true, Port: 8080})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	expectedSrc := "/file/" + strings.TrimPrefix(filepath.Join(tmpDir, "images/diagram.png"), "/")
	if !contains(html, expectedSrc) {
		t.Errorf("Render() should rewrite image src to %q, got HTML:\n%s", expectedSrc, html)
	}
}

func TestRenderWithCodeBlockHasCopyButton(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := "# Hello\n\n```go\nfmt.Println(\"hello\")\n```\n"
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	html, err := Render(mdFile, Options{NoReload: true, Port: 8080})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Should have the copy button script
	if !contains(html, "md-view-copy-btn") {
		t.Error("Render() should include copy-to-clipboard script")
	}
	if !contains(html, "md-view-code-container") {
		t.Error("Render() should include code container CSS class")
	}
}

func TestRenderWithMermaidBlock(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := "# Diagram\n\n```mermaid\ngraph TD; A-->B;\n```\n"
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	html, err := Render(mdFile, Options{NoReload: true, Port: 8080})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Should have mermaid init script
	if !contains(html, "mermaid.initialize") {
		t.Error("Render() should include mermaid init script")
	}
	if !contains(html, "mermaid.min.js") {
		t.Error("Render() should include mermaid.js library")
	}
}

func TestRenderBody(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := "# Hello\n\nThis is **bold** text.\n"
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	body, err := RenderBody(mdFile, Options{})
	if err != nil {
		t.Fatalf("RenderBody() error = %v", err)
	}

	// Title should be the filename, NOT prefixed with "md-view: "
	if body.Title != "test.md" {
		t.Errorf("Title = %q, want %q", body.Title, "test.md")
	}

	// Body should contain rendered markdown, not the page chrome
	for _, want := range []string{"<h1>Hello</h1>", "<strong>bold</strong>"} {
		if !contains(body.Body, want) {
			t.Errorf("Body missing %q, got: %s", want, body.Body)
		}
	}

	// A body fragment must NOT contain page chrome
	for _, unwanted := range []string{"<!DOCTYPE", "<html", "<head", "md-view:", "chroma"} {
		if contains(body.Body, unwanted) {
			t.Errorf("Body should not contain page chrome %q", unwanted)
		}
	}

	// No frontmatter → empty Frontmatter field
	if body.Frontmatter != "" {
		t.Errorf("Frontmatter = %q, want empty", body.Frontmatter)
	}
}

func TestRenderBodyWithFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "fm.md")
	content := "---\nTitle: My Doc\nStatus: draft\n---\n# Content\n"
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	body, err := RenderBody(mdFile, Options{})
	if err != nil {
		t.Fatalf("RenderBody() error = %v", err)
	}

	if body.Title != "My Doc" {
		t.Errorf("Title = %q, want %q (from frontmatter)", body.Title, "My Doc")
	}
	if !contains(body.Frontmatter, "md-view-frontmatter") {
		t.Errorf("Frontmatter should contain the details block, got: %s", body.Frontmatter)
	}
	if contains(body.Body, "---") {
		t.Errorf("Body should strip frontmatter delimiters, got: %s", body.Body)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
