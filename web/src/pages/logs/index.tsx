import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { RefreshCw, ScrollText, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useLogs } from '@/hooks/use-logs'

export function LogsPage() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()

  const autoRefresh = searchParams.get('autoRefresh') === 'true'
  const errorsOnly = searchParams.get('errorsOnly') === 'true'
  const selectedProvider = searchParams.get('provider') || 'all'

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

  const { data, isLoading, refetch } = useLogs(
    {
      provider: selectedProvider === 'all' ? undefined : selectedProvider,
      errors_only: errorsOnly,
      limit: 100,
    },
    autoRefresh ? 5000 : undefined
  )

  const formatTimestamp = (ts: string) => {
    return new Date(ts).toLocaleString()
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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{t('logs.title')}</h1>
          <p className="text-muted-foreground">{t('logs.description')}</p>
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
            <Label htmlFor="provider-filter">{t('logs.provider')}</Label>
            <Select value={selectedProvider} onValueChange={(v) => updateParams({ provider: v })}>
              <SelectTrigger id="provider-filter" className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('logs.allProviders')}</SelectItem>
                {data?.providers?.map((provider) => (
                  <SelectItem key={provider} value={provider}>
                    {provider}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            <Switch id="errors-only" checked={errorsOnly} onCheckedChange={(v) => updateParams({ errorsOnly: v.toString() })} />
            <Label htmlFor="errors-only">{t('logs.errorsOnly')}</Label>
          </div>

          <div className="flex items-center gap-2">
            <Switch id="auto-refresh" checked={autoRefresh} onCheckedChange={(v) => updateParams({ autoRefresh: v.toString() })} />
            <Label htmlFor="auto-refresh">{t('logs.autoRefresh')}</Label>
          </div>
        </CardContent>
      </Card>

      {/* Logs Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ScrollText className="h-5 w-5" />
            {t('logs.title')} ({data?.total ?? 0})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center py-8">{t('common.loading')}</div>
          ) : data?.entries && data.entries.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="px-4 py-3 text-left font-medium">{t('logs.timestamp')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('logs.provider')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('logs.model')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('logs.status')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('logs.latency')}</th>
                    <th className="px-4 py-3 text-left font-medium">{t('logs.tokens')}</th>
                  </tr>
                </thead>
                <tbody>
                  {data.entries.map((entry) => (
                    <tr key={entry.id} className="border-b hover:bg-muted/50">
                      <td className="px-4 py-3 text-muted-foreground">{formatTimestamp(entry.timestamp)}</td>
                      <td className="px-4 py-3">{entry.provider}</td>
                      <td className="px-4 py-3 text-muted-foreground">{entry.model || '-'}</td>
                      <td className="px-4 py-3">{getStatusBadge(entry.status)}</td>
                      <td className="px-4 py-3 text-muted-foreground">{entry.latency_ms}ms</td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {entry.input_tokens}/{entry.output_tokens}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12">
              <AlertCircle className="mb-4 h-12 w-12 text-muted-foreground" />
              <p className="text-muted-foreground">{t('logs.noLogs')}</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
