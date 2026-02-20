import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { format } from 'date-fns'
import {
  BarChart3,
  DollarSign,
  Zap,
  TrendingUp,
  Calendar,
} from 'lucide-react'
import {
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { useUsageSummary, useHourlyUsage, useBudgetStatus } from '@/hooks/use-usage'
import type { HourlyUsageByDimension } from '@/types/api'

// Chart colors
const COLORS = [
  '#5eead4', // teal
  '#c4b5fd', // lavender
  '#86efac', // sage
  '#fbbf24', // amber
  '#93c5fd', // blue
  '#fb7185', // red
  '#a78bfa', // purple
  '#34d399', // emerald
]

type TimeRange = 'today' | 'week' | 'month' | 'custom'

export function UsagePage() {
  const { t } = useTranslation()
  const [timeRange, setTimeRange] = useState<TimeRange>('today')
  const [customStart, setCustomStart] = useState('')
  const [customEnd, setCustomEnd] = useState('')
  const [selectedProviders, setSelectedProviders] = useState<string[]>([])
  const [selectedModels, setSelectedModels] = useState<string[]>([])

  // Calculate time range params
  const timeParams = useMemo(() => {
    if (timeRange === 'custom' && customStart && customEnd) {
      return {
        since: new Date(customStart).toISOString(),
        until: new Date(customEnd + 'T23:59:59').toISOString(),
      }
    }
    // For preset ranges, use period param
    const periodMap: Record<string, 'today' | 'week' | 'month'> = {
      today: 'today',
      week: 'week',
      month: 'month',
    }
    return { period: periodMap[timeRange] || 'today' }
  }, [timeRange, customStart, customEnd])

  // Fetch data
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

  // Get unique providers and models
  const providers = useMemo(() => {
    if (!usage?.by_provider) return []
    return Object.keys(usage.by_provider)
  }, [usage])

  const models = useMemo(() => {
    if (!usage?.by_model) return []
    return Object.keys(usage.by_model)
  }, [usage])

  // Initialize selected filters
  useMemo(() => {
    if (providers.length > 0 && selectedProviders.length === 0) {
      setSelectedProviders(providers)
    }
  }, [providers, selectedProviders.length])

  useMemo(() => {
    if (models.length > 0 && selectedModels.length === 0) {
      setSelectedModels(models)
    }
  }, [models, selectedModels.length])

  // Prepare pie chart data
  const providerPieData = useMemo(() => {
    if (!usage?.by_provider) return []
    return Object.entries(usage.by_provider).map(([name, stats]) => ({
      name,
      tokens: stats.input_tokens + stats.output_tokens,
      cost: stats.cost,
    }))
  }, [usage])

  const modelPieData = useMemo(() => {
    if (!usage?.by_model) return []
    return Object.entries(usage.by_model).map(([name, stats]) => ({
      name: name.length > 20 ? name.slice(0, 20) + '...' : name,
      fullName: name,
      tokens: stats.input_tokens + stats.output_tokens,
      cost: stats.cost,
    }))
  }, [usage])

  // Prepare line chart data - transform hourly data by dimension into chart format
  const prepareLineChartData = (
    data: HourlyUsageByDimension[] | undefined,
    selectedDimensions: string[],
    valueKey: 'tokens' | 'cost'
  ) => {
    if (!data || data.length === 0) return []

    // Group by hour
    const byHour: Record<string, Record<string, number>> = {}
    data.forEach((item) => {
      if (!selectedDimensions.includes(item.dimension)) return
      const hourKey = item.hour
      if (!byHour[hourKey]) byHour[hourKey] = {}
      byHour[hourKey][item.dimension] =
        valueKey === 'tokens'
          ? item.input_tokens + item.output_tokens
          : item.cost
    })

    // Convert to array format for recharts
    return Object.entries(byHour)
      .map(([hour, values]) => ({
        hour: format(new Date(hour), 'MM/dd HH:mm'),
        ...values,
      }))
      .sort((a, b) => a.hour.localeCompare(b.hour))
  }

  const providerTokensChartData = useMemo(
    () => prepareLineChartData(hourlyByProvider, selectedProviders, 'tokens'),
    [hourlyByProvider, selectedProviders]
  )

  const providerCostChartData = useMemo(
    () => prepareLineChartData(hourlyByProvider, selectedProviders, 'cost'),
    [hourlyByProvider, selectedProviders]
  )

  const modelTokensChartData = useMemo(
    () => prepareLineChartData(hourlyByModel, selectedModels, 'tokens'),
    [hourlyByModel, selectedModels]
  )

  const modelCostChartData = useMemo(
    () => prepareLineChartData(hourlyByModel, selectedModels, 'cost'),
    [hourlyByModel, selectedModels]
  )

  const formatCost = (cost: number) => `$${cost.toFixed(4)}`
  const formatTokens = (tokens: number) => {
    if (tokens >= 1000000) return `${(tokens / 1000000).toFixed(2)}M`
    if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}K`
    return tokens.toString()
  }

  const toggleProvider = (provider: string) => {
    setSelectedProviders((prev) =>
      prev.includes(provider)
        ? prev.filter((p) => p !== provider)
        : [...prev, provider]
    )
  }

  const toggleModel = (model: string) => {
    setSelectedModels((prev) =>
      prev.includes(model) ? prev.filter((m) => m !== model) : [...prev, model]
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">{t('usage.title')}</h1>
          <p className="text-muted-foreground">{t('usage.description')}</p>
        </div>

        {/* Time Range Selector */}
        <div className="flex items-center gap-2">
          <Button
            variant={timeRange === 'today' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setTimeRange('today')}
          >
            {t('usage.today')}
          </Button>
          <Button
            variant={timeRange === 'week' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setTimeRange('week')}
          >
            {t('usage.thisWeek')}
          </Button>
          <Button
            variant={timeRange === 'month' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setTimeRange('month')}
          >
            {t('usage.thisMonth')}
          </Button>
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant={timeRange === 'custom' ? 'default' : 'outline'}
                size="sm"
              >
                <Calendar className="mr-2 h-4 w-4" />
                {t('usage.custom')}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80">
              <div className="space-y-4">
                <div className="grid gap-2">
                  <Label>{t('usage.startDate')}</Label>
                  <Input
                    type="date"
                    value={customStart}
                    onChange={(e) => setCustomStart(e.target.value)}
                    max={customEnd || undefined}
                  />
                </div>
                <div className="grid gap-2">
                  <Label>{t('usage.endDate')}</Label>
                  <Input
                    type="date"
                    value={customEnd}
                    onChange={(e) => setCustomEnd(e.target.value)}
                    min={customStart || undefined}
                    max={format(new Date(), 'yyyy-MM-dd')}
                  />
                </div>
                <Button
                  className="w-full"
                  onClick={() => setTimeRange('custom')}
                  disabled={!customStart || !customEnd}
                >
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
                <span>
                  {formatCost(budgetStatus.monthly_used)} /{' '}
                  {formatCost(budgetStatus.monthly_limit)}
                </span>
              </div>
              <Progress
                value={(budgetStatus.monthly_used / budgetStatus.monthly_limit) * 100}
                className="h-2"
              />
            </div>
            {budgetStatus.daily_limit > 0 && (
              <div>
                <div className="mb-2 flex justify-between text-sm">
                  <span>{t('usage.today')}</span>
                  <span>
                    {formatCost(budgetStatus.daily_used)} /{' '}
                    {formatCost(budgetStatus.daily_limit)}
                  </span>
                </div>
                <Progress
                  value={(budgetStatus.daily_used / budgetStatus.daily_limit) * 100}
                  className="h-2"
                />
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {t('usage.totalRequests')}
            </CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{usage?.request_count ?? 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {t('usage.inputTokens')}
            </CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatTokens(usage?.total_input_tokens ?? 0)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {t('usage.outputTokens')}
            </CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatTokens(usage?.total_output_tokens ?? 0)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {t('usage.totalCost')}
            </CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCost(usage?.total_cost ?? 0)}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Pie Charts Row */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Provider Distribution */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{t('usage.byProvider')}</CardTitle>
          </CardHeader>
          <CardContent>
            {providerPieData.length > 0 ? (
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="mb-2 text-center text-sm text-muted-foreground">
                    {t('usage.tokens')}
                  </p>
                  <ResponsiveContainer width="100%" height={200}>
                    <PieChart>
                      <Pie
                        data={providerPieData}
                        dataKey="tokens"
                        nameKey="name"
                        cx="50%"
                        cy="50%"
                        outerRadius={70}
                        label={({ percent }) =>
                          `${name} ${((percent ?? 0) * 100).toFixed(0)}%`
                        }
                        labelLine={false}
                      >
                        {providerPieData.map((_, index) => (
                          <Cell
                            key={`cell-${index}`}
                            fill={COLORS[index % COLORS.length]}
                          />
                        ))}
                      </Pie>
                      <Tooltip
                        formatter={(value: number | undefined) => formatTokens(value ?? 0)}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <div>
                  <p className="mb-2 text-center text-sm text-muted-foreground">
                    {t('usage.cost')}
                  </p>
                  <ResponsiveContainer width="100%" height={200}>
                    <PieChart>
                      <Pie
                        data={providerPieData}
                        dataKey="cost"
                        nameKey="name"
                        cx="50%"
                        cy="50%"
                        outerRadius={70}
                        label={({ percent }) =>
                          `${name} ${((percent ?? 0) * 100).toFixed(0)}%`
                        }
                        labelLine={false}
                      >
                        {providerPieData.map((_, index) => (
                          <Cell
                            key={`cell-${index}`}
                            fill={COLORS[index % COLORS.length]}
                          />
                        ))}
                      </Pie>
                      <Tooltip formatter={(value: number | undefined) => formatCost(value ?? 0)} />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>
            ) : (
              <p className="py-8 text-center text-muted-foreground">
                {t('usage.noData')}
              </p>
            )}
          </CardContent>
        </Card>

        {/* Model Distribution */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{t('usage.byModel')}</CardTitle>
          </CardHeader>
          <CardContent>
            {modelPieData.length > 0 ? (
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="mb-2 text-center text-sm text-muted-foreground">
                    {t('usage.tokens')}
                  </p>
                  <ResponsiveContainer width="100%" height={200}>
                    <PieChart>
                      <Pie
                        data={modelPieData}
                        dataKey="tokens"
                        nameKey="name"
                        cx="50%"
                        cy="50%"
                        outerRadius={70}
                        label={({ percent }) => `${((percent ?? 0) * 100).toFixed(0)}%`}
                        labelLine={false}
                      >
                        {modelPieData.map((_, index) => (
                          <Cell
                            key={`cell-${index}`}
                            fill={COLORS[index % COLORS.length]}
                          />
                        ))}
                      </Pie>
                      <Tooltip
                        formatter={(value: number | undefined, _name: string | undefined, props: any) => [
                          formatTokens(value ?? 0),
                          props.payload.fullName,
                        ]}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <div>
                  <p className="mb-2 text-center text-sm text-muted-foreground">
                    {t('usage.cost')}
                  </p>
                  <ResponsiveContainer width="100%" height={200}>
                    <PieChart>
                      <Pie
                        data={modelPieData}
                        dataKey="cost"
                        nameKey="name"
                        cx="50%"
                        cy="50%"
                        outerRadius={70}
                        label={({ percent }) => `${((percent ?? 0) * 100).toFixed(0)}%`}
                        labelLine={false}
                      >
                        {modelPieData.map((_, index) => (
                          <Cell
                            key={`cell-${index}`}
                            fill={COLORS[index % COLORS.length]}
                          />
                        ))}
                      </Pie>
                      <Tooltip
                        formatter={(value: number | undefined, _name: string | undefined, props: any) => [
                          formatCost(value ?? 0),
                          props.payload.fullName,
                        ]}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>
            ) : (
              <p className="py-8 text-center text-muted-foreground">
                {t('usage.noData')}
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Line Charts - Tokens by Provider */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm">{t('usage.tokensByProvider')}</CardTitle>
            <div className="flex flex-wrap gap-2">
              {providers.map((provider, index) => (
                <Button
                  key={provider}
                  variant={selectedProviders.includes(provider) ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => toggleProvider(provider)}
                  style={{
                    backgroundColor: selectedProviders.includes(provider)
                      ? COLORS[index % COLORS.length]
                      : undefined,
                    borderColor: COLORS[index % COLORS.length],
                  }}
                >
                  {provider}
                </Button>
              ))}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {providerTokensChartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={providerTokensChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                <XAxis dataKey="hour" stroke="#888" fontSize={12} />
                <YAxis stroke="#888" fontSize={12} tickFormatter={formatTokens} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1e2030', border: '1px solid #333' }}
                  formatter={(value: number | undefined) => formatTokens(value ?? 0)}
                />
                <Legend />
                {selectedProviders.map((provider) => (
                  <Line
                    key={provider}
                    type="monotone"
                    dataKey={provider}
                    stroke={COLORS[providers.indexOf(provider) % COLORS.length]}
                    strokeWidth={2}
                    dot={false}
                  />
                ))}
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="py-8 text-center text-muted-foreground">{t('usage.noData')}</p>
          )}
        </CardContent>
      </Card>

      {/* Line Charts - Cost by Provider */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t('usage.costByProvider')}</CardTitle>
        </CardHeader>
        <CardContent>
          {providerCostChartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={providerCostChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                <XAxis dataKey="hour" stroke="#888" fontSize={12} />
                <YAxis stroke="#888" fontSize={12} tickFormatter={(v) => `$${v.toFixed(2)}`} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1e2030', border: '1px solid #333' }}
                  formatter={(value: number | undefined) => formatCost(value ?? 0)}
                />
                <Legend />
                {selectedProviders.map((provider) => (
                  <Line
                    key={provider}
                    type="monotone"
                    dataKey={provider}
                    stroke={COLORS[providers.indexOf(provider) % COLORS.length]}
                    strokeWidth={2}
                    dot={false}
                  />
                ))}
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="py-8 text-center text-muted-foreground">{t('usage.noData')}</p>
          )}
        </CardContent>
      </Card>

      {/* Line Charts - Tokens by Model */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm">{t('usage.tokensByModel')}</CardTitle>
            <div className="flex flex-wrap gap-2">
              {models.map((model, index) => (
                <Button
                  key={model}
                  variant={selectedModels.includes(model) ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => toggleModel(model)}
                  style={{
                    backgroundColor: selectedModels.includes(model)
                      ? COLORS[index % COLORS.length]
                      : undefined,
                    borderColor: COLORS[index % COLORS.length],
                  }}
                >
                  {model.length > 25 ? model.slice(0, 25) + '...' : model}
                </Button>
              ))}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {modelTokensChartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={modelTokensChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                <XAxis dataKey="hour" stroke="#888" fontSize={12} />
                <YAxis stroke="#888" fontSize={12} tickFormatter={formatTokens} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1e2030', border: '1px solid #333' }}
                  formatter={(value: number | undefined) => formatTokens(value ?? 0)}
                />
                <Legend />
                {selectedModels.map((model) => (
                  <Line
                    key={model}
                    type="monotone"
                    dataKey={model}
                    stroke={COLORS[models.indexOf(model) % COLORS.length]}
                    strokeWidth={2}
                    dot={false}
                  />
                ))}
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="py-8 text-center text-muted-foreground">{t('usage.noData')}</p>
          )}
        </CardContent>
      </Card>

      {/* Line Charts - Cost by Model */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t('usage.costByModel')}</CardTitle>
        </CardHeader>
        <CardContent>
          {modelCostChartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={modelCostChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                <XAxis dataKey="hour" stroke="#888" fontSize={12} />
                <YAxis stroke="#888" fontSize={12} tickFormatter={(v) => `$${v.toFixed(2)}`} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1e2030', border: '1px solid #333' }}
                  formatter={(value: number | undefined) => formatCost(value ?? 0)}
                />
                <Legend />
                {selectedModels.map((model) => (
                  <Line
                    key={model}
                    type="monotone"
                    dataKey={model}
                    stroke={COLORS[models.indexOf(model) % COLORS.length]}
                    strokeWidth={2}
                    dot={false}
                  />
                ))}
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="py-8 text-center text-muted-foreground">{t('usage.noData')}</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
