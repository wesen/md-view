package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-go-golems/md-view/pkg/daemon"
	"github.com/go-go-golems/md-view/pkg/protocol"
	"github.com/go-go-golems/md-view/pkg/renderer"
	"github.com/go-go-golems/md-view/pkg/watcher"
)

// Server is the md-view HTTP + Unix socket server.
type Server struct {
	httpServer  *http.Server
	port        int
	watcher     *watcher.FileWatcher
	mu          sync.Mutex
	sseClients  map[string]map[<-chan struct{}]struct{} // file path → set of watch channels
	browser     string                                  // override browser command
	noReload    bool                                    // disable live reload
	allowedDirs map[string]bool                         // directories allowed for /file/ serving
}

// NewServer creates a new Server bound to localhost on the given port (0 = random).
func NewServer(port int, browser string, noReload bool) (*Server, error) {
	fw, err := watcher.New()
	if err != nil {
		return nil, err
	}

	s := &Server{
		port:        port,
		watcher:     fw,
		sseClients:  make(map[string]map[<-chan struct{}]struct{}),
		browser:     browser,
		noReload:    noReload,
		allowedDirs: make(map[string]bool),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/render", s.handleRender)
	mux.HandleFunc("/raw", s.handleRaw)
	mux.HandleFunc("/events", s.handleEvents)
	mux.HandleFunc("/static/", s.handleStatic)
	mux.HandleFunc("/file/", s.handleFileServing)
	mux.HandleFunc("/upload-remarkable", s.handleUploadRemarkable)
	mux.HandleFunc("/favicon.ico", s.handleFavicon)

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s, nil
}

// Start starts the HTTP server and Unix socket listener.
// This is a blocking call — it returns when the server shuts down.
func (s *Server) Start(ctx context.Context) error {
	// Start file watcher
	s.watcher.Start()

	// Start HTTP listener
	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("cannot listen on %s: %w", s.httpServer.Addr, err)
	}

	// Get the actual port (important when port=0)
	s.port = listener.Addr().(*net.TCPAddr).Port

	// Write state files
	if err := daemon.WritePort(s.port); err != nil {
		return fmt.Errorf("cannot write port file: %w", err)
	}

	// Start Unix socket listener
	socketPath, err := daemon.SocketPath()
	if err != nil {
		return fmt.Errorf("cannot determine socket path: %w", err)
	}
	// Remove stale socket
	_ = os.Remove(socketPath)

	socketListener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("cannot listen on unix socket: %w", err)
	}
	// Restrict socket permissions
	_ = os.Chmod(socketPath, 0600)

	go s.acceptUnixConnections(ctx, socketListener)

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case <-sigCh:
			log.Println("Received shutdown signal")
			s.Shutdown()
		case <-ctx.Done():
			s.Shutdown()
		}
	}()

	log.Printf("md-view server listening on http://localhost:%d (socket: %s)", s.port, socketPath)

	// Serve HTTP
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() {
	_ = s.watcher.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.httpServer.Shutdown(ctx)
	_ = daemon.Cleanup()
}

// Port returns the actual HTTP port the server is listening on.
func (s *Server) Port() int {
	return s.port
}

// acceptUnixConnections accepts connections on the Unix domain socket.
func (s *Server) acceptUnixConnections(ctx context.Context, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("Unix socket accept error: %v", err)
				continue
			}
		}
		go s.handleSocketConn(ctx, conn)
	}
}

// handleSocketConn handles a single Unix socket connection.
func (s *Server) handleSocketConn(_ context.Context, conn net.Conn) {
	defer func() { _ = conn.Close() }()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	var cmd protocol.Command
	line := strings.TrimSpace(string(buf[:n]))
	if line == "" {
		return
	}
	if err := json.Unmarshal([]byte(line), &cmd); err != nil {
		s.writeSocketResponse(conn, protocol.Response{Status: "error", Message: fmt.Sprintf("invalid command: %v", err)})
		return
	}

	switch cmd.Command {
	case "view":
		url := fmt.Sprintf("http://localhost:%d/render?file=%s", s.port, urlEncodePath(cmd.Path))
		if cmd.Dark {
			url += "&theme=dark"
		}
		s.writeSocketResponse(conn, protocol.Response{Status: "ok", URL: url})

		// Open browser in background (if browser command provided)
		if cmd.Browser != "" {
			go s.openBrowserWith(url, cmd.Browser)
		} else {
			go s.openBrowser(url)
		}

	case "ping":
		s.writeSocketResponse(conn, protocol.Response{Status: "pong"})

	case "stop":
		s.writeSocketResponse(conn, protocol.Response{Status: "ok", Message: "shutting down"})
		go func() {
			time.Sleep(100 * time.Millisecond)
			s.Shutdown()
			os.Exit(0)
		}()

	default:
		s.writeSocketResponse(conn, protocol.Response{Status: "error", Message: fmt.Sprintf("unknown command: %s", cmd.Command)})
	}
}

// --- HTTP Handlers ---

// renderErrorPage returns a styled HTML error page.
func renderErrorPage(code int, title, message string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>md-view: %s</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif; color: #24292e; max-width: 600px; margin: 80px auto; padding: 0 20px; }
h1 { font-size: 24px; color: #cf222e; margin-bottom: 8px; }
p { color: #656d76; font-size: 16px; line-height: 1.5; }
code { background: #f6f8fa; padding: 2px 6px; border-radius: 3px; font-size: 14px; }
.status { font-size: 72px; font-weight: 700; color: #d0d7de; margin-bottom: 16px; }
</style>
</head>
<body>
<div class="status">%d</div>
<h1>%s</h1>
<p>%s</p>
</body>
</html>`, title, code, title, message)
}

func (s *Server) handleRender(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		s.writeErrorHTML(w, 400, "Bad Request", "Missing <code>file</code> query parameter. Usage: <code>/render?file=/path/to/file.md</code>")
		return
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		s.writeErrorHTML(w, 400, "Bad Request", fmt.Sprintf("Invalid path: %v", err))
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		s.writeErrorHTML(w, 404, "Not Found", fmt.Sprintf("File not found: <code>%s</code>", htmlEscape(absPath)))
		return
	}
	if !info.Mode().IsRegular() {
		s.writeErrorHTML(w, 400, "Bad Request", fmt.Sprintf("<code>%s</code> is not a regular file", htmlEscape(absPath)))
		return
	}

	// Check for theme parameter
	dark := r.URL.Query().Get("theme") == "dark"

	// Register the file's directory and all ancestor directories as allowed
	// for /file/ serving. This is necessary because markdown can reference
	// images with ../ paths that resolve outside the immediate parent dir.
	s.mu.Lock()
	dir := filepath.Dir(absPath)
	for {
		s.allowedDirs[dir] = true
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	s.mu.Unlock()

	opts := renderer.Options{
		File:     absPath,
		Port:     s.port,
		NoReload: s.noReload,
		Dark:     dark,
	}

	html, err := renderer.Render(absPath, opts)
	if err != nil {
		s.writeErrorHTML(w, 500, "Internal Server Error", fmt.Sprintf("Render error: %v", err))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html)) // #nosec G705 -- HTML is server-rendered from markdown
}

func (s *Server) handleRaw(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "missing file parameter", http.StatusBadRequest)
		return
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid path: %v", err), http.StatusBadRequest)
		return
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read file: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(data) // #nosec G705 -- raw file content served as text/plain
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "missing file parameter", http.StatusBadRequest)
		return
	}

	// Resolve absolute path for consistent key
	absPath, _ := filepath.Abs(filePath)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch, err := s.watcher.Watch(absPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot watch file: %v", err), http.StatusInternalServerError)
		return
	}

	// Track SSE client for cleanup
	s.mu.Lock()
	if s.sseClients[absPath] == nil {
		s.sseClients[absPath] = make(map[<-chan struct{}]struct{})
	}
	s.sseClients[absPath][ch] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.sseClients[absPath], ch)
		if len(s.sseClients[absPath]) == 0 {
			delete(s.sseClients, absPath)
		}
		s.mu.Unlock()
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial comment to establish connection
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-ch:
			fmt.Fprintf(w, "event: reload\ndata: reload\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/static/base.css":
		w.Header().Set("Content-Type", "text/css")
		_, _ = w.Write(renderer.CSS())
	case "/static/reload.js":
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write(renderer.ReloadJS())
	case "/static/mermaid.min.js":
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write(renderer.MermaidJS())
	case "/static/copy-button.js":
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write(renderer.CopyButtonJS())
	case "/static/remarkable-button.js":
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write(renderer.RemarkableButtonJS())
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleFileServing(w http.ResponseWriter, r *http.Request) {
	// URL path: /file/<path-without-leading-slash>
	// e.g. /file/tmp/md-test/images/diagram.png (absolute path was /tmp/md-test/...)
	filePath := strings.TrimPrefix(r.URL.Path, "/file/")
	if filePath == "" {
		http.NotFound(w, r)
		return
	}

	// Re-add the leading / stripped to avoid // in URLs
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	// Security: only serve files under an allowed directory
	s.mu.Lock()
	allowed := false
	for dir := range s.allowedDirs {
		if strings.HasPrefix(absPath, dir+string(filepath.Separator)) || absPath == dir {
			allowed = true
			break
		}
	}
	s.mu.Unlock()

	if !allowed {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}

	// Read and serve the file directly (don't use http.ServeFile — it
	// redirects to "clean" the URL, breaking absolute paths with leading /).
	f, err := os.Open(absPath)
	if err != nil {
		http.Error(w, "cannot open file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = f.Close() }()

	http.ServeContent(w, r, filepath.Base(absPath), info.ModTime(), f)
}

func (s *Server) handleUploadRemarkable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeErrorJSON(w, http.StatusMethodNotAllowed, "use POST")
		return
	}

	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		s.writeErrorJSON(w, http.StatusBadRequest, "missing file parameter")
		return
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		s.writeErrorJSON(w, http.StatusBadRequest, fmt.Sprintf("invalid path: %v", err))
		return
	}

	if _, err := os.Stat(absPath); err != nil {
		s.writeErrorJSON(w, http.StatusNotFound, fmt.Sprintf("file not found: %s", absPath))
		return
	}

	// Run remarquee upload md <file> --non-interactive
	cmd := exec.Command("remarquee", "upload", "md", absPath, "--non-interactive") // #nosec G702 -- fixed args, file path is validated
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		log.Printf("reMarkable upload failed for %s: %s", absPath, errMsg)
		s.writeErrorJSON(w, http.StatusInternalServerError, errMsg)
		return
	}

	output := strings.TrimSpace(stdout.String())
	log.Printf("reMarkable upload succeeded for %s: %s", absPath, output)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": output,
	})
}

func (s *Server) writeErrorJSON(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "error",
		"message": message,
	})
}

func (s *Server) handleFavicon(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// --- Helpers ---

// openBrowser opens a URL in the user's browser.
//
//nolint:gosec // G702: browser command comes from CLI flags/env, not untrusted input
func (s *Server) openBrowser(url string) {
	browser := s.browser
	if browser == "" {
		browser = os.Getenv("BROWSER")
	}
	if browser == "" {
		for _, b := range []string{"xdg-open", "firefox", "google-chrome", "chromium"} {
			if _, err := exec.LookPath(b); err == nil {
				browser = b
				break
			}
		}
	}
	if browser == "" {
		log.Println("Warning: no browser found (set $BROWSER)")
		return
	}

	var cmd *exec.Cmd
	switch browser {
	case "xdg-open":
		cmd = exec.Command(browser, url) // #nosec G702 -- browser comes from CLI flags/env, not untrusted input
	case "firefox":
		cmd = exec.Command(browser, "--new-window", url) // #nosec G702
	default:
		cmd = exec.Command(browser, "--new-window", url) // #nosec G702
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		log.Printf("Cannot open browser: %v", err)
	}
}

// openBrowserWith opens a URL using the given browser command string.
// The command string can contain arguments (e.g. "firefox --new-window")
// which are split before execution.
//
//nolint:gosec // G702: browserCmd comes from CLI flags, not untrusted input
func (s *Server) openBrowserWith(url, browserCmd string) {
	parts := strings.Fields(browserCmd)
	if len(parts) == 0 {
		return
	}
	args := append(parts[1:], url)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		log.Printf("Cannot open browser %q: %v", browserCmd, err)
	}
}

func urlEncodePath(p string) string {
	return strings.ReplaceAll(p, " ", "%20")
}

// writeErrorHTML writes a styled HTML error page.
func (s *Server) writeErrorHTML(w http.ResponseWriter, code int, title, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(renderErrorPage(code, title, message))) // #nosec G705 -- error page is template-rendered
}

func (s *Server) writeSocketResponse(conn net.Conn, resp protocol.Response) {
	data, _ := json.Marshal(resp)
	_, _ = conn.Write(append(data, '\n')) // #nosec G702 -- data is json.Marshal output
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
