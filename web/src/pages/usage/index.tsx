import { useTranslation } from 'react-i18next'
import { BarChart3, DollarSign, Zap, TrendingUp } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useUsageSummary, useBudgetStatus } from '@/hooks/use-usage'

export function UsagePage() {
  const { t } = useTranslation()
  const { data: todayUsage } = useUsageSummary('today')
  const { data: weekUsage } = useUsageSummary('week')
  const { data: monthUsage } = useUsageSummary('month')
  const { data: budgetStatus } = useBudgetStatus()

  const formatCost = (cost: number) => `$${cost.toFixed(4)}`
  const formatTokens = (tokens: number) => {
    if (tokens >= 1000000) return `${(tokens / 1000000).toFixed(2)}M`
    if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}K`
    return tokens.toString()
  }

  const renderUsageCards = (usage: typeof todayUsage) => (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">{t('usage.totalRequests')}</CardTitle>
          <Zap className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{usage?.total_requests ?? 0}</div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">{t('usage.inputTokens')}</CardTitle>
          <TrendingUp className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{formatTokens(usage?.total_input_tokens ?? 0)}</div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">{t('usage.outputTokens')}</CardTitle>
          <TrendingUp className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{formatTokens(usage?.total_output_tokens ?? 0)}</div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">{t('usage.totalCost')}</CardTitle>
          <DollarSign className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{formatCost(usage?.total_cost ?? 0)}</div>
        </CardContent>
      </Card>
    </div>
  )

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">{t('usage.title')}</h1>
        <p className="text-muted-foreground">{t('usage.description')}</p>
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
                  {formatCost(budgetStatus.monthly_used)} / {formatCost(budgetStatus.monthly_limit)}
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
                    {formatCost(budgetStatus.daily_used)} / {formatCost(budgetStatus.daily_limit)}
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

      {/* Usage Tabs */}
      <Tabs defaultValue="today">
        <TabsList>
          <TabsTrigger value="today">{t('usage.today')}</TabsTrigger>
          <TabsTrigger value="week">{t('usage.thisWeek')}</TabsTrigger>
          <TabsTrigger value="month">{t('usage.thisMonth')}</TabsTrigger>
        </TabsList>
        <TabsContent value="today" className="mt-4">
          {renderUsageCards(todayUsage)}
        </TabsContent>
        <TabsContent value="week" className="mt-4">
          {renderUsageCards(weekUsage)}
        </TabsContent>
        <TabsContent value="month" className="mt-4">
          {renderUsageCards(monthUsage)}
        </TabsContent>
      </Tabs>

      {/* Usage by Provider */}
      {todayUsage?.by_provider && Object.keys(todayUsage.by_provider).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>{t('nav.providers')}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {Object.entries(todayUsage.by_provider).map(([name, usage]) => (
                <div key={name} className="flex items-center justify-between rounded-lg border p-3">
                  <span className="font-medium">{name}</span>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <span>{usage.requests} requests</span>
                    <span>{formatTokens(usage.input_tokens + usage.output_tokens)} tokens</span>
                    <span className="font-medium text-foreground">{formatCost(usage.cost)}</span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
