import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Bell } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useUpdateBot } from '@/hooks/use-bot'
import type { BotConfig } from '@/types/api'
import type { TabProps } from './types'

const COMMON_TIMEZONES = [
  'Pacific/Honolulu', 'America/Anchorage', 'America/Los_Angeles', 'America/Denver',
  'America/Chicago', 'America/New_York', 'America/Sao_Paulo',
  'Europe/London', 'Europe/Paris', 'Europe/Berlin', 'Europe/Moscow',
  'Asia/Dubai', 'Asia/Kolkata', 'Asia/Bangkok', 'Asia/Singapore',
  'Asia/Shanghai', 'Asia/Tokyo', 'Asia/Seoul',
  'Australia/Sydney', 'Pacific/Auckland',
]

function getLocalTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch {
    return 'UTC'
  }
}

function formatUtcOffset(tz: string): string {
  try {
    const now = new Date()
    const formatter = new Intl.DateTimeFormat('en-US', { timeZone: tz, timeZoneName: 'shortOffset' })
    const parts = formatter.formatToParts(now)
    const offsetPart = parts.find((p) => p.type === 'timeZoneName')
    return offsetPart?.value?.replace('GMT', 'UTC') || ''
  } catch {
    return ''
  }
}

function formatTzLabel(tz: string): string {
  const offset = formatUtcOffset(tz)
  return offset ? `${tz} (${offset})` : tz
}

export function NotifyTab({ config, setConfig }: TabProps) {
  const { t } = useTranslation()
  const updateBot = useUpdateBot()
  const localTz = useMemo(() => getLocalTimezone(), [])

  // Build timezone list: local tz first (if not already in list), then common ones
  const timezones = useMemo(() => {
    const list = COMMON_TIMEZONES.includes(localTz)
      ? COMMON_TIMEZONES
      : [localTz, ...COMMON_TIMEZONES]
    return list
  }, [localTz])

  const handleSave = async () => {
    try {
      await updateBot.mutateAsync(config)
      toast.success(t('common.success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const updateNotify = (updates: Partial<NonNullable<BotConfig['notify']>>) => {
    setConfig((c) => ({
      ...c,
      notify: { ...c.notify, ...updates },
    }))
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bell className="h-5 w-5" />
          {t('bot.notify')}
        </CardTitle>
        <CardDescription>{t('bot.notifyDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-2">
          <Label>{t('bot.defaultPlatform')}</Label>
          <Select
            value={config.notify?.default_platform || ''}
            onValueChange={(value) => updateNotify({ default_platform: value })}
          >
            <SelectTrigger>
              <SelectValue placeholder={t('bot.selectPlatform')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="telegram">Telegram</SelectItem>
              <SelectItem value="discord">Discord</SelectItem>
              <SelectItem value="slack">Slack</SelectItem>
              <SelectItem value="lark">Lark</SelectItem>
              <SelectItem value="fbmessenger">Facebook Messenger</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="grid gap-2">
          <Label>{t('bot.defaultChatId')}</Label>
          <Input
            value={config.notify?.default_chat_id || ''}
            onChange={(e) => updateNotify({ default_chat_id: e.target.value })}
            placeholder="Chat ID"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="grid gap-2">
            <Label>{t('bot.quietHoursStart')}</Label>
            <Input
              type="time"
              value={config.notify?.quiet_hours_start || ''}
              onChange={(e) => updateNotify({ quiet_hours_start: e.target.value })}
            />
          </div>
          <div className="grid gap-2">
            <Label>{t('bot.quietHoursEnd')}</Label>
            <Input
              type="time"
              value={config.notify?.quiet_hours_end || ''}
              onChange={(e) => updateNotify({ quiet_hours_end: e.target.value })}
            />
          </div>
        </div>

        <div className="grid gap-2">
          <Label>{t('bot.quietHoursZone')}</Label>
          <Select
            value={config.notify?.quiet_hours_zone || localTz}
            onValueChange={(value) => updateNotify({ quiet_hours_zone: value })}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {timezones.map((tz) => (
                <SelectItem key={tz} value={tz}>{formatTzLabel(tz)}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <Button onClick={handleSave} disabled={updateBot.isPending}>
          {t('common.save')}
        </Button>
      </CardContent>
    </Card>
  )
}
