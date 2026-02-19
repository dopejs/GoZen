import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import {
  Bot,
  MessageSquare,
  Users,
  Bell,
  ChevronDown,
  ChevronRight,
  Plus,
  Trash2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { useBot, useUpdateBot } from '@/hooks/use-bot'
import { useProfiles } from '@/hooks/use-profiles'
import type { BotConfig } from '@/types/api'

export function BotPage() {
  const { t } = useTranslation()
  const { data: bot, isLoading } = useBot()
  const [localConfig, setLocalConfig] = useState<Partial<BotConfig>>({})

  useEffect(() => {
    if (bot) {
      setLocalConfig(bot)
    }
  }, [bot])

  if (isLoading) {
    return (
      <div className="flex justify-center py-8">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">{t('bot.title')}</h1>
        <p className="text-muted-foreground">{t('bot.description')}</p>
      </div>

      <Tabs defaultValue="general">
        <TabsList>
          <TabsTrigger value="general">{t('bot.general')}</TabsTrigger>
          <TabsTrigger value="platforms">{t('bot.platforms')}</TabsTrigger>
          <TabsTrigger value="interaction">{t('bot.interaction')}</TabsTrigger>
          <TabsTrigger value="aliases">{t('bot.aliases')}</TabsTrigger>
          <TabsTrigger value="notify">{t('bot.notify')}</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-4">
          <GeneralTab config={localConfig} setConfig={setLocalConfig} />
        </TabsContent>

        <TabsContent value="platforms" className="mt-4">
          <PlatformsTab config={localConfig} setConfig={setLocalConfig} />
        </TabsContent>

        <TabsContent value="interaction" className="mt-4">
          <InteractionTab config={localConfig} setConfig={setLocalConfig} />
        </TabsContent>

        <TabsContent value="aliases" className="mt-4">
          <AliasesTab config={localConfig} setConfig={setLocalConfig} />
        </TabsContent>

        <TabsContent value="notify" className="mt-4">
          <NotifyTab config={localConfig} setConfig={setLocalConfig} />
        </TabsContent>
      </Tabs>
    </div>
  )
}

interface TabProps {
  config: Partial<BotConfig>
  setConfig: React.Dispatch<React.SetStateAction<Partial<BotConfig>>>
}

function GeneralTab({ config, setConfig }: TabProps) {
  const { t } = useTranslation()
  const { data: profiles } = useProfiles()
  const updateBot = useUpdateBot()

  const handleSave = async () => {
    try {
      await updateBot.mutateAsync({
        enabled: config.enabled,
        profile: config.profile,
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
        <div className="flex items-center justify-between">
          <Label htmlFor="bot-enabled">{t('bot.enabled')}</Label>
          <Switch
            id="bot-enabled"
            checked={config.enabled ?? false}
            onCheckedChange={(checked) => setConfig((c) => ({ ...c, enabled: checked }))}
          />
        </div>

        <div className="grid gap-2">
          <Label htmlFor="bot-profile">{t('bot.nluProfile')}</Label>
          <Select
            value={config.profile || ''}
            onValueChange={(value) => setConfig((c) => ({ ...c, profile: value }))}
          >
            <SelectTrigger id="bot-profile">
              <SelectValue placeholder={t('bot.selectProfile')} />
            </SelectTrigger>
            <SelectContent>
              {profiles?.map((p) => (
                <SelectItem key={p.name} value={p.name}>
                  {p.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="grid gap-2">
          <Label htmlFor="socket-path">{t('bot.socketPath')}</Label>
          <Input
            id="socket-path"
            value={config.socket_path || ''}
            onChange={(e) => setConfig((c) => ({ ...c, socket_path: e.target.value }))}
            placeholder="/tmp/zen-gateway.sock"
          />
        </div>

        <Button onClick={handleSave} disabled={updateBot.isPending}>
          {t('common.save')}
        </Button>
      </CardContent>
    </Card>
  )
}

function PlatformsTab({ config, setConfig }: TabProps) {
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
              <Input
                type="password"
                value={config.platforms?.telegram?.token || ''}
                onChange={(e) => updatePlatform('telegram', { token: e.target.value })}
                placeholder="Bot Token"
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedUsers')}</Label>
              <Input
                value={config.platforms?.telegram?.allowed_users?.join(', ') || ''}
                onChange={(e) => updatePlatform('telegram', { allowed_users: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
                placeholder="user1, user2"
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedChats')}</Label>
              <Input
                value={config.platforms?.telegram?.allowed_chats?.join(', ') || ''}
                onChange={(e) => updatePlatform('telegram', { allowed_chats: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
                placeholder="-100123456789"
              />
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
              <Input
                type="password"
                value={config.platforms?.discord?.token || ''}
                onChange={(e) => updatePlatform('discord', { token: e.target.value })}
                placeholder="Bot Token"
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedUsers')}</Label>
              <Input
                value={config.platforms?.discord?.allowed_users?.join(', ') || ''}
                onChange={(e) => updatePlatform('discord', { allowed_users: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedChannels')}</Label>
              <Input
                value={config.platforms?.discord?.allowed_channels?.join(', ') || ''}
                onChange={(e) => updatePlatform('discord', { allowed_channels: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedGuilds')}</Label>
              <Input
                value={config.platforms?.discord?.allowed_guilds?.join(', ') || ''}
                onChange={(e) => updatePlatform('discord', { allowed_guilds: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
              />
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
              <Input
                type="password"
                value={config.platforms?.slack?.bot_token || ''}
                onChange={(e) => updatePlatform('slack', { bot_token: e.target.value })}
                placeholder="xoxb-..."
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.appToken')}</Label>
              <Input
                type="password"
                value={config.platforms?.slack?.app_token || ''}
                onChange={(e) => updatePlatform('slack', { app_token: e.target.value })}
                placeholder="xapp-..."
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedUsers')}</Label>
              <Input
                value={config.platforms?.slack?.allowed_users?.join(', ') || ''}
                onChange={(e) => updatePlatform('slack', { allowed_users: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedChannels')}</Label>
              <Input
                value={config.platforms?.slack?.allowed_channels?.join(', ') || ''}
                onChange={(e) => updatePlatform('slack', { allowed_channels: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
              />
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
              <Input
                value={config.platforms?.lark?.app_id || ''}
                onChange={(e) => updatePlatform('lark', { app_id: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.appSecret')}</Label>
              <Input
                type="password"
                value={config.platforms?.lark?.app_secret || ''}
                onChange={(e) => updatePlatform('lark', { app_secret: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedUsers')}</Label>
              <Input
                value={config.platforms?.lark?.allowed_users?.join(', ') || ''}
                onChange={(e) => updatePlatform('lark', { allowed_users: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedChats')}</Label>
              <Input
                value={config.platforms?.lark?.allowed_chats?.join(', ') || ''}
                onChange={(e) => updatePlatform('lark', { allowed_chats: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
              />
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
              <Input
                type="password"
                value={config.platforms?.fbmessenger?.page_token || ''}
                onChange={(e) => updatePlatform('fbmessenger', { page_token: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.verifyToken')}</Label>
              <Input
                value={config.platforms?.fbmessenger?.verify_token || ''}
                onChange={(e) => updatePlatform('fbmessenger', { verify_token: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.appSecret')}</Label>
              <Input
                type="password"
                value={config.platforms?.fbmessenger?.app_secret || ''}
                onChange={(e) => updatePlatform('fbmessenger', { app_secret: e.target.value })}
              />
            </div>
            <div className="grid gap-2">
              <Label>{t('bot.allowedUsers')}</Label>
              <Input
                value={config.platforms?.fbmessenger?.allowed_users?.join(', ') || ''}
                onChange={(e) => updatePlatform('fbmessenger', { allowed_users: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
              />
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

function InteractionTab({ config, setConfig }: TabProps) {
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
        <div className="flex items-center justify-between">
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
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
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
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
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

function AliasesTab({ config, setConfig }: TabProps) {
  const { t } = useTranslation()
  const updateBot = useUpdateBot()
  const [newAlias, setNewAlias] = useState('')
  const [newPath, setNewPath] = useState('')

  const handleSave = async () => {
    try {
      await updateBot.mutateAsync({ aliases: config.aliases })
      toast.success(t('common.success'))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    }
  }

  const addAlias = () => {
    if (!newAlias || !newPath) return
    setConfig((c) => ({
      ...c,
      aliases: { ...c.aliases, [newAlias]: newPath },
    }))
    setNewAlias('')
    setNewPath('')
  }

  const removeAlias = (alias: string) => {
    setConfig((c) => {
      const newAliases = { ...c.aliases }
      delete newAliases[alias]
      return { ...c, aliases: newAliases }
    })
  }

  const aliases = Object.entries(config.aliases || {})

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('bot.aliases')}</CardTitle>
        <CardDescription>{t('bot.aliasesDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="rounded-lg border">
          <div className="grid grid-cols-[1fr_2fr_auto] gap-2 border-b bg-muted/50 p-3 text-sm font-medium">
            <div>{t('bot.alias')}</div>
            <div>{t('bot.projectPath')}</div>
            <div></div>
          </div>
          {aliases.length === 0 ? (
            <div className="p-4 text-center text-muted-foreground">{t('bot.noAliases')}</div>
          ) : (
            aliases.map(([alias, path]) => (
              <div key={alias} className="grid grid-cols-[1fr_2fr_auto] items-center gap-2 border-b p-3 last:border-0">
                <div className="font-mono text-sm">{alias}</div>
                <div className="truncate text-sm text-muted-foreground">{path}</div>
                <Button variant="ghost" size="icon" onClick={() => removeAlias(alias)}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))
          )}
        </div>

        <div className="flex gap-2">
          <Input
            placeholder={t('bot.alias')}
            value={newAlias}
            onChange={(e) => setNewAlias(e.target.value)}
            className="flex-1"
          />
          <Input
            placeholder={t('bot.projectPath')}
            value={newPath}
            onChange={(e) => setNewPath(e.target.value)}
            className="flex-[2]"
          />
          <Button variant="outline" onClick={addAlias}>
            <Plus className="h-4 w-4" />
          </Button>
        </div>

        <Button onClick={handleSave} disabled={updateBot.isPending}>
          {t('common.save')}
        </Button>
      </CardContent>
    </Card>
  )
}

function NotifyTab({ config, setConfig }: TabProps) {
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
