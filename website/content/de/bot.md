---
sidebar_position: 10
title: Bot-Gateway
---

# Bot-Gateway

Beobachte und steuere deine Claude-Code-Sitzungen aus der Ferne über Chat-Plattformen. Der Bot verbindet sich per IPC mit laufenden `zen`-Prozessen und erlaubt dir:

- Verbundene Prozesse und ihren Status anzusehen
- Aufgaben an einen bestimmten Prozess zu schicken
- Benachrichtigungen zu Freigaben, Fehlern und Abschlüssen zu erhalten
- Aufgaben zu steuern (anhalten, fortsetzen, abbrechen)

## Unterstützte Plattformen

| Plattform | Nötige Einrichtung |
|----------|----------------|
| [Telegram](#telegram) | BotFather-Token |
| [Discord](#discord) | Token der Bot-Anwendung |
| [Slack](#slack) | Bot- und App-Token (Socket Mode) |
| [Lark/Feishu](#larkfeishu) | App ID und Secret |
| [Facebook Messenger](#facebook-messenger) | Page-Token und Verify-Token |

## Grundkonfiguration

```json
{
  "bot": {
    "enabled": true,
    "socket_path": "/tmp/zen-bot.sock",
    "platforms": {
      // Platform-specific config (see below)
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

## Bot-Befehle

| Befehl | Beschreibung |
|---------|-------------|
| `list` | Listet alle verbundenen Prozesse auf |
| `status [name]` | Zeigt den Status eines Prozesses |
| `bind <name>` | Bindet an einen Prozess für die folgenden Befehle |
| `pause [name]` | Hält die laufende Aufgabe an |
| `resume [name]` | Setzt eine angehaltene Aufgabe fort |
| `cancel [name]` | Bricht die laufende Aufgabe ab |
| `<name> <task>` | Schickt eine Aufgabe an einen Prozess |
| `help` | Zeigt die verfügbaren Befehle |

### Natürliche Sprache

Der Bot versteht Anfragen in natürlicher Sprache, in mehreren Sprachen:

- „show me the status of gozen“
- „帮我看看 gozen 的状态“
- „list all processes“
- „pause the api project“

## Interaktionsmodi

### Direktnachrichten

`direct_message_mode` legt fest, wie der Bot in Direktnachrichten antwortet:

- `"always"` — antwortet immer (keine Erwähnung nötig)
- `"mention"` — antwortet nur, wenn er erwähnt wird

### Kanalnachrichten

`channel_mode` legt das Verhalten in Gruppenchats fest:

- `"always"` — antwortet auf alle Nachrichten
- `"mention"` — antwortet nur bei Erwähnung (empfohlen)

### Erwähnungs-Schlüsselwörter

Lege fest, was den Bot auslöst:

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## Projekt-Aliasse

Vergib Kurznamen für deine Projekte:

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

Dann nutze sie in Befehlen:

```
api run tests
web build production
status backend
```

## Einrichtung je Plattform

### Telegram

1. Erstelle einen Bot über [@BotFather](https://t.me/botfather):
   - Sende `/newbot` und folge den Anweisungen
   - Kopiere den Token (z. B. `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. Ermittle deine Benutzer-ID:
   - Schreibe [@userinfobot](https://t.me/userinfobot) an
   - Kopiere deine numerische Benutzer-ID

3. Konfiguriere:

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

**Sicherheitsoptionen:**
- `allowed_users` — Benutzernamen oder IDs, die mit dem Bot sprechen dürfen
- `allowed_chats` — IDs der Gruppen, in denen der Bot antwortet (über [@getidsbot](https://t.me/getidsbot))

### Discord

1. Lege eine Discord-Anwendung an:
   - Öffne das [Discord Developer Portal](https://discord.com/developers/applications)
   - Klicke auf „New Application“ und vergib einen Namen
   - Gehe zum Bereich „Bot“ und klicke auf „Add Bot“
   - Kopiere den Token

2. Aktiviere die nötigen Intents:
   - Aktiviere im Bot-Bereich „Message Content Intent“
   - Aktiviere „Server Members Intent“, wenn du nach Nutzern filterst

3. Lade den Bot auf deinen Server ein:
   - Gehe zu OAuth2 → URL Generator
   - Wähle den Scope: `bot`
   - Wähle die Rechte: `Send Messages`, `Read Message History`
   - Nutze die erzeugte URL zum Einladen

4. Konfiguriere:

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

**IDs ermitteln:** Aktiviere den Entwicklermodus in den Discord-Einstellungen und kopiere IDs per Rechtsklick auf Nutzer, Kanäle oder Server.

### Slack

1. Lege eine Slack-App an:
   - Öffne die [Slack API](https://api.slack.com/apps)
   - Klicke auf „Create New App“ → „From scratch“
   - Benenne die App und wähle den Workspace

2. Aktiviere den Socket Mode:
   - Gehe zu „Socket Mode“ und schalte ihn ein
   - Erzeuge ein App-Level-Token mit dem Scope `connections:write`
   - Kopiere den Token (beginnt mit `xapp-`)

3. Konfiguriere den Bot-Token:
   - Gehe zu „OAuth & Permissions“
   - Ergänze die Scopes: `chat:write`, `channels:history`, `groups:history`, `im:history`, `mpim:history`
   - Installiere die App im Workspace und kopiere den Bot-Token (beginnt mit `xoxb-`)

4. Aktiviere Events:
   - Gehe zu „Event Subscriptions“ und aktiviere sie
   - Abonniere: `message.channels`, `message.groups`, `message.im`, `message.mpim`

5. Konfiguriere:

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

### Lark/Feishu

1. Lege eine Lark-App an:
   - Öffne die [Lark Open Platform](https://open.larksuite.com/) oder [Feishu Open Platform](https://open.feishu.cn/)
   - Erstelle eine neue App
   - Kopiere App ID und App Secret

2. Konfiguriere die Rechte:
   - Ergänze das Ereignis `im:message:receive_v1`
   - Ergänze die Berechtigung `im:message:send_v1`

3. Konfiguriere den Webhook:
   - Trage die URL für das Event-Abonnement ein (oder nutze den WebSocket-Modus)

4. Konfiguriere:

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

1. Lege eine Facebook-App an:
   - Öffne [Facebook Developers](https://developers.facebook.com/)
   - Erstelle eine App vom Typ „Business“
   - Füge das Produkt „Messenger“ hinzu

2. Konfiguriere Messenger:
   - Erzeuge ein Page Access Token
   - Richte den Webhook mit einem Verify-Token ein
   - Abonniere das Ereignis `messages`

3. Konfiguriere:

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

**Hinweis:** Facebook Messenger benötigt eine öffentlich erreichbare Webhook-URL. In der Entwicklung bietet sich ein Dienst wie ngrok an.

## Benachrichtigungen

Lege fest, wohin der Bot Benachrichtigungen schickt:

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

### Arten von Benachrichtigungen

- **Freigabeanfragen** — wenn Claude Code eine Erlaubnis für eine Aktion braucht
- **Aufgabenabschluss** — wenn eine Aufgabe erfolgreich endet
- **Fehler** — wenn eine Aufgabe scheitert oder auf einen Fehler läuft
- **Statuswechsel** — wenn sich ein Prozess verbindet oder trennt

### Ruhezeiten

Während der Ruhezeiten werden nicht dringende Benachrichtigungen unterdrückt. Freigabeanfragen werden immer zugestellt.

## Sicherheitsempfehlungen

1. **Nutzer einschränken** — konfiguriere immer `allowed_users`, um den Zugriff auf deine Sitzungen zu begrenzen
2. **Private Kanäle nutzen** — verwende den Bot nicht in öffentlichen Kanälen
3. **Tokens schützen** — Bot-Tokens niemals in die Versionsverwaltung geben
4. **Freigaben prüfen** — Freigabeanfragen sorgfältig lesen, bevor du zustimmst

## Fehlersuche

### Der Bot antwortet nicht

1. Prüfen, ob der Daemon läuft: `zen daemon status`
2. Die Bot-Konfiguration in der Weboberfläche prüfen
3. Die Daemon-Protokolle ansehen: `tail -f ~/.zen/zend.log`

### Verbindungsprobleme

1. Prüfen, ob die Tokens stimmen
2. Die Netzwerkverbindung prüfen
3. Bei Slack und Discord sicherstellen, dass die nötigen Intents aktiviert sind

### Ein Prozess erscheint nicht in der Liste

1. Sicherstellen, dass der Prozess mit `zen` gestartet wurde (nicht direkt mit `claude`)
2. Prüfen, ob der Socket-Pfad zur Bot-Konfiguration passt
