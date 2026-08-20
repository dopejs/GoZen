---
sidebar_position: 10
title: Passerelle de bots
---

# Passerelle de bots

Supervisez et pilotez vos sessions Claude Code à distance depuis une messagerie. Le bot se connecte aux processus `zen` en cours via IPC et vous permet de :

- Voir les processus connectés et leur état
- Envoyer des tâches à un processus précis
- Recevoir des notifications d’approbation, d’erreur et de fin de tâche
- Piloter les tâches (pause, reprise, annulation)

## Plateformes prises en charge

| Plateforme | Configuration requise |
|----------|----------------|
| [Telegram](#telegram) | Jeton BotFather |
| [Discord](#discord) | Jeton d’application bot |
| [Slack](#slack) | Jetons Bot et App (mode Socket) |
| [Lark/Feishu](#larkfeishu) | App ID et Secret |
| [Facebook Messenger](#facebook-messenger) | Jeton de page et jeton de vérification |

## Configuration de base

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

## Commandes du bot

| Commande | Description |
|---------|-------------|
| `list` | Liste tous les processus connectés |
| `status [name]` | Affiche l’état d’un processus |
| `bind <name>` | Se lie à un processus pour les commandes suivantes |
| `pause [name]` | Met la tâche en cours en pause |
| `resume [name]` | Reprend une tâche en pause |
| `cancel [name]` | Annule la tâche en cours |
| `<name> <task>` | Envoie une tâche à un processus |
| `help` | Affiche les commandes disponibles |

### Langage naturel

Le bot comprend des requêtes en langage naturel, dans plusieurs langues :

- « show me the status of gozen »
- « 帮我看看 gozen 的状态 »
- « list all processes »
- « pause the api project »

## Modes d’interaction

### Messages privés

`direct_message_mode` définit la façon dont le bot répond en message privé :

- `"always"` — répond toujours (aucune mention nécessaire)
- `"mention"` — ne répond que s’il est mentionné

### Messages de salon

`channel_mode` définit le comportement dans les discussions de groupe :

- `"always"` — répond à tous les messages
- `"mention"` — ne répond que s’il est mentionné (recommandé)

### Mots-clés de mention

Configurez ce qui déclenche le bot :

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## Alias de projets

Définissez des noms courts pour vos projets :

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

Puis utilisez-les dans les commandes :

```
api run tests
web build production
status backend
```

## Mise en place par plateforme

### Telegram

1. Créez un bot avec [@BotFather](https://t.me/botfather) :
   - Envoyez `/newbot` et suivez les instructions
   - Copiez le jeton (par exemple `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. Récupérez votre identifiant utilisateur :
   - Écrivez à [@userinfobot](https://t.me/userinfobot)
   - Copiez votre identifiant numérique

3. Configurez :

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

**Options de sécurité :**
- `allowed_users` — noms d’utilisateur ou identifiants autorisés à parler au bot
- `allowed_chats` — identifiants des groupes où le bot répond (via [@getidsbot](https://t.me/getidsbot))

### Discord

1. Créez une application Discord :
   - Rendez-vous sur le [portail développeur Discord](https://discord.com/developers/applications)
   - Cliquez sur « New Application » et donnez-lui un nom
   - Allez dans la section « Bot » et cliquez sur « Add Bot »
   - Copiez le jeton

2. Activez les intents nécessaires :
   - Dans la section Bot, activez « Message Content Intent »
   - Activez « Server Members Intent » si vous filtrez par utilisateur

3. Invitez le bot sur votre serveur :
   - Allez dans OAuth2 → URL Generator
   - Cochez la portée : `bot`
   - Cochez les permissions : `Send Messages`, `Read Message History`
   - Utilisez l’URL générée pour l’inviter

4. Configurez :

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

**Récupérer les identifiants :** activez le mode développeur dans les réglages Discord, puis faites un clic droit sur un utilisateur, un salon ou un serveur pour copier son identifiant.

### Slack

1. Créez une application Slack :
   - Rendez-vous sur [Slack API](https://api.slack.com/apps)
   - Cliquez sur « Create New App » → « From scratch »
   - Nommez l’application et choisissez l’espace de travail

2. Activez le mode Socket :
   - Allez dans « Socket Mode » et activez-le
   - Générez un App-Level Token avec la portée `connections:write`
   - Copiez le jeton (il commence par `xapp-`)

3. Configurez le jeton du bot :
   - Allez dans « OAuth & Permissions »
   - Ajoutez les portées : `chat:write`, `channels:history`, `groups:history`, `im:history`, `mpim:history`
   - Installez l’application et copiez le Bot Token (il commence par `xoxb-`)

4. Activez les événements :
   - Allez dans « Event Subscriptions » et activez-les
   - Abonnez-vous à : `message.channels`, `message.groups`, `message.im`, `message.mpim`

5. Configurez :

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

1. Créez une application Lark :
   - Rendez-vous sur la [plateforme ouverte Lark](https://open.larksuite.com/) ou [Feishu](https://open.feishu.cn/)
   - Créez une application
   - Copiez l’App ID et l’App Secret

2. Configurez les permissions :
   - Ajoutez l’événement `im:message:receive_v1`
   - Ajoutez la permission `im:message:send_v1`

3. Configurez le webhook :
   - Renseignez l’URL d’abonnement aux événements (ou utilisez le mode WebSocket)

4. Configurez :

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

1. Créez une application Facebook :
   - Rendez-vous sur [Facebook Developers](https://developers.facebook.com/)
   - Créez une application de type « Business »
   - Ajoutez le produit « Messenger »

2. Configurez Messenger :
   - Générez un Page Access Token
   - Configurez le webhook avec un jeton de vérification
   - Abonnez-vous à l’événement `messages`

3. Configurez :

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

**Remarque :** Facebook Messenger exige une URL de webhook accessible publiquement. Pensez à un service comme ngrok en développement.

## Notifications

Indiquez où le bot envoie ses notifications :

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

### Types de notifications

- **Demandes d’approbation** — quand Claude Code a besoin d’une autorisation
- **Fin de tâche** — quand une tâche se termine correctement
- **Erreurs** — quand une tâche échoue ou rencontre une erreur
- **Changements d’état** — quand un processus se connecte ou se déconnecte

### Heures calmes

Pendant les heures calmes, les notifications non urgentes sont supprimées. Les demandes d’approbation sont toujours envoyées.

## Bonnes pratiques de sécurité

1. **Restreignez les utilisateurs** — configurez toujours `allowed_users` pour limiter qui pilote vos sessions
2. **Utilisez des salons privés** — évitez le bot dans les salons publics
3. **Protégez les jetons** — ne versionnez jamais un jeton de bot
4. **Relisez les approbations** — examinez soigneusement chaque demande avant d’accepter

## Dépannage

### Le bot ne répond pas

1. Vérifiez que le daemon tourne : `zen daemon status`
2. Vérifiez la configuration du bot dans l’interface web
3. Consultez les journaux du daemon : `tail -f ~/.zen/zend.log`

### Problèmes de connexion

1. Vérifiez que les jetons sont corrects
2. Vérifiez la connectivité réseau
3. Pour Slack et Discord, vérifiez que les intents nécessaires sont activés

### Un processus n’apparaît pas dans la liste

1. Vérifiez que le processus a été lancé avec `zen` (et non directement avec `claude`)
2. Vérifiez que le chemin du socket correspond à la configuration du bot
