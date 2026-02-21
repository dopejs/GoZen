import { ReactNode } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { CHART_COLORS } from './constants'

interface LineChartCardProps {
  title: string
  data: Record<string, unknown>[]
  dataKeys: string[]
  xAxisKey?: string
  height?: number
  formatValue?: (value: number) => string
  formatYAxis?: (value: number) => string
  headerExtra?: ReactNode
  noDataText?: string
}

export function LineChartCard({
  title,
  data,
  dataKeys,
  xAxisKey = 'hour',
  height = 300,
  formatValue = (v) => String(v),
  formatYAxis,
  headerExtra,
  noDataText = 'No data',
}: LineChartCardProps) {
  return (
    <Card>
      <CardHeader>
        {headerExtra ? (
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm">{title}</CardTitle>
            {headerExtra}
          </div>
        ) : (
          <CardTitle className="text-sm">{title}</CardTitle>
        )}
      </CardHeader>
      <CardContent>
        {data.length > 0 ? (
          <ResponsiveContainer width="100%" height={height}>
            <LineChart data={data}>
              <CartesianGrid strokeDasharray="3 3" stroke="#333" />
              <XAxis dataKey={xAxisKey} stroke="#888" fontSize={12} />
              <YAxis stroke="#888" fontSize={12} tickFormatter={formatYAxis} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1e2030', border: '1px solid #333' }}
                formatter={(value: number | undefined) => formatValue(value ?? 0)}
              />
              <Legend />
              {dataKeys.map((key, index) => (
                <Line
                  key={key}
                  type="monotone"
                  dataKey={key}
                  stroke={CHART_COLORS[index % CHART_COLORS.length]}
                  strokeWidth={2}
                  dot={false}
                />
              ))}
            </LineChart>
          </ResponsiveContainer>
        ) : (
          <p className="py-8 text-center text-muted-foreground">{noDataText}</p>
        )}
      </CardContent>
    </Card>
  )
}
