import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Bot } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useUpdateBot } from '@/hooks/use-bot'
import { useProfiles } from '@/hooks/use-profiles'
import { useSettings } from '@/hooks/use-settings'
import type { TabProps } from './types'

export function GeneralTab({ config, setConfig }: TabProps) {
  const { t } = useTranslation()
  const { data: profiles } = useProfiles()
  const { data: settings } = useSettings()
  const updateBot = useUpdateBot()

  const effectiveProfile = config.profile || settings?.default_profile || ''

  const handleSave = async () => {
    try {
      await updateBot.mutateAsync({
        enabled: config.enabled,
        profile: effectiveProfile,
        socket_path: config.socket_path,
      })
      toast.success(t('common.success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bot className="h-5 w-5" />
          {t('bot.general')}
        </CardTitle>
        <CardDescription>{t('bot.generalDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-2">
          <Label htmlFor="bot-enabled">{t('bot.enabled')}</Label>
          <Switch
            id="bot-enabled"
            checked={config.enabled ?? false}
            onCheckedChange={(checked) => setConfig((c) => ({ ...c, enabled: checked }))}
          />
        </div>

        <div className="grid gap-2">
          <Label htmlFor="bot-profile">{t('bot.profile')}</Label>
          <Select
            value={effectiveProfile}
            onValueChange={(value) => setConfig((c) => ({ ...c, profile: value }))}
          >
            <SelectTrigger id="bot-profile">
              <SelectValue placeholder={t('bot.selectProfile')} />
            </SelectTrigger>
            <SelectContent>
              {profiles?.map((p) => (
                <SelectItem key={p.name} value={p.name}>{p.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">{t('bot.profileHint')}</p>
        </div>

        <div className="grid gap-2">
          <Label htmlFor="socket-path">{t('bot.socketPath')}</Label>
          <Input
            id="socket-path"
            value={config.socket_path || ''}
            onChange={(e) => setConfig((c) => ({ ...c, socket_path: e.target.value }))}
            placeholder="/tmp/zen-gateway.sock"
          />
          <p className="text-xs text-muted-foreground">{t('bot.socketPathHint')}</p>
        </div>

        <Button onClick={handleSave} disabled={updateBot.isPending}>
          {t('common.save')}
        </Button>
      </CardContent>
    </Card>
  )
}
