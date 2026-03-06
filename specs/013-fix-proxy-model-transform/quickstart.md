# Quickstart: Fix Proxy Model Transform

## Prerequisites

- Go 1.21+
- Dev environment: `~/.zen-dev/zen.json` with `cctq-codex` and `yunyi-codex` providers

## Setup

```bash
# Start dev daemon
./scripts/dev.sh restart

# Verify dev daemon is running
./scripts/dev.sh status
```

## Development Flow

### 1. Run existing tests (verify green baseline)

```bash
go test ./internal/proxy/... -v -count=1
go test ./internal/proxy/transform/... -v -count=1
```

### 2. Write failing tests (Red phase)

Add to `internal/proxy/server_test.go`:
- Path dedup tests in `TestSingleJoiningSlash` (or new test function)
- Model default tests for OpenAI provider type

### 3. Implement fixes (Green phase)

**Bug 1** — `internal/proxy/server.go` (forwardRequest, ~line 465):
```go
// Before singleJoiningSlash, strip overlapping /v1 prefix
basePath := strings.TrimSuffix(p.BaseURL.Path, "/")
if strings.HasSuffix(basePath, "/v1") && strings.HasPrefix(targetPath, "/v1") {
    targetPath = targetPath[3:] // strip "/v1"
}
```

**Bug 2** — `internal/proxy/profile_proxy.go` (buildProviders, ~line 186):
```go
isAnthropic := pc.GetType() == config.ProviderTypeAnthropic

reasoningModel := pc.ReasoningModel
if reasoningModel == "" && isAnthropic {
    reasoningModel = "claude-sonnet-4-5-thinking"
}
// ... same for haiku, opus, sonnet
```

### 4. Verify (Green phase)

```bash
go test ./internal/proxy/... -v -count=1
go test ./internal/proxy/transform/... -v -count=1
```

### 5. E2E verification

```bash
./scripts/dev.sh restart

# Test Claude client → OpenAI provider (was 404)
curl -s -X POST "http://127.0.0.1:29841/codex/test/v1/messages" \
  -H "Content-Type: application/json" \
  -H "x-api-key: test" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"claude-sonnet-4-6","max_tokens":50,"messages":[{"role":"user","content":"Say hello"}]}'
```

### 6. Coverage check

```bash
go test -cover ./internal/proxy/...
# Target: ≥80%
```
