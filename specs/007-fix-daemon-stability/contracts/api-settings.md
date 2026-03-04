# API Contract: Settings (extended)

**Feature**: 007-fix-daemon-stability

## GET /api/v1/settings

### Response (extended)

```json
{
  "default_profile": "string",
  "default_client": "string",
  "web_port": 19840,
  "proxy_port": 19841,
  "profiles": ["string"],
  "clients": ["string"]
}
```

**New field**: `proxy_port` (int) — read-only, reflects the current proxy port from config.

### Notes

- `proxy_port` is NOT accepted in PUT requests. Attempting to set it via the API is silently ignored.
- The Web UI displays `proxy_port` as a disabled/read-only field with the message: "To change the proxy port, use `zen config set proxy_port <port>` in the terminal."

## GET /api/v1/monitoring/requests

### Response (duration field fix)

```json
{
  "requests": [
    {
      "id": "string",
      "timestamp": "2026-03-04T10:00:00Z",
      "duration_ms": 2500,
      "failover_chain": [
        {
          "provider": "string",
          "status_code": 200,
          "duration_ms": 1200
        }
      ]
    }
  ]
}
```

**Change**: `duration_ms` values are now actual milliseconds (e.g., `2500` for 2.5s), not nanoseconds (previously `2500000000`).

**Breaking change for consumers**: Any client that was dividing by 1,000,000 to convert the old nanosecond values to milliseconds will now get incorrect results. Since this is a bug fix and the field was always documented as milliseconds, this is intentional.
