---
sidebar_position: 7
title: Web Interface
---

# Web Interface

Visually manage all configurations through your browser. The daemon starts automatically when needed.

## Usage

```bash
# Open in browser (auto-starts daemon if needed)
zen web
```

## Features

- Provider and Profile management
- Project binding management
- Global settings (default client, default Profile, ports)
- Config sync settings
- Request log viewer with auto-refresh
- Model field autocomplete

## Security

When the daemon starts for the first time, it auto-generates an access password. Non-local requests (outside 127.0.0.1/::1) require login.

- Session-based auth with configurable expiry
- Brute-force protection with exponential backoff
- RSA encryption for sensitive token transport (API keys encrypted in-browser)
- Local access (127.0.0.1) bypasses authentication

### Password Management

```bash
# Reset the Web UI password
zen config reset-password

# Change password via Web UI
zen web  # Settings → Change Password
```
