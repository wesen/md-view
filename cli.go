package main

import (
	"os"
	"path/filepath"
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

// absolutizeFileArg resolves `file` against the current process's working
// directory and, when it was relative, rewrites the matching entry in os.Args
// to the absolute form. runDesktop calls this before wails.Run so that:
//
//   - the first instance opens the right file (PendingOpen becomes absolute),
//   - and if Wails' SingleInstanceLock forwards THIS process to an already-
//     running instance, the forwarded os.Args carry the absolute path.
//
// The os.Args rewrite is necessary because Wails' SingleInstanceLock forwards
// os.Args[1:] verbatim, and on macOS the accompanying WorkingDirectory is the
// executable's directory (not the caller's cwd), so a relative path would
// resolve incorrectly in OnSecondInstanceLaunch. An already-absolute `file`
// and any `file` not found verbatim in os.Args are returned unchanged (the
// latter still resolves correctly for the first instance via PendingOpen).
func absolutizeFileArg(file string) string {
	if file == "" || filepath.IsAbs(file) {
		return file
	}
	abs, err := filepath.Abs(file)
	if err != nil {
		return file
	}
	for i, arg := range os.Args {
		if arg == file {
			os.Args[i] = abs
			break
		}
	}
	return abs
}
