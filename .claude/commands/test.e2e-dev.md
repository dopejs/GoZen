---
description: Run E2E tests against the dev daemon with real providers, safely isolated from production.
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Run end-to-end tests against a real dev daemon (`scripts/dev.sh`) with real provider API calls. This tests the full proxy pipeline (request → transform → provider → response → transform) using the dev environment, completely isolated from production.

## Safety Rules (MANDATORY)

1. **NEVER read, write, or modify `~/.zen/zen.json`** (production config)
2. **NEVER interact with the production daemon** (ports 19840/19841 or whatever `~/.zen/zen.json` configures)
3. **All operations use `~/.zen-dev/zen.json`** (dev config, ports 29840/29841)
4. **If `~/.zen-dev/zen.json` lacks providers**, you may READ `~/.zen/zen.json` ONLY to copy provider definitions (base_url, type, model mappings) into the dev config — but NEVER modify the production file
5. **All curl/HTTP requests go to `127.0.0.1:29841`** (dev proxy) or `127.0.0.1:29840` (dev web API), never production ports
6. **Environment variable**: Always use `GOZEN_CONFIG_DIR=~/.zen-dev` when running zen commands directly

## Steps

### 1. Ensure dev daemon is running

```bash
cd /path/to/GoZen && ./scripts/dev.sh status
```

- If **not running**: rebuild and start with `./scripts/dev.sh restart`
- If **running**: check if a code rebuild is needed
  - Run `./scripts/dev.sh restart` if Go source files have been modified since the last build (check `bin/zen-dev` mtime vs latest `.go` file mtime)
  - Otherwise, leave it running

### 2. Ensure dev config has providers

Read `~/.zen-dev/zen.json` and check if it has at least one provider with an `auth_token`.

If **no usable providers** exist:
1. READ `~/.zen/zen.json` (production) to get provider definitions
2. Pick 1-2 providers that are suitable for testing (prefer ones with known-working `base_url` and `auth_token`)
3. Write them into `~/.zen-dev/zen.json`, preserving the dev ports (29840/29841) and any existing dev config
4. Ensure at least one profile references the added providers
5. Reload the dev daemon config: `curl -s -X POST http://127.0.0.1:29840/api/v1/reload`

### 3. Determine test scope

Based on user input, determine what to test:

- **No input / "all"**: Run the full test suite (steps 4-8)
- **Specific feature** (e.g., "streaming", "responses api", "failover"): Run only relevant tests
- **Specific provider** (e.g., "yunyi-codex"): Test only that provider

### 4. Basic connectivity test

For each configured provider, send a minimal non-streaming request:

```bash
curl -s -X POST http://127.0.0.1:29841/default/e2e-test/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -H "x-api-key: test" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 50,
    "messages": [{"role": "user", "content": "Say hi in 3 words"}]
  }'
```

- Check: HTTP 200, valid JSON response with `content[0].text`
- If provider is OpenAI-type, adjust model accordingly

### 5. Streaming test

Send a streaming request and verify SSE format:

```bash
curl -s -N -X POST http://127.0.0.1:29841/default/e2e-test/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -H "x-api-key: test" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 50,
    "stream": true,
    "messages": [{"role": "user", "content": "Say hi in 3 words"}]
  }'
```

- Check: Response contains `event: message_start`, `event: content_block_delta`, `event: message_stop`
- Check: Each SSE line follows `event: ...\ndata: {...}\n\n` format

### 6. Multi-turn with cache_control test

Send a multi-turn conversation with `cache_control` fields (tests Anthropic-specific field handling):

```bash
curl -s -X POST http://127.0.0.1:29841/default/e2e-test/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -H "x-api-key: test" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 50,
    "messages": [
      {"role": "user", "content": [{"type": "text", "text": "Remember: sky is blue", "cache_control": {"type": "ephemeral"}}]},
      {"role": "assistant", "content": [{"type": "text", "text": "OK"}]},
      {"role": "user", "content": "What color is the sky?"}
    ]
  }'
```

- Check: HTTP 200, valid response (cache_control should be handled gracefully whether provider supports it or not)

### 7. Tool use test

Send a request with tool definitions:

```bash
curl -s -X POST http://127.0.0.1:29841/default/e2e-test/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -H "x-api-key: test" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "max_tokens": 200,
    "tools": [{"name": "get_weather", "description": "Get weather", "input_schema": {"type": "object", "properties": {"city": {"type": "string"}}, "required": ["city"]}}],
    "messages": [{"role": "user", "content": "What is the weather in Tokyo?"}]
  }'
```

- Check: HTTP 200, response contains either `tool_use` content block or text response

### 8. Scenario routing and failover tests (if multiple providers configured)

**Scenario routing** (if profiles have routing config):
- Send a request through a profile with scenario routing configured
- Verify the request is routed to the expected provider for that scenario

**Failover** (if profile has 2+ providers):
- This is only testable if one provider is known to be down; skip if all providers are healthy
- Document which providers were tested and their status

### 9. Report results

Produce a summary table:

```
=== E2E Test Results (Dev Daemon) ===
Provider: [name] ([type])
Proxy:    127.0.0.1:29841

| Test                        | Status | Duration | Notes           |
|-----------------------------|--------|----------|-----------------|
| Basic connectivity          | PASS   | 1.2s     |                 |
| Streaming                   | PASS   | 2.1s     |                 |
| Multi-turn + cache_control  | PASS   | 1.5s     |                 |
| Tool use                    | PASS   | 1.8s     |                 |
| Scenario routing            | SKIP   | -        | Not configured  |
| Failover                    | SKIP   | -        | Single provider |

Result: 4/4 PASSED, 2 SKIPPED
```

If any test fails:
1. Show the full curl output (request + response)
2. Check dev daemon logs: `curl -s http://127.0.0.1:29840/api/v1/logs?limit=10`
3. Suggest possible fixes

## Notes

- The dev daemon uses `pnpm run build` (not npm) for frontend builds via `scripts/dev.sh`
- Dev binary is at `bin/zen-dev`, config at `~/.zen-dev/`
- If the user says "test with [provider-name]", create a temporary profile in dev config pointing to that provider, run tests, then clean up
- For OpenAI-type providers (e.g., cctq-codex, yunyi-codex), the proxy handles Chat Completions → Responses API transform automatically; tests should still use Anthropic Messages API format on the client side
- Timeout each curl request at 30 seconds to avoid hanging
- Use unique session IDs per test run to avoid session cache interference (e.g., `e2e-test-{timestamp}`)
