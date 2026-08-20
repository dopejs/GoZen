---
sidebar_position: 10
title: Bot 게이트웨이
---

# Bot 게이트웨이

채팅 플랫폼을 통해 Claude Code 세션을 원격으로 모니터링하고 제어합니다. Bot은 IPC를 통해 실행 중인 `zen` 프로세스에 연결되어 다음을 수행할 수 있습니다:

- 연결된 프로세스 및 상태 확인
- 특정 프로세스에 작업 전송
- 승인, 오류 및 완료 알림 수신
- 작업 제어 (일시 중지/재개/취소)

## 지원 플랫폼

| 플랫폼 | 필요한 설정 |
|----------|----------------|
| [Telegram](#telegram) | BotFather token |
| [Discord](#discord) | Bot 애플리케이션 token |
| [Slack](#slack) | Bot + App tokens (Socket Mode) |
| [Lark/飞书](#lark飞书) | App ID + Secret |
| [Facebook Messenger](#facebook-messenger) | Page token + Verify token |

## 기본 설정

```json
{
  "bot": {
    "enabled": true,
    "socket_path": "/tmp/zen-bot.sock",
    "platforms": {
      // 플랫폼별 설정 (아래 참조)
    },
    "interaction": {
      "require_mention": true,
      "mention_keywords": ["@zen", "/zen"],
      "direct_message_mode": "always",
      "channel_mode": "mention"
    },
    "aliases": {
      "api": "/path/to/api-project",
      "web": "/path/to/web-project"
    },
    "notify": {
      "default_platform": "telegram",
      "default_chat_id": "-100123456789",
      "quiet_hours_start": "23:00",
      "quiet_hours_end": "07:00",
      "quiet_hours_zone": "Asia/Shanghai"
    }
  }
}
```

## Bot 명령어

| 명령어 | 설명 |
|---------|-------------|
| `list` | 연결된 모든 프로세스 나열 |
| `status [name]` | 프로세스 상태 표시 |
| `bind <name>` | 후속 명령을 위해 프로세스에 바인딩 |
| `pause [name]` | 현재 작업 일시 중지 |
| `resume [name]` | 일시 중지된 작업 재개 |
| `cancel [name]` | 현재 작업 취소 |
| `<name> <task>` | 프로세스에 작업 전송 |
| `help` | 사용 가능한 명령어 표시 |

### 자연어 지원

Bot은 여러 언어의 자연어 쿼리를 이해합니다:

- "show me the status of gozen"
- "gozen의 상태를 보여줘"
- "list all processes"
- "pause the api project"

## 상호작용 모드

### 다이렉트 메시지

`direct_message_mode`를 설정하여 다이렉트 메시지에서 bot의 응답 방식을 제어합니다:

- `"always"` — 항상 응답 (멘션 불필요)
- `"mention"` — 멘션된 경우에만 응답

### 채널 메시지

`channel_mode`를 설정하여 그룹 채팅에서의 동작을 제어합니다:

- `"always"` — 모든 메시지에 응답
- `"mention"` — 멘션된 경우에만 응답 (권장)

### 멘션 키워드

bot을 트리거하는 키워드를 설정합니다:

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## 프로젝트 별칭

프로젝트에 대한 짧은 이름을 정의합니다:

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

명령어에서 사용:

```
api run tests
web build production
status backend
```

## 플랫폼 설정

### Telegram

1. [@BotFather](https://t.me/botfather)를 통해 bot 생성:
   - `/newbot`을 전송하고 안내를 따름
   - token 복사 (예: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. 사용자 ID 가져오기:
   - [@userinfobot](https://t.me/userinfobot)에 메시지 전송
   - 숫자 사용자 ID 복사

3. 설정:

```json
{
  "platforms": {
    "telegram": {
      "enabled": true,
      "token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
      "allowed_users": ["your_username", "123456789"],
      "allowed_chats": ["-100123456789"]
    }
  }
}
```

**보안 옵션:**
- `allowed_users` — bot과 상호작용할 수 있는 사용자명 또는 사용자 ID
- `allowed_chats` — bot이 응답하는 그룹 채팅 ID ([@getidsbot](https://t.me/getidsbot)을 통해 가져오기)

### Discord

1. Discord 애플리케이션 생성:
   - [Discord 개발자 포털](https://discord.com/developers/applications) 방문
   - "New Application" 클릭 및 이름 지정
   - "Bot" 섹션으로 이동하여 "Add Bot" 클릭
   - token 복사

2. 필요한 intents 활성화:
   - Bot 섹션에서 "Message Content Intent" 활성화
   - 사용자 필터링을 사용하는 경우 "Server Members Intent" 활성화

3. 서버에 bot 초대:
   - OAuth2 → URL Generator로 이동
   - scopes 선택: `bot`
   - 권한 선택: `Send Messages`, `Read Message History`
   - 생성된 URL로 초대

4. 설정:

```json
{
  "platforms": {
    "discord": {
      "enabled": true,
      "token": "MTIzNDU2Nzg5MDEyMzQ1Njc4.XXXXXX.XXXXXXXXXXXXXXXXXXXXXXXX",
      "allowed_users": ["user_id_1", "user_id_2"],
      "allowed_channels": ["channel_id_1"],
      "allowed_guilds": ["guild_id_1"]
    }
  }
}
```

**ID 가져오기:** Discord 설정에서 개발자 모드를 활성화한 다음 사용자/채널/서버를 우클릭하여 ID를 복사합니다.

### Slack

1. Slack 앱 생성:
   - [Slack API](https://api.slack.com/apps) 방문
   - "Create New App" → "From scratch" 클릭
   - 앱 이름 지정 및 워크스페이스 선택

2. Socket Mode 활성화:
   - "Socket Mode"로 이동하여 활성화
   - `connections:write` 범위로 앱 수준 Token 생성
   - token 복사 (`xapp-`로 시작)

3. Bot Token 설정:
   - "OAuth & Permissions"로 이동
   - scopes 추가: `chat:write`, `channels:history`, `groups:history`, `im:history`, `mpim:history`
   - 워크스페이스에 설치하고 Bot Token 복사 (`xoxb-`로 시작)

4. 이벤트 활성화:
   - "Event Subscriptions"로 이동하여 활성화
   - 구독: `message.channels`, `message.groups`, `message.im`, `message.mpim`

5. 설정:

```json
{
  "platforms": {
    "slack": {
      "enabled": true,
      "bot_token": "xoxb-xxx-xxx-xxx",
      "app_token": "xapp-xxx-xxx-xxx",
      "allowed_users": ["U12345678"],
      "allowed_channels": ["C12345678"]
    }
  }
}
```

### Lark/飞书

1. Lark 앱 생성:
   - [Lark 개방 플랫폼](https://open.larksuite.com/) 또는 [飞书开放平台](https://open.feishu.cn/) 방문
   - 새 앱 생성
   - App ID 및 App Secret 복사

2. 권한 설정:
   - `im:message:receive_v1` 이벤트 추가
   - `im:message:send_v1` 권한 추가

3. webhook 설정:
   - 이벤트 구독 URL 설정 (또는 WebSocket 모드 사용)

4. 설정:

```json
{
  "platforms": {
    "lark": {
      "enabled": true,
      "app_id": "cli_xxxxx",
      "app_secret": "xxxxxxxxxxxxx",
      "allowed_users": ["ou_xxxxx"],
      "allowed_chats": ["oc_xxxxx"]
    }
  }
}
```

### Facebook Messenger

1. Facebook 앱 생성:
   - [Facebook 개발자](https://developers.facebook.com/) 방문
   - "Business" 유형의 새 앱 생성
   - "Messenger" 제품 추가

2. Messenger 설정:
   - Page Access Token 생성
   - verify token으로 webhook 설정
   - `messages` 이벤트 구독

3. 설정:

```json
{
  "platforms": {
    "fbmessenger": {
      "enabled": true,
      "page_token": "EAAxxxxx",
      "verify_token": "your_verify_token",
      "app_secret": "xxxxx",
      "allowed_users": ["psid_1", "psid_2"]
    }
  }
}
```

**참고:** Facebook Messenger는 공개적으로 액세스 가능한 webhook URL이 필요합니다. 개발 시 ngrok 등의 서비스를 고려하세요.

## 알림

bot이 알림을 보낼 위치를 설정합니다:

```json
{
  "notify": {
    "default_platform": "telegram",
    "default_chat_id": "-100123456789",
    "quiet_hours_start": "23:00",
    "quiet_hours_end": "07:00",
    "quiet_hours_zone": "UTC"
  }
}
```

### 알림 유형

- **승인 요청** — Claude Code가 작업 권한이 필요할 때
- **작업 완료** — 작업이 성공적으로 완료되었을 때
- **오류** — 작업이 실패하거나 오류가 발생했을 때
- **상태 변경** — 프로세스가 연결/연결 해제되었을 때

### 방해 금지 시간

방해 금지 시간 동안 긴급하지 않은 알림은 억제됩니다. 승인 요청은 항상 전송됩니다.

## 보안 모범 사례

1. **사용자 제한** — 항상 `allowed_users`를 설정하여 세션을 제어할 수 있는 사람을 제한
2. **비공개 채널 사용** — 공개 채널에서 bot 사용 방지
3. **token 보호** — bot token을 버전 관리에 커밋하지 않음
4. **승인 검토** — 수락하기 전에 승인 요청을 신중하게 검토

## 문제 해결

### Bot이 응답하지 않음

1. 데몬이 실행 중인지 확인: `zen daemon status`
2. Web UI에서 bot 설정 확인
3. 데몬 로그 확인: `tail -f ~/.zen/zend.log`

### 연결 문제

1. token이 올바른지 확인
2. 네트워크 연결 확인
3. Slack/Discord의 경우 필요한 intents가 활성화되어 있는지 확인

### 프로세스가 목록에 표시되지 않음

1. 프로세스가 `zen`으로 시작되었는지 확인 (`claude`로 직접 시작하지 않음)
2. socket 경로가 bot 설정과 일치하는지 확인
