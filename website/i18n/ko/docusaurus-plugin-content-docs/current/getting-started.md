---
sidebar_position: 1
title: 시작하기
---

# 시작하기

## 설치

원라인 스크립트로 설치할 수 있습니다:

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

제거:

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## 첫 실행

먼저 첫 번째 제공자를 추가합니다:

```bash
zen config add provider
```

기본 프로필로 시작:

```bash
zen
```

특정 프로필로 시작:

```bash
zen -p work
```

특정 CLI로 시작:

```bash
zen --cli codex
```
