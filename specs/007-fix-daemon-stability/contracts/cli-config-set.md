# CLI Contract: `zen config set`

**Feature**: 007-fix-daemon-stability

## Command Signature

```
zen config set <key> <value>
```

## Supported Keys

| Key | Type | Validation | Side Effects |
|-----|------|------------|--------------|
| `proxy_port` | int | 1024-65535 | Saves to config, stops running daemon, starts daemon on new port, prints warning to restart client processes |

## Behavior

### Success

```
$ zen config set proxy_port 29841
Proxy port updated to 29841.
Restarting daemon on new port...
Daemon restarted on port 29841.
⚠ Please restart all running zen client processes (their ANTHROPIC_BASE_URL still points to the old port).
```

### Validation Error

```
$ zen config set proxy_port 80
Error: port must be between 1024 and 65535

$ zen config set proxy_port abc
Error: invalid value for proxy_port: "abc" (expected integer)
```

### Unknown Key

```
$ zen config set unknown_key value
Error: unknown config key: "unknown_key"
Supported keys: proxy_port
```

### Daemon Not Running

```
$ zen config set proxy_port 29841
Proxy port updated to 29841.
Daemon is not running. New port will take effect on next start.
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Validation error, unknown key, or daemon restart failure |
