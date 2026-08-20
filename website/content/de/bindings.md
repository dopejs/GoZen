---
sidebar_position: 5
title: Projektbindungen
---

# Projektbindungen

Binde Verzeichnisse an bestimmte Profile und/oder CLIs, damit die Konfiguration je Projekt automatisch greift.

## Verwendung

```bash
cd ~/work/company-project

# Profil binden
zen bind work-profile

# CLI binden
zen bind --cli codex

# Beides binden
zen bind work-profile --cli codex

# Status prüfen
zen status

# Bindung lösen
zen unbind
```

## Priorität

CLI-Argumente > Projektbindungen > globale Standardwerte
