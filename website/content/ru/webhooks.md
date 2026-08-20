---
sidebar_position: 13
title: Webhooks
---

# Webhooks

Получайте уведомления в реальном времени — предупреждения о бюджете, изменения состояния провайдеров, ежедневные сводки — через Slack, Discord или собственный webhook.

## Возможности

- **Несколько форматов** — Slack, Discord или обычный JSON
- **Фильтрация событий** — подписка на нужные типы событий
- **Свои заголовки** — добавьте аутентификацию или собственные заголовки
- **Асинхронная отправка** — доставка не блокирует запросы
- **Автоматическое оформление** — сообщения с эмодзи и цветами
- **Проверка** — протестируйте настройку до включения

## Настройка

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": [
        "budget_warning",
        "budget_exceeded",
        "provider_down",
        "provider_up",
        "failover",
        "daily_summary"
      ],
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  ]
}
```

## Типы событий

| Событие | Описание | Когда срабатывает |
|-------|-------------|----------------|
| `budget_warning` | Достигнут порог бюджета | Когда расходы достигают 80 % лимита |
| `budget_exceeded` | Лимит бюджета превышен | Когда расходы превышают заданный лимит |
| `provider_down` | Провайдер стал нездоровым | Когда доля успешных запросов падает ниже 70 % |
| `provider_up` | Провайдер восстановился | Когда нездоровый провайдер снова становится здоровым |
| `failover` | Запрос переключён | Когда запрос уходит к резервному провайдеру |
| `daily_summary` | Ежедневная сводка использования | Раз в сутки, в полночь UTC |

## Форматы webhook

### Slack

Определяется автоматически, если URL содержит `slack.com`.

**Пример сообщения:**
```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

**Формат:**
```json
{
  "text": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)"
      }
    }
  ]
}
```

### Discord

Определяется автоматически, если URL содержит `discord.com`.

**Пример embed:**
- **Заголовок:** budget_warning
- **Описание:** ⚠️ Предупреждение о бюджете: дневной бюджет на 85,0 % ($8.50 / $10.00)
- **Цвет:** янтарный (#FBBF24)
- **Время:** 2026-03-05T10:30:00Z

**Формат:**
```json
{
  "content": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
  "embeds": [
    {
      "title": "budget_warning",
      "description": "⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)",
      "timestamp": "2026-03-05T10:30:00Z",
      "color": 16432932
    }
  ]
}
```

### Обычный JSON

Используется для всех остальных URL.

**Формат:**
```json
{
  "event": "budget_warning",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "period": "daily",
    "spent": 8.5,
    "limit": 10.0,
    "percent": 85.0,
    "project": ""
  }
}
```

## Структуры данных событий

### Предупреждение о бюджете / превышение

```json
{
  "event": "budget_warning",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "period": "daily",
    "spent": 8.5,
    "limit": 10.0,
    "percent": 85.0,
    "action": "warn",
    "project": "my-project"
  }
}
```

### Провайдер недоступен / восстановился

```json
{
  "event": "provider_down",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "provider": "anthropic-primary",
    "status": "unhealthy",
    "error": "connection timeout",
    "latency_ms": 0
  }
}
```

### Переключение

```json
{
  "event": "failover",
  "timestamp": "2026-03-05T10:30:00Z",
  "data": {
    "from_provider": "anthropic-primary",
    "to_provider": "anthropic-backup",
    "reason": "rate limit exceeded",
    "session_id": "sess_abc123"
  }
}
```

### Ежедневная сводка

```json
{
  "event": "daily_summary",
  "timestamp": "2026-03-05T00:00:00Z",
  "data": {
    "date": "2026-03-04",
    "total_cost": 25.50,
    "total_requests": 150,
    "total_input_tokens": 125000,
    "total_output_tokens": 35000,
    "by_provider": {
      "anthropic": 18.20,
      "openai": 7.30
    }
  }
}
```

## Настройка по платформам

### Slack

1. Откройте [Slack API](https://api.slack.com/apps)
2. Создайте приложение или выберите существующее
3. Включите «Incoming Webhooks»
4. Добавьте webhook в рабочее пространство
5. Скопируйте URL (он начинается с `https://hooks.slack.com/`)

**Настройка:**
```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
      "events": ["budget_warning", "provider_down"]
    }
  ]
}
```

### Discord

1. Откройте настройки сервера Discord
2. Перейдите в «Интеграции» → «Вебхуки»
3. Нажмите «New Webhook»
4. Выберите канал и скопируйте URL

**Настройка:**
```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://discord.com/api/webhooks/123456789/XXXXXXXXXXXXXXXXXXXX",
      "events": ["budget_exceeded", "failover"]
    }
  ]
}
```

### Собственный webhook

Для своих интеграций используйте обычный JSON:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://your-server.com/webhook",
      "events": ["budget_warning", "daily_summary"],
      "headers": {
        "Authorization": "Bearer YOUR_SECRET_TOKEN",
        "X-Custom-Header": "value"
      }
    }
  ]
}
```

## Настройка в веб-интерфейсе

Настройки webhook находятся по адресу `http://localhost:19840/settings`:

1. Откройте вкладку «Webhooks»
2. Нажмите «Add Webhook»
3. Введите URL
4. Выберите события для подписки
5. (Необязательно) Добавьте свои заголовки
6. Нажмите «Test», чтобы проверить настройку
7. Нажмите «Save»

## Точки API

### Список webhook

```bash
GET /api/v1/webhooks
```

### Добавить webhook

```bash
POST /api/v1/webhooks
Content-Type: application/json

{
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["budget_warning", "provider_down"]
}
```

### Обновить webhook

```bash
PUT /api/v1/webhooks/{id}
Content-Type: application/json

{
  "enabled": false
}
```

### Удалить webhook

```bash
DELETE /api/v1/webhooks/{id}
```

### Проверить webhook

```bash
POST /api/v1/webhooks/{id}/test
```

Отправляет тестовое сообщение для проверки настройки.

## Примеры сообщений

### Предупреждение о бюджете (Slack)

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### Бюджет превышен (Discord)

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### Провайдер недоступен (Slack)

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### Провайдер восстановился (Discord)

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### Переключение (Slack)

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### Ежедневная сводка (Discord)

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## Рекомендации

1. **Разделяйте webhook** — заводите отдельные адреса под разные типы событий
2. **Тестируйте до включения** — всегда проверяйте настройку перед сохранением
3. **Защищайте свои webhook** — используйте HTTPS и заголовки аутентификации
4. **Следите за сбоями** — если уведомления прекратились, смотрите журналы демона
5. **Не кладите секреты в URL** — никаких ключей API и токенов в адресе webhook
6. **Настройте оповещения** — подпишитесь на критичные события `budget_exceeded` и `provider_down`

## Устранение неполадок

### Webhook не получает сообщений

1. Убедитесь, что webhook включён в конфигурации
2. Проверьте правильность URL (протестируйте curl)
3. Проверьте, что события настроены верно
4. Поищите ошибки webhook в журналах демона: `tail -f ~/.zen/zend.log`
5. Протестируйте через API: `POST /api/v1/webhooks/{id}/test`

### Не работает webhook Slack

1. Проверьте, что URL начинается с `https://hooks.slack.com/`
2. Проверьте, не отозван ли webhook в настройках Slack
3. Убедитесь, что в рабочем пространстве не отключены входящие webhook
4. Проверьте через curl:
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### Не работает webhook Discord

1. Проверьте, что URL начинается с `https://discord.com/api/webhooks/`
2. Проверьте, не удалён ли webhook в настройках Discord
3. Убедитесь, что у бота есть право писать в канал
4. Проверьте через curl:
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### Не работает собственный webhook

1. Убедитесь, что адрес доступен (проверьте curl)
2. Проверьте заголовки аутентификации
3. Убедитесь, что адрес принимает POST-запросы
4. Проверьте, что он возвращает код 2xx
5. Посмотрите его журналы на предмет ошибок

## Вопросы безопасности

1. **Берегите URL webhook** — относитесь к ним как к секретам
2. **Используйте HTTPS** — всегда для адресов webhook
3. **Проверяйте подписи** — реализуйте проверку подписи для своих webhook
4. **Ограничивайте частоту** — введите rate limiting на своих адресах
5. **Не пишите секреты в журнал** — не логируйте полные полезные нагрузки

## Расширенная настройка

### Условные webhook

Отправляйте разные события на разные адреса:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/CRITICAL/ALERTS",
      "events": ["budget_exceeded", "provider_down"]
    },
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/DAILY/REPORTS",
      "events": ["daily_summary"]
    },
    {
      "enabled": true,
      "url": "https://discord.com/api/webhooks/MONITORING",
      "events": ["failover", "provider_up"]
    }
  ]
}
```

### Свои заголовки для аутентификации

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://your-server.com/webhook",
      "events": ["budget_warning"],
      "headers": {
        "Authorization": "Bearer YOUR_SECRET_TOKEN",
        "X-API-Key": "your-api-key",
        "X-Webhook-Source": "gozen"
      }
    }
  ]
}
```
