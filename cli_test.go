package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseViewArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want ViewArgs
	}{
		{
			name: "bare view shorthand",
			args: []string{"README.md"},
			want: ViewArgs{File: "README.md"},
		},
		{
			name: "explicit view verb",
			args: []string{"view", "README.md"},
			want: ViewArgs{File: "README.md"},
		},
		{
			name: "view with dark after file",
			args: []string{"view", "notes.md", "--dark"},
			want: ViewArgs{File: "notes.md", Dark: true},
		},
		{
			name: "view with dark before file",
			args: []string{"view", "--dark", "notes.md"},
			want: ViewArgs{File: "notes.md", Dark: true},
		},
		{
			name: "dark only, no file",
			args: []string{"view", "--dark"},
			want: ViewArgs{Dark: true},
		},
		{
			name: "no args",
			args: []string{},
			want: ViewArgs{},
		},
		{
			// Note: ParseViewArgs is flag-value-unaware. `--port` is skipped as a
			// flag, but `8080` (its value) is then treated as the first non-flag
			// argument and becomes the file. Callers that need real flag parsing
			// should use Cobra; ParseViewArgs is the lenient 2nd-instance path.
			name: "unknown flag value becomes the file (lenient parser)",
			args: []string{"view", "--port", "8080", "doc.md"},
			want: ViewArgs{File: "8080"},
		},
		{
			name: "only the first non-flag is the file",
			args: []string{"view", "first.md", "second.md"},
			want: ViewArgs{File: "first.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseViewArgs(tt.args)
			if got.File != tt.want.File || got.Dark != tt.want.Dark {
				t.Errorf("ParseViewArgs(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}

// TestAbsolutizeFileArgRewritesOsArgs verifies that a relative file argument
// is resolved against the process cwd AND that the matching os.Args entry is
// rewritten to absolute — the property runDesktop relies on so a Wails
// SingleInstanceLock handoff forwards an absolute path (macOS ships the
// executable dir as WorkingDirectory, not the caller's cwd).
func TestAbsolutizeFileArgRewritesOsArgs(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	savedArgs := os.Args
	t.Cleanup(func() { os.Args = savedArgs })
	os.Args = []string{"md-view", "view", "notes.md"}

	got := absolutizeFileArg("notes.md")
	want := filepath.Join(dir, "notes.md")
	if got != want {
		t.Errorf("absolutizeFileArg returned %q, want %q", got, want)
	}
	if os.Args[2] != want {
		t.Errorf("os.Args[2] = %q, want it rewritten to %q", os.Args[2], want)
	}
}

// TestAbsolutizeFileArgPreservesAbsolute verifies an already-absolute path is
// returned unchanged with no os.Args scanning side effects.
func TestAbsolutizeFileArgPreservesAbsolute(t *testing.T) {
	savedArgs := os.Args
	t.Cleanup(func() { os.Args = savedArgs })
	os.Args = []string{"md-view", "view", "/abs/path/notes.md"}

	got := absolutizeFileArg("/abs/path/notes.md")
	if got != "/abs/path/notes.md" {
		t.Errorf("absolutizeFileArg returned %q, want /abs/path/notes.md", got)
	}
}

// TestAbsolutizeFileArgEmpty verifies the no-file case is a no-op.
func TestAbsolutizeFileArgEmpty(t *testing.T) {
	if got := absolutizeFileArg(""); got != "" {
		t.Errorf("absolutizeFileArg(\"\") = %q, want \"\"", got)
	}
}

// TestAddAllowedDirTree verifies the allow-list registers a directory and its
// ancestors but NEVER the filesystem root — so /etc/passwd stays forbidden
// even after a file deep under /home/... is opened.
func TestAddAllowedDirTree(t *testing.T) {
	a := NewApp()
	sep := string(filepath.Separator)

	// Simulate opening /home/user/repo/docs/intro.md: addAllowedDirTree is
	// called with the file's directory (/home/user/repo/docs).
	a.addAllowedDirTree(sep + filepath.Join("home", "user", "repo", "docs"))

	checks := []struct {
		path string
		want bool
	}{
		// the file's own dir + contents: allowed
		{sep + filepath.Join("home", "user", "repo", "docs", "img", "a.png"), true},
		// a sibling-assets parent (../assets): allowed (ancestor registered)
		{sep + filepath.Join("home", "user", "repo", "assets", "logo.png"), true},
		// a deeper ancestor sibling: allowed (under /home/user)
		{sep + filepath.Join("home", "user", "other", "x.png"), true},
		// a system path under a DIFFERENT root subtree: forbidden
		{sep + filepath.Join("etc", "passwd"), false},
		// the root itself: forbidden
		{sep, false},
	}
	for _, c := range checks {
		if got := a.isAllowed(c.path); got != c.want {
			t.Errorf("isAllowed(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
