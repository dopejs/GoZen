package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/dopejs/gozen/internal/config"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Open the web configuration interface",
	Long:  "Ensure the zend daemon is running and open the web UI in the default browser.",
	RunE:  runWeb,
}

func runWeb(cmd *cobra.Command, args []string) error {
	// Ensure daemon is running (auto-start if needed)
	if err := ensureDaemonRunning(); err != nil {
		return fmt.Errorf("failed to start zend: %w", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d", config.GetWebPort())
	fmt.Printf("Opening %s\n", url)
	openBrowser(url)
	return nil
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", url).Start()
	case "linux":
		_ = exec.Command("xdg-open", url).Start()
	}
}
