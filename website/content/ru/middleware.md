---
sidebar_position: 15
title: Конвейер middleware (БЕТА)
---

# Конвейер middleware (БЕТА)

:::warning Функция в стадии БЕТА
Конвейер middleware находится в бете. По умолчанию он выключен и требует явной настройки.
:::

Расширяйте GoZen подключаемыми middleware: преобразование запросов и ответов, журналирование, ограничение частоты и своя обработка.

## Возможности

- **Подключаемая архитектура** — добавляйте свою логику, не трогая ядро
- **Выполнение по приоритету** — управляйте порядком выполнения
- **Хуки запроса и ответа** — вмешивайтесь до отправки и после получения
- **Встроенные middleware** — вставка контекста, журналирование, ограничение частоты, сжатие
- **Загрузчик плагинов** — загружайте middleware из локальных файлов или по URL
- **Обработка ошибок** — аккуратная обработка с запасным поведением

## Архитектура

```
Client Request
    ↓
[Middleware 1: Priority 100]
    ↓
[Middleware 2: Priority 200]
    ↓
[Middleware 3: Priority 300]
    ↓
Provider API
    ↓
[Middleware 3: Response]
    ↓
[Middleware 2: Response]
    ↓
[Middleware 1: Response]
    ↓
Client Response
```

## Настройка

### Включение конвейера middleware

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "context-injection",
        "enabled": true,
        "priority": 100,
        "config": {}
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info"
        }
      }
    ]
  }
}
```

**Параметры:**

| Параметр | Описание |
|--------|-------------|
| `enabled` | Включает конвейер middleware |
| `pipeline` | Массив конфигураций middleware |
| `name` | Идентификатор middleware |
| `priority` | Порядок выполнения (меньше — раньше) |
| `config` | Конфигурация конкретного middleware |

## Встроенные middleware

### 1. Вставка контекста

Вставляет собственный контекст в запросы.

```json
{
  "name": "context-injection",
  "enabled": true,
  "priority": 100,
  "config": {
    "system_prompt": "You are a helpful coding assistant.",
    "metadata": {
      "session_id": "sess_123",
      "user_id": "user_456"
    }
  }
}
```

**Сценарии:**
- Добавить системные подсказки
- Вставить метаданные сессии
- Добавить пользовательский контекст

### 2. Журнал запросов

Пишет в журнал все запросы и ответы.

```json
{
  "name": "request-logger",
  "enabled": true,
  "priority": 200,
  "config": {
    "log_level": "info",
    "log_body": false,
    "log_headers": true
  }
}
```

**Сценарии:**
- Отладка
- Аудит
- Наблюдение за производительностью

### 3. Ограничитель частоты

Ограничивает частоту запросов по провайдеру или глобально.

```json
{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60,
    "burst": 10,
    "per_provider": true
  }
}
```

**Сценарии:**
- Избежать ошибок из-за лимитов
- Контролировать расход API
- Защититься от злоупотреблений

### 4. Сжатие (БЕТА)

Сжимает контекст, когда число токенов превышает порог.

```json
{
  "name": "compression",
  "enabled": true,
  "priority": 400,
  "config": {
    "threshold_tokens": 50000,
    "target_tokens": 20000
  }
}
```

Подробности — в разделе [Сжатие контекста](./compression.md).

### 5. Память сессии (БЕТА)

Сохраняет память беседы между сессиями.

```json
{
  "name": "session-memory",
  "enabled": true,
  "priority": 150,
  "config": {
    "max_memories": 100,
    "ttl_hours": 24,
    "storage": "sqlite"
  }
}
```

**Сценарии:**
- Запоминать предпочтения пользователя
- Вести историю беседы
- Сохранять контекст между сессиями

### 6. Оркестрация (БЕТА)

Направляет запросы нескольким провайдерам и объединяет ответы.

```json
{
  "name": "orchestration",
  "enabled": true,
  "priority": 500,
  "config": {
    "strategy": "parallel",
    "providers": ["anthropic", "openai"],
    "consensus": "longest"
  }
}
```

**Сценарии:**
- Сравнивать ответы моделей
- Резервирование для критичных запросов
- Повышение качества через консенсус

## Собственные middleware

### Интерфейс middleware

```go
type Middleware interface {
    Name() string
    Priority() int
    ProcessRequest(ctx *RequestContext) error
    ProcessResponse(ctx *ResponseContext) error
}

type RequestContext struct {
    Provider  string
    Model     string
    Messages  []Message
    Metadata  map[string]interface{}
}

type ResponseContext struct {
    Provider  string
    Model     string
    Response  *APIResponse
    Latency   time.Duration
    Metadata  map[string]interface{}
}
```

### Пример: вставка своего заголовка

```go
package main

import (
    "github.com/dopejs/gozen/internal/middleware"
)

type CustomHeaderMiddleware struct {
    headers map[string]string
}

func (m *CustomHeaderMiddleware) Name() string {
    return "custom-headers"
}

func (m *CustomHeaderMiddleware) Priority() int {
    return 250
}

func (m *CustomHeaderMiddleware) ProcessRequest(ctx *middleware.RequestContext) error {
    for k, v := range m.headers {
        ctx.Metadata[k] = v
    }
    return nil
}

func (m *CustomHeaderMiddleware) ProcessResponse(ctx *middleware.ResponseContext) error {
    // No response processing needed
    return nil
}

func init() {
    middleware.Register("custom-headers", func(config map[string]interface{}) middleware.Middleware {
        return &CustomHeaderMiddleware{
            headers: config["headers"].(map[string]string),
        }
    })
}
```

### Загрузка своего middleware

#### Локальный плагин

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "local",
        "path": "/path/to/custom-middleware.so",
        "config": {
          "headers": {
            "X-Custom-Header": "value"
          }
        }
      }
    ]
  }
}
```

#### Удалённый плагин

```json
{
  "middleware": {
    "enabled": true,
    "plugins": [
      {
        "type": "remote",
        "url": "https://example.com/middleware/custom-headers.so",
        "checksum": "sha256:abc123...",
        "config": {}
      }
    ]
  }
}
```

## Веб-интерфейс

Настройки middleware находятся по адресу `http://localhost:19840/settings`:

1. Откройте вкладку «Middleware» (помечена значком BETA)
2. Включите «Enable Middleware Pipeline»
3. Добавляйте и удаляйте middleware в конвейере
4. Настройте приоритет и конфигурацию
5. Включайте и отключайте отдельные middleware
6. Нажмите «Save»

## Точки API

### Список middleware

```bash
GET /api/v1/middleware
```

Ответ:
```json
{
  "enabled": true,
  "pipeline": [
    {
      "name": "context-injection",
      "enabled": true,
      "priority": 100,
      "type": "builtin"
    },
    {
      "name": "request-logger",
      "enabled": true,
      "priority": 200,
      "type": "builtin"
    }
  ]
}
```

### Добавить middleware

```bash
POST /api/v1/middleware
Content-Type: application/json

{
  "name": "rate-limiter",
  "enabled": true,
  "priority": 300,
  "config": {
    "requests_per_minute": 60
  }
}
```

### Обновить middleware

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### Удалить middleware

```bash
DELETE /api/v1/middleware/{name}
```

### Перезагрузить конвейер

```bash
POST /api/v1/middleware/reload
```

## Сценарии использования

### Среда разработки

Добавьте отладочное журналирование и разбор запросов:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 100,
        "config": {
          "log_level": "debug",
          "log_body": true
        }
      }
    ]
  }
}
```

### Боевая среда

Добавьте ограничение частоты и наблюдение:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "rate-limiter",
        "enabled": true,
        "priority": 100,
        "config": {
          "requests_per_minute": 100,
          "burst": 20
        }
      },
      {
        "name": "request-logger",
        "enabled": true,
        "priority": 200,
        "config": {
          "log_level": "info",
          "log_body": false
        }
      }
    ]
  }
}
```

### Сравнение нескольких провайдеров

Используйте оркестрацию, чтобы сравнить ответы:

```json
{
  "middleware": {
    "enabled": true,
    "pipeline": [
      {
        "name": "orchestration",
        "enabled": true,
        "priority": 500,
        "config": {
          "strategy": "parallel",
          "providers": ["anthropic", "openai", "google"],
          "consensus": "longest"
        }
      }
    ]
  }
}
```

## Рекомендации

1. **Задавайте разумные приоритеты** — меньшие числа выполняются раньше
2. **Держите middleware узкими** — каждый делает одно дело хорошо
3. **Аккуратно обрабатывайте ошибки** — ошибка не должна ломать конвейер
4. **Тщательно тестируйте** — проверьте поведение до продакшена
5. **Следите за производительностью** — измеряйте накладные расходы
6. **Документируйте настройки** — ясно опишите параметры

## Ограничения

1. **Накладные расходы** — каждый middleware добавляет задержку
2. **Сложность** — слишком много middleware затрудняет отладку
3. **Безопасность плагинов** — удалённые плагины требуют доверия и проверки
4. **Распространение ошибок** — ошибка в middleware может задеть все запросы
5. **Сложность конфигурации** — запутанные конвейеры труднее поддерживать

## Устранение неполадок

### Middleware не выполняется

1. Убедитесь, что `middleware.enabled` равно `true`
2. Проверьте, что middleware включён в конвейере
3. Проверьте правильность приоритета
4. Изучите журналы демона на предмет ошибок middleware

### Неожиданное поведение

1. Проверьте порядок выполнения (приоритет)
2. Проверьте конфигурацию
3. Протестируйте middleware отдельно
4. Просмотрите его журналы

### Проблемы с производительностью

1. Найдите медленный middleware (по журналам)
2. Сократите их число
3. Оптимизируйте реализацию
4. Отключите необязательные middleware

### Плагин не загружается

1. Проверьте путь к плагину
2. Проверьте, что он собран под нужную архитектуру
3. Сверьте контрольную сумму (для удалённых плагинов)
4. Просмотрите журналы плагина

## Вопросы безопасности

1. **Проверяйте плагины** — загружайте только доверенные
2. **Сверяйте контрольные суммы** — всегда для удалённых плагинов
3. **Изолируйте плагины** — рассмотрите запуск в изолированной среде
4. **Проводите аудит** — читайте код middleware перед развёртыванием
5. **Следите за поведением** — обращайте внимание на необычное поведение

## Планы развития

- Поддержка плагинов на WebAssembly для кроссплатформенности
- Каталог для обмена плагинами сообщества
- Визуальный редактор конвейера в веб-интерфейсе
- Профилирование производительности middleware
- Горячая перезагрузка при обновлении плагинов
- Фреймворк для тестирования middleware
