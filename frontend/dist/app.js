// MarkDown Viewer — Frontend Application Logic
//
// This file calls Go methods exposed via Wails bindings and updates the UI.
// Wails injects window.go and window.runtime into the WebView context at runtime.
//
// Communication patterns:
//   - JS → Go:  window['go']['main']['App']['MethodName']()  (returns Promise)
//   - Go → JS:  runtime.EventsOn('event-name', callback)     (Wails events)

// ---- DOM References ----
const openBtn = document.getElementById('open-btn');
const themeBtn = document.getElementById('theme-btn');
const contentDiv = document.getElementById('content');
const dropzone = document.getElementById('dropzone');
const filenameSpan = document.getElementById('filename');
const errorDiv = document.getElementById('error');

// ---- Open File Button (toolbar) ----
openBtn.addEventListener('click', () => {
    clearError();
    window['go']['main']['App']['OpenFile']()
        .then((html) => {
            if (html) {
                showContent(html);
            }
        })
        .catch((err) => {
            showError('Failed to open file: ' + err);
        });
});

// ---- Theme Toggle Button (toolbar) ----
themeBtn.addEventListener('click', () => {
    window['go']['main']['App']['ToggleTheme']()
        .then((theme) => {
            applyTheme(theme);
        })
        .catch((err) => {
            console.error('Failed to toggle theme:', err);
        });
});

// ---- Drag and Drop (visual feedback; Wails handles the actual event) ----
document.addEventListener('dragover', (e) => {
    e.preventDefault();
    e.stopPropagation();
    dropzone.classList.add('drag-over');
});

document.addEventListener('dragleave', (e) => {
    e.preventDefault();
    e.stopPropagation();
    dropzone.classList.remove('drag-over');
});

document.addEventListener('drop', (e) => {
    e.preventDefault();
    e.stopPropagation();
    dropzone.classList.remove('drag-over');
});

// ---- Listen for events from Go (menu callbacks, drag-and-drop) ----
// These events are emitted by the Go backend when:
//   - Menu > File > Open is clicked (Ctrl+O)
//   - A file is dragged and dropped onto the window
//   - Menu > View > Toggle Theme is clicked
//   - Menu > File > Close Window is clicked

runtime.EventsOn('file-opened', (data) => {
    if (data && data.html) {
        showContent(data.html);
        if (data.title) {
            filenameSpan.textContent = data.title;
        }
    }
});

runtime.EventsOn('file-error', (errMsg) => {
    showError(errMsg);
});

runtime.EventsOn('theme-changed', (theme) => {
    applyTheme(theme);
});

runtime.EventsOn('close-file', () => {
    contentDiv.innerHTML = '';
    contentDiv.style.display = 'none';
    dropzone.style.display = 'flex';
    filenameSpan.textContent = 'No file open';
    clearError();
});

// Live reload: the Go backend watches the open file with fsnotify and emits
// `file-changed` on write. Re-render the current file and swap the content
// (showContent re-runs copy/mermaid augmentation). This replaces SSE /events.
runtime.EventsOn('file-changed', (data) => {
    if (!data || !data.path) return;
    clearError();
    window['go']['main']['App']['ReopenCurrent']()
        .then((html) => { if (html) showContent(html); })
        .catch((err) => { showError('Reload failed: ' + err); });
});

// ---- Helper: Display Rendered HTML ----
function showContent(html) {
    contentDiv.innerHTML = html;
    contentDiv.style.display = 'block';
    dropzone.style.display = 'none';
    clearError();

    // Re-run content augmentation (copy buttons + mermaid) against the new DOM.
    // The page chrome is stable; only #content is swapped per file / reload.
    if (window.MDSAugmentPage) window.MDSAugmentPage();

    // Update the filename display in the toolbar
    window['go']['main']['App']['GetCurrentFile']()
        .then((path) => {
            if (path) {
                const parts = path.split(/[/\\]/);
                filenameSpan.textContent = parts[parts.length - 1];
            }
        });

    // Scroll to the top
    document.getElementById('main').scrollTop = 0;

    // Refresh recent files list
    loadRecentFiles();
}

// ---- Helper: Apply Theme ----
function applyTheme(theme) {
    // md-view's CSS targets [data-theme="dark"] on an ancestor of .markdown-body;
    // set it on <html> (md-view's convention) and mirror on <body>.
    document.documentElement.setAttribute('data-theme', theme);
    document.body.setAttribute('data-theme', theme);
    themeBtn.textContent = theme === 'dark' ? '☀️ Light' : '🌙 Dark';
    // Re-render mermaid diagrams for the new theme.
    if (window.MDSMermaidRerender) window.MDSMermaidRerender(theme);
}

// ---- Helper: Show Error ----
function showError(message) {
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
}

// ---- Helper: Clear Error ----
function clearError() {
    errorDiv.textContent = '';
    errorDiv.style.display = 'none';
}

// ---- Helper: Load Recent Files ----
function loadRecentFiles() {
    window['go']['main']['App']['GetRecentFiles']()
        .then((files) => {
            const sidebar = document.getElementById('sidebar');
            const list = document.getElementById('recent-list');

            if (!files || files.length === 0) {
                sidebar.style.display = 'none';
                return;
            }

            list.innerHTML = '';

            files.forEach((path) => {
                const li = document.createElement('li');
                const parts = path.split(/[/\\]/);
                li.textContent = parts[parts.length - 1];
                li.title = path;

                li.addEventListener('click', () => {
                    clearError();
                    window['go']['main']['App']['OpenFileAtPath'](path)
                        .then((html) => {
                            if (html) showContent(html);
                        })
                        .catch((err) => {
                            showError('Failed to open file: ' + err);
                        });
                });

                list.appendChild(li);
            });

            sidebar.style.display = 'block';
        });
}

// ---- Initialization ----
window.addEventListener('DOMContentLoaded', () => {
    const checkReady = setInterval(() => {
        if (window['go'] && window['go']['main']) {
            clearInterval(checkReady);

            // Load theme on startup
            window['go']['main']['App']['GetTheme']()
                .then((theme) => { applyTheme(theme); });

            // Load recent files on startup
            loadRecentFiles();
        }
    }, 50);
});
