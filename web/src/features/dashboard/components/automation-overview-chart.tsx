import {
  Bar,
  CartesianGrid,
  ComposedChart,
  Legend,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { formatTrendDay } from '@/features/automation/data/utils'
import type { DashboardTrend } from '../api/dashboard-api'

export function AutomationOverviewChart({
  trends,
}: {
  trends: DashboardTrend[]
}) {
  const data = trends.map((trend) => ({
    ...trend,
    day: formatTrendDay(trend.stat_date),
  }))

  return (
    <ResponsiveContainer width='100%' height={340}>
      <ComposedChart data={data}>
        <CartesianGrid vertical={false} strokeDasharray='3 3' />
        <XAxis
          dataKey='day'
          stroke='currentColor'
          fontSize={12}
          tickLine={false}
          axisLine={false}
        />
        <YAxis
          stroke='currentColor'
          fontSize={12}
          tickLine={false}
          axisLine={false}
          allowDecimals={false}
        />
        <Tooltip />
        <Legend />
        <Bar
          dataKey='execution_total'
          name='Executions'
          fill='hsl(var(--primary))'
          radius={[6, 6, 0, 0]}
        />
        <Line
          type='monotone'
          dataKey='success_cases'
          name='Passed Cases'
          stroke='#16a34a'
          strokeWidth={2}
          dot={{ r: 3 }}
        />
        <Line
          type='monotone'
          dataKey='failure_cases'
          name='Failed Cases'
          stroke='#dc2626'
          strokeWidth={2}
          dot={{ r: 3 }}
        />
      </ComposedChart>
    </ResponsiveContainer>
  )
}
