---
sidebar_position: 5
title: 프로젝트 바인딩
---

# 프로젝트 바인딩

디렉터리를 특정 프로필 또는 CLI에 연결하여 프로젝트 단위 설정을 자동으로 적용할 수 있습니다.

## 사용법

```bash
cd ~/work/company-project

# 프로필 바인딩
zen bind work-profile

# CLI 바인딩
zen bind --cli codex

# 둘 다 바인딩
zen bind work-profile --cli codex

# 상태 확인
zen status

# 바인딩 해제
zen unbind
```

## 우선순위

CLI 인자 > 프로젝트 바인딩 > 전역 기본값
