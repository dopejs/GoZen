import { useTranslation } from 'react-i18next'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Shield } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { autoPermissionApi } from '@/lib/api'
import { CLAUDE_PERMISSION_MODES, CODEX_PERMISSION_MODES } from '@/types/api'
import type { AutoPermissionConfig } from '@/types/api'

const CLIENT_CONFIGS = [
  {
    key: 'claude' as const,
    label: 'Claude Code',
    modes: CLAUDE_PERMISSION_MODES,
    defaultMode: 'bypassPermissions',
  },
  {
    key: 'codex' as const,
    label: 'Codex CLI',
    modes: CODEX_PERMISSION_MODES,
    defaultMode: 'full-auto',
  },
  {
    key: 'opencode' as const,
    label: 'OpenCode',
    modes: ['auto'] as const,
    defaultMode: 'auto',
  },
] as const

export function PermissionSettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const { data: permissions, isLoading } = useQuery({
    queryKey: ['auto-permission'],
    queryFn: autoPermissionApi.getAll,
  })

  const updatePermission = useMutation({
    mutationFn: ({ client, config }: { client: string; config: { enabled: boolean; mode: string } }) =>
      autoPermissionApi.update(client, config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auto-permission'] })
      queryClient.invalidateQueries({ queryKey: ['settings'] })
      toast.success(t('common.success'))
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : t('common.error'))
    },
  })

  if (isLoading) {
    return <div className="flex justify-center py-4">{t('common.loading')}</div>
  }

  const getConfig = (client: 'claude' | 'codex' | 'opencode'): AutoPermissionConfig => {
    return permissions?.[client] ?? { enabled: false, mode: '' }
  }

  const handleToggle = (client: typeof CLIENT_CONFIGS[number]) => {
    const current = getConfig(client.key)
    updatePermission.mutate({
      client: client.key,
      config: {
        enabled: !current.enabled,
        mode: current.mode || client.defaultMode,
      },
    })
  }

  const handleModeChange = (client: typeof CLIENT_CONFIGS[number], mode: string) => {
    const current = getConfig(client.key)
    updatePermission.mutate({
      client: client.key,
      config: {
        enabled: current.enabled,
        mode,
      },
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5" />
          {t('settings.permissions')}
        </CardTitle>
        <CardDescription>{t('settings.permissionsDesc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {CLIENT_CONFIGS.map((client) => {
          const config = getConfig(client.key)
          return (
            <div key={client.key} className="space-y-3 rounded-lg border p-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label className="text-base font-medium">{client.label}</Label>
                  <p className="text-sm text-muted-foreground">
                    {t('settings.autoPermissionHint', { client: client.label })}
                  </p>
                </div>
                <Switch
                  checked={config.enabled}
                  onCheckedChange={() => handleToggle(client)}
                  disabled={updatePermission.isPending}
                />
              </div>
              {config.enabled && (
                <div className="grid gap-2 pl-1">
                  <Label>{t('settings.permissionMode')}</Label>
                  <Select
                    value={config.mode || client.defaultMode}
                    onValueChange={(mode) => handleModeChange(client, mode)}
                    disabled={updatePermission.isPending}
                  >
                    <SelectTrigger className="w-[240px]">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {client.modes.map((mode) => (
                        <SelectItem key={mode} value={mode}>
                          {mode}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )}
            </div>
          )
        })}
        <p className="text-xs text-muted-foreground">
          {t('settings.permissionPriority')}
        </p>
      </CardContent>
    </Card>
  )
}
