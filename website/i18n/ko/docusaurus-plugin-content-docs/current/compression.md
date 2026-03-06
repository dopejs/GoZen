---
sidebar_position: 14
title: 컨텍스트 압축 (BETA)
---

# 컨텍스트 압축 (BETA)

:::warning BETA 기능
컨텍스트 압축은 현재 베타 단계입니다. 기본적으로 비활성화되어 있으며 명시적인 설정이 필요합니다.
:::

토큰 수가 임계값을 초과할 때 대화 컨텍스트를 자동으로 압축하여 대화 품질을 유지하면서 비용을 절감합니다.

## 기능

- **자동 압축** — 토큰 수가 임계값을 초과할 때 트리거
- **지능형 요약** — 저렴한 모델(claude-3-haiku)을 사용하여 오래된 메시지 요약
- **최근 메시지 보존** — 컨텍스트 연속성을 위해 최근 메시지를 완전하게 유지
- **토큰 추정** — API 호출 전 정확한 토큰 수 계산
- **통계 추적** — 압축 효과 모니터링
- **투명한 작동** — 모든 AI 클라이언트와 원활하게 작동

## 작동 방식

1. **토큰 추정** — 대화 기록의 토큰 수 계산
2. **임계값 확인** — 설정된 임계값과 비교 (기본값: 50,000)
3. **메시지 선택** — 압축이 필요한 오래된 메시지 식별
4. **요약 생성** — 저렴한 모델을 사용하여 간결한 요약 생성
5. **컨텍스트 교체** — 오래된 메시지를 요약으로 교체
6. **요청 전달** — 압축된 컨텍스트를 대상 모델로 전송

## 설정

### 압축 활성화

```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 50000,
    "target_tokens": 20000,
    "summarizer_model": "claude-3-haiku-20240307",
    "preserve_recent_messages": 5,
    "tokens_per_char": 0.25
  }
}
```

**옵션:**

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `enabled` | `false` | 컨텍스트 압축 활성화 |
| `threshold_tokens` | `50000` | 이 값을 초과하면 압축 트리거 |
| `target_tokens` | `20000` | 압축 후 목표 토큰 수 |
| `summarizer_model` | `claude-3-haiku-20240307` | 요약에 사용할 모델 |
| `preserve_recent_messages` | `5` | 완전하게 유지할 최근 메시지 수 |
| `tokens_per_char` | `0.25` | 토큰 계산을 위한 추정 비율 |

### 프로필별 설정

특정 프로필에 대해 압축 활성화:

```json
{
  "profiles": {
    "long-context": {
      "providers": ["anthropic"],
      "compression": {
        "enabled": true,
        "threshold_tokens": 100000,
        "target_tokens": 40000
      }
    },
    "short-context": {
      "providers": ["openai"],
      "compression": {
        "enabled": false
      }
    }
  }
}
```

## 토큰 추정

GoZen은 빠른 토큰 계산을 위해 문자 기반 추정을 사용합니다:

```
estimated_tokens = character_count * tokens_per_char
```

**기본 비율:** 문자당 0.25 토큰 (1 토큰 ≈ 4 문자)

**정확도:** 영어 텍스트 ±10%, 다른 언어는 다를 수 있음

정확한 토큰 계산을 위해 GoZen은 사용 가능한 경우 `tiktoken-go` 라이브러리를 사용합니다.

## 압축 전략

### 메시지 선택

1. **시스템 메시지** — 항상 보존
2. **최근 메시지** — 마지막 N개 메시지 보존 (기본값: 5)
3. **오래된 메시지** — 압축 후보

### 요약 프롬프트

```
다음 대화 기록을 간결하게 요약하되 핵심 정보, 결정 사항 및 컨텍스트를 보존하세요:

[오래된 메시지]

요점을 포착하는 짧은 요약을 제공하세요.
```

### 결과

```
원본: 45,000 토큰 (30개 메시지)
압축 후: 22,000 토큰 (요약 + 5개 최근 메시지)
절감: 23,000 토큰 (51%)
```

## Web UI

`http://localhost:19840/settings`에서 압축 설정에 액세스:

1. "Compression" 탭으로 이동 (BETA 배지 표시)
2. "Enable Compression" 토글
3. 임계값 및 목표 토큰 수 조정
4. 요약 모델 선택
5. 보존할 최근 메시지 수 설정
6. "Save" 클릭

### 통계 대시보드

압축 통계 확인:

- **총 압축 횟수** — 압축이 트리거된 횟수
- **절감된 토큰** — 모든 압축에서 절감된 총 토큰 수
- **평균 절감** — 압축당 평균 토큰 감소량
- **압축 비율** — 압축이 트리거된 요청 비율

## API 엔드포인트

### 압축 통계 가져오기

```bash
GET /api/v1/compression/stats
```

응답:
```json
{
  "enabled": true,
  "total_compressions": 42,
  "tokens_saved": 1250000,
  "average_savings": 29761,
  "compression_rate": 0.15,
  "last_compression": "2026-03-05T10:30:00Z"
}
```

### 압축 설정 업데이트

```bash
PUT /api/v1/compression/settings
Content-Type: application/json

{
  "enabled": true,
  "threshold_tokens": 60000,
  "target_tokens": 25000
}
```

### 통계 재설정

```bash
POST /api/v1/compression/stats/reset
```

## 사용 사례

### 장시간 코딩 세션

**시나리오:** Claude Code를 사용한 여러 시간의 코딩 세션

**설정:**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 80000,
    "target_tokens": 30000,
    "preserve_recent_messages": 10
  }
}
```

**이점:** 컨텍스트 제한에 도달하지 않고 대화 연속성 유지

### 배치 처리

**시나리오:** AI를 사용한 여러 문서 처리

**설정:**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 40000,
    "target_tokens": 15000,
    "preserve_recent_messages": 3
  }
}
```

**이점:** 대량 문서 세트 처리 시 비용 절감

### 연구 및 분석

**시나리오:** 여러 주제를 다루는 장시간 연구 세션

**설정:**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 100000,
    "target_tokens": 40000,
    "preserve_recent_messages": 8
  }
}
```

**이점:** 초기 컨텍스트를 보존하면서 최근 주제에 대화 집중

## 모범 사례

1. **기본값으로 시작** — 기본 설정은 대부분의 사용 사례에 적합
2. **통계 모니터링** — 정기적으로 압축 비율 및 절감량 확인
3. **임계값 조정** — 장기 컨텍스트 모델(Claude Opus)의 경우 증가, 단기 컨텍스트의 경우 감소
4. **충분한 메시지 보존** — 컨텍스트 연속성을 위해 5-10개의 최근 메시지 보존
5. **저렴한 요약기 사용** — Haiku는 빠르고 비용 효율적이며 요약에 적합
6. **프로덕션 전 테스트** — 특정 사용 사례로 압축 품질 검증

## 제한 사항

1. **품질 손실** — 요약은 미묘한 세부 사항을 놓칠 수 있음
2. **지연 시간 증가** — 요약 API 호출 오버헤드 추가
3. **비용 트레이드오프** — 요약 비용 vs. 토큰 절감
4. **언어 지원** — 영어에 가장 적합하며 다른 언어는 다를 수 있음
5. **컨텍스트 윈도우** — 모델의 최대 컨텍스트 윈도우를 초과할 수 없음

## 문제 해결

### 압축이 트리거되지 않음

1. `compression.enabled`가 `true`인지 확인
2. 토큰 수가 임계값을 초과하는지 확인
3. 대화에 압축할 충분한 메시지가 있는지 확인
4. 데몬 로그에서 압축 오류 확인

### 요약 품질 저하

1. 다른 요약 모델 시도 (예: claude-3-sonnet)
2. `preserve_recent_messages`를 늘려 더 많은 컨텍스트 보존
3. `target_tokens`를 조정하여 더 긴 요약 허용
4. 요약 모델이 사용 가능하고 정상 작동하는지 확인

### 지연 시간 증가

1. 압축은 추가 API 호출(요약) 추가
2. 더 빠른 요약 모델 사용 (haiku가 가장 빠름)
3. 임계값을 높여 압축 빈도 감소
4. 지연 시간에 민감한 애플리케이션의 경우 압축 비활성화 고려

### 예상치 못한 비용

1. 사용 대시보드에서 요약 비용 모니터링
2. 절감량 vs. 요약 비용 비교
3. 임계값을 조정하여 압축 빈도 감소
4. 요약에 가장 저렴한 사용 가능 모델 사용

## 성능 영향

- **토큰 추정** — 요청당 약 1ms (무시 가능)
- **요약 생성** — 1-3초 (모델 및 메시지 수에 따라 다름)
- **메모리 오버헤드** — 최소 (압축당 약 1KB)
- **비용 절감** — 일반적으로 토큰 30-50% 감소

## 고급 설정

### 사용자 정의 요약 프롬프트

```json
{
  "compression": {
    "enabled": true,
    "custom_prompt": "다음 대화의 기술 요약을 작성하되 코드 변경 사항, 결정 사항 및 작업 항목에 중점을 두세요:\n\n{messages}\n\n요약:"
  }
}
```

### 조건부 압축

특정 시나리오에 대해서만 압축 활성화:

```json
{
  "profiles": {
    "default": {
      "scenarios": {
        "longContext": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": true,
            "threshold_tokens": 100000
          }
        },
        "default": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": false
          }
        }
      }
    }
  }
}
```

### 다단계 압축

매우 긴 대화를 위한 여러 압축:

```json
{
  "compression": {
    "enabled": true,
    "stages": [
      {
        "threshold_tokens": 50000,
        "target_tokens": 30000
      },
      {
        "threshold_tokens": 80000,
        "target_tokens": 40000
      }
    ]
  }
}
```

## 향후 개선 사항

- 지능형 메시지 선택을 위한 의미론적 유사성 매칭
- 품질 비교를 위한 다중 모델 요약
- 압축 품질 메트릭 및 피드백
- 각 사용 사례에 맞춤화된 압축 전략
- 외부 컨텍스트 저장을 위한 RAG 통합
