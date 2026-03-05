# Quickstart: Provider & Model Tag in Proxy Responses

**Feature**: 005-provider-model-tag
**Date**: 2026-03-02

## Build & Test

```bash
# Build
go build ./...

# Run all tests
go test ./...

# Run specific package tests
go test ./internal/proxy/ -v
go test ./internal/config/ -v
go test ./internal/web/ -v

# Check coverage for affected packages
go test -cover ./internal/proxy/
go test -cover ./internal/config/
go test -cover ./internal/web/

# Web UI tests
cd web && npm run test && cd ..
```

## Manual Verification

### Prerequisites

1. Dev daemon running: `./scripts/dev.sh restart`
2. At least one provider configured with a valid API key

### Test 1: Non-Streaming Tag Injection (Anthropic Format)

```bash
# Enable the tag via API
curl -X PUT http://localhost:29840/api/v1/settings \
  -H 'Content-Type: application/json' \
  -d '{"show_provider_tag": true}'

# Send a non-streaming Anthropic request through the dev proxy
curl http://localhost:29841/v1/messages \
  -H 'Content-Type: application/json' \
  -H 'x-api-key: test' \
  -H 'anthropic-version: 2023-06-01' \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 50,
    "messages": [{"role": "user", "content": "Say hello"}]
  }'

# Expected: First text block starts with [provider: <name>, model: <model>]\n
```

### Test 2: SSE Streaming Tag Injection (Anthropic Format)

```bash
# Send a streaming Anthropic request
curl http://localhost:29841/v1/messages \
  -H 'Content-Type: application/json' \
  -H 'x-api-key: test' \
  -H 'anthropic-version: 2023-06-01' \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 50,
    "stream": true,
    "messages": [{"role": "user", "content": "Say hello"}]
  }'

# Expected: First content_block_delta with text_delta starts with [provider: <name>, model: <model>]\n
```

### Test 3: OpenAI Format (Non-Streaming)

```bash
# Send a non-streaming OpenAI request
curl http://localhost:29841/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer test' \
  -d '{
    "model": "gpt-4",
    "max_tokens": 50,
    "messages": [{"role": "user", "content": "Say hello"}]
  }'

# Expected: choices[0].message.content starts with [provider: <name>, model: <model>]\n
```

### Test 4: Tag Disabled (Default)

```bash
# Disable the tag
curl -X PUT http://localhost:29840/api/v1/settings \
  -H 'Content-Type: application/json' \
  -d '{"show_provider_tag": false}'

# Send same request as Test 1
# Expected: Response is unmodified (no tag)
```

### Test 5: Web UI Toggle

1. Open `http://localhost:29840` in browser
2. Navigate to Settings → General
3. Verify "Show provider info in responses" toggle is visible and OFF by default
4. Toggle ON, save
5. Repeat Test 1 — verify tag appears
6. Toggle OFF, save
7. Repeat Test 1 — verify tag is gone

### Test 6: Error Response (No Tag)

```bash
# Enable tag, then send a request that will fail (invalid model)
curl -X PUT http://localhost:29840/api/v1/settings \
  -H 'Content-Type: application/json' \
  -d '{"show_provider_tag": true}'

# With all providers misconfigured to trigger 502
# Expected: Error response has NO tag injected
```
