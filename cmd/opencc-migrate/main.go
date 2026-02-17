// opencc-migrate is a standalone migration tool published as the "opencc" v2.0.0 binary.
// When a user runs `opencc` (or `opencc <anything>`), it performs a one-time migration
// to zen and removes itself. Zero dependency on the main zen codebase.
package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("╭──────────────────────────────────────────────────────────╮")
	fmt.Println("│  opencc has been renamed to GoZen (zen)                  │")
	fmt.Println("│  Starting migration...                                   │")
	fmt.Println("╰──────────────────────────────────────────────────────────╯")
	fmt.Println()

	home, err := os.UserHomeDir()
	if err != nil {
		fatalf("cannot determine home directory: %v", err)
	}

	oldDir := filepath.Join(home, ".opencc")
	newDir := filepath.Join(home, ".zen")

	// Step 1: Migrate config (JSON and logs only; .db files after daemon stop)
	migrateConfig(oldDir, newDir)

	// Step 2: Download and install zen
	downloadZen()

	// Step 3: Remove opencc web system service (before stopping daemon,
	// otherwise the service may auto-restart the daemon)
	removed := removeService()

	// Step 4: Stop opencc web daemon
	stopDaemon(oldDir, newDir)

	// Now safe to copy .db files (daemon is stopped, no active writes)
	migrateDBFiles(oldDir, newDir)

	// Re-enable service under new name if old one existed
	if removed {
		reEnableService()
	}

	// Step 5: Remove self
	removeSelf()
	// Step 6: Final message
	fmt.Println()
	fmt.Println("╭──────────────────────────────────────────────────────────╮")
	fmt.Println("│  Migration complete!                                     │")
	fmt.Println("╰──────────────────────────────────────────────────────────╯")
	fmt.Println()
	fmt.Println("Quick Start:")
	fmt.Println("  zen                          Start with default profile")
	fmt.Println("  zen use <provider>           Use a specific provider")
	fmt.Println("  zen web                      Open web UI")
	fmt.Println("  zen config add provider      Add a new provider")
	fmt.Println("  zen config add profile       Add a new profile")
	fmt.Println("  zen daemon enable            Auto-start daemon on login")
	fmt.Println("  zen daemon disable           Remove auto-start")
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func step(n int, total int, msg string) {
	fmt.Printf("[%d/%d] %s", n, total, msg)
}

// --- Step 1: Migrate config ---

func migrateConfig(oldDir, newDir string) {
	step(1, 5, "Migrating config ~/.opencc/ → ~/.zen/ ... ")

	// Skip if new config already exists
	if _, err := os.Stat(filepath.Join(newDir, "zen.json")); err == nil {
		fmt.Println("already exists, skipping")
		return
	}

	// Skip if old dir doesn't exist
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		fmt.Println("~/.opencc/ not found, skipping")
		return
	}

	if err := os.MkdirAll(newDir, 0755); err != nil {
		fmt.Printf("warning: cannot create ~/.zen/: %v\n", err)
		return
	}

	// Copy opencc.json → zen.json
	if err := copyFile(filepath.Join(oldDir, "opencc.json"), filepath.Join(newDir, "zen.json")); err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("warning: %v\n", err)
		}
		// Continue to copy auxiliary files even if opencc.json is missing
	}

	// Copy log files only (skip .pid — stale; skip .db — daemon may be writing)
	entries, _ := os.ReadDir(oldDir)
	for _, e := range entries {
		name := e.Name()
		if name == "opencc.json" {
			continue
		}
		if filepath.Ext(name) == ".log" {
			copyFile(filepath.Join(oldDir, name), filepath.Join(newDir, name))
		}
	}

	fmt.Println("done")
}

// migrateDBFiles copies .db files after the daemon has been stopped.
func migrateDBFiles(oldDir, newDir string) {
	entries, _ := os.ReadDir(oldDir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".db" {
			src := filepath.Join(oldDir, e.Name())
			dst := filepath.Join(newDir, e.Name())
			if _, err := os.Stat(dst); err == nil {
				continue // already exists
			}
			copyFile(src, dst)
		}
	}
}

// --- Step 2: Download and install zen ---

func downloadZen() {
	step(2, 5, "Downloading zen ... ")

	zenPath := "/usr/local/bin/zen"
	if runtime.GOOS == "windows" {
		zenPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "zen", "zen.exe")
	}

	// Skip if zen already exists and is executable
	if info, err := os.Stat(zenPath); err == nil && info.Mode()&0111 != 0 {
		fmt.Println("already installed, skipping")
		return
	}

	// Fetch latest release tag
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/dopejs/gozen/releases/latest")
	if err != nil {
		fatalf("failed to fetch release info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fatalf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fatalf("failed to parse release info: %v", err)
	}

	tag := release.TagName
	if tag == "" {
		fatalf("no release tag found")
	}
	assetName := fmt.Sprintf("zen-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://github.com/dopejs/gozen/releases/download/%s/%s", tag, assetName)

	// Download tarball
	dlResp, err := client.Get(url)
	if err != nil {
		fatalf("download failed: %v", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != 200 {
		fatalf("download failed: HTTP %d", dlResp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "zen-migrate-*.tar.gz")
	if err != nil {
		fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, dlResp.Body); err != nil {
		tmpFile.Close()
		fatalf("download failed: %v", err)
	}
	tmpFile.Close()

	// Extract zen binary from tarball
	binPath, err := extractZenFromTarGz(tmpFile.Name())
	if err != nil {
		fatalf("extraction failed: %v", err)
	}
	defer os.Remove(binPath)

	os.Chmod(binPath, 0755)

	// Install to target path
	if err := installBinary(binPath, zenPath); err != nil {
		fatalf("install failed: %v", err)
	}

	// macOS codesign
	if runtime.GOOS == "darwin" {
		exec.Command("codesign", "--force", "--sign", "-", zenPath).Run()
	}

	fmt.Println("done")
}

// --- Step 3: Remove service (platform-specific, before stopping daemon) ---

func removeService() bool {
	step(3, 5, "Removing legacy system services ... ")

	removed := false

	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		// Remove all legacy plist variants: opencc-web, zen-web
		for _, label := range []string{"com.dopejs.opencc-web", "com.dopejs.zen-web"} {
			plist := filepath.Join(home, "Library", "LaunchAgents", label+".plist")
			if _, err := os.Stat(plist); err == nil {
				exec.Command("launchctl", "unload", plist).Run()
				os.Remove(plist)
				removed = true
			}
		}

	case "linux":
		home, _ := os.UserHomeDir()
		// Remove all legacy unit variants: opencc-web, zen-web
		for _, name := range []string{"opencc-web.service", "zen-web.service"} {
			unit := filepath.Join(home, ".config", "systemd", "user", name)
			if _, err := os.Stat(unit); err == nil {
				exec.Command("systemctl", "--user", "stop", name).Run()
				exec.Command("systemctl", "--user", "disable", name).Run()
				os.Remove(unit)
				removed = true
			}
		}
		if removed {
			exec.Command("systemctl", "--user", "daemon-reload").Run()
		}

	case "windows":
		for _, tn := range []string{"opencc-web", "zen-web"} {
			if err := exec.Command("schtasks", "/query", "/tn", tn).Run(); err == nil {
				exec.Command("schtasks", "/delete", "/tn", tn, "/f").Run()
				removed = true
			}
		}
	}

	if removed {
		fmt.Println("done")
	} else {
		fmt.Println("not found")
	}
	return removed
}

// --- Step 4: Stop daemon ---

func stopDaemon(oldDir, newDir string) {
	step(4, 5, "Stopping legacy daemons ... ")

	stopped := false
	for _, dir := range []string{oldDir, newDir} {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			name := e.Name()
			// Match legacy web*.pid files AND zend.pid
			isLegacyWeb := strings.HasPrefix(name, "web") && strings.HasSuffix(name, ".pid")
			isZend := name == "zend.pid"
			if !isLegacyWeb && !isZend {
				continue
			}
			pidFile := filepath.Join(dir, name)
			data, err := os.ReadFile(pidFile)
			if err != nil {
				continue
			}
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				continue
			}
			proc, err := os.FindProcess(pid)
			if err != nil {
				continue
			}
			if err := proc.Signal(os.Interrupt); err == nil {
				stopped = true
			}
			os.Remove(pidFile)
		}
	}

	if stopped {
		fmt.Println("done")
	} else {
		fmt.Println("not running")
	}
}

func reEnableService() {
	zenPath := "/usr/local/bin/zen"
	if runtime.GOOS == "windows" {
		zenPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "zen", "zen.exe")
	}
	if _, err := os.Stat(zenPath); err == nil {
		exec.Command(zenPath, "daemon", "enable").Run()
	}
}

// --- Step 5: Remove self ---

func removeSelf() {
	step(5, 5, "Removing opencc binary ... ")

	exe, err := os.Executable()
	if err != nil {
		fmt.Println("skipped (cannot determine path)")
		return
	}

	// Do NOT resolve symlinks — if opencc is a symlink to zen,
	// we want to remove the symlink, not the zen binary.

	if runtime.GOOS == "windows" {
		fmt.Printf("please delete %s manually\n", exe)
		return
	}

	if err := os.Remove(exe); err != nil {
		fmt.Println()
		if sudoErr := exec.Command("sudo", "rm", "-f", exe).Run(); sudoErr != nil {
			fmt.Printf("please delete %s manually\n", exe)
			return
		}
	}
	fmt.Println("done")
}

// --- Helpers ---

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func extractZenFromTarGz(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if header.Typeflag == tar.TypeReg && (header.Name == "zen" || strings.HasSuffix(header.Name, "/zen")) {
			tmp, err := os.CreateTemp("", "zen-extracted-*")
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(tmp, tr); err != nil {
				tmp.Close()
				os.Remove(tmp.Name())
				return "", err
			}
			tmp.Close()
			return tmp.Name(), nil
		}
	}
	return "", fmt.Errorf("zen binary not found in archive")
}

func installBinary(src, dst string) error {
	// Try direct copy first
	if err := copyFile(src, dst); err == nil {
		return os.Chmod(dst, 0755)
	}
	// Fall back to sudo cp + chmod
	if err := exec.Command("sudo", "cp", src, dst).Run(); err != nil {
		return err
	}
	return exec.Command("sudo", "chmod", "+x", dst).Run()
}
