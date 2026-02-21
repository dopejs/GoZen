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

export function NotifyTab({ config, setConfig }: TabProps) {
  const { t } = useTranslation()
  const updateBot = useUpdateBot()

  const handleSave = async () => {
    try {
      await updateBot.mutateAsync({ notify: config.notify })
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
          <Input
            value={config.notify?.quiet_hours_zone || ''}
            onChange={(e) => updateNotify({ quiet_hours_zone: e.target.value })}
            placeholder="Asia/Shanghai"
          />
        </div>

        <Button onClick={handleSave} disabled={updateBot.isPending}>
          {t('common.save')}
        </Button>
      </CardContent>
    </Card>
  )
}
