# Quickstart: Request Monitoring Feature

## Build & Test

### Prerequisites
- Go 1.21+
- Running GoZen daemon (dev or prod)
- Web browser for UI testing

### Build
```bash
# Build the project
go build ./...

# Run tests
go test ./...

# Run specific package tests
go test ./internal/proxy -v
go test ./internal/web -v
```

### Dev Environment Setup
```bash
# Start dev daemon (uses ports 29840/29841)
./scripts/dev.sh

# Check dev daemon status
./scripts/dev.sh status

# Restart after code changes
./scripts/dev.sh restart

# Stop dev daemon
./scripts/dev.sh stop
```

## Manual Verification

### Part 1: Verify Tag Removal

**Objective**: Confirm provider tag injection code is completely removed and responses are unmodified.

**Steps:**

1. **Start dev daemon**
   ```bash
   ./scripts/dev.sh restart
   ```

2. **Send test request with thinking block**
   ```bash
   curl -X POST http://localhost:29841/v1/messages \
     -H "Content-Type: application/json" \
     -H "anthropic-version: 2023-06-01" \
     -d '{
       "model": "claude-sonnet-4-20250514",
       "thinking": {"type": "enabled"},
       "max_tokens": 100,
       "messages": [{"role": "user", "content": "Hello"}]
     }'
   ```

3. **Verify response**
   - Response should NOT contain `[provider: xxx, model: xxx]` tag
   - Response should be valid JSON
   - No Bedrock validation errors

4. **Check config backward compatibility**
   ```bash
   # Add deprecated field to dev config
   echo '{"show_provider_tag": true}' > ~/.zen-dev/test-config.json

   # Verify daemon loads without error
   ./scripts/dev.sh restart

   # Check logs for errors
   tail -f ~/.zen-dev/logs/daemon.log
   ```

**Expected Results:**
- ✅ No provider tags in responses
- ✅ Thinking blocks work correctly
- ✅ Deprecated config field is ignored (no errors)

---

### Part 2: Verify Request Monitoring

**Objective**: Confirm request metadata is captured and accessible via API.

**Steps:**

1. **Send multiple test requests**
   ```bash
   # Request 1: Success
   curl -X POST http://localhost:29841/v1/messages \
     -H "Content-Type: application/json" \
     -H "anthropic-version: 2023-06-01" \
     -d '{
       "model": "claude-sonnet-4-20250514",
       "max_tokens": 50,
       "messages": [{"role": "user", "content": "Test 1"}]
     }'

   # Request 2: Different model
   curl -X POST http://localhost:29841/v1/messages \
     -H "Content-Type: application/json" \
     -H "anthropic-version: 2023-06-01" \
     -d '{
       "model": "claude-haiku-3-5-20241022",
       "max_tokens": 50,
       "messages": [{"role": "user", "content": "Test 2"}]
     }'
   ```

2. **Query monitoring API**
   ```bash
   # Get recent requests
   curl http://localhost:29840/api/v1/monitoring/requests?limit=10

   # Filter by provider
   curl http://localhost:29840/api/v1/monitoring/requests?provider=aws-bedrock

   # Filter by status code
   curl http://localhost:29840/api/v1/monitoring/requests?status_min=200&status_max=299
   ```

3. **Verify response data**
   ```bash
   # Pretty print JSON
   curl -s http://localhost:29840/api/v1/monitoring/requests?limit=5 | jq .
   ```

**Expected Results:**
- ✅ API returns request records in JSON format
- ✅ Each record contains: timestamp, provider, model, tokens, cost, duration
- ✅ Records are in reverse chronological order (newest first)
- ✅ Filters work correctly

---

### Part 3: Verify Web UI

**Objective**: Confirm monitoring page displays requests and auto-refreshes.

**Steps:**

1. **Open Web UI**
   ```bash
   open http://localhost:29840
   ```

2. **Navigate to Requests page**
   - Click "Requests" in navigation
   - Or go directly to: http://localhost:29840/requests

3. **Verify table display**
   - Table shows recent requests
   - Columns: Timestamp, Provider, Model, Status, Duration, Tokens, Cost
   - Data matches API response

4. **Test auto-refresh**
   - Send new request via curl (see Part 2)
   - Wait 5-10 seconds
   - Verify new request appears in table

5. **Test filters (if implemented)**
   - Filter by provider
   - Filter by status code
   - Verify table updates

6. **Test detail view (if implemented)**
   - Click on a request row
   - Verify detail modal/panel opens
   - Check failover history, token breakdown, timing

**Expected Results:**
- ✅ Requests page loads without errors
- ✅ Table displays request data correctly
- ✅ Auto-refresh works (new requests appear)
- ✅ Filters work (if implemented)
- ✅ Detail view shows complete metadata (if implemented)

---

## Test Scenarios

### Scenario 1: Single Provider Success
```bash
# Configure single provider
zen config add-provider test-provider https://api.anthropic.com YOUR_KEY

# Send request
curl -X POST http://localhost:29841/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model": "claude-sonnet-4-20250514", "max_tokens": 50, "messages": [{"role": "user", "content": "Hi"}]}'

# Check monitoring
curl http://localhost:29840/api/v1/monitoring/requests?limit=1 | jq '.requests[0]'
```

**Expected:**
- Provider: test-provider
- Status: 200
- FailoverChain: empty or single entry
- Tokens and cost populated

### Scenario 2: Failover (Primary Fails, Backup Succeeds)
```bash
# Configure two providers (first one invalid)
zen config add-provider bad-provider https://invalid.example.com fake-key
zen config add-provider good-provider https://api.anthropic.com YOUR_KEY

# Send request
curl -X POST http://localhost:29841/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model": "claude-sonnet-4-20250514", "max_tokens": 50, "messages": [{"role": "user", "content": "Hi"}]}'

# Check monitoring
curl http://localhost:29840/api/v1/monitoring/requests?limit=1 | jq '.requests[0].failover_chain'
```

**Expected:**
- Provider: good-provider
- Status: 200
- FailoverChain: 2 entries (bad-provider failed, good-provider succeeded)
- First attempt shows error message

### Scenario 3: All Providers Fail
```bash
# Configure only invalid providers
zen config add-provider bad1 https://invalid1.example.com fake-key
zen config add-provider bad2 https://invalid2.example.com fake-key

# Send request
curl -X POST http://localhost:29841/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model": "claude-sonnet-4-20250514", "max_tokens": 50, "messages": [{"role": "user", "content": "Hi"}]}'

# Check monitoring
curl http://localhost:29840/api/v1/monitoring/requests?limit=1 | jq '.requests[0]'
```

**Expected:**
- Provider: empty or last attempted
- Status: 502 (Bad Gateway)
- FailoverChain: 2 entries (both failed)
- ErrorMessage: "all providers failed"

### Scenario 4: Streaming Request
```bash
# Send streaming request
curl -X POST http://localhost:29841/v1/messages \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model": "claude-sonnet-4-20250514", "max_tokens": 50, "stream": true, "messages": [{"role": "user", "content": "Hi"}]}'

# Check monitoring
curl http://localhost:29840/api/v1/monitoring/requests?limit=1 | jq '.requests[0]'
```

**Expected:**
- Provider: (successful provider)
- Status: 200
- InputTokens: 0 or N/A (streaming doesn't report tokens in response)
- OutputTokens: 0 or N/A
- Cost: 0.0 (can't calculate without tokens)

---

## Debugging

### Check Daemon Logs
```bash
# Dev daemon logs
tail -f ~/.zen-dev/logs/daemon.log

# Production daemon logs
tail -f ~/.zen/logs/daemon.log
```

### Check SQLite Database (Optional Persistence)

If SQLite persistence is enabled, you can query the database directly:

```bash
# Dev database
sqlite3 ~/.zen-dev/logs/zen.db

# Query request records
sqlite3 ~/.zen-dev/logs/zen.db "SELECT * FROM request_records ORDER BY timestamp DESC LIMIT 10;"

# Check table schema
sqlite3 ~/.zen-dev/logs/zen.db ".schema request_records"
```

### Check In-Memory Buffer
```bash
# Use Web API to inspect buffer
curl http://localhost:29840/api/v1/monitoring/requests?limit=1000 | jq '.total'
```

### Common Issues

**Issue: API returns empty array**
- Cause: No requests have been proxied yet
- Solution: Send test requests via curl

**Issue: Tokens/cost are 0**
- Cause: Streaming request or response doesn't include usage
- Solution: Expected behavior for streaming; non-streaming should have tokens

**Issue: Web UI doesn't auto-refresh**
- Cause: JavaScript error or polling not started
- Solution: Check browser console for errors

**Issue: Failover chain is empty**
- Cause: First provider succeeded immediately
- Solution: Expected behavior; failover only recorded when providers fail

---

## Performance Verification

### Memory Usage
```bash
# Check daemon memory usage
ps aux | grep zen

# Should be <100MB for daemon with 1000 request buffer
```

### API Response Time
```bash
# Measure API latency
time curl -s http://localhost:29840/api/v1/monitoring/requests?limit=50 > /dev/null

# Should be <100ms for 50 records
```

### Request Overhead
```bash
# Compare request latency with/without monitoring
# (Monitoring should add <5ms overhead)

# Send 10 requests and measure average latency
for i in {1..10}; do
  time curl -s -X POST http://localhost:29841/v1/messages \
    -H "Content-Type: application/json" \
    -H "anthropic-version: 2023-06-01" \
    -d '{"model": "claude-sonnet-4-20250514", "max_tokens": 10, "messages": [{"role": "user", "content": "Hi"}]}' > /dev/null
done
```

---

## Cleanup

### Reset Dev Environment
```bash
# Stop dev daemon
./scripts/dev.sh stop

# Clear dev database
rm ~/.zen-dev/logs/zen.db

# Clear dev config
rm ~/.zen-dev/zen.json

# Restart
./scripts/dev.sh
```

### Clear Request History
```bash
# Via SQLite
sqlite3 ~/.zen-dev/logs/zen.db "DELETE FROM request_records;"

# Or delete entire database
rm ~/.zen-dev/logs/zen.db
```
