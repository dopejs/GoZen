---
sidebar_position: 5
title: Project Bindings
---

# Project Bindings

Bind directories to specific profiles and/or CLIs for automatic project-level configuration.

## Usage

```bash
cd ~/work/company-project

# Bind profile
zen bind work-profile

# Bind CLI
zen bind --cli codex

# Bind both
zen bind work-profile --cli codex

# Check status
zen status

# Unbind
zen unbind
```

## Priority

CLI arguments > Project bindings > Global defaults
