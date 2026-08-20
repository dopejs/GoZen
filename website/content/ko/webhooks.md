---
sidebar_position: 13
title: Webhook
---

# Webhook

Slack, Discord 또는 사용자 정의 Webhook을 통해 예산 경고, 제공자 상태 변경 및 일일 요약에 대한 실시간 알림을 받습니다.

## 기능

- **다양한 형식** — Slack, Discord 또는 일반 JSON
- **이벤트 필터링** — 특정 이벤트 유형 구독
- **사용자 정의 헤더** — 인증 또는 사용자 정의 헤더 추가
- **비동기 전달** — 논블로킹 Webhook 전달
- **자동 포맷팅** — 이모지 및 색상이 포함된 풍부한 메시지
- **테스트 기능** — 활성화 전 Webhook 설정 검증

## 설정

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

## 이벤트 유형

| 이벤트 | 설명 | 트리거 시점 |
|------|------|----------|
| `budget_warning` | 예산 임계값 도달 | 지출이 한도의 80%에 도달할 때 |
| `budget_exceeded` | 예산 한도 초과 | 지출이 설정된 한도를 초과할 때 |
| `provider_down` | 제공자가 비정상 상태로 전환 | 성공률이 70% 미만일 때 |
| `provider_up` | 제공자 복구 | 비정상 제공자가 다시 정상 상태가 될 때 |
| `failover` | 요청 장애 조치 | 요청이 백업 제공자로 전환될 때 |
| `daily_summary` | 일일 사용 요약 | 매일 UTC 자정에 한 번 |

## Webhook 형식

### Slack

URL에 `slack.com`이 포함되어 있으면 자동으로 감지됩니다.

**예제 메시지:**
```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

**형식:**
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

URL에 `discord.com`이 포함되어 있으면 자동으로 감지됩니다.

**예제 임베드:**
- **제목:** budget_warning
- **설명:** ⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
- **색상:** 앰버 (#FBBF24)
- **타임스탬프:** 2026-03-05T10:30:00Z

**형식:**
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

### 일반 JSON

다른 모든 URL에 사용됩니다.

**형식:**
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

## 이벤트 데이터 구조

### 예산 경고 / 초과

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

### 제공자 다운 / 복구

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

### 장애 조치

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

### 일일 요약

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

## 플랫폼 설정

### Slack

1. [Slack API](https://api.slack.com/apps) 방문
2. 새 앱 생성 또는 기존 앱 선택
3. "Incoming Webhooks" 활성화
4. 워크스페이스에 Webhook 추가
5. Webhook URL 복사 (`https://hooks.slack.com/`으로 시작)

**설정:**
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

1. Discord 서버 설정 열기
2. Integrations → Webhooks로 이동
3. "New Webhook" 클릭
4. 채널 선택 및 Webhook URL 복사

**설정:**
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

### 사용자 정의 Webhook

사용자 정의 통합의 경우 일반 JSON 형식 사용:

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

## Web UI 설정

`http://localhost:19840/settings`에서 Webhook 설정에 액세스:

1. "Webhooks" 탭으로 이동
2. "Add Webhook" 클릭
3. Webhook URL 입력
4. 구독할 이벤트 선택
5. (선택 사항) 사용자 정의 헤더 추가
6. "Test" 클릭하여 설정 검증
7. "Save" 클릭

## API 엔드포인트

### Webhook 목록

```bash
GET /api/v1/webhooks
```

### Webhook 추가

```bash
POST /api/v1/webhooks
Content-Type: application/json

{
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
  "events": ["budget_warning", "provider_down"]
}
```

### Webhook 업데이트

```bash
PUT /api/v1/webhooks/{id}
Content-Type: application/json

{
  "enabled": false
}
```

### Webhook 삭제

```bash
DELETE /api/v1/webhooks/{id}
```

### Webhook 테스트

```bash
POST /api/v1/webhooks/{id}/test
```

설정을 검증하기 위해 테스트 메시지를 전송합니다.

## 메시지 예제

### 예산 경고 (Slack)

```
⚠️ Budget Warning: daily budget at 85.0% ($8.50 / $10.00)
```

### 예산 초과 (Discord)

```
🚫 Budget Exceeded: monthly limit of $200.00 reached (spent: $205.50). Action: block
```

### 제공자 다운 (Slack)

```
🔴 Provider Down: anthropic-primary is unhealthy. Error: connection timeout
```

### 제공자 복구 (Discord)

```
🟢 Provider Up: anthropic-primary is healthy again (latency: 1250ms)
```

### 장애 조치 (Slack)

```
🔄 Failover: Switched from anthropic-primary to anthropic-backup. Reason: rate limit exceeded
```

### 일일 요약 (Discord)

```
📊 Daily Summary (2026-03-04): 150 requests, $25.50 total cost, 125000 input / 35000 output tokens
```

## 모범 사례

1. **별도의 Webhook 사용** — 다른 이벤트 유형에 대해 다른 Webhook 생성
2. **활성화 전 테스트** — 저장하기 전에 항상 Webhook 설정 테스트
3. **사용자 정의 Webhook 보호** — HTTPS 및 인증 헤더 사용
4. **Webhook 실패 모니터링** — 알림이 중단되면 데몬 로그 확인
5. **민감한 데이터 방지** — Webhook URL에 API 키 또는 토큰 포함하지 않음
6. **경고 설정** — `budget_exceeded` 및 `provider_down`과 같은 중요한 이벤트 구독

## 문제 해결

### Webhook이 메시지를 받지 못함

1. 설정에서 Webhook이 활성화되어 있는지 확인
2. URL이 올바른지 확인 (curl로 테스트)
3. 이벤트 설정이 올바른지 확인
4. 데몬 로그에서 Webhook 오류 확인: `tail -f ~/.zen/zend.log`
5. API를 통해 Webhook 테스트: `POST /api/v1/webhooks/{id}/test`

### Slack Webhook 실패

1. Webhook URL이 `https://hooks.slack.com/`으로 시작하는지 확인
2. Slack 설정에서 Webhook이 취소되지 않았는지 확인
3. 워크스페이스에서 수신 Webhook이 비활성화되지 않았는지 확인
4. curl로 테스트:
   ```bash
   curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

### Discord Webhook 실패

1. Webhook URL이 `https://discord.com/api/webhooks/`로 시작하는지 확인
2. Discord 설정에서 Webhook이 삭제되지 않았는지 확인
3. 봇이 채널에 게시할 권한이 있는지 확인
4. curl로 테스트:
   ```bash
   curl -X POST https://discord.com/api/webhooks/YOUR/WEBHOOK/URL \
     -H 'Content-Type: application/json' \
     -d '{"content":"test"}'
   ```

### 사용자 정의 Webhook이 작동하지 않음

1. 엔드포인트에 액세스할 수 있는지 확인 (curl로 테스트)
2. 인증 헤더가 올바른지 확인
3. 엔드포인트가 POST 요청을 수락하는지 확인
4. 엔드포인트가 2xx 상태 코드를 반환하는지 확인
5. 엔드포인트 로그에서 오류 확인

## 보안 고려 사항

1. **Webhook URL 보호** — Webhook URL을 기밀로 취급
2. **HTTPS 사용** — Webhook 엔드포인트에 항상 HTTPS 사용
3. **서명 검증** — 사용자 정의 Webhook에 대한 서명 검증 구현
4. **속도 제한** — Webhook 엔드포인트에 속도 제한 구현
5. **민감한 데이터 로깅 금지** — 전체 Webhook 페이로드 로깅 방지

## 고급 설정

### 조건부 Webhook

다른 이벤트를 다른 Webhook으로 전송:

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

### 인증을 위한 사용자 정의 헤더

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
