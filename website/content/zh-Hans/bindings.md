---
sidebar_position: 5
title: 项目绑定
---

# 项目绑定

将目录绑定到特定 profile 和/或 CLI，实现项目级自动配置。

## 使用方法

```bash
cd ~/work/company-project

# Bind profile
zen bind work-profile

# Bind CLI
zen bind --cli codex

# Bind both
zen bind work-profile --cli codex

# Check status
zen status

# Unbind
zen unbind
```

## 优先级

命令行参数 > 项目绑定 > 全局默认
