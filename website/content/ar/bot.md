---
sidebar_position: 10
title: بوابة الروبوتات
---

# بوابة الروبوتات

راقب جلسات Claude Code وتحكّم بها عن بُعد من منصات المحادثة. يتصل الروبوت بعمليات `zen` الجارية عبر IPC ويتيح لك:

- عرض العمليات المتصلة وحالتها
- إرسال المهام إلى عملية بعينها
- استقبال إشعارات الموافقات والأخطاء وحالات الإنجاز
- التحكم بالمهام (إيقاف مؤقت/استئناف/إلغاء)

## المنصات المدعومة

| المنصة | ما يلزم إعداده |
|----------|----------------|
| [Telegram](#telegram) | رمز من BotFather |
| [Discord](#discord) | رمز تطبيق الروبوت |
| [Slack](#slack) | رمزا Bot و App (وضع Socket) |
| [Lark/Feishu](#larkfeishu) | ‏App ID و Secret |
| [Facebook Messenger](#facebook-messenger) | رمز الصفحة ورمز التحقق |

## الإعداد الأساسي

```json
{
  "bot": {
    "enabled": true,
    "socket_path": "/tmp/zen-bot.sock",
    "platforms": {
      // Platform-specific config (see below)
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

## أوامر الروبوت

| الأمر | الوصف |
|---------|-------------|
| `list` | يسرد كل العمليات المتصلة |
| `status [name]` | يعرض حالة عملية |
| `bind <name>` | يرتبط بعملية لتنفيذ الأوامر التالية عليها |
| `pause [name]` | يوقف المهمة الحالية مؤقتًا |
| `resume [name]` | يستأنف مهمة موقوفة |
| `cancel [name]` | يلغي المهمة الحالية |
| `<name> <task>` | يرسل مهمة إلى عملية |
| `help` | يعرض الأوامر المتاحة |

### دعم اللغة الطبيعية

يفهم الروبوت الاستفسارات بلغة طبيعية وبعدة لغات:

- "show me the status of gozen"
- "帮我看看 gozen 的状态"
- "list all processes"
- "pause the api project"

## أنماط التفاعل

### الرسائل المباشرة

يحدّد `direct_message_mode` كيفية رد الروبوت في الرسائل المباشرة:

- `"always"` — يرد دائمًا (دون حاجة إلى إشارة)
- `"mention"` — يرد فقط عند الإشارة إليه

### رسائل القنوات

يحدّد `channel_mode` السلوك في المحادثات الجماعية:

- `"always"` — يرد على كل الرسائل
- `"mention"` — يرد فقط عند الإشارة إليه (موصى به)

### كلمات الإشارة

اضبط ما الذي يستدعي الروبوت:

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## أسماء مختصرة للمشاريع

عرّف أسماء قصيرة لمشاريعك:

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

ثم استخدمها في الأوامر:

```
api run tests
web build production
status backend
```

## الإعداد حسب المنصة

### Telegram

1. أنشئ روبوتًا عبر [@BotFather](https://t.me/botfather):
   - أرسل `/newbot` واتبع التعليمات
   - انسخ الرمز (مثل `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. احصل على معرّف المستخدم الخاص بك:
   - أرسل رسالة إلى [@userinfobot](https://t.me/userinfobot)
   - انسخ المعرّف الرقمي

3. اضبط الإعدادات:

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

**خيارات الأمان:**
- `allowed_users` — أسماء أو معرّفات المستخدمين المسموح لهم بالتفاعل مع الروبوت
- `allowed_chats` — معرّفات المحادثات الجماعية التي يرد فيها الروبوت (تُعرف عبر [@getidsbot](https://t.me/getidsbot))

### Discord

1. أنشئ تطبيق Discord:
   - انتقل إلى [بوابة مطوّري Discord](https://discord.com/developers/applications)
   - اضغط "New Application" وسمِّه
   - انتقل إلى قسم "Bot" واضغط "Add Bot"
   - انسخ الرمز

2. فعّل الـ intents المطلوبة:
   - في قسم Bot فعّل "Message Content Intent"
   - فعّل "Server Members Intent" إذا كنت تصفّي حسب المستخدمين

3. ادعُ الروبوت إلى خادمك:
   - انتقل إلى OAuth2 ← URL Generator
   - اختر النطاق: `bot`
   - اختر الصلاحيات: `Send Messages` و `Read Message History`
   - استخدم الرابط الناتج للدعوة

4. اضبط الإعدادات:

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

**الحصول على المعرّفات:** فعّل وضع المطوّر في إعدادات Discord، ثم انقر بزر الفأرة الأيمن على المستخدمين أو القنوات أو الخوادم لنسخ المعرّفات.

### Slack

1. أنشئ تطبيق Slack:
   - انتقل إلى [Slack API](https://api.slack.com/apps)
   - اضغط "Create New App" ← "From scratch"
   - سمِّ التطبيق واختر مساحة العمل

2. فعّل وضع Socket:
   - انتقل إلى "Socket Mode" وفعّله
   - أنشئ App-Level Token بالصلاحية `connections:write`
   - انسخ الرمز (يبدأ بـ `xapp-`)

3. اضبط رمز الروبوت:
   - انتقل إلى "OAuth & Permissions"
   - أضف الصلاحيات: `chat:write` و `channels:history` و `groups:history` و `im:history` و `mpim:history`
   - ثبّت التطبيق في مساحة العمل وانسخ Bot Token (يبدأ بـ `xoxb-`)

4. فعّل الأحداث:
   - انتقل إلى "Event Subscriptions" وفعّلها
   - اشترك في: `message.channels` و `message.groups` و `message.im` و `message.mpim`

5. اضبط الإعدادات:

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

### Lark/Feishu

1. أنشئ تطبيق Lark:
   - انتقل إلى [Lark Open Platform](https://open.larksuite.com/) أو [Feishu Open Platform](https://open.feishu.cn/)
   - أنشئ تطبيقًا جديدًا
   - انسخ App ID و App Secret

2. اضبط الصلاحيات:
   - أضف الحدث `im:message:receive_v1`
   - أضف الصلاحية `im:message:send_v1`

3. اضبط الـ webhook:
   - حدّد عنوان الاشتراك في الأحداث (أو استخدم وضع WebSocket)

4. اضبط الإعدادات:

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

1. أنشئ تطبيق Facebook:
   - انتقل إلى [Facebook Developers](https://developers.facebook.com/)
   - أنشئ تطبيقًا من نوع "Business"
   - أضف منتج "Messenger"

2. اضبط Messenger:
   - أنشئ Page Access Token
   - اضبط الـ webhook مع رمز تحقق
   - اشترك في الحدث `messages`

3. اضبط الإعدادات:

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

**ملاحظة:** يتطلب Facebook Messenger عنوان webhook متاحًا للعموم. وللتطوير يمكن الاستعانة بخدمة مثل ngrok.

## الإشعارات

حدّد إلى أين يرسل الروبوت الإشعارات:

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

### أنواع الإشعارات

- **طلبات الموافقة** — حين يحتاج Claude Code إذنًا لتنفيذ إجراء
- **إنجاز المهمة** — حين تنتهي المهمة بنجاح
- **الأخطاء** — حين تفشل المهمة أو تواجه خطأ
- **تغيّر الحالة** — حين تتصل عملية أو ينقطع اتصالها

### ساعات الهدوء

خلال ساعات الهدوء تُكتم الإشعارات غير العاجلة، أما طلبات الموافقة فتُرسل دائمًا.

## ممارسات أمنية مفضّلة

1. **قيّد المستخدمين** — اضبط `allowed_users` دائمًا لتحديد من يمكنه التحكم بجلساتك
2. **استخدم قنوات خاصة** — تجنّب استخدام الروبوت في القنوات العامة
3. **احمِ الرموز** — لا تضع رموز الروبوت في نظام إدارة الإصدارات أبدًا
4. **راجع الموافقات** — اقرأ طلبات الموافقة بعناية قبل القبول

## معالجة المشكلات

### الروبوت لا يستجيب

1. تحقّق من أن الخدمة تعمل: `zen daemon status`
2. تأكد من إعداد الروبوت في واجهة الويب
3. راجع سجلات الخدمة: `tail -f ~/.zen/zend.log`

### مشكلات في الاتصال

1. تأكد من صحة الرموز
2. تحقّق من الاتصال بالشبكة
3. في Slack و Discord تأكد من تفعيل الـ intents المطلوبة

### العملية لا تظهر في القائمة

1. تأكد أن العملية بدأت عبر `zen` وليس مباشرة عبر `claude`
2. تأكد أن مسار الـ socket يطابق إعداد الروبوت
