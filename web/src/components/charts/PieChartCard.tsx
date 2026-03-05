import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { CHART_COLORS } from './constants'

interface PieChartData {
  name: string
  value: number
  fullName?: string
}

interface PieChartCardProps {
  title: string
  data: PieChartData[]
  height?: number
  formatValue?: (value: number) => string
  noDataText?: string
}

export function PieChartCard({
  title,
  data,
  height = 200,
  formatValue = (v) => String(v),
  noDataText = 'No data',
}: PieChartCardProps) {
  return (
    <div>
      <p className="mb-2 text-center text-sm text-muted-foreground">{title}</p>
      {data.length > 0 ? (
        <ResponsiveContainer width="100%" height={height}>
          <PieChart>
            <Pie
              data={data}
              dataKey="value"
              nameKey="name"
              cx="50%"
              cy="50%"
              outerRadius={70}
              label={({ percent }) => `${((percent ?? 0) * 100).toFixed(0)}%`}
              labelLine={false}
            >
              {data.map((_, index) => (
                <Cell key={`cell-${index}`} fill={CHART_COLORS[index % CHART_COLORS.length]} />
              ))}
            </Pie>
            <Tooltip
              formatter={(value: number | undefined, _name: string | undefined, props: any) => [
                formatValue(value ?? 0),
                props.payload.fullName || props.payload.name,
              ]}
            />
          </PieChart>
        </ResponsiveContainer>
      ) : (
        <p className="py-8 text-center text-muted-foreground">{noDataText}</p>
      )}
    </div>
  )
}

interface DualPieChartCardProps {
  title: string
  leftTitle: string
  rightTitle: string
  leftData: PieChartData[]
  rightData: PieChartData[]
  leftFormatValue?: (value: number) => string
  rightFormatValue?: (value: number) => string
  noDataText?: string
}

export function DualPieChartCard({
  title,
  leftTitle,
  rightTitle,
  leftData,
  rightData,
  leftFormatValue = (v) => String(v),
  rightFormatValue = (v) => String(v),
  noDataText = 'No data',
}: DualPieChartCardProps) {
  const hasData = leftData.length > 0 || rightData.length > 0

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {hasData ? (
          <div className="grid grid-cols-2 gap-4">
            <PieChartCard
              title={leftTitle}
              data={leftData}
              formatValue={leftFormatValue}
              noDataText={noDataText}
            />
            <PieChartCard
              title={rightTitle}
              data={rightData}
              formatValue={rightFormatValue}
              noDataText={noDataText}
            />
          </div>
        ) : (
          <p className="py-8 text-center text-muted-foreground">{noDataText}</p>
        )}
      </CardContent>
    </Card>
  )
}
