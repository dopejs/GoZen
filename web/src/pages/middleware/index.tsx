import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { RefreshCw, Power, PowerOff, Settings2, Puzzle } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import {
  useMiddleware,
  useUpdateMiddleware,
  useEnableMiddleware,
  useDisableMiddleware,
  useReloadMiddleware,
} from '@/hooks/use-middleware'
import type { MiddlewareEntry } from '@/types/api'

export function MiddlewarePage() {
  const { t } = useTranslation()
  const { data: config, isLoading } = useMiddleware()
  const updateMiddleware = useUpdateMiddleware()
  const enableMiddleware = useEnableMiddleware()
  const disableMiddleware = useDisableMiddleware()
  const reloadMiddleware = useReloadMiddleware()

  const handleGlobalToggle = async (enabled: boolean) => {
    if (!config) return
    try {
      await updateMiddleware.mutateAsync({
        ...config,
        enabled,
      })
      toast.success(enabled ? t('middleware.enabled') : t('middleware.disabled'))
    } catch {
      toast.error(t('errors.unknown'))
    }
  }

  const handleMiddlewareToggle = async (middleware: MiddlewareEntry) => {
    try {
      if (middleware.enabled) {
        await disableMiddleware.mutateAsync(middleware.name)
        toast.success(t('middleware.middlewareDisabled', { name: middleware.name }))
      } else {
        await enableMiddleware.mutateAsync(middleware.name)
        toast.success(t('middleware.middlewareEnabled', { name: middleware.name }))
      }
    } catch {
      toast.error(t('errors.unknown'))
    }
  }

  const handleReload = async () => {
    try {
      await reloadMiddleware.mutateAsync()
      toast.success(t('middleware.reloaded'))
    } catch {
      toast.error(t('errors.unknown'))
    }
  }

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  const middlewares = config?.middlewares || []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{t('middleware.title')}</h1>
          <p className="text-muted-foreground">{t('middleware.description')}</p>
        </div>
        <div className="flex items-center gap-4">
          <Button
            variant="outline"
            size="sm"
            onClick={handleReload}
            disabled={reloadMiddleware.isPending}
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${reloadMiddleware.isPending ? 'animate-spin' : ''}`} />
            {t('middleware.reload')}
          </Button>
        </div>
      </div>

      {/* Global Enable/Disable */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings2 className="h-5 w-5" />
            {t('middleware.globalSettings')}
          </CardTitle>
          <CardDescription>{t('middleware.globalSettingsDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>{t('middleware.enablePipeline')}</Label>
              <p className="text-sm text-muted-foreground">
                {t('middleware.enablePipelineDesc')}
              </p>
            </div>
            <Switch
              checked={config?.enabled || false}
              onCheckedChange={handleGlobalToggle}
              disabled={updateMiddleware.isPending}
            />
          </div>
        </CardContent>
      </Card>

      {/* Middleware List */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Puzzle className="h-5 w-5" />
            {t('middleware.middlewares')}
          </CardTitle>
          <CardDescription>{t('middleware.middlewaresDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          {middlewares.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">
              {t('middleware.noMiddlewares')}
            </p>
          ) : (
            <div className="space-y-4">
              {middlewares.map((middleware) => (
                <div
                  key={middleware.name}
                  className="flex items-center justify-between rounded-lg border p-4"
                >
                  <div className="flex items-center gap-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                      {middleware.enabled ? (
                        <Power className="h-5 w-5 text-primary" />
                      ) : (
                        <PowerOff className="h-5 w-5 text-muted-foreground" />
                      )}
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{middleware.name}</span>
                        <Badge variant={middleware.source === 'builtin' ? 'secondary' : 'outline'}>
                          {middleware.source}
                        </Badge>
                        {middleware.version && (
                          <Badge variant="outline">v{middleware.version}</Badge>
                        )}
                        {middleware.priority !== undefined && (
                          <Badge variant="outline">
                            {t('middleware.priority')}: {middleware.priority}
                          </Badge>
                        )}
                      </div>
                      {middleware.description && (
                        <p className="text-sm text-muted-foreground mt-1">
                          {middleware.description}
                        </p>
                      )}
                    </div>
                  </div>
                  <Switch
                    checked={middleware.enabled}
                    onCheckedChange={() => handleMiddlewareToggle(middleware)}
                    disabled={!config?.enabled || enableMiddleware.isPending || disableMiddleware.isPending}
                  />
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Beta Notice */}
      <Card className="border-amber-500/50 bg-amber-500/5">
        <CardContent className="pt-6">
          <p className="text-sm text-amber-600 dark:text-amber-400">
            <strong>{t('middleware.betaNotice')}</strong> {t('middleware.betaNoticeDesc')}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
