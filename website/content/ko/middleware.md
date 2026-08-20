---
sidebar_position: 15
title: 미들웨어 파이프라인 (BETA)
---

# 미들웨어 파이프라인 (BETA)

:::warning BETA 기능
미들웨어 파이프라인은 현재 베타 단계입니다. 기본적으로 비활성화되어 있으며 명시적인 설정이 필요합니다.
:::

플러그형 미들웨어를 사용하여 GoZen을 확장하고 요청/응답 변환, 로깅, 속도 제한 및 사용자 정의 처리를 구현합니다.

## 기능

- **플러그형 아키텍처** — 핵심 코드 수정 없이 사용자 정의 처리 로직 추가
- **우선순위 기반 실행** — 미들웨어 실행 순서 제어
- **요청/응답 훅** — 전송 전 요청 처리, 수신 후 응답 처리
- **내장 미들웨어** — 컨텍스트 주입, 로깅, 속도 제한, 압축
- **플러그인 로더** — 로컬 파일 또는 원격 URL에서 미들웨어 로드
- **오류 처리** — 우아한 오류 처리 및 폴백 동작

## 아키텍처

```
클라이언트 요청
    ↓
[미들웨어 1: 우선순위 100]
    ↓
[미들웨어 2: 우선순위 200]
    ↓
[미들웨어 3: 우선순위 300]
    ↓
제공자 API
    ↓
[미들웨어 3: 응답]
    ↓
[미들웨어 2: 응답]
    ↓
[미들웨어 1: 응답]
    ↓
클라이언트 응답
```

## 설정

### 미들웨어 파이프라인 활성화

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

**옵션:**

| 옵션 | 설명 |
|------|------|
| `enabled` | 미들웨어 파이프라인 활성화 |
| `pipeline` | 미들웨어 설정 배열 |
| `name` | 미들웨어 식별자 |
| `priority` | 실행 순서 (낮을수록 먼저 실행) |
| `config` | 미들웨어별 설정 |

## 내장 미들웨어

### 1. 컨텍스트 주입

요청에 사용자 정의 컨텍스트를 주입합니다.

```json
{
  "name": "context-injection",
  "enabled": true,
  "priority": 100,
  "config": {
    "system_prompt": "당신은 유용한 코딩 도우미입니다.",
    "metadata": {
      "session_id": "sess_123",
      "user_id": "user_456"
    }
  }
}
```

**사용 사례:**
- 시스템 프롬프트 추가
- 세션 메타데이터 주입
- 사용자 컨텍스트 추가

### 2. 요청 로거

모든 요청과 응답을 기록합니다.

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

**사용 사례:**
- 디버깅
- 감사 추적
- 성능 모니터링

### 3. 속도 제한기

제공자별 또는 전역 요청 속도를 제한합니다.

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

**사용 사례:**
- 속도 제한 오류 방지
- API 사용량 제어
- 남용 방지

### 4. 압축 (BETA)

토큰 수가 임계값을 초과할 때 컨텍스트를 압축합니다.

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

자세한 내용은 [컨텍스트 압축](./compression.md)을 참조하세요.

### 5. 세션 메모리 (BETA)

세션 간 대화 기억을 유지합니다.

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

**사용 사례:**
- 사용자 선호도 기억
- 대화 기록 추적
- 세션 간 컨텍스트 유지

### 6. 오케스트레이션 (BETA)

여러 제공자에게 요청을 라우팅하고 응답을 집계합니다.

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

**사용 사례:**
- 모델 출력 비교
- 중요한 요청의 중복성
- 합의를 통한 품질 향상

## 사용자 정의 미들웨어

### 미들웨어 인터페이스

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

### 예제: 사용자 정의 헤더 주입

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
    // 응답 처리 불필요
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

### 사용자 정의 미들웨어 로드

#### 로컬 플러그인

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

#### 원격 플러그인

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

## Web UI

`http://localhost:19840/settings`에서 미들웨어 설정에 액세스:

1. "Middleware" 탭으로 이동 (BETA 배지 표시)
2. "Enable Middleware Pipeline" 토글
3. 파이프라인에서 미들웨어 추가/제거
4. 우선순위 및 설정 조정
5. 개별 미들웨어 활성화/비활성화
6. "Save" 클릭

## API 엔드포인트

### 미들웨어 목록

```bash
GET /api/v1/middleware
```

응답:
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

### 미들웨어 추가

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

### 미들웨어 업데이트

```bash
PUT /api/v1/middleware/{name}
Content-Type: application/json

{
  "enabled": false
}
```

### 미들웨어 삭제

```bash
DELETE /api/v1/middleware/{name}
```

### 파이프라인 재로드

```bash
POST /api/v1/middleware/reload
```

## 사용 사례

### 개발 환경

디버그 로깅 및 요청 검사 추가:

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

### 프로덕션 환경

속도 제한 및 모니터링 추가:

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

### 다중 제공자 비교

오케스트레이션을 사용하여 출력 비교:

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

## 모범 사례

1. **적절한 우선순위 사용** — 낮은 숫자가 먼저 실행됨
2. **미들웨어를 집중적으로 유지** — 각 미들웨어는 한 가지 작업을 잘 수행해야 함
3. **오류를 우아하게 처리** — 오류로 인해 파이프라인이 중단되지 않도록 함
4. **철저한 테스트** — 프로덕션 전에 미들웨어 동작 검증
5. **성능 모니터링** — 미들웨어 오버헤드 추적
6. **설정 문서화** — 설정 옵션을 명확하게 문서화

## 제한 사항

1. **성능 오버헤드** — 각 미들웨어는 지연 시간을 추가함
2. **복잡성** — 너무 많은 미들웨어는 디버깅을 어렵게 만듦
3. **플러그인 보안** — 원격 플러그인은 신뢰와 검증이 필요함
4. **오류 전파** — 미들웨어 오류는 모든 요청에 영향을 미침
5. **설정 복잡성** — 복잡한 파이프라인은 유지 관리가 더 어려움

## 문제 해결

### 미들웨어가 실행되지 않음

1. `middleware.enabled`가 `true`인지 확인
2. 파이프라인에서 미들웨어가 활성화되어 있는지 확인
3. 우선순위가 올바르게 설정되었는지 확인
4. 데몬 로그에서 미들웨어 오류 확인

### 예상치 못한 동작

1. 미들웨어 실행 순서(우선순위) 확인
2. 설정이 올바른지 확인
3. 미들웨어를 개별적으로 테스트
4. 미들웨어 로그 확인

### 성능 문제

1. 느린 미들웨어 식별 (로그 확인)
2. 미들웨어 수 줄이기
3. 미들웨어 구현 최적화
4. 필수적이지 않은 미들웨어 비활성화 고려

### 플러그인 로드 실패

1. 플러그인 경로가 올바른지 확인
2. 플러그인이 올바른 아키텍처로 컴파일되었는지 확인
3. 체크섬이 일치하는지 확인 (원격 플러그인의 경우)
4. 플러그인 로그에서 오류 확인

## 보안 고려 사항

1. **플러그인 검증** — 신뢰할 수 있는 플러그인만 로드
2. **체크섬 확인** — 항상 원격 플러그인 체크섬 검증
3. **플러그인 샌드박스** — 격리된 환경에서 플러그인 실행 고려
4. **미들웨어 감사** — 배포 전 미들웨어 코드 검토
5. **동작 모니터링** — 예상치 못한 미들웨어 동작 주의

## 향후 개선 사항

- 크로스 플랫폼 호환성을 위한 WebAssembly 플러그인 지원
- 커뮤니티 플러그인 공유를 위한 미들웨어 마켓플레이스
- Web UI의 시각적 파이프라인 편집기
- 미들웨어 성능 프로파일링
- 플러그인 업데이트의 핫 리로드
- 미들웨어 테스트 프레임워크
