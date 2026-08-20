---
sidebar_position: 1
title: Erste Schritte
---

# Erste Schritte

## Installation

Installation mit dem Einzeiler:

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

Deinstallation:

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## Erster Start

Den ersten Anbieter hinzufügen:

```bash
zen config add provider
```

Mit dem Standardprofil starten:

```bash
zen
```

Mit einem bestimmten Profil starten:

```bash
zen -p work
```

Mit einer bestimmten CLI starten:

```bash
zen --cli codex
```
