import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { format } from 'date-fns'
import { BarChart3, DollarSign, Zap, TrendingUp, Calendar } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import {
  StatCard,
  LineChartCard,
  DualPieChartCard,
  CHART_COLORS,
  formatCost,
  formatTokens,
} from '@/components/charts'
import { useUsageSummary, useHourlyUsage, useBudgetStatus } from '@/hooks/use-usage'
import type { HourlyUsageByDimension } from '@/types/api'

type TimeRange = 'today' | 'week' | 'month' | 'custom'

export function UsagePage() {
  const { t } = useTranslation()
  const [timeRange, setTimeRange] = useState<TimeRange>('today')
  const [customStart, setCustomStart] = useState('')
  const [customEnd, setCustomEnd] = useState('')
  const [selectedProviders, setSelectedProviders] = useState<string[]>([])
  const [selectedModels, setSelectedModels] = useState<string[]>([])

  const timeParams = useMemo(() => {
    if (timeRange === 'custom' && customStart && customEnd) {
      return {
        since: new Date(customStart).toISOString(),
        until: new Date(customEnd + 'T23:59:59').toISOString(),
      }
    }
    const periodMap: Record<string, 'today' | 'week' | 'month'> = {
      today: 'today',
      week: 'week',
      month: 'month',
    }
    return { period: periodMap[timeRange] || 'today' }
  }, [timeRange, customStart, customEnd])

  const { data: usage } = useUsageSummary(timeParams)
  const { data: budgetStatus } = useBudgetStatus()
  const { data: hourlyByProvider } = useHourlyUsage({
    ...('since' in timeParams ? timeParams : {}),
    ...('period' in timeParams ? { hours: timeRange === 'today' ? 24 : timeRange === 'week' ? 168 : 720 } : {}),
    groupBy: 'provider',
  }) as { data: HourlyUsageByDimension[] | undefined }
  const { data: hourlyByModel } = useHourlyUsage({
    ...('since' in timeParams ? timeParams : {}),
    ...('period' in timeParams ? { hours: timeRange === 'today' ? 24 : timeRange === 'week' ? 168 : 720 } : {}),
    groupBy: 'model',
  }) as { data: HourlyUsageByDimension[] | undefined }

  const providers = useMemo(() => (usage?.by_provider ? Object.keys(usage.by_provider) : []), [usage])
  const models = useMemo(() => (usage?.by_model ? Object.keys(usage.by_model) : []), [usage])

  useMemo(() => {
    if (providers.length > 0 && selectedProviders.length === 0) setSelectedProviders(providers)
  }, [providers, selectedProviders.length])

  useMemo(() => {
    if (models.length > 0 && selectedModels.length === 0) setSelectedModels(models)
  }, [models, selectedModels.length])

  // Prepare pie chart data
  const providerPieData = useMemo(() => {
    if (!usage?.by_provider) return { tokens: [], cost: [] }
    const entries = Object.entries(usage.by_provider)
    return {
      tokens: entries.map(([name, stats]) => ({ name, value: stats.input_tokens + stats.output_tokens })),
      cost: entries.map(([name, stats]) => ({ name, value: stats.cost })),
    }
  }, [usage])

  const modelPieData = useMemo(() => {
    if (!usage?.by_model) return { tokens: [], cost: [] }
    const entries = Object.entries(usage.by_model)
    return {
      tokens: entries.map(([name, stats]) => ({
        name: name.length > 20 ? name.slice(0, 20) + '...' : name,
        fullName: name,
        value: stats.input_tokens + stats.output_tokens,
      })),
      cost: entries.map(([name, stats]) => ({
        name: name.length > 20 ? name.slice(0, 20) + '...' : name,
        fullName: name,
        value: stats.cost,
      })),
    }
  }, [usage])

  // Prepare line chart data
  const prepareLineChartData = (
    data: HourlyUsageByDimension[] | undefined,
    selectedDimensions: string[],
    valueKey: 'tokens' | 'cost'
  ) => {
    if (!data || data.length === 0) return []
    const byHour: Record<string, Record<string, number>> = {}
    data.forEach((item) => {
      if (!selectedDimensions.includes(item.dimension)) return
      if (!byHour[item.hour]) byHour[item.hour] = {}
      byHour[item.hour][item.dimension] = valueKey === 'tokens'
        ? item.input_tokens + item.output_tokens
        : item.cost
    })
    return Object.entries(byHour)
      .map(([hour, values]) => ({ hour: format(new Date(hour), 'MM/dd HH:mm'), ...values }))
      .sort((a, b) => a.hour.localeCompare(b.hour))
  }

  const providerTokensData = useMemo(() => prepareLineChartData(hourlyByProvider, selectedProviders, 'tokens'), [hourlyByProvider, selectedProviders])
  const providerCostData = useMemo(() => prepareLineChartData(hourlyByProvider, selectedProviders, 'cost'), [hourlyByProvider, selectedProviders])
  const modelTokensData = useMemo(() => prepareLineChartData(hourlyByModel, selectedModels, 'tokens'), [hourlyByModel, selectedModels])
  const modelCostData = useMemo(() => prepareLineChartData(hourlyByModel, selectedModels, 'cost'), [hourlyByModel, selectedModels])

  const toggleProvider = (p: string) => setSelectedProviders((prev) => prev.includes(p) ? prev.filter((x) => x !== p) : [...prev, p])
  const toggleModel = (m: string) => setSelectedModels((prev) => prev.includes(m) ? prev.filter((x) => x !== m) : [...prev, m])

  const renderFilterButtons = (items: string[], selected: string[], toggle: (item: string) => void) => (
    <div className="flex flex-wrap gap-2">
      {items.map((item, index) => (
        <Button
          key={item}
          variant={selected.includes(item) ? 'default' : 'outline'}
          size="sm"
          onClick={() => toggle(item)}
          style={{
            backgroundColor: selected.includes(item) ? CHART_COLORS[index % CHART_COLORS.length] : undefined,
            borderColor: CHART_COLORS[index % CHART_COLORS.length],
          }}
        >
          {item.length > 25 ? item.slice(0, 25) + '...' : item}
        </Button>
      ))}
    </div>
  )

  return (
    <div className="space-y-6">
      {/* Header with Time Range */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{t('usage.title')}</h1>
          <p className="text-muted-foreground">{t('usage.description')}</p>
        </div>
        <div className="flex items-center gap-2">
          {(['today', 'week', 'month'] as const).map((range) => (
            <Button
              key={range}
              variant={timeRange === range ? 'default' : 'outline'}
              size="sm"
              onClick={() => setTimeRange(range)}
            >
              {t(`usage.${range === 'week' ? 'thisWeek' : range === 'month' ? 'thisMonth' : 'today'}`)}
            </Button>
          ))}
          <Popover>
            <PopoverTrigger asChild>
              <Button variant={timeRange === 'custom' ? 'default' : 'outline'} size="sm">
                <Calendar className="mr-2 h-4 w-4" />
                {t('usage.custom')}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80">
              <div className="space-y-4">
                <div className="grid gap-2">
                  <Label>{t('usage.startDate')}</Label>
                  <Input type="date" value={customStart} onChange={(e) => setCustomStart(e.target.value)} max={customEnd || undefined} />
                </div>
                <div className="grid gap-2">
                  <Label>{t('usage.endDate')}</Label>
                  <Input type="date" value={customEnd} onChange={(e) => setCustomEnd(e.target.value)} min={customStart || undefined} max={format(new Date(), 'yyyy-MM-dd')} />
                </div>
                <Button className="w-full" onClick={() => setTimeRange('custom')} disabled={!customStart || !customEnd}>
                  {t('usage.apply')}
                </Button>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>

      {/* Budget Status */}
      {budgetStatus && budgetStatus.monthly_limit > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              {t('usage.budget')}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="mb-2 flex justify-between text-sm">
                <span>{t('usage.thisMonth')}</span>
                <span>{formatCost(budgetStatus.monthly_used)} / {formatCost(budgetStatus.monthly_limit)}</span>
              </div>
              <Progress value={(budgetStatus.monthly_used / budgetStatus.monthly_limit) * 100} className="h-2" />
            </div>
            {budgetStatus.daily_limit > 0 && (
              <div>
                <div className="mb-2 flex justify-between text-sm">
                  <span>{t('usage.today')}</span>
                  <span>{formatCost(budgetStatus.daily_used)} / {formatCost(budgetStatus.daily_limit)}</span>
                </div>
                <Progress value={(budgetStatus.daily_used / budgetStatus.daily_limit) * 100} className="h-2" />
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard title={t('usage.totalRequests')} value={usage?.request_count ?? 0} icon={<Zap className="h-4 w-4" />} />
        <StatCard title={t('usage.inputTokens')} value={formatTokens(usage?.total_input_tokens ?? 0)} icon={<TrendingUp className="h-4 w-4" />} />
        <StatCard title={t('usage.outputTokens')} value={formatTokens(usage?.total_output_tokens ?? 0)} icon={<TrendingUp className="h-4 w-4" />} />
        <StatCard title={t('usage.totalCost')} value={formatCost(usage?.total_cost ?? 0)} icon={<DollarSign className="h-4 w-4" />} />
      </div>

      {/* Pie Charts */}
      <div className="grid gap-4 md:grid-cols-2">
        <DualPieChartCard
          title={t('usage.byProvider')}
          leftTitle={t('usage.tokens')}
          rightTitle={t('usage.cost')}
          leftData={providerPieData.tokens}
          rightData={providerPieData.cost}
          leftFormatValue={formatTokens}
          rightFormatValue={formatCost}
          noDataText={t('usage.noData')}
        />
        <DualPieChartCard
          title={t('usage.byModel')}
          leftTitle={t('usage.tokens')}
          rightTitle={t('usage.cost')}
          leftData={modelPieData.tokens}
          rightData={modelPieData.cost}
          leftFormatValue={formatTokens}
          rightFormatValue={formatCost}
          noDataText={t('usage.noData')}
        />
      </div>

      {/* Line Charts */}
      <LineChartCard
        title={t('usage.tokensByProvider')}
        data={providerTokensData}
        dataKeys={selectedProviders}
        formatValue={formatTokens}
        formatYAxis={formatTokens}
        headerExtra={renderFilterButtons(providers, selectedProviders, toggleProvider)}
        noDataText={t('usage.noData')}
      />
      <LineChartCard
        title={t('usage.costByProvider')}
        data={providerCostData}
        dataKeys={selectedProviders}
        formatValue={formatCost}
        formatYAxis={(v) => `$${v.toFixed(2)}`}
        noDataText={t('usage.noData')}
      />
      <LineChartCard
        title={t('usage.tokensByModel')}
        data={modelTokensData}
        dataKeys={selectedModels}
        formatValue={formatTokens}
        formatYAxis={formatTokens}
        headerExtra={renderFilterButtons(models, selectedModels, toggleModel)}
        noDataText={t('usage.noData')}
      />
      <LineChartCard
        title={t('usage.costByModel')}
        data={modelCostData}
        dataKeys={selectedModels}
        formatValue={formatCost}
        formatYAxis={(v) => `$${v.toFixed(2)}`}
        noDataText={t('usage.noData')}
      />
    </div>
  )
}
