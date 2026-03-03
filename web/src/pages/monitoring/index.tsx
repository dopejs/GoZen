import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { RefreshCw, Activity, AlertCircle, Clock, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useRequests } from '@/hooks/use-requests'
import type { RequestRecord } from '@/types/api'

export function MonitoringPage() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()

  const autoRefresh = searchParams.get('autoRefresh') === 'true'
  const selectedProvider = searchParams.get('provider') || 'all'
  const statusFilter = searchParams.get('status') || 'all'

  const updateParams = (updates: Record<string, string | null>) => {
    const newParams = new URLSearchParams(searchParams)
    for (const [key, value] of Object.entries(updates)) {
      if (value === null || value === 'false' || value === 'all') {
        newParams.delete(key)
      } else {
        newParams.set(key, value)
      }
    }
    setSearchParams(newParams)
  }

  // Build filter params
  const filterParams: any = {
    limit: 100,
  }
  if (selectedProvider !== 'all') {
    filterParams.provider = selectedProvider
  }
  if (statusFilter === 'errors') {
    filterParams.status_min = 400
  } else if (statusFilter === 'success') {
    filterParams.status_min = 200
    filterParams.status_max = 299
  }

  const { data, isLoading, refetch } = useRequests(
    filterParams,
    autoRefresh ? 5000 : undefined
  )

  const formatTimestamp = (ts: string) => {
    return new Date(ts).toLocaleString()
  }

  const formatDuration = (ms: number) => {
    if (ms < 1000) {
      return `${ms}ms`
    }
    return `${(ms / 1000).toFixed(2)}s`
  }

  const formatCost = (cost: number) => {
    return `$${cost.toFixed(6)}`
  }

  const getStatusBadge = (status: number) => {
    if (status >= 200 && status < 300) {
      return <Badge variant="success">{status}</Badge>
    } else if (status >= 400 && status < 500) {
      return <Badge variant="warning">{status}</Badge>
    } else if (status >= 500) {
      return <Badge variant="destructive">{status}</Badge>
    }
    return <Badge variant="secondary">{status}</Badge>
  }

  // Extract unique providers from data
  const providers = data?.requests
    ? Array.from(new Set(data.requests.map((r) => r.provider)))
    : []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{t('monitoring.title')}</h1>
          <p className="text-muted-foreground">{t('monitoring.description')}</p>
        </div>
        <Button variant="outline" onClick={() => refetch()}>
          <RefreshCw className="mr-2 h-4 w-4" />
          {t('common.refresh')}
        </Button>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="flex flex-wrap items-center gap-6 pt-6">
          <div className="flex items-center gap-2">
            <Label htmlFor="provider-filter">{t('monitoring.provider')}</Label>
            <Select value={selectedProvider} onValueChange={(v) => updateParams({ provider: v })}>
              <SelectTrigger id="provider-filter" className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('monitoring.allProviders')}</SelectItem>
                {providers.map((provider) => (
                  <SelectItem key={provider} value={provider}>
                    {provider}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            <Label htmlFor="status-filter">{t('monitoring.status')}</Label>
            <Select value={statusFilter} onValueChange={(v) => updateParams({ status: v })}>
              <SelectTrigger id="status-filter" className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('monitoring.allStatus')}</SelectItem>
                <SelectItem value="success">{t('monitoring.successOnly')}</SelectItem>
                <SelectItem value="errors">{t('monitoring.errorsOnly')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            <Switch id="auto-refresh" checked={autoRefresh} onCheckedChange={(v) => updateParams({ autoRefresh: v.toString() })} />
            <Label htmlFor="auto-refresh">{t('monitoring.autoRefresh')}</Label>
          </div>
        </CardContent>
      </Card>

      {/* Requests Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            {t('monitoring.requests')} ({data?.total ?? 0})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center py-8">{t('common.loading')}</div>
          ) : data?.requests && data.requests.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="px-4 py-3 text-left font-medium">{t('monitoring.timestamp')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('monitoring.provider')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('monitoring.model')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('monitoring.status')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('monitoring.duration')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('monitoring.tokens')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('monitoring.cost')}</th>
                  </tr>
                </thead>
                <tbody>
                  {data.requests.map((request: RequestRecord) => (
                    <tr key={request.id} className="border-b hover:bg-muted/50">
                      <td className="px-4 py-3 text-muted-foreground">{formatTimestamp(request.timestamp)}</td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          {request.provider}
                          {request.failover_chain && request.failover_chain.length > 0 && (
                            <Badge variant="outline" className="text-xs">
                              <Zap className="mr-1 h-3 w-3" />
                              {request.failover_chain.length} failover
                            </Badge>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">{request.model || '-'}</td>
                      <td className="px-4 py-3">{getStatusBadge(request.status_code)}</td>
                      <td className="px-4 py-3 text-muted-foreground">
                        <div className="flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          {formatDuration(request.duration_ms)}
                        </div>
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {request.input_tokens}/{request.output_tokens}
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">{formatCost(request.cost_usd)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12">
              <AlertCircle className="mb-4 h-12 w-12 text-muted-foreground" />
              <p className="text-muted-foreground">{t('monitoring.noRequests')}</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
