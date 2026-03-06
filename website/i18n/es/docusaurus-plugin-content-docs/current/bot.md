---
sidebar_position: 10
title: Gateway de Bot
---

# Gateway de Bot

Monitorea y controla remotamente tus sesiones de Claude Code a través de plataformas de chat. El bot se conecta a procesos `zen` en ejecución mediante IPC, permitiéndote:

- Ver procesos conectados y su estado
- Enviar tareas a procesos específicos
- Recibir notificaciones de aprobaciones, errores y completaciones
- Controlar tareas (pausar/reanudar/cancelar)

## Plataformas Soportadas

| Plataforma | Configuración Requerida |
|----------|----------------|
| [Telegram](#telegram) | Token de BotFather |
| [Discord](#discord) | Token de aplicación Bot |
| [Slack](#slack) | Tokens de Bot + App (Socket Mode) |
| [Lark/飞书](#lark飞书) | App ID + Secret |
| [Facebook Messenger](#facebook-messenger) | Token de página + Token de verificación |

## Configuración Básica

```json
{
  "bot": {
    "enabled": true,
    "socket_path": "/tmp/zen-bot.sock",
    "platforms": {
      // Configuración específica de plataforma (ver abajo)
    },
    "interaction": {
      "require_mention": true,
      "mention_keywords": ["@zen", "/zen"],
      "direct_message_mode": "always",
      "channel_mode": "mention"
    },
    "aliases": {
      "api": "/path/to/api-project",
      "web": "/path/to/web-project"
    },
    "notify": {
      "default_platform": "telegram",
      "default_chat_id": "-100123456789",
      "quiet_hours_start": "23:00",
      "quiet_hours_end": "07:00",
      "quiet_hours_zone": "Asia/Shanghai"
    }
  }
}
```

## Comandos del Bot

| Comando | Descripción |
|---------|-------------|
| `list` | Lista todos los procesos conectados |
| `status [name]` | Muestra el estado del proceso |
| `bind <name>` | Vincula a un proceso para comandos posteriores |
| `pause [name]` | Pausa la tarea actual |
| `resume [name]` | Reanuda la tarea pausada |
| `cancel [name]` | Cancela la tarea actual |
| `<name> <task>` | Envía una tarea al proceso |
| `help` | Muestra los comandos disponibles |

### Soporte de Lenguaje Natural

El bot entiende consultas en lenguaje natural en múltiples idiomas:

- "show me the status of gozen"
- "帮我看看 gozen 的状态"
- "list all processes"
- "pause the api project"

## Modos de Interacción

### Mensajes Directos

Configura `direct_message_mode` para controlar cómo responde el bot en mensajes directos:

- `"always"` — Siempre responde (no requiere mención)
- `"mention"` — Solo responde cuando es mencionado

### Mensajes de Canal

Configura `channel_mode` para controlar el comportamiento en chats grupales:

- `"always"` — Responde a todos los mensajes
- `"mention"` — Solo responde cuando es mencionado (recomendado)

### Palabras Clave de Mención

Configura las palabras clave que activan el bot:

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## Alias de Proyectos

Define nombres cortos para tus proyectos:

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

Luego úsalos en comandos:

```
api run tests
web build production
status backend
```

## Configuración de Plataformas

### Telegram

1. Crea un bot a través de [@BotFather](https://t.me/botfather):
   - Envía `/newbot` y sigue las instrucciones
   - Copia el token (ej: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. Obtén tu ID de usuario:
   - Envía un mensaje a [@userinfobot](https://t.me/userinfobot)
   - Copia tu ID de usuario numérico

3. Configura:

```json
{
  "platforms": {
    "telegram": {
      "enabled": true,
      "token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
      "allowed_users": ["your_username", "123456789"],
      "allowed_chats": ["-100123456789"]
    }
  }
}
```

**Opciones de Seguridad:**
- `allowed_users` — Nombres de usuario o IDs de usuario que pueden interactuar con el bot
- `allowed_chats` — IDs de chat grupal donde el bot responde (obtén mediante [@getidsbot](https://t.me/getidsbot))

### Discord

1. Crea una aplicación de Discord:
   - Visita el [Portal de Desarrolladores de Discord](https://discord.com/developers/applications)
   - Haz clic en "New Application" y nómbrala
   - Ve a la sección "Bot" y haz clic en "Add Bot"
   - Copia el token

2. Habilita los intents requeridos:
   - En la sección Bot, habilita "Message Content Intent"
   - Si usas filtrado de usuarios, habilita "Server Members Intent"

3. Invita el bot a tu servidor:
   - Ve a OAuth2 → URL Generator
   - Selecciona scopes: `bot`
   - Selecciona permisos: `Send Messages`, `Read Message History`
   - Usa la URL generada para invitar

4. Configura:

```json
{
  "platforms": {
    "discord": {
      "enabled": true,
      "token": "MTIzNDU2Nzg5MDEyMzQ1Njc4.XXXXXX.XXXXXXXXXXXXXXXXXXXXXXXX",
      "allowed_users": ["user_id_1", "user_id_2"],
      "allowed_channels": ["channel_id_1"],
      "allowed_guilds": ["guild_id_1"]
    }
  }
}
```

**Obtener IDs:** Habilita el modo desarrollador en la configuración de Discord, luego haz clic derecho en usuarios/canales/servidores para copiar IDs.

### Slack

1. Crea una aplicación de Slack:
   - Visita [Slack API](https://api.slack.com/apps)
   - Haz clic en "Create New App" → "From scratch"
   - Nombra tu aplicación y selecciona el workspace

2. Habilita Socket Mode:
   - Ve a "Socket Mode" y habilítalo
   - Genera un token a nivel de aplicación con scope `connections:write`
   - Copia el token (comienza con `xapp-`)

3. Configura el Bot Token:
   - Ve a "OAuth & Permissions"
   - Agrega scopes: `chat:write`, `channels:history`, `groups:history`, `im:history`, `mpim:history`
   - Instala en el workspace y copia el Bot Token (comienza con `xoxb-`)

4. Habilita eventos:
   - Ve a "Event Subscriptions" y habilítalo
   - Suscríbete a: `message.channels`, `message.groups`, `message.im`, `message.mpim`

5. Configura:

```json
{
  "platforms": {
    "slack": {
      "enabled": true,
      "bot_token": "xoxb-xxx-xxx-xxx",
      "app_token": "xapp-xxx-xxx-xxx",
      "allowed_users": ["U12345678"],
      "allowed_channels": ["C12345678"]
    }
  }
}
```

### Lark/飞书

1. Crea una aplicación de Lark:
   - Visita [Lark Open Platform](https://open.larksuite.com/) o [飞书开放平台](https://open.feishu.cn/)
   - Crea una nueva aplicación
   - Copia el App ID y App Secret

2. Configura permisos:
   - Agrega el evento `im:message:receive_v1`
   - Agrega el permiso `im:message:send_v1`

3. Configura webhook:
   - Establece la URL de suscripción de eventos (o usa modo WebSocket)

4. Configura:

```json
{
  "platforms": {
    "lark": {
      "enabled": true,
      "app_id": "cli_xxxxx",
      "app_secret": "xxxxxxxxxxxxx",
      "allowed_users": ["ou_xxxxx"],
      "allowed_chats": ["oc_xxxxx"]
    }
  }
}
```

### Facebook Messenger

1. Crea una aplicación de Facebook:
   - Visita [Facebook Developers](https://developers.facebook.com/)
   - Crea una nueva aplicación de tipo "Business"
   - Agrega el producto "Messenger"

2. Configura Messenger:
   - Genera un Page Access Token
   - Configura el webhook con un token de verificación
   - Suscríbete al evento `messages`

3. Configura:

```json
{
  "platforms": {
    "fbmessenger": {
      "enabled": true,
      "page_token": "EAAxxxxx",
      "verify_token": "your_verify_token",
      "app_secret": "xxxxx",
      "allowed_users": ["psid_1", "psid_2"]
    }
  }
}
```

**Nota:** Facebook Messenger requiere una URL de webhook públicamente accesible. Considera usar servicios como ngrok para desarrollo.

## Notificaciones

Configura dónde el bot envía notificaciones:

```json
{
  "notify": {
    "default_platform": "telegram",
    "default_chat_id": "-100123456789",
    "quiet_hours_start": "23:00",
    "quiet_hours_end": "07:00",
    "quiet_hours_zone": "UTC"
  }
}
```

### Tipos de Notificaciones

- **Solicitudes de Aprobación** — Cuando Claude Code necesita permiso para una operación
- **Completación de Tareas** — Cuando una tarea se completa exitosamente
- **Errores** — Cuando una tarea falla o encuentra un error
- **Cambios de Estado** — Cuando los procesos se conectan/desconectan

### Horario de Silencio

Durante el horario de silencio, las notificaciones no urgentes se suprimen. Las solicitudes de aprobación siempre se envían.

## Mejores Prácticas de Seguridad

1. **Restringir Usuarios** — Siempre configura `allowed_users` para limitar quién puede controlar tus sesiones
2. **Usar Canales Privados** — Evita usar el bot en canales públicos
3. **Proteger Tokens** — Nunca hagas commit de tokens de bot en control de versiones
4. **Revisar Aprobaciones** — Revisa cuidadosamente las solicitudes de aprobación antes de aceptar

## Solución de Problemas

### El Bot No Responde

1. Verifica que el daemon esté en ejecución: `zen daemon status`
2. Valida la configuración del bot en la Web UI
3. Revisa los logs del daemon: `tail -f ~/.zen/zend.log`

### Problemas de Conexión

1. Verifica que el token sea correcto
2. Revisa la conexión de red
3. Para Slack/Discord, asegúrate de que los intents requeridos estén habilitados

### Los Procesos No Aparecen en la Lista

1. Asegúrate de que el proceso se inició con `zen` (no directamente con `claude`)
2. Verifica que la ruta del socket coincida con la configuración del bot
