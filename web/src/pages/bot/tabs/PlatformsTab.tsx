import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { MessageSquare, ChevronDown, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { useUpdateBot } from '@/hooks/use-bot'
import type { BotConfig } from '@/types/api'
import type { TabProps } from './types'

export function PlatformsTab({ config, setConfig }: TabProps) {
  const { t } = useTranslation()
  const updateBot = useUpdateBot()
  const [openPlatforms, setOpenPlatforms] = useState<Record<string, boolean>>({})

  const togglePlatform = (platform: string) => {
    setOpenPlatforms((prev) => ({ ...prev, [platform]: !prev[platform] }))
  }

  const handleSave = async () => {
    try {
      await updateBot.mutateAsync({ platforms: config.platforms })
      toast.success(t('common.success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const updatePlatform = <K extends keyof NonNullable<BotConfig['platforms']>>(
    platform: K,
    updates: Partial<NonNullable<BotConfig['platforms']>[K]>
  ) => {
    setConfig((c) => ({
      ...c,
      platforms: {
        ...c.platforms,
        [platform]: { ...c.platforms?.[platform], ...updates },
      },
    }))
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <MessageSquare className="h-5 w-5" />
          {t('bot.platforms')}
        </CardTitle>
        <CardDescription>{t('bot.platformsDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Telegram */}
        <Collapsible open={openPlatforms.telegram} onOpenChange={() => togglePlatform('telegram')}>
          <div className="flex items-center justify-between rounded-lg border p-3">
            <CollapsibleTrigger className="flex items-center gap-2">
              {openPlatforms.telegram ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
              <span className="font-medium">Telegram</span>
            </CollapsibleTrigger>
            <Switch
              checked={config.platforms?.telegram?.enabled ?? false}
              onCheckedChange={(checked) => updatePlatform('telegram', { enabled: checked })}
            />
          </div>
          <CollapsibleContent className="mt-2 space-y-3 rounded-lg border p-4">
            <div className="grid gap-2">
              <Label>{t('bot.token')}</Label>
              <Input type="password" value={config.platforms?.telegram?.token || ''} onChange={(e) => updatePlatform('telegram', { token: e.target.value })} placeholder="Bot Token" />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedUsers')}</Label>
              <Input value={config.platforms?.telegram?.allowed_users?.join(', ') || ''} onChange={(e) => updatePlatform('telegram', { allowed_users: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })} placeholder="user1, user2" />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedChats')}</Label>
              <Input value={config.platforms?.telegram?.allowed_chats?.join(', ') || ''} onChange={(e) => updatePlatform('telegram', { allowed_chats: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })} placeholder="-100123456789" />
            </div>
          </CollapsibleContent>
        </Collapsible>

        {/* Discord */}
        <Collapsible open={openPlatforms.discord} onOpenChange={() => togglePlatform('discord')}>
          <div className="flex items-center justify-between rounded-lg border p-3">
            <CollapsibleTrigger className="flex items-center gap-2">
              {openPlatforms.discord ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
              <span className="font-medium">Discord</span>
            </CollapsibleTrigger>
            <Switch
              checked={config.platforms?.discord?.enabled ?? false}
              onCheckedChange={(checked) => updatePlatform('discord', { enabled: checked })}
            />
          </div>
          <CollapsibleContent className="mt-2 space-y-3 rounded-lg border p-4">
            <div className="grid gap-2">
              <Label>{t('bot.token')}</Label>
              <Input type="password" value={config.platforms?.discord?.token || ''} onChange={(e) => updatePlatform('discord', { token: e.target.value })} placeholder="Bot Token" />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedUsers')}</Label>
              <Input value={config.platforms?.discord?.allowed_users?.join(', ') || ''} onChange={(e) => updatePlatform('discord', { allowed_users: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })} />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedChannels')}</Label>
              <Input value={config.platforms?.discord?.allowed_channels?.join(', ') || ''} onChange={(e) => updatePlatform('discord', { allowed_channels: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })} />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedGuilds')}</Label>
              <Input value={config.platforms?.discord?.allowed_guilds?.join(', ') || ''} onChange={(e) => updatePlatform('discord', { allowed_guilds: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })} />
            </div>
          </CollapsibleContent>
        </Collapsible>

        {/* Slack */}
        <Collapsible open={openPlatforms.slack} onOpenChange={() => togglePlatform('slack')}>
          <div className="flex items-center justify-between rounded-lg border p-3">
            <CollapsibleTrigger className="flex items-center gap-2">
              {openPlatforms.slack ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
              <span className="font-medium">Slack</span>
            </CollapsibleTrigger>
            <Switch
              checked={config.platforms?.slack?.enabled ?? false}
              onCheckedChange={(checked) => updatePlatform('slack', { enabled: checked })}
            />
          </div>
          <CollapsibleContent className="mt-2 space-y-3 rounded-lg border p-4">
            <div className="grid gap-2">
              <Label>{t('bot.botToken')}</Label>
              <Input type="password" value={config.platforms?.slack?.bot_token || ''} onChange={(e) => updatePlatform('slack', { bot_token: e.target.value })} placeholder="xoxb-..." />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.appToken')}</Label>
              <Input type="password" value={config.platforms?.slack?.app_token || ''} onChange={(e) => updatePlatform('slack', { app_token: e.target.value })} placeholder="xapp-..." />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedUsers')}</Label>
              <Input value={config.platforms?.slack?.allowed_users?.join(', ') || ''} onChange={(e) => updatePlatform('slack', { allowed_users: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })} />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedChannels')}</Label>
              <Input value={config.platforms?.slack?.allowed_channels?.join(', ') || ''} onChange={(e) => updatePlatform('slack', { allowed_channels: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })} />
            </div>
          </CollapsibleContent>
        </Collapsible>

        {/* Lark */}
        <Collapsible open={openPlatforms.lark} onOpenChange={() => togglePlatform('lark')}>
          <div className="flex items-center justify-between rounded-lg border p-3">
            <CollapsibleTrigger className="flex items-center gap-2">
              {openPlatforms.lark ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
              <span className="font-medium">Lark / Feishu</span>
            </CollapsibleTrigger>
            <Switch
              checked={config.platforms?.lark?.enabled ?? false}
              onCheckedChange={(checked) => updatePlatform('lark', { enabled: checked })}
            />
          </div>
          <CollapsibleContent className="mt-2 space-y-3 rounded-lg border p-4">
            <div className="grid gap-2">
              <Label>{t('bot.appId')}</Label>
              <Input value={config.platforms?.lark?.app_id || ''} onChange={(e) => updatePlatform('lark', { app_id: e.target.value })} />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.appSecret')}</Label>
              <Input type="password" value={config.platforms?.lark?.app_secret || ''} onChange={(e) => updatePlatform('lark', { app_secret: e.target.value })} />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedUsers')}</Label>
              <Input value={config.platforms?.lark?.allowed_users?.join(', ') || ''} onChange={(e) => updatePlatform('lark', { allowed_users: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })} />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedChats')}</Label>
              <Input value={config.platforms?.lark?.allowed_chats?.join(', ') || ''} onChange={(e) => updatePlatform('lark', { allowed_chats: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })} />
            </div>
          </CollapsibleContent>
        </Collapsible>

        {/* Facebook Messenger */}
        <Collapsible open={openPlatforms.fbmessenger} onOpenChange={() => togglePlatform('fbmessenger')}>
          <div className="flex items-center justify-between rounded-lg border p-3">
            <CollapsibleTrigger className="flex items-center gap-2">
              {openPlatforms.fbmessenger ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
              <span className="font-medium">Facebook Messenger</span>
            </CollapsibleTrigger>
            <Switch
              checked={config.platforms?.fbmessenger?.enabled ?? false}
              onCheckedChange={(checked) => updatePlatform('fbmessenger', { enabled: checked })}
            />
          </div>
          <CollapsibleContent className="mt-2 space-y-3 rounded-lg border p-4">
            <div className="grid gap-2">
              <Label>{t('bot.pageToken')}</Label>
              <Input type="password" value={config.platforms?.fbmessenger?.page_token || ''} onChange={(e) => updatePlatform('fbmessenger', { page_token: e.target.value })} />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.verifyToken')}</Label>
              <Input value={config.platforms?.fbmessenger?.verify_token || ''} onChange={(e) => updatePlatform('fbmessenger', { verify_token: e.target.value })} />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.appSecret')}</Label>
              <Input type="password" value={config.platforms?.fbmessenger?.app_secret || ''} onChange={(e) => updatePlatform('fbmessenger', { app_secret: e.target.value })} />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedUsers')}</Label>
              <Input value={config.platforms?.fbmessenger?.allowed_users?.join(', ') || ''} onChange={(e) => updatePlatform('fbmessenger', { allowed_users: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })} />
            </div>
          </CollapsibleContent>
        </Collapsible>

        <Button onClick={handleSave} disabled={updateBot.isPending}>
          {t('common.save')}
        </Button>
      </CardContent>
    </Card>
  )
}
