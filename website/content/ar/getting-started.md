---
sidebar_position: 1
title: البدء
---

# البدء

## التثبيت

ثبِّت بسطر واحد:

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

إلغاء التثبيت:

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## أول تشغيل

أضف أول مزوّد لديك:

```bash
zen config add provider
```

التشغيل بملف التعريف الافتراضي:

```bash
zen
```

التشغيل بملف تعريف محدد:

```bash
zen -p work
```

التشغيل بواجهة سطر أوامر محددة:

```bash
zen --cli codex
```
