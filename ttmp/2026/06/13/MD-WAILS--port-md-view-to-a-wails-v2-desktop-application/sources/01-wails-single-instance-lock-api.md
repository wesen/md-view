---
title: "Wails v2 SingleInstanceLock API (authoritative source)"
source: "https://raw.githubusercontent.com/wailsapp/wails/master/v2/pkg/options/options.go"
guide: "https://wails.io/docs/guides/single-instance-lock/"
retrieved: 2026-06-13
ticket: MD-WAILS
why: "Drop-in CLI compatibility requires that a second `md-view view` invocation routes to the already-running app instead of starting a new window. Wails v2.7.0+ has a built-in SingleInstanceLock option that does exactly this — no external plugin, no hand-rolled socket/PID files."
---

# Wails v2 SingleInstanceLock API

> **Why this is the key resource for MD-WAILS.** The current md-view achieves "one server, many views" with a Unix-socket daemon + PID/port files (`pkg/daemon`, `pkg/protocol`). In a Wails replacement we need the same UX: running `md-view view a.md` then `md-view view b.md` should reuse one app and open `b.md`, not spawn a second window. Wails' built-in `SingleInstanceLock` provides this directly — the second invocation's `os.Args` are forwarded to the first instance via a callback. This **replaces** the entire daemon/socket/PID subsystem with ~15 lines of config.

## How it works (conceptual)

Wails provides functionality to ensure that only one instance of an application can run at a time. If a user attempts to launch a second instance, the application can be configured to either:

1. bring the existing instance to the foreground, **or**
2. pass command-line arguments from the second instance to the first.

The mechanism is OS-native (a named lock + IPC channel keyed on a `UniqueId` you choose). No file-based PID/socket bookkeeping is required.

## The API (verbatim from `pkg/options/options.go`, Wails master)

```go
// App contains options for creating the App
type App struct {
    Title              string
    Width              int
    Height             int
    DisableResize      bool
    Fullscreen         bool
    Frameless          bool
    SingleInstanceLock *SingleInstanceLock   // <-- here

    Windows *windows.Options
    Mac     *mac.Options
    Linux   *linux.Options
    // ...
    DragAndDrop *DragAndDrop
    // ...
}

// SingleInstanceLock — added in Wails v2.7.0
type SingleInstanceLock struct {
    // uniqueId that will be used for setting up messaging between instances
    UniqueId               string
    OnSecondInstanceLaunch func(secondInstanceData SecondInstanceData)
}

type SecondInstanceData struct {
    Args             []string
    WorkingDirectory string
}

// NewSecondInstanceData is a helper used internally to capture the
// second invocation's os.Args and CWD.
func NewSecondInstanceData() (*SecondInstanceData, error) {
    ex, err := os.Executable()
    if err != nil {
        return nil, err
    }
    workingDirectory := filepath.Dir(ex)
    return &SecondInstanceData{
        Args:             os.Args[1:],
        WorkingDirectory: workingDirectory,
    }, nil
}
```

## Minimal usage

```go
err := wails.Run(&options.App{
    Title: "md-view",
    // ...
    SingleInstanceLock: &options.SingleInstanceLock{
        UniqueId:               "github.com/go-go-golems/md-view",
        OnSecondInstanceLaunch: onSecondInstance,
    },
})

func onSecondInstance(data options.SecondInstanceData) {
    // data.Args == os.Args[1:] of the SECOND process.
    // e.g. ["view", "/abs/path/to/b.md", "--dark"]
    // Parse them (Cobra, or by hand) and open the file in THIS (first) instance.
    runtime.EventsEmit(app.ctx, "open-from-cli", data.Args)
}
```

## How this maps onto md-view's drop-in CLI

| Current (daemon model) | Wails (single-instance model) |
|------------------------|-------------------------------|
| `md-view view a.md` spawns/reuses daemon, opens browser | `md-view view a.md` starts the app (first instance) |
| 2nd `md-view view b.md` → socket command to daemon | 2nd process exits immediately; `OnSecondInstanceLaunch` fires in instance #1 with `Args=["view","b.md"]` |
| `pkg/daemon` (PID/port/socket files) | Not needed — Wails owns the lock |
| `pkg/protocol` (Unix-socket JSON RPC) | Not needed — `os.Args` forwarded directly |
| Browser tab | Native app window (`runtime.WindowSetTitle`) |

## Notes / caveats

- **Availability:** `SingleInstanceLock` was added in **Wails v2.7.0**. The demo at `2026-06-13--wails-demo` pins `wails/v2 v2.12.0`, which includes it.
- **First vs. second instance:** The first process blocks in `wails.Run(...)` and owns the window. The second process calls the same `main()`, but Wails detects the lock, triggers `OnSecondInstanceLaunch` in instance #1, and then instance #2 **exits**. So instance #2 must not do irreversible work (e.g. must not overwrite config) before Wails hands off.
- **Foregrounding:** Wails automatically brings the first instance's window to the front on a second launch (platform-dependent); you still must do the file-open work in the callback.
- **UniqueId:** Choose a stable, unique string (the module path works). Collisions with other Wails apps on the same machine would break the lock.

## Provenance

- Source file (authoritative types): `https://raw.githubusercontent.com/wailsapp/wails/master/v2/pkg/options/options.go` — retrieved 2026-06-13.
- User guide (Cloudflare-protected, could not fetch raw; summary derived from search snippets + source): `https://wails.io/docs/guides/single-instance-lock/`
- pkg.go.dev reference: `https://pkg.go.dev/github.com/wailsapp/wails/v2/pkg/options#SingleInstanceLock`
