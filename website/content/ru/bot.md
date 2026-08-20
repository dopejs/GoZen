---
sidebar_position: 10
title: Бот-шлюз
---

# Бот-шлюз

Наблюдайте за сессиями Claude Code и управляйте ими удалённо через мессенджеры. Бот подключается к запущенным процессам `zen` по IPC и позволяет:

- Смотреть подключённые процессы и их состояние
- Отправлять задачи конкретному процессу
- Получать уведомления о подтверждениях, ошибках и завершении
- Управлять задачами (пауза, продолжение, отмена)

## Поддерживаемые платформы

| Платформа | Что нужно настроить |
|----------|----------------|
| [Telegram](#telegram) | Токен от BotFather |
| [Discord](#discord) | Токен бот-приложения |
| [Slack](#slack) | Токены Bot и App (Socket Mode) |
| [Lark/Feishu](#larkfeishu) | App ID и Secret |
| [Facebook Messenger](#facebook-messenger) | Токен страницы и проверочный токен |

## Базовая настройка

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

## Команды бота

| Команда | Описание |
|---------|-------------|
| `list` | Список всех подключённых процессов |
| `status [name]` | Состояние процесса |
| `bind <name>` | Привязка к процессу для последующих команд |
| `pause [name]` | Пауза текущей задачи |
| `resume [name]` | Продолжение приостановленной задачи |
| `cancel [name]` | Отмена текущей задачи |
| `<name> <task>` | Отправка задачи процессу |
| `help` | Показать доступные команды |

### Естественный язык

Бот понимает запросы на естественном языке, на нескольких языках:

- «show me the status of gozen»
- «帮我看看 gozen 的状态»
- «list all processes»
- «pause the api project»

## Режимы взаимодействия

### Личные сообщения

`direct_message_mode` задаёт поведение бота в личных сообщениях:

- `"always"` — отвечать всегда (упоминание не нужно)
- `"mention"` — отвечать только при упоминании

### Сообщения в каналах

`channel_mode` задаёт поведение в групповых чатах:

- `"always"` — отвечать на все сообщения
- `"mention"` — отвечать только при упоминании (рекомендуется)

### Ключевые слова для упоминания

Задайте, что вызывает бота:

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## Псевдонимы проектов

Задайте короткие имена для проектов:

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

Затем используйте их в командах:

```
api run tests
web build production
status backend
```

## Настройка по платформам

### Telegram

1. Создайте бота через [@BotFather](https://t.me/botfather):
   - Отправьте `/newbot` и следуйте подсказкам
   - Скопируйте токен (например, `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. Узнайте свой user ID:
   - Напишите [@userinfobot](https://t.me/userinfobot)
   - Скопируйте числовой идентификатор

3. Настройте:

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

**Параметры безопасности:**
- `allowed_users` — имена или идентификаторы пользователей, которым разрешено обращаться к боту
- `allowed_chats` — идентификаторы групп, где бот отвечает (узнать через [@getidsbot](https://t.me/getidsbot))

### Discord

1. Создайте приложение Discord:
   - Откройте [портал разработчика Discord](https://discord.com/developers/applications)
   - Нажмите «New Application» и задайте имя
   - Перейдите в раздел «Bot» и нажмите «Add Bot»
   - Скопируйте токен

2. Включите нужные intents:
   - В разделе Bot включите «Message Content Intent»
   - Включите «Server Members Intent», если используете фильтр по пользователям

3. Пригласите бота на сервер:
   - Откройте OAuth2 → URL Generator
   - Выберите scope: `bot`
   - Выберите права: `Send Messages`, `Read Message History`
   - Пригласите бота по полученной ссылке

4. Настройте:

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

**Как узнать ID:** включите режим разработчика в настройках Discord и копируйте идентификаторы правым щелчком по пользователю, каналу или серверу.

### Slack

1. Создайте приложение Slack:
   - Откройте [Slack API](https://api.slack.com/apps)
   - Нажмите «Create New App» → «From scratch»
   - Задайте имя и выберите рабочее пространство

2. Включите Socket Mode:
   - Перейдите в «Socket Mode» и включите его
   - Создайте App-Level Token со scope `connections:write`
   - Скопируйте токен (начинается с `xapp-`)

3. Настройте Bot Token:
   - Перейдите в «OAuth & Permissions»
   - Добавьте scopes: `chat:write`, `channels:history`, `groups:history`, `im:history`, `mpim:history`
   - Установите приложение и скопируйте Bot Token (начинается с `xoxb-`)

4. Включите события:
   - Перейдите в «Event Subscriptions» и включите их
   - Подпишитесь на: `message.channels`, `message.groups`, `message.im`, `message.mpim`

5. Настройте:

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

1. Создайте приложение Lark:
   - Откройте [Lark Open Platform](https://open.larksuite.com/) или [Feishu Open Platform](https://open.feishu.cn/)
   - Создайте новое приложение
   - Скопируйте App ID и App Secret

2. Настройте права:
   - Добавьте событие `im:message:receive_v1`
   - Добавьте право `im:message:send_v1`

3. Настройте webhook:
   - Укажите URL подписки на события (или используйте режим WebSocket)

4. Настройте:

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

1. Создайте приложение Facebook:
   - Откройте [Facebook Developers](https://developers.facebook.com/)
   - Создайте приложение типа «Business»
   - Добавьте продукт «Messenger»

2. Настройте Messenger:
   - Создайте Page Access Token
   - Настройте webhook с проверочным токеном
   - Подпишитесь на событие `messages`

3. Настройте:

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

**Примечание:** Facebook Messenger требует публично доступный URL webhook. Для разработки подойдёт сервис вроде ngrok.

## Уведомления

Укажите, куда бот отправляет уведомления:

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

### Типы уведомлений

- **Запросы подтверждения** — когда Claude Code нужно разрешение на действие
- **Завершение задачи** — когда задача успешно закончилась
- **Ошибки** — когда задача упала или столкнулась с ошибкой
- **Смена состояния** — когда процесс подключается или отключается

### Тихие часы

В тихие часы несрочные уведомления подавляются. Запросы подтверждения отправляются всегда.

## Рекомендации по безопасности

1. **Ограничьте пользователей** — всегда задавайте `allowed_users`, чтобы ограничить доступ к вашим сессиям
2. **Используйте приватные каналы** — не используйте бота в публичных каналах
3. **Берегите токены** — никогда не коммитьте токены бота в репозиторий
4. **Проверяйте подтверждения** — внимательно читайте запрос, прежде чем согласиться

## Устранение неполадок

### Бот не отвечает

1. Проверьте, запущен ли демон: `zen daemon status`
2. Проверьте настройку бота в веб-интерфейсе
3. Посмотрите журналы демона: `tail -f ~/.zen/zend.log`

### Проблемы с подключением

1. Проверьте правильность токенов
2. Проверьте сетевую доступность
3. Для Slack и Discord убедитесь, что нужные intents включены

### Процесс не появляется в списке

1. Убедитесь, что процесс запущен через `zen`, а не напрямую через `claude`
2. Проверьте, что путь к сокету совпадает с настройкой бота
