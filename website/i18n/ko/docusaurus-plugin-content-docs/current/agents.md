---
title: 에이전트
---

# 에이전트

GoZen은 Claude Code, Codex, 기타 CLI 기반 도우미를 위한 코딩 에이전트 운영 레이어로 동작할 수 있습니다. 기존 워크플로를 바꾸지 않고도 에이전트 작업 조정, 세션 모니터링, 런타임 안전 제어를 적용할 수 있습니다.

## GoZen이 추가하는 기능

- **조정**: 여러 에이전트가 같은 프로젝트를 다룰 때 충돌을 줄입니다.
- **가시성**: 세션, 비용, 오류, 활동을 한곳에서 추적합니다.
- **가드레일**: 지출, 요청 빈도, 민감한 작업에 제한을 둡니다.
- **작업 라우팅**: 작업 유형에 따라 다른 제공자나 프로필로 보냅니다.

## 설정 예시

```json
{
  "agent": {
    "enabled": true,
    "coordinator": {
      "enabled": true,
      "lock_timeout_sec": 300,
      "inject_warnings": true
    },
    "observatory": {
      "enabled": true,
      "stuck_threshold": 5,
      "idle_timeout_min": 30
    },
    "guardrails": {
      "enabled": true,
      "session_spending_cap": 5.0,
      "request_rate_limit": 30
    }
  }
}
```

## 일반적인 워크플로

### 멀티 에이전트 조정

여러 에이전트가 같은 저장소에서 작업할 때 GoZen은 파일 활동을 추적하고 경고를 표시해 충돌을 줄이도록 도와줍니다.

### 세션 모니터링

대시보드와 API를 사용해 활성 세션, 토큰 사용량, 오류 수, 실행 시간을 확인할 수 있습니다.

### 안전 제어 적용

가드레일은 폭주하는 세션을 일시 중지하고, 위험한 작업을 표시하고, 비용이 커지기 전에 과도한 재시도 루프를 늦출 수 있습니다.

## 관련 문서

- [에이전트 인프라](/docs/agent-infrastructure)는 새로운 런타임, 관측소, 코디네이터, 가드레일 구조를 더 자세히 설명합니다.
- [Bot 게이트웨이](/docs/bot)는 Telegram, Slack, Discord 같은 채팅 플랫폼에서 실행 중인 세션을 제어하는 방법을 설명합니다.
- [사용량 추적](/docs/usage-tracking)과 [헬스 모니터링](/docs/health-monitoring)은 에이전트 운영에 필요한 메트릭을 다룹니다.
