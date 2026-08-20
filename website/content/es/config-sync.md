---
sidebar_position: 9
title: Sincronización de Config
---

# Sincronización de Config

Sincroniza proveedores, perfiles, perfil predeterminado y cliente predeterminado entre dispositivos. Los tokens de autenticación se cifran con AES-256-GCM (derivación de clave PBKDF2-SHA256) antes de la carga.

## Backends Soportados

| Backend | Descripción |
|---------|-------------|
| `webdav` | Cualquier servidor WebDAV (ej. Nextcloud, ownCloud) |
| `s3` | AWS S3 o almacenamiento compatible con S3 (ej. MinIO, Cloudflare R2) |
| `gist` | Gist privado (requiere PAT con alcance gist) |
| `repo` | Archivo de repositorio vía Contents API (requiere PAT con alcance repo) |

## Configuración

Configura la sincronización a través de la página de ajustes de la Web UI:

```bash
# Open Web UI settings
zen web  # Settings → Config Sync
```

O extrae manualmente vía CLI:

```bash
zen config sync
```

## Ejemplo de Configuración

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

## Cifrado

Cuando se establece una frase de contraseña, todos los tokens de autenticación en la carga de sincronización se cifran con AES-256-GCM usando una clave derivada vía PBKDF2-SHA256 (600k iteraciones). La sal de cifrado se almacena junto con la carga. Sin frase de contraseña, los datos se cargan como JSON en texto plano.

## Resolución de Conflictos

- Fusión por marca de tiempo por entidad: la modificación más reciente gana
- Las entidades eliminadas usan lápidas (expiran después de 30 días)
- Escalares (perfil/cliente predeterminado): la marca de tiempo más reciente gana

## Alcance de Sincronización

**Sincronizado:** Proveedores (con tokens cifrados), Perfiles, Perfil predeterminado, Cliente predeterminado

**No Sincronizado:** Configuración de puertos, Contraseña Web, Vinculaciones de proyecto, La configuración de sincronización en sí
