---
sidebar_position: 5
title: Liaisons de projet
---

# Liaisons de projet

Liez des répertoires à des profils et/ou des CLI précis pour obtenir une configuration automatique au niveau du projet.

## Utilisation

```bash
cd ~/work/company-project

# Lier un profil
zen bind work-profile

# Lier une CLI
zen bind --cli codex

# Lier les deux
zen bind work-profile --cli codex

# Vérifier l’état
zen status

# Délier
zen unbind
```

## Priorité

Arguments de la CLI > liaisons de projet > valeurs par défaut globales
