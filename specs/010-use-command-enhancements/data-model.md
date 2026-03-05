# Data Model: zen use Command Enhancements

**Date**: 2026-03-05
**Feature**: 010-use-command-enhancements

## Configuration Schema

### AutoPermissionConfig (New Type)

Represents per-client auto-permission configuration.

**Fields**:
- `enabled` (boolean): Whether auto-permission mode is enabled for this client
- `mode` (string): The permission mode to use (client-specific values)

**Validation Rules**:
- `enabled`: Must be boolean
- `mode`: Must be one of the valid modes for the client type:
  - Claude Code: `"default"`, `"acceptEdits"`, `"bypassPermissions"`, `"dontAsk"`, `"plan"`, or empty string
  - Codex: `"untrusted"`, `"on-request"`, `"never"`, or empty string
  - OpenCode: TBD (pending research)
- If `enabled` is `true`, `mode` must not be empty
- If `enabled` is `false`, `mode` is ignored

**Default Values**:
```go
&AutoPermissionConfig{
    Enabled: false,
    Mode:    "",
}
```

---

### OpenCCConfig (Modified)

The main configuration structure. Adding new fields for auto-permission settings.

**New Fields**:
- `claude_auto_permission` (*AutoPermissionConfig): Auto-permission config for Claude Code
- `codex_auto_permission` (*AutoPermissionConfig): Auto-permission config for Codex
- `opencode_auto_permission` (*AutoPermissionConfig): Auto-permission config for OpenCode

**JSON Representation**:
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
  }
}
```

**Config Version**: Bump from current version to next (e.g., 11 → 12)

---

## State Transitions

### Permission Mode Resolution

The system resolves the final permission mode through a priority chain:

```
1. Check if `--` parameters contain permission flags
   ├─ YES → Use flags from `--`, skip steps 2-4
   └─ NO → Continue to step 2

2. Check if `--yes` flag is present
   ├─ YES → Use most permissive mode for client
   │        (bypassPermissions for Claude, never for Codex)
   └─ NO → Continue to step 3

3. Check if Web UI auto-permission is enabled for current client
   ├─ YES → Use configured mode from Web UI
   └─ NO → Continue to step 4

4. Use default behavior (no permission flags added)
```

**State Diagram**:
```
[User Command] → [Parse Flags] → [Detect `--`?]
                                      ├─ YES → [Use `--` params] → [Launch Client]
                                      └─ NO → [Check `--yes`?]
                                               ├─ YES → [Use bypassPermissions/never] → [Launch Client]
                                               └─ NO → [Check Web UI config?]
                                                        ├─ Enabled → [Use configured mode] → [Launch Client]
                                                        └─ Disabled → [No flags] → [Launch Client]
```

---

## Relationships

### Client Type → Permission Modes

**Claude Code**:
- `default`: Prompt for each tool on first use
- `acceptEdits`: Auto-approve file edits only
- `bypassPermissions`: Skip all prompts
- `dontAsk`: Auto-deny unless pre-approved
- `plan`: Analyze only, no modifications

**Codex**:
- `untrusted`: Ask for untrusted operations
- `on-request`: Ask in interactive runs
- `never`: Never ask (for automation)

**OpenCode**:
- TBD (pending research)

### Config → Client Launch

```
OpenCCConfig
├─ claude_auto_permission → prependAutoApproveArgs("claude", ...)
├─ codex_auto_permission → prependAutoApproveArgs("codex", ...)
└─ opencode_auto_permission → prependAutoApproveArgs("opencode", ...)
```

---

## Validation Rules

### At Config Load Time

1. **Version Check**: If `version < CurrentConfigVersion`, apply migration
2. **Auto-Permission Validation**:
   - If `claude_auto_permission.enabled == true`, validate `mode` is valid Claude Code mode
   - If `codex_auto_permission.enabled == true`, validate `mode` is valid Codex mode
   - If `opencode_auto_permission.enabled == true`, validate `mode` is valid OpenCode mode
3. **Nil Handling**: If any `*_auto_permission` field is nil, initialize with default values

### At Runtime (Command Execution)

1. **Client Type Resolution**: Determine which client is being launched (claude/codex/opencode)
2. **Permission Flag Detection**: Scan `--` parameters for permission-related flags:
   - Claude Code: `--permission-mode`
   - Codex: `-a` or `--ask-for-approval`
   - OpenCode: TBD
3. **Priority Resolution**: Apply priority order (see State Transitions above)
4. **Flag Construction**: Build final argument list with permission flags prepended (if applicable)

---

## Migration Strategy

### From Version 11 to 12

**Changes**:
- Add `claude_auto_permission` field
- Add `codex_auto_permission` field
- Add `opencode_auto_permission` field

**Migration Logic**:
```go
// In UnmarshalJSON or after unmarshal:
if cfg.Version < 12 {
    // Initialize new fields with defaults
    if cfg.ClaudeAutoPermission == nil {
        cfg.ClaudeAutoPermission = &AutoPermissionConfig{Enabled: false, Mode: ""}
    }
    if cfg.CodexAutoPermission == nil {
        cfg.CodexAutoPermission = &AutoPermissionConfig{Enabled: false, Mode: ""}
    }
    if cfg.OpenCodeAutoPermission == nil {
        cfg.OpenCodeAutoPermission = &AutoPermissionConfig{Enabled: false, Mode: ""}
    }
    cfg.Version = 12
}
```

**Backward Compatibility**:
- Old configs (version < 12) without new fields: Work fine, auto-permission disabled
- New configs (version 12) read by old zen versions: Gracefully ignored (unknown fields skipped)

**Forward Compatibility**:
- Future versions can add more fields without breaking version 12 configs

---

## Summary

**New Types**: 1 (`AutoPermissionConfig`)
**Modified Types**: 1 (`OpenCCConfig`)
**Config Version**: Bump to 12
**Migration Complexity**: Low (add optional fields with defaults)
**Validation Points**: 2 (config load, runtime execution)
