---
sidebar_position: 16
title: 에이전트 인프라 (BETA)
---

# 에이전트 인프라 (BETA)

:::warning BETA 기능
에이전트 인프라는 현재 베타 단계입니다. 기본적으로 비활성화되어 있으며 명시적인 설정이 필요합니다.
:::

세션 관리, 파일 조정, 실시간 모니터링 및 보안 제어를 포함한 자율 에이전트 워크플로우에 대한 내장 지원.

## 기능

- **에이전트 런타임** — 완전한 생명주기 관리를 통한 자율 에이전트 작업 실행
- **관측소** — 에이전트 세션 및 활동의 실시간 모니터링
- **가드레일** — 에이전트 동작에 대한 보안 제어 및 제약
- **코디네이터** — 파일 기반 다중 에이전트 워크플로우 조정
- **작업 큐** — 우선순위 및 종속성이 있는 에이전트 작업 관리
- **세션 관리** — 여러 프로젝트에 걸쳐 에이전트 세션 추적

## 아키텍처

```
에이전트 클라이언트 (Claude Code, Codex 등)
    ↓
에이전트 런타임
    ↓
┌─────────────┬──────────────┬─────────────┐
│ 관측소      │ 가드레일     │ 코디네이터  │
│ (모니터링)  │ (보안)       │ (동기화)    │
└─────────────┴──────────────┴─────────────┘
    ↓
작업 큐 → 제공자 API
```

## 설정

### 에이전트 인프라 활성화

```json
{
  "agent": {
    "enabled": true,
    "runtime": {
      "max_concurrent_tasks": 5,
      "task_timeout": "30m",
      "auto_cleanup": true
    },
    "observatory": {
      "enabled": true,
      "update_interval": "5s",
      "history_retention": "7d"
    },
    "guardrails": {
      "enabled": true,
      "max_file_operations": 100,
      "max_api_calls": 1000,
      "allowed_paths": ["/Users/john/projects"],
      "blocked_commands": ["rm -rf", "sudo"]
    },
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    }
  }
}
```

## 구성 요소

### 1. 에이전트 런타임

에이전트 작업 실행 생명주기를 관리합니다.

**기능:**
- 작업 스케줄링 및 실행
- 동시 작업 관리
- 타임아웃 처리
- 자동 정리
- 오류 복구

**설정:**
```json
{
  "runtime": {
    "max_concurrent_tasks": 5,
    "task_timeout": "30m",
    "auto_cleanup": true,
    "retry_failed_tasks": true,
    "max_retries": 3
  }
}
```

**API:**
```bash
# 에이전트 작업 시작
POST /api/v1/agent/tasks
Content-Type: application/json

{
  "name": "code-review",
  "description": "Review pull request #123",
  "priority": 1,
  "config": {
    "model": "claude-opus-4",
    "max_tokens": 100000
  }
}

# 작업 상태 가져오기
GET /api/v1/agent/tasks/{task_id}

# 작업 취소
DELETE /api/v1/agent/tasks/{task_id}
```

### 2. 관측소

에이전트 활동을 실시간으로 모니터링합니다.

**기능:**
- 세션 추적
- 활동 로그
- 성능 메트릭
- 상태 업데이트
- 기록 데이터

**설정:**
```json
{
  "observatory": {
    "enabled": true,
    "update_interval": "5s",
    "history_retention": "7d",
    "metrics": {
      "track_tokens": true,
      "track_costs": true,
      "track_latency": true
    }
  }
}
```

**모니터링 메트릭:**
- 활성 세션
- 진행 중인 작업
- 토큰 사용량
- API 호출
- 파일 작업
- 오류율
- 평균 지연 시간

**API:**
```bash
# 모든 활성 세션 가져오기
GET /api/v1/agent/sessions

# 세션 세부 정보 가져오기
GET /api/v1/agent/sessions/{session_id}

# 세션 메트릭 가져오기
GET /api/v1/agent/sessions/{session_id}/metrics
```

### 3. 가드레일

에이전트 동작에 대한 보안 제어 및 제약.

**기능:**
- 작업 제한
- 경로 제한
- 명령 차단
- 리소스 할당량
- 승인 워크플로우

**설정:**
```json
{
  "guardrails": {
    "enabled": true,
    "max_file_operations": 100,
    "max_api_calls": 1000,
    "max_tokens_per_session": 1000000,
    "allowed_paths": [
      "/Users/john/projects",
      "/tmp/agent-workspace"
    ],
    "blocked_paths": [
      "/etc",
      "/System",
      "~/.ssh"
    ],
    "blocked_commands": [
      "rm -rf /",
      "sudo",
      "chmod 777"
    ],
    "require_approval": {
      "file_delete": true,
      "system_commands": true,
      "network_requests": false
    }
  }
}
```

**적용 메커니즘:**
- 실행 전 검증
- 실시간 모니터링
- 자동 차단
- 승인 프롬프트
- 감사 로그

**API:**
```bash
# 가드레일 상태 가져오기
GET /api/v1/agent/guardrails

# 가드레일 규칙 업데이트
PUT /api/v1/agent/guardrails
Content-Type: application/json

{
  "max_file_operations": 200,
  "blocked_commands": ["rm -rf", "sudo", "dd"]
}
```

### 4. 코디네이터

파일 기반 다중 에이전트 워크플로우 조정.

**기능:**
- 파일 잠금
- 변경 감지
- 충돌 해결
- 상태 동기화
- 이벤트 알림

**설정:**
```json
{
  "coordinator": {
    "enabled": true,
    "lock_timeout": "5m",
    "change_detection": true,
    "conflict_resolution": "last-write-wins",
    "notification_webhook": "https://hooks.slack.com/..."
  }
}
```

**사용 사례:**
- 여러 에이전트가 동일한 파일 편집
- 동시 수정 방지
- 외부 파일 변경 감지
- 에이전트 워크플로우 조정

**API:**
```bash
# 파일 잠금 획득
POST /api/v1/agent/locks
Content-Type: application/json

{
  "path": "/path/to/file.go",
  "session_id": "sess_123",
  "timeout": "5m"
}

# 파일 잠금 해제
DELETE /api/v1/agent/locks/{lock_id}

# 파일 변경 이벤트 가져오기
GET /api/v1/agent/changes?since=2026-03-05T10:00:00Z
```

### 5. 작업 큐

우선순위 및 종속성이 있는 에이전트 작업을 관리합니다.

**기능:**
- 우선순위 스케줄링
- 작업 종속성
- 큐 관리
- 상태 추적
- 재시도 로직

**설정:**
```json
{
  "task_queue": {
    "enabled": true,
    "max_queue_size": 100,
    "priority_levels": 5,
    "enable_dependencies": true,
    "retry_policy": {
      "max_retries": 3,
      "backoff": "exponential"
    }
  }
}
```

**API:**
```bash
# 큐에 작업 추가
POST /api/v1/agent/queue
Content-Type: application/json

{
  "name": "run-tests",
  "priority": 2,
  "depends_on": ["build-project"],
  "config": {}
}

# 큐 상태 가져오기
GET /api/v1/agent/queue

# 큐에서 작업 제거
DELETE /api/v1/agent/queue/{task_id}
```

## Web UI

에이전트 대시보드 액세스: `http://localhost:19840/agent`

### 세션 탭

- **활성 세션** — 현재 실행 중인 에이전트 세션
- **세션 세부 정보** — 작업 진행 상황, 메트릭, 로그
- **세션 제어** — 일시 중지, 재개, 취소

### 작업 탭

- **작업 큐** — 대기 중 및 진행 중인 작업
- **작업 기록** — 완료 및 실패한 작업
- **작업 세부 정보** — 설정, 로그, 결과

### 가드레일 탭

- **작업 제한** — 현재 사용량 vs. 제한
- **차단된 작업** — 최근 차단된 시도
- **승인 큐** — 승인 대기 중인 작업

### 메트릭 탭

- **토큰 사용량** — 세션별 및 총계
- **API 호출** — 요청 수 및 비율
- **파일 작업** — 읽기/쓰기/삭제 수
- **성능** — 지연 시간 및 처리량

## Claude Code와의 통합

GoZen은 Claude Code 세션을 자동으로 감지하고 에이전트 인프라를 제공합니다:

```bash
# 에이전트 지원으로 Claude Code 시작
zen --agent

# 에이전트 기능이 자동으로 활성화됨:
# - 세션 추적
# - 파일 조정
# - 가드레일 적용
# - 실시간 모니터링
```

**이점:**
- 동시 파일 수정 방지
- 토큰 사용량 및 비용 추적
- 보안 제약 적용
- 에이전트 활동 모니터링
- 다중 에이전트 워크플로우 조정

## 사용 사례

### 다중 에이전트 개발

동일한 코드베이스에서 작업하는 여러 에이전트:

```json
{
  "agent": {
    "coordinator": {
      "enabled": true,
      "lock_timeout": "5m",
      "change_detection": true
    },
    "guardrails": {
      "max_file_operations": 200,
      "allowed_paths": ["/Users/john/project"]
    }
  }
}
```

### 장시간 실행 작업

장시간 실행되는 에이전트 작업 모니터링 및 제어:

```json
{
  "agent": {
    "runtime": {
      "task_timeout": "2h",
      "auto_cleanup": false
    },
    "observatory": {
      "update_interval": "10s",
      "history_retention": "30d"
    }
  }
}
```

### 보안 중요 작업

엄격한 보안 제어 적용:

```json
{
  "agent": {
    "guardrails": {
      "enabled": true,
      "max_file_operations": 50,
      "blocked_commands": ["rm", "sudo", "chmod"],
      "require_approval": {
        "file_delete": true,
        "system_commands": true,
        "network_requests": true
      }
    }
  }
}
```

## 모범 사례

1. **가드레일 활성화** — 프로덕션 환경에서 항상 가드레일 사용
2. **적절한 제한 설정** — 사용 사례에 따라 제한 구성
3. **적극적으로 모니터링** — 정기적으로 관측소 대시보드 확인
4. **파일 잠금 사용** — 다중 에이전트 워크플로우에 대해 코디네이터 활성화
5. **승인 구성** — 파괴적인 작업에 대해 승인 요구
6. **로그 검토** — 정기적으로 에이전트 활동 감사

## 제한 사항

1. **성능 오버헤드** — 모니터링 및 조정이 지연 시간 추가
2. **파일 잠금** — 다중 에이전트 시나리오에서 지연 발생 가능
3. **메모리 사용** — 세션 기록이 메모리 소비
4. **복잡성** — 에이전트 워크플로우에 대한 이해 필요
5. **베타 상태** — 향후 버전에서 기능이 변경될 수 있음

## 문제 해결

### 에이전트 세션이 추적되지 않음

1. `agent.enabled`가 `true`인지 확인
2. 관측소가 활성화되어 있는지 확인
3. 에이전트 클라이언트가 지원되는지 확인 (Claude Code, Codex)
4. 데몬 로그에서 오류 확인

### 파일 잠금 문제

1. 코디네이터가 활성화되어 있는지 확인
2. 잠금 타임아웃이 적절한지 확인
3. 활성 잠금 확인: `GET /api/v1/agent/locks`
4. 필요한 경우 고착된 잠금 수동 해제

### 가드레일이 적용되지 않음

1. 가드레일이 활성화되어 있는지 확인
2. 규칙 설정이 올바른지 확인
3. 차단된 작업 로그 확인
4. 에이전트 클라이언트가 가드레일을 준수하는지 확인

### 높은 메모리 사용량

1. 기록 보존 기간 줄이기
2. 업데이트 간격 감소
3. 최대 동시 작업 수 제한
4. 자동 정리 활성화

## 보안 고려 사항

1. **경로 제한** — 항상 허용/차단 경로 구성
2. **명령 차단** — 위험한 명령 차단
3. **승인 워크플로우** — 민감한 작업에 대해 승인 요구
4. **감사 로그** — 포괄적인 로깅 활성화
5. **리소스 제한** — 적절한 작업 제한 설정

## 향후 개선 사항

- 다중 에이전트 협업 프로토콜
- 고급 충돌 해결 전략
- 이상 감지를 위한 머신 러닝
- 외부 모니터링 도구와의 통합
- 에이전트 동작 분석
- 자동 보안 정책 생성
