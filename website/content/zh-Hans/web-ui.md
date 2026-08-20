---
sidebar_position: 7
title: Web 管理界面
---

# Web 管理界面

通过浏览器可视化管理所有配置。守护进程会在需要时自动启动。

## 使用方法

```bash
# Open in browser (auto-starts daemon if needed)
zen web
```

## 功能

- Provider 和 Profile 管理
- 项目绑定管理
- 全局设置（默认客户端、默认 Profile、端口）
- 配置同步设置
- 请求日志查看（支持自动刷新）
- 模型字段自动补全

## 安全

守护进程首次启动时自动生成访问密码。非本地请求（127.0.0.1/::1 以外）需要登录。

- 会话认证，支持可配置的过期时间
- 暴力破解保护，指数级退避
- RSA 加密敏感令牌传输（API 密钥在浏览器端加密）
- 本地访问（127.0.0.1）免认证

### Password Management

```bash
# Reset the Web UI password
zen config reset-password

# Change password via Web UI
zen web  # Settings → Change Password
```
