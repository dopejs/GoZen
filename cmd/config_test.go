package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/dopejs/gozen/internal/config"
)

func TestConfigSet(t *testing.T) {
	t.Run("valid proxy_port saves to config", func(t *testing.T) {
		setTestHome(t)

		rootCmd.SetArgs([]string{"config", "set", "proxy_port", "9999"})
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		buf.ReadFrom(r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := config.GetProxyPort()
		if got != 9999 {
			t.Errorf("proxy_port = %d, want 9999", got)
		}
	})

	t.Run("invalid port below 1024 returns error", func(t *testing.T) {
		setTestHome(t)

		rootCmd.SetArgs([]string{"config", "set", "proxy_port", "80"})
		old := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old

		if err == nil {
			t.Error("expected error for port below 1024")
		}
	})

	t.Run("invalid port above 65535 returns error", func(t *testing.T) {
		setTestHome(t)

		rootCmd.SetArgs([]string{"config", "set", "proxy_port", "70000"})
		old := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old

		if err == nil {
			t.Error("expected error for port above 65535")
		}
	})

	t.Run("non-numeric port returns error", func(t *testing.T) {
		setTestHome(t)

		rootCmd.SetArgs([]string{"config", "set", "proxy_port", "abc"})
		old := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old

		if err == nil {
			t.Error("expected error for non-numeric port")
		}
	})

	t.Run("unknown key returns error", func(t *testing.T) {
		setTestHome(t)

		rootCmd.SetArgs([]string{"config", "set", "unknown_key", "value"})
		old := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old

		if err == nil {
			t.Error("expected error for unknown key")
		}
	})

	t.Run("missing args returns error", func(t *testing.T) {
		setTestHome(t)

		rootCmd.SetArgs([]string{"config", "set"})
		old := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		err := rootCmd.Execute()

		w.Close()
		os.Stdout = old

		if err == nil {
			t.Error("expected error for missing args")
		}
	})
}
