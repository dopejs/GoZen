---
sidebar_position: 3
title: Profils et bascule
---

# Profils et bascule

Un profil est une liste ordonnée de fournisseurs servant à la bascule. Quand le premier fournisseur est indisponible, le suivant prend automatiquement le relais.

## Exemple de configuration

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}
```

## Utiliser les profils

```bash
# Utiliser le profil par défaut
zen

# Utiliser un profil précis
zen -p work

# Sélection interactive
zen -p
```
