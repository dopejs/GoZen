# CLI Command Contracts

**Date**: 2026-03-05
**Feature**: 010-use-command-enhancements

## Command: `zen --yes` / `zen -y`

**Purpose**: Launch client with all permissions auto-approved

**Syntax**:
```bash
zen --yes [other-flags]
zen -y [other-flags]
```

**Behavior**:
- Prepends `--permission-mode bypassPermissions` for Claude Code
- Prepends `-a never` for Codex
- Prepends appropriate flags for OpenCode (TBD)
- Overrides Web UI auto-permission config if present
- Does NOT add flags if `--` parameters contain permission flags

**Exit Codes**:
- `0`: Success (client launched successfully)
- `1`: Error (client not found, config error, etc.)

**Examples**:
```bash
# Claude Code with bypass permissions
zen --yes
# Equivalent to: claude --permission-mode bypassPermissions [zen-managed-args]

# Codex with never ask
zen --yes --cli codex
# Equivalent to: codex -a never [zen-managed-args]

# With profile selection
zen -y -p work
# Launches work profile with auto-approve permissions
```

---

## Command: `zen -- [client-params]`

**Purpose**: Pass arbitrary parameters to the underlying client

**Syntax**:
```bash
zen -- [client-specific-flags-and-args]
zen [zen-flags] -- [client-specific-flags-and-args]
```

**Behavior**:
- Everything after `--` is passed directly to the client
- `--` parameters have highest priority (override `--yes` and Web UI config)
- If `--` contains permission flags, zen does NOT add its own permission flags
- Order and format of `--` parameters are preserved exactly

**Exit Codes**:
- `0`: Success (client launched successfully)
- `1`: Error (client not found, config error, etc.)
- Client's exit code is propagated if client fails

**Examples**:
```bash
# Pass verbose flag to client
zen -- --verbose

# Override --yes with specific permission mode
zen --yes -- --permission-mode acceptEdits
# Result: Uses acceptEdits (not bypassPermissions) because `--` has priority

# Pass multiple flags
zen -- --verbose --debug --log-level trace

# With profile selection
zen -p work -- --permission-mode default
```

---

## Command: `zen use <profile> -- [client-params]`

**Purpose**: Switch to a profile and pass parameters to client

**Syntax**:
```bash
zen use <profile-name> -- [client-specific-flags-and-args]
```

**Behavior**:
- Switches to the specified profile
- Passes parameters after `--` to the client
- Same priority rules as `zen --` apply

**Exit Codes**:
- `0`: Success
- `1`: Error (profile not found, client error, etc.)

**Examples**:
```bash
# Use work profile with custom flags
zen use work -- --verbose

# Use work profile with specific permission mode
zen use work -- --permission-mode plan
```

---

## Web UI API Contract

### GET `/api/v1/config`

**Purpose**: Retrieve current configuration including auto-permission settings

**Response**:
```json
{
  "version": 12,
  "claude_auto_permission": {
    "enabled": true,
    "mode": "bypassPermissions"
  },
  "codex_auto_permission": {
    "enabled": false,
    "mode": ""
  },
  "opencode_auto_permission": {
    "enabled": false,
    "mode": ""
  },
  ...
}
```

---

### POST `/api/v1/config`

**Purpose**: Update configuration including auto-permission settings

**Request Body**:
```json
{
  "claude_auto_permission": {
    "enabled": true,
    "mode": "acceptEdits"
  }
}
```

**Response**:
- `200 OK`: Configuration updated successfully
- `400 Bad Request`: Invalid permission mode for client type
- `500 Internal Server Error`: Failed to save configuration

**Validation**:
- If `enabled` is `true`, `mode` must be valid for the client type
- If `enabled` is `false`, `mode` is ignored but can be any value

---

## Priority Order Contract

**Guarantee**: The system MUST resolve permission modes in this exact order:

1. **`--` parameters** (highest priority)
   - If permission flags detected in `--`, use them
   - Skip all other sources

2. **`--yes` flag**
   - If present and no `--` permission flags, use most permissive mode
   - Claude Code: `bypassPermissions`
   - Codex: `never`
   - OpenCode: TBD

3. **Web UI auto-permission config**
   - If enabled for current client and no `--yes`, use configured mode
   - Client-specific mode from config

4. **Default behavior** (lowest priority)
   - No permission flags added
   - Client uses its own default behavior

**Invariant**: Only ONE source provides permission flags. No mixing or merging.

---

## Error Handling Contract

### Invalid Permission Mode in Config

**Scenario**: Web UI config has invalid mode for client type

**Behavior**:
- Log warning: "Invalid permission mode '{mode}' for client '{client}', ignoring auto-permission config"
- Fall back to default behavior (no flags added)
- Do NOT block client launch

### Client Not Found

**Scenario**: Client binary not in PATH

**Behavior**:
- Print error: "{client} not found in PATH"
- Exit with code `1`
- Do NOT attempt to launch

### Permission Flag Detection Ambiguity

**Scenario**: `--` contains partial match (e.g., `--permission-something`)

**Behavior**:
- Use exact flag name matching:
  - Claude Code: `--permission-mode`
  - Codex: `-a` or `--ask-for-approval`
- Partial matches do NOT count as permission flags
- zen may add its own flags if no exact match found

---

## Backward Compatibility Contract

### Old Configs (version < 12)

**Guarantee**: Configs without `*_auto_permission` fields MUST work without modification

**Behavior**:
- Missing fields treated as disabled (default behavior)
- No migration prompt or user intervention required
- Seamless upgrade experience

### Old zen Versions Reading New Configs

**Guarantee**: New configs (version 12) MUST NOT break old zen versions

**Behavior**:
- Old versions ignore unknown fields
- Core functionality (profiles, providers) continues to work
- Auto-permission feature simply unavailable in old versions

---

## Summary

**Commands Modified**: 2 (`zen`, `zen use`)
**New Flags**: 1 (`--yes` / `-y` behavior changed)
**New Separator**: 1 (`--` for client parameters)
**API Endpoints Modified**: 2 (GET/POST `/api/v1/config`)
**Priority Levels**: 4 (with strict ordering guarantee)
**Error Handling**: Graceful degradation, no blocking errors for config issues
