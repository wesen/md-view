package main

import (
	"strings"
)

// ViewArgs holds the file path and theme parsed from a `md-view view ...`
// invocation (used both for the first-instance launch and a second instance's
// forwarded os.Args via SingleInstanceLock).
type ViewArgs struct {
	File string // absolute path, or "" if none given
	Dark bool   // --dark flag
}

// ParseViewArgs parses the argument vector of a `md-view` invocation and
// returns the requested file (if any) and theme. It accepts:
//
//	md-view <file> [--dark]            # bare/shorthand (root command)
//	md-view view <file> [--dark]       # explicit `view` subcommand
//	md-view view --dark <file>
//
// Unknown flags are tolerated (forward-compatible); the file is the first
// non-flag, non-"view" argument. Relative paths are returned as-is; callers
// resolve them against the working directory (ParseViewArgs is cwd-agnostic,
// but SingleInstanceData carries WorkingDirectory for the 2nd-instance path).
func ParseViewArgs(args []string) ViewArgs {
	var out ViewArgs
	for _, a := range args {
		switch {
		case a == "view":
			// the explicit subcommand verb; ignore
		case a == "--dark":
			out.Dark = true
		case strings.HasPrefix(a, "-"):
			// unknown flag; tolerated (DR-8: removed flags error at the Cobra
			// layer, but 2nd-instance forwarding is lenient here)
		default:
			if out.File == "" {
				out.File = a
			}
		}
	}
	return out
}
