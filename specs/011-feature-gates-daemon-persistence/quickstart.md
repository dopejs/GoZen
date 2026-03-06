# Quickstart: Feature Gates & Daemon Persistence

**Feature**: 011-feature-gates-daemon-persistence
**Audience**: Developers and power users

## What's New

This feature adds:
1. **Feature Gates**: Hidden `zen experience` command to enable/disable experimental features
2. **Enhanced Daemon Persistence**: Improved reliability for daemon survival across sleep/wake cycles

## Feature Gates

### Overview

Feature gates allow you to enable experimental (BETA) features that are not yet ready for general availability. The `zen experience` command is intentionally hidden from help text to avoid confusing regular users.

### Available Features

| Feature | Description | Status |
|---------|-------------|--------|
| `bot` | Bot gateway for Slack/Discord integration | BETA |
| `compression` | Context compression for long conversations | BETA |
| `middleware` | Middleware pipeline for request/response transformation | BETA |
| `agent` | Agent infrastructure for multi-agent workflows | BETA |

### Quick Start

**List all features:**
```bash
zen experience
```

**Enable a feature:**
```bash
zen experience bot
zen daemon restart  # Required to apply changes
```

**Disable a feature:**
```bash
zen experience bot -c
zen daemon restart  # Required to apply changes
```

### Example Workflow

```bash
# Check current feature status
$ zen experience
Experimental Features:
  bot          [disabled]  Bot gateway (BETA)
  compression  [disabled]  Context compression (BETA)
  middleware   [disabled]  Middleware pipeline (BETA)
  agent        [disabled]  Agent infrastructure (BETA)

# Enable bot feature
$ zen experience bot
Feature 'bot' enabled.
Changes will take effect after daemon restart: zen daemon restart

# Restart daemon to apply changes
$ zen daemon restart
Stopping daemon...
Daemon stopped.
Starting daemon...
Daemon started (PID: 12346)
Feature gate changes applied.

# Verify feature is enabled
$ zen daemon status
Daemon Status: running
PID: 12346
Uptime: 5s
Web UI: http://localhost:19840
Proxy: http://localhost:19841

Feature Gates:
  bot:         enabled
  compression: disabled
  middleware:  disabled
  agent:       disabled
```

### Important Notes

- **Hidden Command**: `zen experience` does not appear in `zen --help` output
- **Restart Required**: Changes take effect only after `zen daemon restart`
- **Audit Logging**: All enable/disable actions are logged to `~/.zen/zend.log`
- **Persistence**: Feature gate settings persist across daemon restarts and system reboots

## Daemon Persistence

### Overview

The daemon now has enhanced persistence mechanisms to ensure it survives:
- System sleep/wake cycles (resumes within 10 seconds)
- CLI process termination (daemon runs independently)
- System reboots (auto-starts on login when enabled)

### Verify Daemon Persistence

**Check daemon status:**
```bash
zen daemon status
```

**Expected output:**
```
Daemon Status: running
PID: 12345
Uptime: 2h 15m
Web UI: http://localhost:19840
Proxy: http://localhost:19841
```

### Test Sleep/Wake Survival

1. Start daemon: `zen daemon start`
2. Put computer to sleep (close laptop or use system menu)
3. Wait 5+ minutes
4. Wake computer
5. Check daemon status: `zen daemon status`
6. Expected: Daemon still running on same ports

### Test Process Independence

1. Start daemon: `zen daemon start`
2. Launch Claude Code: `zen`
3. Find CLI process: `ps aux | grep zen`
4. Kill CLI process: `kill <pid>`
5. Check daemon status: `zen daemon status`
6. Expected: Daemon still running

### Platform-Specific Details

**macOS (launchd)**:
- Service file: `~/Library/LaunchAgents/com.dopejs.zend.plist`
- Auto-restart: `KeepAlive=true`
- Auto-start on login: `RunAtLoad=true`

**Linux (systemd)**:
- Service file: `~/.config/systemd/user/zend.service`
- Auto-restart: `Restart=always` (enhanced from `on-failure`)
- Auto-start on login: `WantedBy=default.target`

### Enable Daemon as System Service

**Install service:**
```bash
zen daemon enable
```

**Verify service is running:**
```bash
# macOS
launchctl list | grep zend

# Linux
systemctl --user status zend
```

**Disable service:**
```bash
zen daemon disable
```

## Configuration

### Config File Location

`~/.zen/zen.json`

### Feature Gates Schema (Version 13)

```json
{
  "version": 13,
  "feature_gates": {
    "bot": false,
    "compression": false,
    "middleware": false,
    "agent": false
  },
  "providers": {...},
  "profiles": {...}
}
```

### Manual Configuration

You can also edit `~/.zen/zen.json` directly:

```json
{
  "version": 13,
  "feature_gates": {
    "bot": true,
    "compression": false,
    "middleware": false,
    "agent": true
  }
}
```

After manual edits, restart the daemon:
```bash
zen daemon restart
```

## Troubleshooting

### Feature gate changes not applied

**Problem**: Enabled a feature but it's not working

**Solution**: Restart the daemon
```bash
zen daemon restart
```

### Daemon not surviving sleep/wake

**Problem**: Daemon stops after waking from sleep

**Solution**: Install daemon as system service
```bash
zen daemon enable
```

### Daemon stops when terminal closes

**Problem**: Daemon terminates when closing terminal

**Solution**: Install daemon as system service
```bash
zen daemon enable
```

### Check daemon logs

**View recent logs:**
```bash
tail -f ~/.zen/zend.log
```

**Search for feature gate changes:**
```bash
grep "feature gates changed" ~/.zen/zend.log
```

**Search for audit logs:**
```bash
grep "\[AUDIT\]" ~/.zen/zend.log
```

## Advanced Usage

### Scripting Feature Gates

```bash
#!/bin/bash
# Enable multiple features
for feature in bot agent; do
  zen experience "$feature"
done

# Restart daemon once
zen daemon restart
```

### Check Feature Status Programmatically

```bash
# Parse config file
cat ~/.zen/zen.json | jq '.feature_gates'
```

**Output:**
```json
{
  "bot": true,
  "compression": false,
  "middleware": false,
  "agent": true
}
```

### Monitor Daemon Uptime

```bash
# Check daemon status every 60 seconds
watch -n 60 zen daemon status
```

## Migration from Version 12

No action required. Version 12 configs are automatically compatible:
- Missing `feature_gates` field: All features disabled by default
- First `zen experience` command creates the field automatically

## Security

### Access Control

- Feature gates require write access to `~/.zen/zen.json`
- No additional authentication required (local user only)
- All changes logged with username to `~/.zen/zend.log`

### Audit Trail

All feature gate changes are logged:

```bash
$ grep "\[AUDIT\]" ~/.zen/zend.log
[zend] 2026-03-05 10:30:45 [AUDIT] action=enable_feature_gate resource=bot user=john
[zend] 2026-03-05 10:31:12 [AUDIT] action=disable_feature_gate resource=compression user=john
```

## Next Steps

- **Enable experimental features**: `zen experience <feature>`
- **Install daemon as service**: `zen daemon enable`
- **Read feature documentation**: Check `website/docs/` for BETA feature guides
- **Monitor daemon logs**: `tail -f ~/.zen/zend.log`

## Related Documentation

- [Bot Gateway](../../../website/docs/bot-gateway.md)
- [Context Compression](../../../website/docs/compression.md)
- [Middleware Pipeline](../../../website/docs/middleware.md)
- [Agent Infrastructure](../../../website/docs/agent-infrastructure.md)
