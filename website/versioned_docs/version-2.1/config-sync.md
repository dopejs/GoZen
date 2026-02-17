---
sidebar_position: 9
title: Config Sync
---

# Config Sync

Sync providers, profiles, default profile, and default client across devices. Auth tokens are encrypted with AES-256-GCM (PBKDF2-SHA256 key derivation) before upload.

## Supported Backends

| Backend | Description |
|---------|-------------|
| `webdav` | Any WebDAV server (e.g. Nextcloud, ownCloud) |
| `s3` | AWS S3 or S3-compatible storage (e.g. MinIO, Cloudflare R2) |
| `gist` | Private gist (requires PAT with gist scope) |
| `repo` | Repository file via Contents API (requires PAT with repo scope) |

## Setup

Configure sync through the Web UI settings page:

```bash
# Open Web UI settings
zen web  # Settings → Config Sync
```

Or pull manually via CLI:

```bash
zen config sync
```

## Configuration Example

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

## Encryption

When a passphrase is set, all auth tokens in the sync payload are encrypted with AES-256-GCM using a key derived via PBKDF2-SHA256 (600k iterations). The encryption salt is stored alongside the payload. Without a passphrase, data is uploaded as plaintext JSON.

## Conflict Resolution

- Per-entity timestamp merge: newer modification wins
- Deleted entities use tombstones (expire after 30 days)
- Scalars (default profile/client): newer timestamp wins

## Sync Scope

**Synced:** Providers (with encrypted tokens), Profiles, Default profile, Default client

**Not Synced:** Port settings, Web password, Project bindings, Sync configuration itself
