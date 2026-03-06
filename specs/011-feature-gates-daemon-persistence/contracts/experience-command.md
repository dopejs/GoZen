# CLI Contract: zen experience

**Feature**: 011-feature-gates-daemon-persistence
**Command**: `zen experience`
**Type**: Hidden command (not shown in `zen --help`)

## Overview

The `zen experience` command manages experimental feature gates. It allows power users to enable/disable BETA features without exposing them to general users through help text.

## Command Syntax

```bash
zen experience [feature] [-c|--close]
```

## Subcommands

### List Features (No Arguments)

**Syntax**: `zen experience`

**Description**: Display all available experimental features with their current status.

**Output Format**:
```
Experimental Features:
  bot          [enabled]   Bot gateway (BETA)
  compression  [disabled]  Context compression (BETA)
  middleware   [disabled]  Middleware pipeline (BETA)
  agent        [enabled]   Agent infrastructure (BETA)
```

**Exit Codes**:
- `0`: Success
- `1`: Config file read error

**Example**:
```bash
$ zen experience
Experimental Features:
  bot          [disabled]  Bot gateway (BETA)
  compression  [disabled]  Context compression (BETA)
  middleware   [disabled]  Middleware pipeline (BETA)
  agent        [disabled]  Agent infrastructure (BETA)
```

### Enable Feature

**Syntax**: `zen experience <feature>`

**Description**: Enable an experimental feature.

**Arguments**:
- `<feature>`: Feature name (bot, compression, middleware, agent)

**Output Format**:
```
Feature '<feature>' enabled.
Changes will take effect after daemon restart: zen daemon restart
```

**Exit Codes**:
- `0`: Success
- `1`: Invalid feature name
- `1`: Config file write error

**Example**:
```bash
$ zen experience bot
Feature 'bot' enabled.
Changes will take effect after daemon restart: zen daemon restart
```

**Error Example**:
```bash
$ zen experience invalid
Error: unknown feature 'invalid'
Valid features: bot, compression, middleware, agent
```

### Disable Feature

**Syntax**: `zen experience <feature> -c` or `zen experience <feature> --close`

**Description**: Disable an experimental feature.

**Arguments**:
- `<feature>`: Feature name (bot, compression, middleware, agent)

**Flags**:
- `-c, --close`: Disable the feature

**Output Format**:
```
Feature '<feature>' disabled.
Changes will take effect after daemon restart: zen daemon restart
```

**Exit Codes**:
- `0`: Success
- `1`: Invalid feature name
- `1`: Config file write error

**Example**:
```bash
$ zen experience bot -c
Feature 'bot' disabled.
Changes will take effect after daemon restart: zen daemon restart
```

## Valid Feature Names

| Feature | Description | Config Field |
|---------|-------------|--------------|
| `bot` | Bot gateway (BETA) | `feature_gates.bot` |
| `compression` | Context compression (BETA) | `feature_gates.compression` |
| `middleware` | Middleware pipeline (BETA) | `feature_gates.middleware` |
| `agent` | Agent infrastructure (BETA) | `feature_gates.agent` |

**Case Sensitivity**: Feature names are case-insensitive (bot = Bot = BOT)

## Behavior

### Hidden Command

- Command does NOT appear in `zen --help` output
- Command does NOT appear in `zen help` output
- Command IS accessible via `zen experience` for power users
- Command IS included in shell completion (bash/zsh/fish)

### Config Modification

1. Load config from `~/.zen/zen.json`
2. Parse `OpenCCConfig` (version 13)
3. Create `FeatureGates` struct if nil
4. Modify specified feature field
5. Write config back to `~/.zen/zen.json`
6. Log audit entry to stderr

### Daemon Interaction

- Config change triggers daemon reload (via ConfigWatcher, 2s polling)
- Daemon logs feature gate changes to `~/.zen/zend.log`
- Daemon continues serving with OLD feature gate values
- User must manually restart daemon for changes to take effect

**Rationale**: Infrastructure-level features require daemon restart to initialize/teardown properly.

### Audit Logging

Every enable/disable action logs to stderr:

```
[zen] 2026-03-05 10:30:45 [AUDIT] action=enable_feature_gate resource=bot user=john
```

**Format**: `[zen] TIMESTAMP [AUDIT] action=ACTION resource=RESOURCE user=USER`

## Error Handling

### Invalid Feature Name

**Input**: `zen experience invalid`

**Output**:
```
Error: unknown feature 'invalid'
Valid features: bot, compression, middleware, agent
```

**Exit Code**: 1

### Config File Not Found

**Input**: `zen experience bot` (when `~/.zen/zen.json` missing)

**Output**:
```
Error: config file not found: ~/.zen/zen.json
Run 'zen config add' to create initial configuration.
```

**Exit Code**: 1

### Config File Corrupted

**Input**: `zen experience bot` (when `~/.zen/zen.json` has invalid JSON)

**Output**:
```
Error: failed to parse config file: invalid character '}' looking for beginning of value
```

**Exit Code**: 1

### Config File Not Writable

**Input**: `zen experience bot` (when `~/.zen/zen.json` is read-only)

**Output**:
```
Error: failed to write config file: permission denied
```

**Exit Code**: 1

## Integration with Other Commands

### zen daemon status

**Enhancement**: Show feature gate status in daemon status output.

**Output**:
```
Daemon Status: running
PID: 12345
Uptime: 2h 15m
Web UI: http://localhost:19840
Proxy: http://localhost:19841

Feature Gates:
  bot:         enabled
  compression: disabled
  middleware:  disabled
  agent:       enabled
```

### zen daemon restart

**Behavior**: Restart daemon to apply feature gate changes.

**Output**:
```
Stopping daemon...
Daemon stopped.
Starting daemon...
Daemon started (PID: 12346)
Feature gate changes applied.
```

## Backward Compatibility

### Version 12 Configs

- Configs without `feature_gates` field: All features disabled
- Command works normally: Creates `feature_gates` field on first enable
- No migration needed

### Version 13 Configs

- Configs with `feature_gates` field: Read/write normally
- Partial `feature_gates` object: Missing fields default to false

## Security Considerations

### Access Control

- Command requires write access to `~/.zen/zen.json`
- No additional authentication required (local user only)
- Audit log records user from `$USER` environment variable

### Privilege Escalation

- Command does NOT require sudo/root
- Command operates on user-level config only
- Daemon runs as user (not system service)

## Testing Contract

### Unit Tests

```bash
# Test enable feature
zen experience bot
# Verify: config contains "feature_gates": {"bot": true}

# Test disable feature
zen experience bot -c
# Verify: config contains "feature_gates": {"bot": false}

# Test list features
zen experience
# Verify: output contains all 4 features with status

# Test invalid feature
zen experience invalid
# Verify: exit code 1, error message with valid features
```

### Integration Tests

```bash
# Test daemon reload detection
zen experience bot
sleep 3  # Wait for ConfigWatcher poll
# Verify: daemon log contains "feature gates changed"

# Test daemon restart applies changes
zen experience bot
zen daemon restart
# Verify: bot feature is active in daemon
```

## Examples

### Enable bot feature

```bash
$ zen experience bot
Feature 'bot' enabled.
Changes will take effect after daemon restart: zen daemon restart

$ zen daemon restart
Stopping daemon...
Daemon stopped.
Starting daemon...
Daemon started (PID: 12346)
Feature gate changes applied.
```

### Disable compression feature

```bash
$ zen experience compression -c
Feature 'compression' disabled.
Changes will take effect after daemon restart: zen daemon restart
```

### List all features

```bash
$ zen experience
Experimental Features:
  bot          [enabled]   Bot gateway (BETA)
  compression  [disabled]  Context compression (BETA)
  middleware   [disabled]  Middleware pipeline (BETA)
  agent        [disabled]  Agent infrastructure (BETA)
```

### Check daemon status with feature gates

```bash
$ zen daemon status
Daemon Status: running
PID: 12345
Uptime: 2h 15m
Web UI: http://localhost:19840
Proxy: http://localhost:19841

Feature Gates:
  bot:         enabled
  compression: disabled
  middleware:  disabled
  agent:       disabled
```
