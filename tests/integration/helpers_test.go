package integration

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/dopejs/gozen/internal/proxy"
)

// mustParseURL parses a URL or panics
func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("invalid URL: %s", s))
	}
	return u
}

// testLogger returns a logger for tests
func testLogger() *log.Logger {
	return log.New(os.Stderr, "[test] ", log.LstdFlags)
}

// createTestProvider creates a test provider with the given base URL
func createTestProvider(baseURL string) *proxy.Provider {
	return &proxy.Provider{
		Name:    "test-provider",
		BaseURL: mustParseURL(baseURL),
		Token:   "test-token",
		Model:   "claude-sonnet-4-5",
		Healthy: true,
	}
}
