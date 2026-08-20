---
sidebar_position: 9
title: Konfigurationssynchronisierung
---

# Konfigurationssynchronisierung

Synchronisiere Anbieter, Profile, Standardprofil und Standardclient zwischen Geräten. Auth-Tokens werden vor dem Hochladen mit AES-256-GCM verschlüsselt (Schlüsselableitung per PBKDF2-SHA256).

## Unterstützte Backends

| Backend | Beschreibung |
|---------|--------------|
| `webdav` | Beliebiger WebDAV-Server (z. B. Nextcloud, ownCloud) |
| `s3` | AWS S3 oder S3-kompatibler Speicher (z. B. MinIO, Cloudflare R2) |
| `gist` | Privater Gist (benötigt ein PAT mit gist-Scope) |
| `repo` | Repository-Datei über die Contents-API (benötigt ein PAT mit repo-Scope) |

## Einrichtung

Richte die Synchronisierung auf der Einstellungsseite der Weboberfläche ein:

```bash
# Einstellungen der Weboberfläche öffnen
zen web  # Einstellungen → Konfigurationssynchronisierung
```

Oder manuell über die CLI holen:

```bash
zen config sync
```

## Beispielkonfiguration

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

## Verschlüsselung

Ist eine Passphrase gesetzt, werden alle Auth-Tokens in den synchronisierten Daten mit AES-256-GCM verschlüsselt; der Schlüssel wird über PBKDF2-SHA256 (600.000 Iterationen) abgeleitet. Das Salt liegt neben den Daten. Ohne Passphrase werden die Daten als Klartext-JSON hochgeladen.

## Konfliktlösung

- Zusammenführung je Entität nach Zeitstempel: die neuere Änderung gewinnt
- Gelöschte Entitäten hinterlassen Grabsteine (verfallen nach 30 Tagen)
- Skalare Werte (Standardprofil/-client): der neuere Zeitstempel gewinnt

## Umfang der Synchronisierung

**Synchronisiert:** Anbieter (mit verschlüsselten Tokens), Profile, Standardprofil, Standardclient

**Nicht synchronisiert:** Porteinstellungen, Web-Passwort, Projektbindungen, die Sync-Konfiguration selbst
