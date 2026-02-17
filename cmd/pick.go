package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dopejs/gozen/internal/config"
	"github.com/dopejs/gozen/tui"
	"github.com/spf13/cobra"
)

var pickCmd = &cobra.Command{
	Use:           "pick [cli args...]",
	Short:         "Select providers interactively and start proxy",
	Long:          "Launch a checkbox picker to select providers for this session, then start the proxy.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runPick,
}

var pickClientFlag string

func init() {
	pickCmd.Flags().StringVarP(&pickClientFlag, "client", "c", "", "client to use (claude, codex, opencode)")
	pickCmd.Flags().String("cli", "", "alias for --client (deprecated)")
	pickCmd.Flags().Lookup("cli").Hidden = true
}

func runPick(cmd *cobra.Command, args []string) error {
	available := config.ProviderNames()
	if len(available) == 0 {
		fmt.Println("No providers configured. Run 'zen config' to set up providers.")
		return nil
	}

	selected, err := tui.RunPick()
	if err != nil {
		if err.Error() == "cancelled" {
			return nil
		}
		return err
	}
	if len(selected) == 0 {
		return nil
	}

	// Resolve client
	client := pickClientFlag
	if client == "" {
		client, _ = cmd.Flags().GetString("cli")
	}
	if client == "" {
		client = config.GetDefaultClient()
	}

	// Ensure daemon is running
	if err := ensureDaemonRunning(); err != nil {
		return fmt.Errorf("failed to start zend: %w", err)
	}

	// Register temporary profile via daemon API
	tempID, err := registerTempProfile(selected)
	if err != nil {
		// Fallback to legacy mode if daemon API fails
		return startLegacyProxy(selected, nil, client, args)
	}

	// Use the temp profile through the daemon
	return startViaDaemon(tempID, client, selected, args, yesFlag)
}

// registerTempProfile registers a temporary profile with the daemon and returns its ID.
func registerTempProfile(providers []string) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"providers": providers,
	})

	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/profiles/temp", config.GetWebPort())
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to register temp profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to register temp profile: status %d", resp.StatusCode)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse temp profile response: %w", err)
	}
	return result.ID, nil
}
