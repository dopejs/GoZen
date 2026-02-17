---
sidebar_position: 1
title: 快速开始
---

# 快速开始

## 安装

使用安装脚本一键安装：

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

卸载：

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## 首次使用

添加第一个 provider：

```bash
zen config add provider
```

使用默认 profile 启动：

```bash
zen
```

使用指定 profile 启动：

```bash
zen -p work
```

使用指定 CLI 启动：

```bash
zen --cli codex
```
