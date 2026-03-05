# Quickstart: zen use Command Enhancements

**Date**: 2026-03-05
**Feature**: 010-use-command-enhancements

## For Users

### Using `--yes` Flag

**Problem**: Current `--yes` only auto-approves file edits, still prompts for reads and bash commands.

**Solution**: Updated `--yes` now bypasses ALL permission prompts.

```bash
# Before (still prompts for reads/bash)
zen --yes
# Claude Code receives: --permission-mode acceptEdits

# After (no prompts at all)
zen --yes
# Claude Code receives: --permission-mode bypassPermissions
# Codex receives: -a never
```

**Use Cases**:
- Automation scripts and CI/CD pipelines
- Power users who trust the AI assistant
- Rapid prototyping without interruptions

---

### Configuring Auto-Permission in Web UI

**Problem**: Typing `--yes` every time is tedious for frequent users.

**Solution**: Configure auto-permission mode once in Web UI.

**Steps**:
1. Open Web UI: `zen web`
2. Navigate to Settings → Permission Configuration
3. Select client (Claude Code, Codex, or OpenCode)
4. Enable auto-permission mode
5. Choose permission level:
   - **Claude Code**: default, acceptEdits, bypassPermissions, dontAsk, plan
   - **Codex**: untrusted, on-request, never
6. Save configuration

**Result**: All subsequent `zen` commands automatically use the configured permission mode.

**Example**:
```bash
# Configure once in Web UI: Claude Code → bypassPermissions

# Then just run zen normally
zen
# Automatically receives: --permission-mode bypassPermissions

# Override with --yes if needed (uses same mode)
zen --yes
# Still uses: --permission-mode bypassPermissions
```

---

### Passing Custom Client Parameters

**Problem**: Can't pass client-specific flags through zen wrapper.

**Solution**: Use `--` separator to pass arbitrary parameters.

```bash
# Pass verbose flag to client
zen -- --verbose

# Pass multiple flags
zen -- --verbose --debug --log-level trace

# Override permission mode
zen --yes -- --permission-mode acceptEdits
# Uses acceptEdits (not bypassPermissions) because `--` has priority

# With profile selection
zen -p work -- --permission-mode plan
```

**Priority Order**:
1. `--` parameters (highest)
2. `--yes` flag
3. Web UI config
4. Default behavior (lowest)

---

## For Developers

### Implementation Checklist

**Phase 1: Backend (Go)**

1. **Update `cmd/root.go`**:
   - Modify `prependAutoApproveArgs()` to use `bypassPermissions` instead of `acceptEdits`
   - Add permission flag detection logic for `--` parameters
   - Implement priority resolution (--  > --yes > Web UI > default)

2. **Update `internal/config/config.go`**:
   - Add `AutoPermissionConfig` type
   - Add fields to `OpenCCConfig`: `claude_auto_permission`, `codex_auto_permission`, `opencode_auto_permission`
   - Bump `CurrentConfigVersion` to 12

3. **Add config migration**:
   - Initialize new fields with defaults if missing
   - Test old config parsing

4. **Write tests**:
   - Test permission flag detection
   - Test priority resolution
   - Test config migration
   - Table-driven tests in existing `*_test.go` files

**Phase 2: Frontend (React)**

1. **Create `PermissionConfig.tsx` component**:
   - Client selector dropdown
   - Enable/disable toggle
   - Permission mode dropdown (client-specific options)
   - Save button

2. **Update Settings page**:
   - Add PermissionConfig component
   - Wire up to config API

3. **Write tests**:
   - Component rendering tests
   - Client-specific options tests
   - Save functionality tests

**Phase 3: Documentation**

1. **Update all four READMEs**:
   - `README.md` (English)
   - `docs/README.zh-CN.md` (简体中文)
   - `docs/README.zh-TW.md` (繁體中文)
   - `docs/README.es.md` (Español)

2. **Update help text**:
   - `zen --help` output
   - Document `--yes` behavior change
   - Document `--` separator usage
   - Document priority order

---

### Testing Strategy

**Unit Tests** (Go):
```go
// Test permission flag detection
func TestDetectPermissionFlags(t *testing.T) {
    tests := []struct {
        name     string
        client   string
        args     []string
        expected bool
    }{
        {"claude with permission-mode", "claude", []string{"--permission-mode", "default"}, true},
        {"codex with -a", "codex", []string{"-a", "never"}, true},
        {"no permission flags", "claude", []string{"--verbose"}, false},
    }
    // ...
}

// Test priority resolution
func TestPermissionPriorityResolution(t *testing.T) {
    // Test: -- > --yes > Web UI > default
}

// Test config migration
func TestConfigMigrationV11ToV12(t *testing.T) {
    // Test old config without new fields
    // Test new fields initialized with defaults
}
```

**Integration Tests** (Go):
```bash
# Test --yes flag
./zen --yes 2>&1 | grep "bypassPermissions"

# Test -- separator
./zen -- --verbose 2>&1 | grep "verbose"

# Test priority order
./zen --yes -- --permission-mode acceptEdits 2>&1 | grep "acceptEdits"
```

**Component Tests** (React):
```typescript
// Test PermissionConfig component
describe('PermissionConfig', () => {
  it('shows Claude Code options when claude selected', () => {
    // Render component with clientType="claude"
    // Assert 5 options visible
  });

  it('shows Codex options when codex selected', () => {
    // Render component with clientType="codex"
    // Assert 3 options visible
  });

  it('saves configuration on submit', async () => {
    // Mock API call
    // Submit form
    // Assert API called with correct data
  });
});
```

---

### TDD Workflow

**Red-Green-Refactor**:

1. **Red**: Write failing test
   ```go
   func TestYesFlagUsesBypassPermissions(t *testing.T) {
       args := prependAutoApproveArgs("claude", []string{}, true, nil)
       if !contains(args, "--permission-mode") || !contains(args, "bypassPermissions") {
           t.Error("Expected bypassPermissions mode")
       }
   }
   ```

2. **Green**: Implement minimal code to pass
   ```go
   func prependAutoApproveArgs(clientBin string, args []string, autoApprove bool, config *AutoPermissionConfig) []string {
       if autoApprove {
           return append([]string{"--permission-mode", "bypassPermissions"}, args...)
       }
       return args
   }
   ```

3. **Refactor**: Clean up and add more cases
   - Add Codex support
   - Add priority resolution
   - Add `--` detection

---

## Common Pitfalls

### Pitfall 1: Forgetting to Bump Config Version

**Problem**: Adding new fields without bumping version breaks migration.

**Solution**: Always bump `CurrentConfigVersion` when modifying `OpenCCConfig` schema.

### Pitfall 2: Not Testing Old Configs

**Problem**: Migration logic not tested, old configs break.

**Solution**: Add test cases for configs without new fields.

### Pitfall 3: Abstracting Permission Modes

**Problem**: Trying to map Claude Code modes to Codex modes.

**Solution**: Keep them separate, no abstraction (per clarification session).

### Pitfall 4: Ignoring Priority Order

**Problem**: `--yes` overrides `--` parameters.

**Solution**: Always check `--` first, then `--yes`, then Web UI, then default.

---

## Summary

**Key Changes**:
1. `--yes` now uses `bypassPermissions` (Claude Code) and `never` (Codex)
2. Web UI allows per-client auto-permission configuration
3. `--` separator passes arbitrary parameters with highest priority
4. Config version bumped to 12 with new auto-permission fields

**Testing Requirements**:
- Unit tests for permission logic
- Integration tests for CLI behavior
- Component tests for Web UI
- Config migration tests

**Documentation Updates**:
- All four README translations
- Help text for `zen --help`
- Web UI tooltips and labels
