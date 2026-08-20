---
sidebar_position: 6
title: 멀티 CLI 지원
---

# 멀티 CLI 지원

GoZen은 세 가지 AI 코딩 지원 CLI를 지원합니다:

| CLI | 설명 | API 형식 |
|-----|------|-----------|
| `claude` | Claude Code (기본값) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

## 기본 CLI 설정

```bash
zen config default-client

# Web UI에서 설정
zen web  # Settings 페이지
```

## 프로젝트별 CLI

```bash
cd ~/work/project
zen bind --cli codex  # 이 디렉터리에서는 Codex 사용
```

## 임시 CLI 오버라이드

```bash
zen --cli opencode  # 이 세션에서만 OpenCode 사용
```
