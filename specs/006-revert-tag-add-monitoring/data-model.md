# Data Model: Request Monitoring

## Request Record

Represents a single proxied API request with complete metadata for monitoring and debugging.

### Fields

| Field | Type | Description | Source |
|-------|------|-------------|--------|
| ID | string | Unique identifier (UUID or timestamp-based) | Generated on capture |
| Timestamp | time.Time | When the request was received | `ServeHTTP()` entry |
| SessionID | string | Client session identifier | Request header `X-Zen-Session` |
| ClientType | string | Client application type (e.g., "claude-code") | Request header `X-Zen-Client` |
| Provider | string | Provider that successfully handled the request | `tryProviders()` success |
| Model | string | Model used for the request | Request body `model` field |
| RequestFormat | string | Request format (anthropic/openai) | Request header `X-Zen-Request-Format` |
| StatusCode | int | HTTP status code of final response | Response status |
| Duration | time.Duration | Total request duration (ms) | End time - start time |
| InputTokens | int | Input tokens consumed | Response body `usage.input_tokens` |
| OutputTokens | int | Output tokens generated | Response body `usage.output_tokens` |
| Cost | float64 | Estimated cost in USD | `UsageTracker.CalculateCost()` |
| RequestSize | int | Request body size in bytes | `len(bodyBytes)` |
| FailoverChain | []ProviderAttempt | List of provider attempts before success | `tryProviders()` loop |
| ErrorMessage | string | Error message if request failed | Final error from all providers |

### Validation Rules

- ID must be unique and non-empty
- Timestamp must be valid and not in the future
- Provider must be non-empty for successful requests
- StatusCode must be in range 100-599
- Duration must be non-negative
- InputTokens and OutputTokens must be non-negative (0 for streaming or unavailable)
- Cost must be non-negative
- RequestSize must be non-negative
- FailoverChain can be empty (single provider success) or contain 1+ attempts

### State Transitions

```
Request Received → Capturing Metadata
                ↓
         Trying Providers (loop)
                ↓
    ┌─────────────────────┐
    │ Provider Attempt    │
    │ - Record timing     │
    │ - Record result     │
    └─────────────────────┘
                ↓
         Success or Failure
                ↓
         Extract Tokens/Cost
                ↓
         Store in Buffer
                ↓
         Persist to SQLite (async)
```

## Provider Attempt

Represents a single attempt to forward a request to a provider (part of failover chain).

### Fields

| Field | Type | Description |
|-------|------|-------------|
| Provider | string | Provider name |
| StatusCode | int | HTTP status code (0 if connection error) |
| ErrorMessage | string | Error message if failed |
| Duration | time.Duration | Time spent on this attempt (ms) |
| Skipped | bool | Whether provider was skipped (unhealthy) |
| SkipReason | string | Reason for skipping (e.g., "backoff 5m") |

### Validation Rules

- Provider must be non-empty
- StatusCode must be 0 or in range 100-599
- Duration must be non-negative
- If Skipped is true, SkipReason must be non-empty

## Request Monitor (In-Memory Buffer)

Thread-safe ring buffer for storing recent request records.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| mu | sync.RWMutex | Protects concurrent access |
| records | []RequestRecord | Slice of request records |
| maxRecords | int | Maximum buffer size (default: 1000) |

### Operations

**Add(record RequestRecord)**
- Acquires write lock
- If buffer is full, evicts oldest 20% (keeps last 80%)
- Appends new record
- Releases lock

**GetRecent(limit int, filter RequestFilter) []RequestRecord**
- Acquires read lock
- Filters records by criteria (provider, session, status, time range)
- Returns newest records first (reverse chronological)
- Limits to specified count
- Releases lock

**Clear()**
- Acquires write lock
- Empties the buffer
- Releases lock

## Request Filter

Criteria for filtering request records.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| Provider | string | Filter by provider name (empty = all) |
| SessionID | string | Filter by session ID (empty = all) |
| MinStatus | int | Minimum status code (0 = no filter) |
| MaxStatus | int | Maximum status code (0 = no filter) |
| StartTime | time.Time | Start of time range (zero = no filter) |
| EndTime | time.Time | End of time range (zero = no filter) |
| Model | string | Filter by model name (empty = all) |

## SQLite Schema

### request_records Table

```sql
CREATE TABLE IF NOT EXISTS request_records (
    id TEXT PRIMARY KEY,
    timestamp INTEGER NOT NULL,
    session_id TEXT,
    client_type TEXT,
    provider TEXT NOT NULL,
    model TEXT,
    request_format TEXT,
    status_code INTEGER NOT NULL,
    duration_ms INTEGER NOT NULL,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    cost_usd REAL DEFAULT 0.0,
    request_size INTEGER DEFAULT 0,
    failover_chain TEXT,  -- JSON array of ProviderAttempt
    error_message TEXT,
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_request_records_timestamp ON request_records(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_request_records_session ON request_records(session_id);
CREATE INDEX IF NOT EXISTS idx_request_records_provider ON request_records(provider);
CREATE INDEX IF NOT EXISTS idx_request_records_status ON request_records(status_code);
```

### Relationships

- **request_records.session_id** → Correlates with `usage` table and `logs` table
- **request_records.provider** → Correlates with `provider_metrics` table
- No foreign keys (loose coupling for flexibility)

## API Response Format

### GET /api/v1/monitoring/requests

**Request Query Parameters:**
```
?limit=50              # Max records to return (default: 50, max: 1000)
&provider=aws-bedrock  # Filter by provider
&session=abc123        # Filter by session ID
&status_min=400        # Min status code
&status_max=599        # Max status code
&start_time=1234567890 # Unix timestamp
&end_time=1234567890   # Unix timestamp
&model=claude-sonnet-4 # Filter by model
```

**Response:**
```json
{
  "requests": [
    {
      "id": "req_abc123",
      "timestamp": "2026-03-03T12:34:56Z",
      "session_id": "session_xyz",
      "client_type": "claude-code",
      "provider": "aws-bedrock",
      "model": "claude-sonnet-4-20250514",
      "request_format": "anthropic",
      "status_code": 200,
      "duration_ms": 1234,
      "input_tokens": 1000,
      "output_tokens": 500,
      "cost_usd": 0.0105,
      "request_size": 4096,
      "failover_chain": [
        {
          "provider": "openrouter",
          "status_code": 429,
          "error_message": "rate limited",
          "duration_ms": 123,
          "skipped": false
        },
        {
          "provider": "aws-bedrock",
          "status_code": 200,
          "duration_ms": 1111,
          "skipped": false
        }
      ],
      "error_message": null
    }
  ],
  "total": 1,
  "limit": 50
}
```

## Memory Estimates

**Single RequestRecord:**
- Fixed fields: ~200 bytes
- Variable fields (strings): ~100-500 bytes average
- FailoverChain: ~50 bytes per attempt × average 1.5 attempts = 75 bytes
- **Total per record: ~400 bytes**

**Buffer with 1000 records:**
- 1000 × 400 bytes = 400 KB
- Negligible memory overhead

**SQLite storage:**
- ~500 bytes per row (with indexes)
- 10,000 records = ~5 MB
- Acceptable for local storage
