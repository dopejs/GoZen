---
sidebar_position: 8
title: 설정 레퍼런스
---

# 설정 레퍼런스

## 파일 위치

| 파일 | 설명 |
|------|------|
| `~/.zen/zen.json` | 메인 설정 파일 |
| `~/.zen/zend.log` | 데몬 로그 |
| `~/.zen/zend.pid` | 데몬 PID 파일 |
| `~/.zen/logs.db` | 요청 로그 데이터베이스 (SQLite) |

## 전체 설정 예시

```json
{
  "version": 7,
  "default_profile": "default",
  "default_client": "claude",
  "proxy_port": 19841,
  "web_port": 19840,
  "providers": {
    "anthropic": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000"
      }
    }
  },
  "profiles": {
    "default": {
      "providers": ["anthropic"]
    }
  },
  "project_bindings": {
    "/path/to/project": {
      "profile": "work",
      "client": "codex"
    }
  }
}
```

## 필드 레퍼런스

| 필드 | 설명 |
|------|------|
| `version` | 설정 파일 버전 번호 |
| `default_profile` | 기본 프로필 이름 |
| `default_client` | 기본 CLI 클라이언트 (claude / codex / opencode) |
| `proxy_port` | 프록시 서버 포트 (기본값: 19841) |
| `web_port` | 웹 관리 인터페이스 포트 (기본값: 19840) |
| `providers` | 제공자 설정 컬렉션 |
| `profiles` | 프로필 설정 컬렉션 |
| `project_bindings` | 프로젝트 바인딩 설정 |
| `sync` | 설정 동기화 옵션 (선택 사항) |
