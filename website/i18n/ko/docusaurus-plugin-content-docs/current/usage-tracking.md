---
sidebar_position: 11
title: 사용량 추적 및 예산 제어
---

# 사용량 추적 및 예산 제어

제공자, 모델 및 프로젝트 전반에 걸쳐 토큰 사용량과 비용을 추적합니다. 지출 한도를 설정하고 자동으로 조치를 실행합니다.

## 기능

- **실시간 추적** — 각 요청의 토큰 사용량 및 비용 모니터링
- **다차원 집계** — 제공자, 모델, 프로젝트 및 기간별 추적
- **예산 한도** — 일일, 주간 및 월간 지출 상한 설정
- **자동 조치** — 한도 초과 시 경고, 다운그레이드 또는 요청 차단
- **비용 추정** — 모든 주요 AI 모델에 대한 정확한 가격 책정
- **기록 데이터** — 성능 향상을 위한 시간별 집계가 포함된 SQLite 저장소

## 설정

### 사용량 추적 활성화

```json
{
  "usage_tracking": {
    "enabled": true,
    "db_path": "~/.zen/usage.db"
  }
}
```

### 모델 가격 설정

```json
{
  "pricing": {
    "models": {
      "claude-opus-4": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet-4": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4o": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    },
    "model_families": {
      "claude-opus": {
        "input_per_mtok": 15.0,
        "output_per_mtok": 75.0
      },
      "claude-sonnet": {
        "input_per_mtok": 3.0,
        "output_per_mtok": 15.0
      },
      "gpt-4": {
        "input_per_mtok": 2.5,
        "output_per_mtok": 10.0
      }
    }
  }
}
```

**모델 매칭:** 먼저 정확한 모델 이름을 매칭한 다음 모델 패밀리 접두사로 폴백합니다.

### 예산 한도 설정

```json
{
  "budget": {
    "daily": {
      "enabled": true,
      "limit": 10.0,
      "action": "warn"
    },
    "weekly": {
      "enabled": true,
      "limit": 50.0,
      "action": "downgrade"
    },
    "monthly": {
      "enabled": true,
      "limit": 200.0,
      "action": "block"
    }
  }
}
```

## 예산 조치

| 조치 | 동작 |
|------|------|
| `warn` | 경고를 기록하고 webhook 알림을 전송하지만 요청 허용 |
| `downgrade` | 더 저렴한 모델로 전환 (예: opus → sonnet → haiku) |
| `block` | 429 상태 코드로 요청 거부 |

## Web UI

`http://localhost:19840/usage`에서 사용량 대시보드에 액세스:

- **개요** — 현재 기간의 총 비용, 요청 및 토큰
- **제공자별** — 각 제공자의 비용 분석
- **모델별** — 각 모델의 사용 통계
- **프로젝트별** — 프로젝트별 비용 추적 (프로젝트 바인딩을 통해)
- **타임라인** — 시간별/일별 비용 추세
- **예산 상태** — 일일/주간/월간 한도의 시각적 표시기

## API 엔드포인트

### 사용량 요약 가져오기

```bash
GET /api/v1/usage/summary?period=daily
```

응답:
```json
{
  "period": "daily",
  "start": "2026-03-05T00:00:00Z",
  "end": "2026-03-05T23:59:59Z",
  "total_cost": 8.45,
  "total_requests": 42,
  "total_input_tokens": 125000,
  "total_output_tokens": 35000,
  "by_provider": {
    "anthropic": 6.20,
    "openai": 2.25
  },
  "by_model": {
    "claude-sonnet-4": 5.10,
    "claude-opus-4": 1.10,
    "gpt-4o": 2.25
  }
}
```

### 예산 상태 가져오기

```bash
GET /api/v1/budget/status
```

응답:
```json
{
  "daily": {
    "enabled": true,
    "limit": 10.0,
    "spent": 8.45,
    "percent": 84.5,
    "action": "warn",
    "exceeded": false
  },
  "weekly": {
    "enabled": true,
    "limit": 50.0,
    "spent": 32.10,
    "percent": 64.2,
    "action": "downgrade",
    "exceeded": false
  },
  "monthly": {
    "enabled": true,
    "limit": 200.0,
    "spent": 145.80,
    "percent": 72.9,
    "action": "block",
    "exceeded": false
  }
}
```

### 예산 한도 업데이트

```bash
PUT /api/v1/budget/limits
Content-Type: application/json

{
  "daily": {
    "enabled": true,
    "limit": 15.0,
    "action": "warn"
  }
}
```

## 프로젝트 수준 추적

디렉토리 바인딩을 사용하여 프로젝트별 비용 추적:

```bash
# 현재 디렉토리를 프로필에 바인딩
zen bind work-profile

# 이 디렉토리의 모든 요청은 프로젝트 경로로 태그됨
# Web UI의 "By Project"에서 비용 확인
```

## Webhook 알림

예산 초과 시 경고 수신:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["budget_warning", "budget_exceeded"]
    }
  ]
}
```

전체 설정은 [Webhooks](./webhooks.md)를 참조하세요.

## 모범 사례

1. **경고로 시작** — 처음에는 `warn` 조치를 사용하여 사용 패턴 파악
2. **실제 한도 설정** — 과거 사용 데이터를 기반으로 한도 설정
3. **개발 시 다운그레이드 사용** — 테스트 시 자동으로 더 저렴한 모델로 전환
4. **프로덕션에서 차단 예약** — 엄격한 지출 상한에만 `block` 조치 사용
5. **일일 모니터링** — 정기적으로 사용량 대시보드를 확인하여 예상치 못한 상황 방지
6. **webhook 활성화** — 한도에 근접할 때 실시간 경고 받기

## 문제 해결

### 사용량이 추적되지 않음

1. 설정에서 `usage_tracking.enabled`가 `true`인지 확인
2. 데이터베이스 경로가 쓰기 가능한지 확인: `~/.zen/usage.db`
3. 데몬 재시작: `zen daemon restart`

### 비용이 올바르지 않음

1. 설정의 모델 가격이 현재 요금과 일치하는지 확인
2. 모델 이름 매칭 확인 (정확한 매칭 vs 패밀리 접두사)
3. 제공자가 요금을 변경한 경우 가격 설정 업데이트

### 예산이 적용되지 않음

1. 예산 설정이 활성화되어 있는지 확인
2. 조치가 설정되어 있는지 확인 (`warn`, `downgrade` 또는 `block`)
3. 데몬 로그에서 예산 검사기 오류 확인

## 성능

- **시간별 집계** — 원시 데이터는 쿼리 부하를 줄이기 위해 시간별로 집계됨
- **인덱스 쿼리** — 데이터베이스는 제공자, 모델, 프로젝트, 타임스탬프를 인덱싱
- **효율적인 저장소** — 요청당 약 1KB, 30,000개 요청에 약 30MB
- **빠른 대시보드** — 일반적인 사용 패턴에서 1초 미만의 쿼리 시간
