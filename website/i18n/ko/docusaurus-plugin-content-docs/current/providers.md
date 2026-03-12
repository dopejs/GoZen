---
sidebar_position: 2
title: 제공자 관리
---

# 제공자 관리

제공자는 base URL, 인증 토큰, 모델 이름 등을 포함하는 API 엔드포인트 설정입니다.

## 설정 예시

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}
```

## 환경 변수

각 제공자에는 CLI별 환경 변수를 설정할 수 있습니다.

### 자주 쓰는 Claude Code 환경 변수

| 변수 | 설명 |
|------|------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | 최대 출력 토큰 수 |
| `MAX_THINKING_TOKENS` | 확장 사고 예산 |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | 최대 컨텍스트 윈도우 |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash 기본 타임아웃 |
