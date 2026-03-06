---
sidebar_position: 12
title: 헬스 모니터링 및 로드 밸런싱
---

# 헬스 모니터링 및 로드 밸런싱

제공자 상태를 실시간으로 모니터링하고 최적의 사용 가능한 제공자로 요청을 자동으로 라우팅합니다.

## 기능

- **실시간 헬스 체크** — 설정 가능한 체크 간격으로 정기적인 헬스 모니터링
- **성공률 추적** — 요청 성공률을 기반으로 제공자 상태 계산
- **지연 시간 모니터링** — 각 제공자의 평균 응답 시간 추적
- **다양한 전략** — 장애 조치, 라운드 로빈, 최저 지연 시간, 최저 비용
- **자동 장애 조치** — 주 제공자가 비정상일 때 백업 제공자로 전환
- **헬스 대시보드** — Web UI의 시각적 상태 표시기

## 설정

### 헬스 모니터링 활성화

```json
{
  "health_check": {
    "enabled": true,
    "interval": "5m",
    "timeout": "10s",
    "endpoint": "/v1/messages",
    "method": "POST"
  }
}
```

**옵션:**
- `interval` — 제공자 상태를 확인하는 빈도 (기본값: 5분)
- `timeout` — 헬스 체크 요청 타임아웃 (기본값: 10초)
- `endpoint` — 테스트할 API 엔드포인트 (기본값: `/v1/messages`)
- `method` — 헬스 체크의 HTTP 메서드 (기본값: `POST`)

### 로드 밸런싱 설정

```json
{
  "load_balancing": {
    "strategy": "least-latency",
    "health_aware": true,
    "cache_ttl": "30s"
  }
}
```

## 로드 밸런싱 전략

### 1. 장애 조치 (기본값)

순서대로 제공자를 사용하고 실패 시 다음으로 전환합니다.

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-primary", "anthropic-backup", "openai"],
      "load_balancing": {
        "strategy": "failover"
      }
    }
  }
}
```

**동작:**
1. `anthropic-primary` 시도
2. 실패하면 `anthropic-backup` 시도
3. 실패하면 `openai` 시도
4. 모두 실패하면 오류 반환

**최적 사용:** 명확한 주/백업 계층 구조가 있는 프로덕션 워크로드

### 2. 라운드 로빈

모든 정상 제공자 간에 요청을 균등하게 분산합니다.

```json
{
  "load_balancing": {
    "strategy": "round-robin"
  }
}
```

**동작:**
- 요청 1 → 제공자 A
- 요청 2 → 제공자 B
- 요청 3 → 제공자 C
- 요청 4 → 제공자 A (순환 반복)

**최적 사용:** 속도 제한을 피하기 위해 여러 계정 간 부하 분산

### 3. 최저 지연 시간

평균 지연 시간이 가장 낮은 제공자로 라우팅합니다.

```json
{
  "load_balancing": {
    "strategy": "least-latency"
  }
}
```

**동작:**
- 각 제공자의 평균 응답 시간 추적
- 가장 빠른 제공자로 라우팅
- 30초마다 메트릭 업데이트 (`cache_ttl`로 설정 가능)

**최적 사용:** 지연 시간에 민감한 애플리케이션, 실시간 상호작용

### 4. 최저 비용

요청된 모델에 대해 가장 저렴한 제공자로 라우팅합니다.

```json
{
  "load_balancing": {
    "strategy": "least-cost"
  }
}
```

**동작:**
- 제공자 간 가격 비교
- 가장 저렴한 옵션으로 라우팅
- 입력 및 출력 토큰 비용 모두 고려

**최적 사용:** 비용 최적화, 배치 처리

## 헬스 상태

제공자는 네 가지 헬스 상태로 분류됩니다:

| 상태 | 성공률 | 동작 |
|------|--------|------|
| **정상** | ≥ 95% | 정상 우선순위 |
| **저하됨** | 70-95% | 낮은 우선순위, 여전히 사용 가능 |
| **비정상** | < 70% | 건너뛰기, 정상 제공자가 없는 경우 제외 |
| **알 수 없음** | 데이터 없음 | 초기에는 정상으로 간주 |

### 헬스 인식 라우팅

`health_aware: true` (기본값)일 때:
- 정상 제공자 우선
- 저하된 제공자는 백업으로 사용
- 비정상 제공자는 다른 모든 제공자가 실패하지 않는 한 건너뜀

## Web UI 대시보드

`http://localhost:19840/health`에서 헬스 대시보드에 액세스:

### 제공자 상태

- **상태 표시기** — 녹색(정상), 노란색(저하됨), 빨간색(비정상)
- **성공률** — 성공한 요청의 백분율
- **평균 지연 시간** — 평균 응답 시간(밀리초)
- **마지막 체크** — 가장 최근 헬스 체크의 타임스탬프
- **오류 수** — 최근 실패 횟수

### 메트릭 타임라인

- **지연 시간 차트** — 시간에 따른 응답 시간 추세
- **성공률 차트** — 시간에 따른 헬스 추세
- **요청 볼륨** — 각 제공자의 요청 수

## API 엔드포인트

### 제공자 헬스 가져오기

```bash
GET /api/v1/health/providers
```

응답:
```json
{
  "providers": [
    {
      "name": "anthropic-primary",
      "status": "healthy",
      "success_rate": 98.5,
      "avg_latency_ms": 1250,
      "last_check": "2026-03-05T10:30:00Z",
      "error_count": 2,
      "total_requests": 150
    },
    {
      "name": "openai-backup",
      "status": "degraded",
      "success_rate": 85.0,
      "avg_latency_ms": 2100,
      "last_check": "2026-03-05T10:29:00Z",
      "error_count": 15,
      "total_requests": 100
    }
  ]
}
```

### 제공자 메트릭 가져오기

```bash
GET /api/v1/health/providers/{name}/metrics?period=1h
```

응답:
```json
{
  "provider": "anthropic-primary",
  "period": "1h",
  "metrics": [
    {
      "timestamp": "2026-03-05T10:00:00Z",
      "latency_ms": 1200,
      "success_rate": 99.0,
      "requests": 25
    },
    {
      "timestamp": "2026-03-05T10:05:00Z",
      "latency_ms": 1300,
      "success_rate": 98.0,
      "requests": 28
    }
  ]
}
```

### 수동 헬스 체크 트리거

```bash
POST /api/v1/health/check
Content-Type: application/json

{
  "provider": "anthropic-primary"
}
```

## Webhook 알림

제공자 상태 변경 시 경고 수신:

```json
{
  "webhooks": [
    {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "events": ["provider_down", "provider_up", "failover"]
    }
  ]
}
```

**이벤트 유형:**
- `provider_down` — 제공자가 비정상 상태로 전환
- `provider_up` — 제공자가 정상 상태로 복구
- `failover` — 요청이 백업 제공자로 장애 조치됨

## 시나리오 기반 라우팅

헬스 모니터링을 시나리오 라우팅과 결합하여 지능형 요청 분산:

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-primary", "anthropic-backup"],
      "scenarios": {
        "thinking": {
          "providers": ["anthropic-thinking"],
          "load_balancing": {
            "strategy": "least-latency"
          }
        },
        "image": {
          "providers": ["anthropic-vision", "openai-vision"],
          "load_balancing": {
            "strategy": "failover"
          }
        }
      }
    }
  }
}
```

자세한 내용은 [시나리오 라우팅](./routing.md)을 참조하세요.

## 모범 사례

1. **적절한 간격 설정** — 대부분의 경우 5분이면 충분하며, 중요한 시스템의 경우 1분 사용
2. **헬스 인식 라우팅 사용** — 프로덕션 워크로드에서 항상 활성화
3. **저하된 제공자 모니터링** — 성공률이 95% 미만일 때 조사
4. **전략 결합** — 주/백업에는 장애 조치, 부하 분산에는 라운드 로빈 사용
5. **webhook 활성화** — 제공자 다운 시 즉시 알림 받기
6. **정기적으로 대시보드 확인** — 헬스 추세를 확인하여 패턴 식별

## 문제 해결

### 헬스 체크 실패

1. 제공자 API 키가 유효한지 확인
2. 제공자 엔드포인트에 대한 네트워크 연결 확인
3. 제공자 응답이 느린 경우 타임아웃 증가: `"timeout": "30s"`
4. 데몬 로그에서 구체적인 오류 메시지 확인

### 지연 시간 메트릭이 올바르지 않음

1. 지연 시간에는 네트워크 시간 + API 처리 시간 포함
2. 프록시 또는 VPN이 오버헤드를 추가하는지 확인
3. 메트릭은 기본적으로 30초 동안 캐시됨 (`cache_ttl`로 설정 가능)

### 장애 조치가 작동하지 않음

1. 로드 밸런싱 설정에서 `health_aware: true` 확인
2. 프로필 설정에 백업 제공자가 구성되어 있는지 확인
3. 헬스 체크가 활성화되어 실행 중인지 확인
4. Web UI 또는 로그에서 장애 조치 이벤트 확인

### 제공자가 비정상 상태에 고착됨

1. API를 통해 수동으로 헬스 체크 트리거
2. 제공자가 실제로 다운되었는지 확인 (curl로 테스트)
3. 데몬을 재시작하여 헬스 상태 재설정: `zen daemon restart`
4. 오류 로그를 확인하여 근본 원인 파악

## 성능 영향

- **헬스 체크** — 최소 오버헤드, 백그라운드 goroutine에서 실행
- **메트릭 캐싱** — 30초 TTL로 데이터베이스 쿼리 감소
- **원자적 연산** — 동시 요청을 위한 스레드 안전 카운터
- **논블로킹** — 헬스 체크가 요청 처리를 차단하지 않음

## 고급 설정

### 사용자 정의 헬스 체크 페이로드

```json
{
  "health_check": {
    "enabled": true,
    "custom_payload": {
      "model": "claude-3-haiku-20240307",
      "max_tokens": 10,
      "messages": [
        {
          "role": "user",
          "content": "ping"
        }
      ]
    }
  }
}
```

### 제공자별 헬스 설정

```json
{
  "providers": {
    "anthropic-primary": {
      "health_check": {
        "interval": "1m",
        "timeout": "5s"
      }
    },
    "openai-backup": {
      "health_check": {
        "interval": "5m",
        "timeout": "10s"
      }
    }
  }
}
```
