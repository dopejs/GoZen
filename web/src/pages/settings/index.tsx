import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { GeneralSettings } from './tabs/GeneralSettings'
import { BindingsSettings } from './tabs/BindingsSettings'
import { SyncSettings } from './tabs/SyncSettings'
import { PasswordSettings } from './tabs/PasswordSettings'

export function SettingsPage() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()

  const currentTab = searchParams.get('s') || 'general'
  const setCurrentTab = (tab: string) => {
    setSearchParams({ s: tab })
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">{t('settings.title')}</h1>
        <p className="text-muted-foreground">{t('settings.description')}</p>
      </div>

      <Tabs value={currentTab} onValueChange={setCurrentTab}>
        <TabsList>
          <TabsTrigger value="general">{t('settings.general')}</TabsTrigger>
          <TabsTrigger value="bindings">{t('settings.bindings')}</TabsTrigger>
          <TabsTrigger value="sync">{t('settings.sync')}</TabsTrigger>
          <TabsTrigger value="password">{t('settings.webPassword')}</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-4">
          <GeneralSettings />
        </TabsContent>

        <TabsContent value="bindings" className="mt-4">
          <BindingsSettings />
        </TabsContent>

        <TabsContent value="sync" className="mt-4">
          <SyncSettings />
        </TabsContent>

        <TabsContent value="password" className="mt-4">
          <PasswordSettings />
        </TabsContent>
      </Tabs>
    </div>
  )
}
