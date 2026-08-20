---
title: 로드 밸런싱
---

# 로드 밸런싱

GoZen은 기본 failover 외에도 여러 제공자 선택 전략을 지원합니다. 프로필별로 전략을 선택하고, 헬스 체크와 결합해 가용성, 지연 시간, 비용에 따라 트래픽을 제어할 수 있습니다.

## 사용 가능한 전략

### Failover

성공할 때까지 제공자를 순서대로 시도합니다. 기본 전략이며 주/예비 구성에 잘 맞습니다.

```json
{
  "profiles": {
    "default": {
      "providers": ["primary", "backup"],
      "strategy": "failover"
    }
  }
}
```

### Round robin

여러 동등한 제공자에 요청을 고르게 분산합니다.

```json
{
  "profiles": {
    "balanced": {
      "providers": ["provider-a", "provider-b", "provider-c"],
      "strategy": "round-robin"
    }
  }
}
```

### Least latency

최근 응답 시간이 가장 짧은 제공자를 우선합니다.

```json
{
  "profiles": {
    "fast": {
      "providers": ["us-east", "us-west", "eu"],
      "strategy": "least-latency"
    }
  }
}
```

### Least cost

요청한 모델에 대해 가장 저렴한 제공자를 우선합니다.

```json
{
  "profiles": {
    "budget": {
      "providers": ["cheap-provider", "premium-provider"],
      "strategy": "least-cost"
    }
  }
}
```

## 헬스 인지 라우팅

모든 전략은 헬스 모니터링과 함께 사용할 수 있습니다. `health_aware` 를 활성화하면 비정상 제공자는 복구될 때까지 자동으로 건너뜁니다.

```json
{
  "profiles": {
    "production": {
      "providers": ["primary", "secondary", "tertiary"],
      "strategy": "least-latency",
      "health_aware": true
    }
  }
}
```

## 전략 선택 가이드

- 안정성을 우선하면 `failover`
- 제공자가 서로 대체 가능하면 `round-robin`
- 대화형 또는 시간 민감형 워크로드에는 `least-latency`
- 속도보다 예산이 중요하면 `least-cost`

## 관련 문서

- [프로필](/docs/profiles)은 제공자 그룹을 정의하는 방법을 설명합니다.
- [라우팅](/docs/routing)은 시나리오 기반 제공자 선택을 다룹니다.
- [헬스 모니터링](/docs/health-monitoring)은 헬스 체크가 라우팅에 미치는 영향을 설명합니다.
