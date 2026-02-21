import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Settings as SettingsIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { settingsApi } from '@/lib/api'

export function GeneralSettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const { data: settings, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: settingsApi.get,
  })

  const updateSettings = useMutation({
    mutationFn: settingsApi.update,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
      toast.success(t('common.success'))
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    },
  })

  const [defaultProfile, setDefaultProfile] = useState('')
  const [defaultClient, setDefaultClient] = useState('')

  useState(() => {
    if (settings) {
      setDefaultProfile(settings.default_profile || '')
      setDefaultClient(settings.default_client || '')
    }
  })

  if (isLoading) {
    return <div className="flex justify-center py-4">{t('common.loading')}</div>
  }

  const handleSave = () => {
    updateSettings.mutate({
      default_profile: defaultProfile || settings?.default_profile,
      default_client: defaultClient || settings?.default_client,
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <SettingsIcon className="h-5 w-5" />
          {t('settings.general')}
        </CardTitle>
        <CardDescription>{t('settings.generalDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-2">
          <Label>{t('settings.defaultProfile')}</Label>
          <Select value={defaultProfile || settings?.default_profile || ''} onValueChange={setDefaultProfile}>
            <SelectTrigger>
              <SelectValue placeholder={t('settings.selectProfile')} />
            </SelectTrigger>
            <SelectContent>
              {settings?.profiles?.map((p) => (
                <SelectItem key={p} value={p}>{p}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="grid gap-2">
          <Label>{t('settings.defaultClient')}</Label>
          <Select value={defaultClient || settings?.default_client || ''} onValueChange={setDefaultClient}>
            <SelectTrigger>
              <SelectValue placeholder={t('settings.selectClient')} />
            </SelectTrigger>
            <SelectContent>
              {settings?.clients?.map((c) => (
                <SelectItem key={c} value={c}>{c}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="grid gap-2">
          <Label>{t('settings.webPort')}</Label>
          <Input value={settings?.web_port || ''} disabled />
          <p className="text-xs text-muted-foreground">{t('settings.webPortHint')}</p>
        </div>

        <Button onClick={handleSave} disabled={updateSettings.isPending}>
          {t('common.save')}
        </Button>
      </CardContent>
    </Card>
  )
}
