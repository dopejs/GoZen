package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/internal/daemon"
	"github.com/spf13/cobra"
)

var daemonForegroundFlag bool

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the zend daemon",
	Long:  "Manage the GoZen daemon (zend) that hosts the proxy and web UI servers.",
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the zend daemon",
	RunE:  runDaemonStart,
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the zend daemon",
	RunE:  runDaemonStop,
}

var daemonRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the zend daemon",
	RunE:  runDaemonRestart,
}

var daemonStatusCmd2 = &cobra.Command{
	Use:   "status",
	Short: "Show zend daemon status",
	RunE:  runDaemonStatus,
}

var daemonEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Install zend as a system service (start on login)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := daemon.EnableService(); err != nil {
			return err
		}
		fmt.Println("zend installed as system service.")
		return nil
	},
}

var daemonDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Uninstall zend system service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := daemon.DisableService(); err != nil {
			return err
		}
		fmt.Println("zend system service removed.")
		return nil
	},
}

func init() {
	daemonStartCmd.Flags().BoolVar(&daemonForegroundFlag, "foreground", false, "run in foreground (don't daemonize)")
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonRestartCmd)
	daemonCmd.AddCommand(daemonStatusCmd2)
	daemonCmd.AddCommand(daemonEnableCmd)
	daemonCmd.AddCommand(daemonDisableCmd)
}

func runDaemonStart(cmd *cobra.Command, args []string) error {
	// If this is the daemon child process, run in foreground
	if os.Getenv("GOZEN_DAEMON") == "1" || daemonForegroundFlag {
		return runDaemonForeground()
	}

	// Check if already running
	if pid, running := daemon.IsDaemonRunning(); running {
		fmt.Printf("zend is already running (PID %d).\n", pid)
		return nil
	}

	return startDaemonBackground()
}

// runDaemonForeground runs the zend daemon in the foreground.
func runDaemonForeground() error {
	logFile, logger := setupDaemonLogger()
	if logFile != nil {
		defer logFile.Close()
	}

	d := daemon.NewDaemon(Version, logger)

	// Write PID file
	daemon.WriteDaemonPid(os.Getpid())

	// Graceful shutdown on signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		d.Shutdown(ctx)
	}()

	return d.Start()
}

// startDaemonBackground forks a child process to run the daemon.
func startDaemonBackground() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	logPath := daemon.DaemonLogPath()
	logDir := config.ConfigDirPath()
	os.MkdirAll(logDir, 0755)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("cannot open log file: %w", err)
	}
	defer logFile.Close()

	child := exec.Command(exe, "daemon", "start")
	child.Env = append(os.Environ(), "GOZEN_DAEMON=1")
	child.Stdout = logFile
	child.Stderr = logFile
	child.SysProcAttr = daemon.DaemonSysProcAttr()

	if err := child.Start(); err != nil {
		return fmt.Errorf("failed to start zend: %w", err)
	}

	daemon.WriteDaemonPid(child.Process.Pid)

	// Wait for the daemon to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := waitForDaemonReady(ctx); err != nil {
		return fmt.Errorf("zend started but did not become ready: %w", err)
	}

	fmt.Printf("zend started (PID %d) — proxy=:%d web=:%d\n",
		child.Process.Pid, config.GetProxyPort(), config.GetWebPort())
	return nil
}

// waitForDaemonReady polls the daemon status endpoint until ready.
func waitForDaemonReady(ctx context.Context) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/status", config.GetWebPort())
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func runDaemonStop(cmd *cobra.Command, args []string) error {
	pid, running := daemon.IsDaemonRunning()
	if !running {
		fmt.Println("zend is not running.")
		return nil
	}

	// Check for active sessions and warn
	sessionCount := queryActiveSessions()
	if sessionCount > 0 {
		fmt.Printf("Warning: %d active session(s) will be affected.\n", sessionCount)
		fmt.Print("Continue? [y/N] ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			answer := scanner.Text()
			if answer != "y" && answer != "Y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}
	}

	fmt.Printf("Stopping zend (PID %d)...\n", pid)
	if err := daemon.StopDaemonProcess(30 * time.Second); err != nil {
		return err
	}
	fmt.Println("zend stopped.")
	return nil
}

func runDaemonRestart(cmd *cobra.Command, args []string) error {
	if pid, running := daemon.IsDaemonRunning(); running {
		fmt.Printf("Stopping zend (PID %d)...\n", pid)
		if err := daemon.StopDaemonProcess(30 * time.Second); err != nil {
			return fmt.Errorf("failed to stop zend: %w", err)
		}
		// Brief pause to let ports be released
		time.Sleep(300 * time.Millisecond)
	}

	return startDaemonBackground()
}

func runDaemonStatus(cmd *cobra.Command, args []string) error {
	pid, running := daemon.IsDaemonRunning()
	if !running {
		fmt.Println("zend is not running.")
		return nil
	}

	// Try to get detailed status from the API
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/status", config.GetWebPort())
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		// API not reachable, show basic info
		fmt.Printf("zend is running (PID %d) but API is not reachable.\n", pid)
		return nil
	}
	defer resp.Body.Close()

	var status struct {
		Version        string `json:"version"`
		Uptime         string `json:"uptime"`
		ProxyPort      int    `json:"proxy_port"`
		WebPort        int    `json:"web_port"`
		ActiveSessions int    `json:"active_sessions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		fmt.Printf("zend is running (PID %d).\n", pid)
		return nil
	}

	fmt.Printf("zend is running (PID %d)\n", pid)
	fmt.Printf("  Version:  %s\n", status.Version)
	fmt.Printf("  Uptime:   %s\n", status.Uptime)
	fmt.Printf("  Proxy:    http://127.0.0.1:%d\n", status.ProxyPort)
	fmt.Printf("  Web UI:   http://127.0.0.1:%d\n", status.WebPort)
	fmt.Printf("  Sessions: %d active\n", status.ActiveSessions)
	return nil
}

// queryActiveSessions queries the daemon API for active session count.
// Returns 0 if the API is not reachable.
func queryActiveSessions() int {
	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/daemon/sessions", config.GetWebPort())
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var result struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0
	}
	return result.Count
}

func setupDaemonLogger() (*os.File, *log.Logger) {
	logDir := config.ConfigDirPath()
	os.MkdirAll(logDir, 0755)
	logFile, err := os.OpenFile(daemon.DaemonLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, log.New(os.Stderr, "[zend] ", log.LstdFlags)
	}
	return logFile, log.New(logFile, "[zend] ", log.LstdFlags)
}
