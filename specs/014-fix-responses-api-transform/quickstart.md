# Quickstart: Testing Responses API Transform

**Feature**: 014-fix-responses-api-transform

## Test Scenarios

### Scenario 1: Responses API Retry (Non-streaming)

1. Start a mock OpenAI server that returns `500 {"error":{"message":"input is required"}}` on `/v1/chat/completions` and `200` with Responses API format on `/v1/responses`
2. Send an Anthropic-format request through the proxy
3. Verify: proxy first tries Chat Completions, gets "input is required", retries with Responses API, succeeds
4. Verify: response is correctly transformed to Anthropic format

### Scenario 2: Responses API Retry (Streaming)

1. Same mock server setup, but `/v1/responses` returns `text/event-stream` with Responses API SSE events
2. Send an Anthropic-format streaming request through the proxy
3. Verify: Responses API SSE events are transformed to Anthropic SSE events
4. Verify: `message_start`, `content_block_delta`, `message_stop` events are emitted

### Scenario 3: Chat Completions Provider (No Retry)

1. Start a mock server that returns `200` on `/v1/chat/completions`
2. Send an Anthropic-format request through the proxy
3. Verify: no retry with Responses API, response is transformed normally

### Scenario 4: Non-"input is required" Error (No Retry)

1. Start a mock server that returns `500 {"error":{"message":"server error"}}` on `/v1/chat/completions`
2. Send an Anthropic-format request through the proxy
3. Verify: NO Responses API retry, normal failover behavior

### Scenario 5: Both Formats Fail

1. Start a mock server that returns `500 "input is required"` on `/v1/chat/completions` AND `401` on `/v1/responses`
2. Send an Anthropic-format request through the proxy
3. Verify: proxy reports the Responses API error (401), not the Chat Completions error

## Manual Production Test

```bash
# With a Responses API provider configured (e.g., cctq-codex)
# Start the dev daemon
./scripts/dev.sh restart

# Send a test request through Claude Code
# Check logs for:
#   [cctq-codex] 500 "input is required"
#   [cctq-codex] retrying with Responses API format
#   [cctq-codex] path: /v1/responses
#   [cctq-codex] success 200
```
