---
sidebar_position: 4
title: 시나리오 라우팅
---

# 시나리오 라우팅

요청 특성에 따라 서로 다른 제공자로 자동 라우팅합니다.

## 지원 시나리오

| 시나리오 | 설명 |
|----------|------|
| `think` | Thinking mode 활성화 |
| `image` | 이미지 콘텐츠 포함 |
| `longContext` | 콘텐츠가 임계값 초과 |
| `webSearch` | `web_search` 도구 사용 |
| `background` | Haiku 모델 사용 |

## 폴백 메커니즘

특정 시나리오에 할당된 모든 제공자가 실패하면 자동으로 해당 프로필의 기본 제공자로 폴백합니다.

## 설정 예시

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```
