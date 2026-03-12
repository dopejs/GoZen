---
sidebar_position: 9
title: 설정 동기화
---

# 설정 동기화

제공자, 프로필, 기본 프로필, 기본 클라이언트를 여러 디바이스 간에 동기화할 수 있습니다. 인증 토큰은 업로드 전에 AES-256-GCM(PBKDF2-SHA256 키 파생)으로 암호화됩니다.

## 지원 백엔드

| 백엔드 | 설명 |
|--------|------|
| `webdav` | 임의의 WebDAV 서버 (예: Nextcloud, ownCloud) |
| `s3` | AWS S3 또는 S3 호환 스토리지 (예: MinIO, Cloudflare R2) |
| `gist` | 비공개 gist (`gist` 스코프 PAT 필요) |
| `repo` | Contents API를 통한 리포지토리 파일 (`repo` 스코프 PAT 필요) |

## 설정

Web UI의 Settings 페이지에서 동기화를 설정합니다:

```bash
# Web UI 설정 열기
zen web  # Settings → Config Sync
```

또는 CLI에서 수동으로 pull 할 수 있습니다:

```bash
zen config sync
```

## 설정 예시

```json
{
  "sync": {
    "backend": "gist",
    "gist_id": "abc123def456",
    "token": "ghp_xxxxxxxxxxxx",
    "passphrase": "my-secret-passphrase",
    "auto_pull": true,
    "pull_interval": 300
  }
}
```

## 암호화

passphrase를 설정하면 동기화 payload 안의 모든 인증 토큰은 PBKDF2-SHA256(60만 회 반복)으로 파생한 키를 사용해 AES-256-GCM으로 암호화됩니다. 암호화 salt는 payload와 함께 저장됩니다. passphrase가 없으면 데이터는 평문 JSON으로 업로드됩니다.

## 충돌 해결

- 엔터티별 타임스탬프 병합: 더 최근 수정이 우선
- 삭제된 엔터티는 tombstone 사용 (30일 후 만료)
- 스칼라 값(기본 프로필 / 기본 클라이언트): 더 최근 타임스탬프가 우선

## 동기화 범위

**동기화됨:** 제공자(암호화된 토큰 포함), 프로필, 기본 프로필, 기본 클라이언트

**동기화되지 않음:** 포트 설정, Web UI 비밀번호, 프로젝트 바인딩, 동기화 설정 자체
