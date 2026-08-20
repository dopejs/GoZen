---
sidebar_position: 10
title: שער בוטים
---

# שער בוטים

נטרו ושלטו בסשני Claude Code מרחוק דרך פלטפורמות צ׳אט. הבוט מתחבר לתהליכי `zen` פעילים דרך IPC ומאפשר לכם:

- לראות תהליכים מחוברים ואת מצבם
- לשלוח משימות לתהליך מסוים
- לקבל התראות על אישורים, שגיאות וסיומים
- לשלוט במשימות (השהיה/המשך/ביטול)

## פלטפורמות נתמכות

| פלטפורמה | מה נדרש להגדיר |
|----------|----------------|
| [Telegram](#telegram) | אסימון מ-BotFather |
| [Discord](#discord) | אסימון של אפליקציית הבוט |
| [Slack](#slack) | אסימוני Bot ו-App (מצב Socket) |
| [Lark/Feishu](#larkfeishu) | ‏App ID ו-Secret |
| [Facebook Messenger](#facebook-messenger) | אסימון עמוד ואסימון אימות |

## הגדרה בסיסית

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

## פקודות הבוט

| פקודה | תיאור |
|---------|-------------|
| `list` | מציג את כל התהליכים המחוברים |
| `status [name]` | מציג את מצב התהליך |
| `bind <name>` | מתחבר לתהליך עבור הפקודות הבאות |
| `pause [name]` | משהה את המשימה הנוכחית |
| `resume [name]` | ממשיך משימה מושהית |
| `cancel [name]` | מבטל את המשימה הנוכחית |
| `<name> <task>` | שולח משימה לתהליך |
| `help` | מציג את הפקודות הזמינות |

### תמיכה בשפה טבעית

הבוט מבין שאילתות בשפה טבעית בכמה שפות:

- "show me the status of gozen"
- "帮我看看 gozen 的状态"
- "list all processes"
- "pause the api project"

## מצבי אינטראקציה

### הודעות ישירות

‏`direct_message_mode` קובע איך הבוט מגיב בהודעות ישירות:

- `"always"` — מגיב תמיד (אין צורך באזכור)
- `"mention"` — מגיב רק כשמאזכרים אותו

### הודעות בערוץ

‏`channel_mode` קובע את ההתנהגות בצ׳אטים קבוצתיים:

- `"always"` — מגיב לכל ההודעות
- `"mention"` — מגיב רק כשמאזכרים אותו (מומלץ)

### מילות אזכור

הגדירו מה מפעיל את הבוט:

```json
{
  "interaction": {
    "require_mention": true,
    "mention_keywords": ["@zen", "/zen", "zen"]
  }
}
```

## כינויי פרויקטים

הגדירו שמות קצרים לפרויקטים שלכם:

```json
{
  "aliases": {
    "api": "/Users/john/projects/api-server",
    "web": "/Users/john/projects/web-app",
    "backend": "/Users/john/work/backend"
  }
}
```

ואז השתמשו בהם בפקודות:

```
api run tests
web build production
status backend
```

## הגדרה לפי פלטפורמה

### Telegram

1. צרו בוט דרך [@BotFather](https://t.me/botfather):
   - שלחו `/newbot` ועקבו אחרי ההנחיות
   - העתיקו את האסימון (למשל `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. מצאו את מזהה המשתמש שלכם:
   - שלחו הודעה ל-[@userinfobot](https://t.me/userinfobot)
   - העתיקו את המזהה המספרי

3. הגדירו:

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

**אפשרויות אבטחה:**
- `allowed_users` — שמות משתמש או מזהים שרשאים לתקשר עם הבוט
- `allowed_chats` — מזהי קבוצות שבהן הבוט מגיב (ניתן לקבל דרך [@getidsbot](https://t.me/getidsbot))

### Discord

1. צרו אפליקציית Discord:
   - היכנסו ל-[פורטל המפתחים של Discord](https://discord.com/developers/applications)
   - לחצו על "New Application" ותנו לה שם
   - עברו לחלק "Bot" ולחצו על "Add Bot"
   - העתיקו את האסימון

2. הפעילו את ה-intents הנדרשים:
   - בחלק Bot הפעילו "Message Content Intent"
   - הפעילו "Server Members Intent" אם אתם מסננים לפי משתמשים

3. הזמינו את הבוט לשרת:
   - עברו ל-OAuth2 ← URL Generator
   - בחרו scope: ‏`bot`
   - בחרו הרשאות: `Send Messages`, ‏`Read Message History`
   - השתמשו בכתובת שנוצרה כדי להזמין

4. הגדירו:

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

**קבלת מזהים:** הפעילו מצב מפתח בהגדרות Discord, ואז לחצו לחיצה ימנית על משתמשים/ערוצים/שרתים כדי להעתיק מזהים.

### Slack

1. צרו אפליקציית Slack:
   - היכנסו ל-[Slack API](https://api.slack.com/apps)
   - לחצו על "Create New App" ← "From scratch"
   - תנו שם לאפליקציה ובחרו מרחב עבודה

2. הפעילו את Socket Mode:
   - עברו ל-"Socket Mode" והפעילו אותו
   - צרו App-Level Token עם ההרשאה `connections:write`
   - העתיקו את האסימון (מתחיל ב-`xapp-`)

3. הגדירו את אסימון הבוט:
   - עברו ל-"OAuth & Permissions"
   - הוסיפו הרשאות: `chat:write`, ‏`channels:history`, ‏`groups:history`, ‏`im:history`, ‏`mpim:history`
   - התקינו במרחב העבודה והעתיקו את ה-Bot Token (מתחיל ב-`xoxb-`)

4. הפעילו אירועים:
   - עברו ל-"Event Subscriptions" והפעילו
   - הירשמו ל-: `message.channels`, ‏`message.groups`, ‏`message.im`, ‏`message.mpim`

5. הגדירו:

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

1. צרו אפליקציית Lark:
   - היכנסו ל-[Lark Open Platform](https://open.larksuite.com/) או ל-[Feishu Open Platform](https://open.feishu.cn/)
   - צרו אפליקציה חדשה
   - העתיקו את App ID ו-App Secret

2. הגדירו הרשאות:
   - הוסיפו את האירוע `im:message:receive_v1`
   - הוסיפו את ההרשאה `im:message:send_v1`

3. הגדירו webhook:
   - הגדירו כתובת להרשמה לאירועים (או השתמשו במצב WebSocket)

4. הגדירו:

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

1. צרו אפליקציית Facebook:
   - היכנסו ל-[Facebook Developers](https://developers.facebook.com/)
   - צרו אפליקציה מסוג "Business"
   - הוסיפו את המוצר "Messenger"

2. הגדירו את Messenger:
   - צרו Page Access Token
   - הגדירו webhook עם אסימון אימות
   - הירשמו לאירוע `messages`

3. הגדירו:

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

**שימו לב:** ‏Facebook Messenger דורש כתובת webhook נגישה מהאינטרנט. בפיתוח אפשר להיעזר בשירות כמו ngrok.

## התראות

הגדירו לאן הבוט שולח התראות:

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

### סוגי התראות

- **בקשות אישור** — כש-Claude Code זקוק להרשאה לפעולה
- **סיום משימה** — כשמשימה מסתיימת בהצלחה
- **שגיאות** — כשמשימה נכשלת או נתקלת בשגיאה
- **שינויי מצב** — כשתהליך מתחבר או מתנתק

### שעות שקט

בשעות השקט התראות שאינן דחופות מושתקות. בקשות אישור נשלחות תמיד.

## המלצות אבטחה

1. **הגבילו משתמשים** — הגדירו תמיד `allowed_users` כדי להגביל מי יכול לשלוט בסשנים
2. **השתמשו בערוצים פרטיים** — הימנעו משימוש בבוט בערוצים ציבוריים
3. **שמרו על האסימונים** — לעולם אל תעלו אסימוני בוט לניהול גרסאות
4. **בדקו בקשות אישור** — קראו בעיון כל בקשה לפני אישור

## פתרון תקלות

### הבוט אינו מגיב

1. בדקו שהדימון פועל: `zen daemon status`
2. ודאו את הגדרות הבוט בממשק הווב
3. בדקו את יומני הדימון: `tail -f ~/.zen/zend.log`

### בעיות התחברות

1. ודאו שהאסימונים נכונים
2. בדקו קישוריות רשת
3. ב-Slack ו-Discord ודאו שה-intents הנדרשים מופעלים

### תהליך אינו מופיע ברשימה

1. ודאו שהתהליך הופעל דרך `zen` ולא ישירות דרך `claude`
2. בדקו שנתיב ה-socket תואם להגדרת הבוט
