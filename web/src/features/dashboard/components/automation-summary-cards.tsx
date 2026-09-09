import { Activity, BadgeCheck, BugPlay, GitBranch } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  formatDateTime,
  formatDuration,
  formatPassRate,
  statusColor,
} from '@/features/automation/data/utils'
import type { DashboardSummary } from '../api/dashboard-api'

export function AutomationSummaryCards({
  summary,
}: {
  summary: DashboardSummary
}) {
  const cards = [
    {
      title: 'Active Jobs',
      value: summary.active_jobs_count,
      hint: `Snapshot ${formatDateTime(summary.updated_at)}`,
      icon: GitBranch,
      tone: 'text-sky-600 dark:text-sky-400',
    },
    {
      title: 'Running Executions',
      value: summary.running_executions_count,
      hint: `Total executions ${summary.total_executions_count}`,
      icon: Activity,
      tone: 'text-amber-600 dark:text-amber-400',
    },
    {
      title: 'Pass Rate · 7d',
      value: formatPassRate(summary.pass_rate_7d),
      hint: `${summary.success_cases_7d} passed / ${summary.failure_cases_7d} failed`,
      icon: BadgeCheck,
      tone: cn(
        'rounded-full border px-2 py-0.5 text-xs',
        statusColor('success')
      ),
      badge: true,
    },
    {
      title: 'Avg Duration · 7d',
      value: formatDuration(summary.avg_execution_duration_7d),
      hint: `${summary.skipped_cases_7d} skipped cases in last 7 days`,
      icon: BugPlay,
      tone: 'text-rose-600 dark:text-rose-400',
    },
  ]

  return (
    <div className='grid gap-4 sm:grid-cols-2 xl:grid-cols-4'>
      {cards.map((card) => (
        <Card key={card.title}>
          <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
            <CardTitle className='text-sm font-medium'>{card.title}</CardTitle>
            {card.badge ? (
              <span className={card.tone}>
                {formatPassRate(summary.pass_rate_7d)}
              </span>
            ) : (
              <card.icon className={cn('size-4', card.tone)} />
            )}
          </CardHeader>
          <CardContent>
            <div className='text-2xl font-bold'>{card.value}</div>
            <p className='text-xs text-muted-foreground'>{card.hint}</p>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
