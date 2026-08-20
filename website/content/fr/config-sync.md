---
sidebar_position: 9
title: Synchronisation de la configuration
---

# Synchronisation de la configuration

Synchronisez fournisseurs, profils, profil par défaut et client par défaut entre appareils. Les jetons d’authentification sont chiffrés en AES-256-GCM (dérivation de clé PBKDF2-SHA256) avant l’envoi.

## Backends pris en charge

| Backend | Description |
|---------|-------------|
| `webdav` | N’importe quel serveur WebDAV (Nextcloud, ownCloud, etc.) |
| `s3` | AWS S3 ou stockage compatible S3 (MinIO, Cloudflare R2, etc.) |
| `gist` | Gist privé (nécessite un PAT avec la portée gist) |
| `repo` | Fichier de dépôt via l’API Contents (nécessite un PAT avec la portée repo) |

## Mise en place

Configurez la synchronisation depuis la page des réglages de l’interface web :

```bash
# Ouvrir les réglages de l’interface web
zen web  # Réglages → Synchronisation
```

Ou récupérez manuellement via la CLI :

```bash
zen config sync
```

## Exemple de configuration

```json
{
  "sync": {
    "backend": "gist",
    "gist_id": "abc123def456",
    "token": "ghp_xxxxxxxxxxxx",
    "passphrase": "my-secret-passphrase",
    "auto_pull": true,
    "pull_interval": 300
  }
}
```

## Chiffrement

Quand une phrase secrète est définie, tous les jetons d’authentification de la charge synchronisée sont chiffrés en AES-256-GCM avec une clé dérivée par PBKDF2-SHA256 (600 000 itérations). Le sel de chiffrement est stocké à côté de la charge. Sans phrase secrète, les données sont envoyées en JSON clair.

## Résolution des conflits

- Fusion par horodatage et par entité : la modification la plus récente l’emporte
- Les entités supprimées utilisent des pierres tombales (expiration au bout de 30 jours)
- Valeurs scalaires (profil/client par défaut) : l’horodatage le plus récent l’emporte

## Portée de la synchronisation

**Synchronisé :** fournisseurs (jetons chiffrés), profils, profil par défaut, client par défaut

**Non synchronisé :** réglages de ports, mot de passe web, liaisons de projet, configuration de la synchronisation elle-même
