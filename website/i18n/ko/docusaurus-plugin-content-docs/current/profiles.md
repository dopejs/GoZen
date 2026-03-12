---
sidebar_position: 3
title: 프로필과 장애 조치
---

# 프로필과 장애 조치

프로필은 장애 조치를 위해 순서가 정해진 제공자 목록입니다. 첫 번째 제공자를 사용할 수 없으면 자동으로 다음 제공자로 전환합니다.

## 설정 예시

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}
```

## 프로필 사용

```bash
# 기본 프로필 사용
zen

# 지정한 프로필 사용
zen -p work

# 대화형으로 선택
zen -p
```
