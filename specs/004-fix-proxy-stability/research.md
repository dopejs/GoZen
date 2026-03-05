# Research: Reverse Proxy Stability Fix

**Feature**: 004-fix-proxy-stability
**Date**: 2026-03-02

## Research Questions

### RQ-1: Why does ProfileProxy.buildProviders() not set ProxyURL/Client?

**Finding**: `ProfileProxy.buildProviders()` (`internal/proxy/profile_proxy.go:125-167`) was likely written before per-provider proxy support was added. The `cmd/root.go:buildProviders()` function (lines 452-529) correctly sets `ProxyURL` and creates a per-provider `*http.Client` via `proxy.NewHTTPClientWithProxy()`, but this logic was never replicated in the daemon path.

**Decision**: Copy the ProxyURL and Client creation logic from `cmd/root.go` into `ProfileProxy.buildProviders()`.

**Rationale**: The `cmd/root.go` implementation is the reference and has been working correctly in the `zen use` direct path. The daemon path should be functionally identical.

**Alternatives considered**:
- Extract a shared `buildProviders()` function used by both paths → Rejected: would require refactoring the daemon's dependency injection and `config.DefaultStore()` vs parameter passing. Over-engineering for a 10-line fix.
- Add ProxyURL support as a separate middleware layer → Rejected: unnecessary abstraction, violates YAGNI.

### RQ-2: Is the readiness race a real problem or theoretical?

**Finding**: `waitForDaemonReady()` (`cmd/daemon.go:195-213`) polls only the web port. In `Daemon.Start()` (`internal/daemon/server.go:94-177`), the proxy starts first via `startProxy()` at line 142, which calls `net.Listen()` synchronously (binding the port) then `go d.proxyServer.Serve(ln)` in a goroutine. The web server starts later at line 176 with `d.webServer.Start()`.

Since `net.Listen()` binds the port before `Serve()` starts accepting, and the web server starts after the proxy, by the time `waitForDaemonReady()` gets a 200 from the web port, the proxy port is already bound and serving. The race is **extremely unlikely** in practice.

**Decision**: Add a proxy port TCP check to `waitForDaemonReady()` as a defensive measure. It's a 5-line change and eliminates the theoretical race entirely.

**Rationale**: Cheap to implement, provides defense-in-depth. The proxy port check can be a simple TCP dial (no HTTP needed).

**Alternatives considered**:
- Leave as-is since the race is unlikely → Rejected: we're already touching this code path, and the fix is trivial.
- Add a `/health` endpoint on the proxy port → Rejected: over-engineering for a readiness check.

### RQ-3: Should the 502 error response be JSON instead of plain text?

**Finding**: The current 502 response uses `http.Error()` which sends `text/plain`. Claude Code and other clients may expect JSON error responses from the Anthropic API format. However, the clients already handle non-JSON errors from the proxy (they show "Unable to connect" or similar).

**Decision**: Improve the plain text format with clearer per-provider details but keep it as plain text. Do not change to JSON — the proxy is not the Anthropic API and clients already handle this format.

**Rationale**: Changing to JSON could break existing error handling in clients. The improvement is in content (clearer details) not format.

**Alternatives considered**:
- Switch to JSON error format matching Anthropic API → Rejected: could confuse clients that parse API errors differently from proxy errors.
- Add both JSON and plain text (content negotiation) → Rejected: over-engineering.

### RQ-4: Does this fix require a config version bump?

**Finding**: No. The fix only changes runtime behavior (how providers are constructed from existing config fields). The `ProxyURL` field already exists in the config schema and is already parsed correctly. The daemon simply wasn't using it.

**Decision**: No config version bump needed.

### RQ-5: Model default fallback in ProfileProxy.buildProviders()

**Finding**: `ProfileProxy.buildProviders()` doesn't default empty model variants (`ReasoningModel`, `HaikuModel`, `OpusModel`, `SonnetModel`) like `cmd/root.go` does. This is a secondary issue but should be fixed while we're here.

**Decision**: Add model default fallbacks to match `cmd/root.go` behavior.

**Rationale**: Prevents potential model-not-found errors when providers don't explicitly specify all model variants.
