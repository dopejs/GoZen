import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { Server, Activity, DollarSign, Zap } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { healthApi, usageApi, providerHealthApi } from '@/lib/api'

export function DashboardPage() {
  const { t } = useTranslation()

  const { data: health } = useQuery({
    queryKey: ['health'],
    queryFn: healthApi.check,
  })

  const { data: usage } = useQuery({
    queryKey: ['usage', 'summary', 'today'],
    queryFn: () => usageApi.summary('today'),
  })

  const { data: providerHealth } = useQuery({
    queryKey: ['health', 'providers'],
    queryFn: providerHealthApi.list,
  })

  const healthyProviders = providerHealth?.filter((p) => p.status === 'healthy').length ?? 0
  const totalProviders = providerHealth?.length ?? 0

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">{t('nav.dashboard')}</h1>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('common.status')}</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {health?.status === 'ok' ? (
                <Badge variant="success">{t('usage.healthy')}</Badge>
              ) : (
                <Badge variant="destructive">{t('common.error')}</Badge>
              )}
            </div>
            <p className="text-xs text-muted-foreground">v{health?.version}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('nav.providers')}</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {healthyProviders}/{totalProviders}
            </div>
            <p className="text-xs text-muted-foreground">{t('usage.healthy')}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('usage.totalRequests')}</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{usage?.total_requests ?? 0}</div>
            <p className="text-xs text-muted-foreground">{t('usage.today')}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{t('usage.totalCost')}</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${(usage?.total_cost ?? 0).toFixed(4)}</div>
            <p className="text-xs text-muted-foreground">{t('usage.today')}</p>
          </CardContent>
        </Card>
      </div>

      {/* Provider Health */}
      <Card>
        <CardHeader>
          <CardTitle>{t('usage.providerHealth')}</CardTitle>
        </CardHeader>
        <CardContent>
          {providerHealth && providerHealth.length > 0 ? (
            <div className="space-y-3">
              {providerHealth.map((provider) => (
                <div key={provider.name} className="flex items-center justify-between rounded-lg border p-3">
                  <div className="flex items-center gap-3">
                    <Server className="h-5 w-5 text-muted-foreground" />
                    <span className="font-medium">{provider.name}</span>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-sm text-muted-foreground">{provider.latency_ms}ms</span>
                    <Badge
                      variant={
                        provider.status === 'healthy'
                          ? 'success'
                          : provider.status === 'degraded'
                            ? 'warning'
                            : 'destructive'
                      }
                    >
                      {t(`usage.${provider.status}`)}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-muted-foreground">{t('providers.noProviders')}</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
