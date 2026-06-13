package main

import "testing"

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
