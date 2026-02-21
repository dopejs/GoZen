import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useUpdateBot } from '@/hooks/use-bot'
import type { BotConfig } from '@/types/api'
import type { TabProps } from './types'

export function InteractionTab({ config, setConfig }: TabProps) {
  const { t } = useTranslation()
  const updateBot = useUpdateBot()

  const handleSave = async () => {
    try {
      await updateBot.mutateAsync({ interaction: config.interaction })
      toast.success(t('common.success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const updateInteraction = (updates: Partial<NonNullable<BotConfig['interaction']>>) => {
    setConfig((c) => ({
      ...c,
      interaction: { ...c.interaction, ...updates },
    }))
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Users className="h-5 w-5" />
          {t('bot.interaction')}
        </CardTitle>
        <CardDescription>{t('bot.interactionDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-2">
          <Label htmlFor="require-mention">{t('bot.requireMention')}</Label>
          <Switch
            id="require-mention"
            checked={config.interaction?.require_mention ?? true}
            onCheckedChange={(checked) => updateInteraction({ require_mention: checked })}
          />
        </div>

        <div className="grid gap-2">
          <Label>{t('bot.mentionKeywords')}</Label>
          <Input
            value={config.interaction?.mention_keywords?.join(', ') || ''}
            onChange={(e) => updateInteraction({ mention_keywords: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
            placeholder="@zen, /zen, zen"
          />
        </div>

        <div className="grid gap-2">
          <Label>{t('bot.directMsgMode')}</Label>
          <Select
            value={config.interaction?.direct_message_mode || 'always'}
            onValueChange={(value) => updateInteraction({ direct_message_mode: value })}
          >
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="always">{t('bot.modeAlways')}</SelectItem>
              <SelectItem value="mention">{t('bot.modeMention')}</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="grid gap-2">
          <Label>{t('bot.channelMode')}</Label>
          <Select
            value={config.interaction?.channel_mode || 'mention'}
            onValueChange={(value) => updateInteraction({ channel_mode: value })}
          >
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="always">{t('bot.modeAlways')}</SelectItem>
              <SelectItem value="mention">{t('bot.modeMention')}</SelectItem>
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
