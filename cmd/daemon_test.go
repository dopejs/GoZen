package cmd

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// T009: Test that waitForDaemonReady checks both web port (HTTP) and proxy port (TCP).
func TestWaitForDaemonReady(t *testing.T) {
	t.Run("succeeds when both ports ready", func(t *testing.T) {
		home := setTestHome(t)
		_ = home

		// Start a web server on a random port
		webSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer webSrv.Close()

		// Start a TCP listener on a random port (simulates proxy)
		proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to create proxy listener: %v", err)
		}
		defer proxyLn.Close()

		// Extract ports
		webPort := webSrv.Listener.Addr().(*net.TCPAddr).Port
		proxyPort := proxyLn.Addr().(*net.TCPAddr).Port

		config.SetWebPort(webPort)
		config.SetProxyPort(proxyPort)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := waitForDaemonReady(ctx); err != nil {
			t.Errorf("waitForDaemonReady() should succeed when both ports are ready, got: %v", err)
		}
	})

	t.Run("fails when proxy port not ready", func(t *testing.T) {
		home := setTestHome(t)
		_ = home

		// Start only the web server
		webSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer webSrv.Close()

		webPort := webSrv.Listener.Addr().(*net.TCPAddr).Port

		// Use a port that nothing is listening on
		proxyLn, _ := net.Listen("tcp", "127.0.0.1:0")
		proxyPort := proxyLn.Addr().(*net.TCPAddr).Port
		proxyLn.Close() // close immediately so nothing listens

		config.SetWebPort(webPort)
		config.SetProxyPort(proxyPort)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := waitForDaemonReady(ctx)
		if err == nil {
			t.Error("waitForDaemonReady() should fail when proxy port is not ready")
		}
	})

	t.Run("fails when web port not ready", func(t *testing.T) {
		home := setTestHome(t)
		_ = home

		// Start only the proxy listener
		proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to create proxy listener: %v", err)
		}
		defer proxyLn.Close()

		// Use a port that nothing is listening on for web
		webLn, _ := net.Listen("tcp", "127.0.0.1:0")
		webPort := webLn.Addr().(*net.TCPAddr).Port
		webLn.Close()

		proxyPort := proxyLn.Addr().(*net.TCPAddr).Port

		config.SetWebPort(webPort)
		config.SetProxyPort(proxyPort)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = waitForDaemonReady(ctx)
		if err == nil {
			t.Error("waitForDaemonReady() should fail when web port is not ready")
		}
	})
}
