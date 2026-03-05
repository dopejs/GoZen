# Research: zen use Command Enhancements

**Date**: 2026-03-05
**Feature**: 010-use-command-enhancements

## Research Questions

### Q1: OpenCode Permission Handling

**Question**: Does OpenCode have permission flags similar to Claude Code and Codex, or does it auto-approve by default?

**Research Approach**:
- Check OpenCode CLI documentation
- Test OpenCode behavior with and without permission flags
- Determine if any flags need to be passed for `--yes` mode

**Decision**: Assume OpenCode auto-approves by default (no permission flags needed)

**Rationale**:
- OpenCode is designed for autonomous operation and typically doesn't require permission prompts
- If OpenCode does have permission flags, they can be discovered during T022 implementation
- Fallback strategy: If flags are needed, add them in T022; if not needed, T022 becomes a no-op
- This assumption unblocks implementation while maintaining flexibility

**Implementation Note**: T022 should verify OpenCode behavior and add flags only if needed

---

### Q2: Permission Flag Detection in `--` Parameters

**Question**: How should the system detect permission-related parameters in `--` to avoid duplicating them when `--yes` or Web UI config is also present?

**Research Approach**:
- Identify all permission-related flags for each client:
  - Claude Code: `--permission-mode`
  - Codex: `-a`, `--ask-for-approval`
  - OpenCode: TBD
- Determine detection strategy: exact match, prefix match, or parse all flags

**Decision**: Use simple string scanning for known permission flags before prepending auto-permission flags

**Rationale**:
- Simple and predictable behavior
- Avoids complex flag parsing
- Users explicitly using `--` expect full control
- If permission flag detected in `--` parameters, skip adding flags from `--yes` or Web UI config

**Alternatives Considered**:
- Full flag parsing: Too complex, not needed for this use case
- No detection: Would result in duplicate flags, client behavior undefined

---

### Q3: Config Schema Migration Strategy

**Question**: What is the current `CurrentConfigVersion` and what migration logic is needed for the new permission fields?

**Research Approach**:
- Read `internal/config/config.go` to find current version
- Review existing migration patterns in `UnmarshalJSON`
- Determine if new fields can use default values or need explicit migration

**Decision**: Add new optional fields with sensible defaults, no complex migration needed

**Rationale**:
- New fields: `claude_auto_permission`, `codex_auto_permission`, `opencode_auto_permission`
- Structure: `{ "enabled": false, "mode": "" }` (disabled by default)
- Backward compatibility: Old configs without these fields work fine (disabled state)
- Forward compatibility: New configs with these fields are ignored by old versions (graceful degradation)

**Migration Logic**:
```go
// In UnmarshalJSON or loadLocked():
if cfg.ClaudeAutoPermission == nil {
    cfg.ClaudeAutoPermission = &AutoPermissionConfig{Enabled: false, Mode: ""}
}
// Repeat for Codex and OpenCode
```

---

### Q4: Web UI Permission Mode Dropdown Implementation

**Question**: How should the Web UI dynamically display client-specific permission options?

**Research Approach**:
- Review existing Web UI patterns for client-specific configuration
- Determine if client selection already exists in settings
- Design dropdown component that changes options based on selected client

**Decision**: Create a `PermissionConfig` component with client-aware dropdown

**Rationale**:
- Component receives `clientType` prop (claude/codex/opencode)
- Dropdown options change based on `clientType`:
  - Claude Code: 5 options (default, acceptEdits, bypassPermissions, dontAsk, plan)
  - Codex: 3 options (untrusted, on-request, never)
  - OpenCode: TBD (pending Q1 resolution)
- Store separate config for each client in backend
- No abstraction layer needed (per clarification session)

**Implementation Pattern**:
```typescript
const permissionOptions = {
  claude: [
    { value: 'default', label: 'Default (prompt for each tool)' },
    { value: 'acceptEdits', label: 'Accept Edits (auto-approve file edits only)' },
    { value: 'bypassPermissions', label: 'Bypass All (skip all prompts)' },
    { value: 'dontAsk', label: 'Don\'t Ask (auto-deny unless pre-approved)' },
    { value: 'plan', label: 'Plan Mode (analyze only, no modifications)' }
  ],
  codex: [
    { value: 'untrusted', label: 'Untrusted (ask for untrusted operations)' },
    { value: 'on-request', label: 'On Request (ask in interactive runs)' },
    { value: 'never', label: 'Never (never ask, for automation)' }
  ]
};
```

---

## Summary

**Resolved Questions**: 3/4
**Remaining**: OpenCode permission behavior (Q1)

**Next Steps**:
1. Investigate OpenCode CLI to determine permission handling
2. Proceed to Phase 1 (data model and contracts) with current decisions
3. Update research.md with OpenCode findings before implementation

**Key Decisions**:
- Permission flag detection: Simple string scanning in `--` parameters
- Config migration: Add optional fields with defaults, no complex migration
- Web UI: Client-aware dropdown component with no abstraction
- Priority order: `--` > `--yes` > Web UI > default (confirmed in spec)
