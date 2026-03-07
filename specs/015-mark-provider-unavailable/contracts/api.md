# API Contracts: Provider Unavailability

**Feature**: 015-mark-provider-unavailable
**Date**: 2026-03-07

## Web API Endpoints

### POST /api/v1/providers/{name}/disable

Mark a provider as unavailable.

**Request Body**:
```json
{
  "type": "today" | "month" | "permanent"
}
```

**Success Response** (200 OK):
```json
{
  "provider": "openai-backup",
  "disabled": true,
  "type": "today",
  "created_at": "2026-03-07T14:30:00+08:00",
  "expires_at": "2026-03-07T23:59:59+08:00"
}
```

**Error Responses**:
- 400: `{"error": "invalid type: must be 'today', 'month', or 'permanent'"}`
- 404: `{"error": "provider 'xxx' not found"}`

---

### POST /api/v1/providers/{name}/enable

Clear the unavailability marking for a provider.

**Request Body**: None (empty body)

**Success Response** (200 OK):
```json
{
  "provider": "openai-backup",
  "disabled": false
}
```

**Error Responses**:
- 404: `{"error": "provider 'xxx' not found"}`

---

### GET /api/v1/providers/disabled

List all currently disabled providers (active markings only, excludes expired).

**Response** (200 OK):
```json
{
  "disabled_providers": [
    {
      "provider": "openai-backup",
      "type": "today",
      "created_at": "2026-03-07T14:30:00+08:00",
      "expires_at": "2026-03-07T23:59:59+08:00"
    }
  ]
}
```

---

### GET /api/v1/providers (existing, extended)

The existing provider list response is extended with an optional `disabled` field.

**Extended Response**:
```json
[
  {
    "name": "anthropic-main",
    "type": "anthropic",
    "base_url": "https://api.anthropic.com",
    "auth_token": "sk-an...abcd"
  },
  {
    "name": "openai-backup",
    "type": "openai",
    "base_url": "https://api.openai.com",
    "auth_token": "sk-op...efgh",
    "disabled": {
      "type": "today",
      "created_at": "2026-03-07T14:30:00+08:00",
      "expires_at": "2026-03-07T23:59:59+08:00"
    }
  }
]
```

---

### Proxy Error Response (new)

When all providers are disabled, the proxy returns:

**Response** (503 Service Unavailable):
```json
{
  "error": {
    "type": "all_providers_unavailable",
    "message": "All providers are manually marked as unavailable. Please re-enable a provider via Web UI or 'zen enable <provider>'.",
    "disabled_providers": ["anthropic-main", "openai-backup"]
  }
}
```

## CLI Commands

### zen disable

```
zen disable <provider> [--today|--month|--permanent]

Flags:
  --today       Mark unavailable for today only (default)
  --month       Mark unavailable for this calendar month
  --permanent   Mark unavailable permanently until manually cleared

Examples:
  zen disable openai-backup              # unavailable today (default)
  zen disable openai-backup --month      # unavailable this month
  zen disable openai-backup --permanent  # unavailable until cleared
```

**Output**:
```
Provider "openai-backup" marked as unavailable (today, expires 2026-03-07 23:59:59)
```

**Error Output**:
```
Error: provider "xxx" not found
Available providers: anthropic-main, openai-backup
```

### zen enable

```
zen enable <provider>

Examples:
  zen enable openai-backup
```

**Output**:
```
Provider "openai-backup" is now available
```

### zen disable --list

```
zen disable --list

Examples:
  zen disable --list
```

**Output**:
```
Disabled providers:
  openai-backup    today       expires 2026-03-07 23:59:59
  anthropic-alt    permanent   no expiration
```

**Output (none disabled)**:
```
No providers are currently disabled.
```
