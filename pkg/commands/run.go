package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-go-golems/logcopter/pkg/logcopter"

	"github.com/go-go-golems/md-view/pkg/daemon"
	"github.com/go-go-golems/md-view/pkg/protocol"
	"github.com/go-go-golems/md-view/pkg/server"
)

var log = logcopter.Package("md-view.commands")

// RunView implements the `md-view view` command logic:
// 1. Check if daemon is alive
// 2. If not, start it and wait for the socket
// 3. Send a "view" command over the Unix socket
// 4. Return the URL
func RunView(ctx context.Context, s *ViewSettings) (string, error) {
	// Resolve file path to absolute
	absPath, err := filepath.Abs(s.File)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}

	// Check file exists
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("cannot access file %s: %w", absPath, err)
	}

	// Ensure daemon is running
	if err := ensureDaemonRunning(s.Port); err != nil {
		return "", fmt.Errorf("cannot start daemon: %w", err)
	}

	// Wait for socket to be ready
	socketPath, err := daemon.SocketPath()
	if err != nil {
		return "", err
	}
	if err := waitForSocket(socketPath, 5*time.Second); err != nil {
		return "", fmt.Errorf("daemon socket not ready: %w", err)
	}

	// Send view command with browser and no-browser settings
	cmd := protocol.Command{
		Command: "view",
		Path:    absPath,
		Dark:    s.Dark,
	}

	// Pass browser command only if not suppressed
	if !s.NoBrowser {
		cmd.Browser = s.Browser
	}

	resp, err := protocol.SendCommand(socketPath, cmd)
	if err != nil {
		return "", fmt.Errorf("cannot send view command: %w", err)
	}

	if resp.Status != "ok" {
		return "", fmt.Errorf("daemon error: %s", resp.Message)
	}

	return resp.URL, nil
}

// RunServe implements the `md-view serve` command — starts the server in foreground.
func RunServe(ctx context.Context, s *ServeSettings) error {
	// Write PID file
	if err := daemon.WritePID(); err != nil {
		return fmt.Errorf("cannot write PID file: %w", err)
	}
	defer func() { _ = daemon.Cleanup() }()

	srv, err := server.NewServer(s.Port, "", false)
	if err != nil {
		return fmt.Errorf("cannot create server: %w", err)
	}

	return srv.Start(ctx)
}

// RunStop implements the `md-view stop` command.
func RunStop(_ context.Context) error {
	return daemon.Stop()
}

// RunStatus implements the `md-view status` command — prints daemon status to stdout.
func RunStatus() error {
	status, err := daemon.GetStatus()
	if err != nil {
		fmt.Printf("md-view daemon: not running (error: %v)\n", err)
		return nil
	}
	if !status.Running {
		fmt.Println("md-view daemon: not running")
		return nil
	}
	fmt.Printf("md-view daemon: running (PID %d, port %d)\n", status.PID, status.Port)
	if !status.StartTime.IsZero() {
		fmt.Printf("  uptime: %s\n", time.Since(status.StartTime).Round(time.Second))
	}
	return nil
}

// ensureDaemonRunning checks if the daemon is alive, and starts it if not.
func ensureDaemonRunning(port int) error {
	status, err := daemon.GetStatus()
	if err != nil {
		return err
	}
	if status.Running {
		return nil
	}

	// Start daemon in background
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	args := []string{"serve"}
	if port > 0 {
		args = append(args, fmt.Sprintf("--port=%d", port))
	}

	cmd := exec.Command(execPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	// Detach from parent
	cmd.SysProcAttr = getSysProcAttr()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cannot start daemon: %w", err)
	}

	log.Info().Int("pid", cmd.Process.Pid).Msg("Started md-view daemon")

	// Wait for PID file to appear
	if err := waitForPIDFile(5 * time.Second); err != nil {
		return fmt.Errorf("daemon did not start: %w", err)
	}

	return nil
}

func waitForSocket(socketPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for socket at %s", socketPath)
}

func waitForPIDFile(timeout time.Duration) error {
	pidPath, err := func() (string, error) {
		dir, err := daemon.StateDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "md-view.pid"), nil
	}()
	if err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(pidPath); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for PID file")
}
