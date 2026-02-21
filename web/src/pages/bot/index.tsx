import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useBot } from '@/hooks/use-bot'
import type { BotConfig } from '@/types/api'
import { GeneralTab, PlatformsTab, InteractionTab, AliasesTab, NotifyTab } from './tabs'

export function BotPage() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()
  const { data: bot, isLoading } = useBot()
  const [localConfig, setLocalConfig] = useState<Partial<BotConfig>>({})

  const currentTab = searchParams.get('s') || 'general'
  const setCurrentTab = (tab: string) => {
    setSearchParams({ s: tab })
  }

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

      <Tabs value={currentTab} onValueChange={setCurrentTab}>
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
