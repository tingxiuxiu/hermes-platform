import { Link } from '@tanstack/react-router'
import { ExternalLink, PlayCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  formatDateTime,
  formatDuration,
  statusColor,
} from '@/features/automation/data/utils'
import type { DashboardRunningExecution } from '../api/dashboard-api'

export function RunningExecutions({
  items,
  onOpen,
}: {
  items: DashboardRunningExecution[]
  onOpen: (execution: DashboardRunningExecution) => void
}) {
  if (!items.length) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Running Executions</CardTitle>
          <CardDescription>No execution is currently running.</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  return (
    <div className='grid gap-4 xl:grid-cols-2'>
      {items.map((execution) => {
        const total = execution.pre_cases_count || 1
        const progressWidth = `${Math.max(0, Math.min(100, execution.progress_percent))}%`
        return (
          <Card key={execution.execution_id} className='overflow-hidden'>
            <CardHeader className='space-y-3 border-b'>
              <div className='flex flex-wrap items-start justify-between gap-3'>
                <div className='space-y-1'>
                  <CardTitle className='text-lg'>
                    {execution.job_name}
                  </CardTitle>
                  <CardDescription>
                    Execution #{execution.execution_id} · Started{' '}
                    {formatDateTime(execution.started_at)}
                  </CardDescription>
                </div>
                <Badge
                  variant='outline'
                  className={cn('capitalize', statusColor(execution.status))}
                >
                  {execution.status}
                </Badge>
              </div>

              <div className='space-y-2'>
                <div className='flex items-center justify-between text-sm'>
                  <span className='text-muted-foreground'>Progress</span>
                  <span className='font-medium'>
                    {execution.completed_cases_count}/{total} ·{' '}
                    {execution.progress_percent.toFixed(0)}%
                  </span>
                </div>
                <div className='h-2.5 rounded-full bg-muted'>
                  <div
                    className='h-2.5 rounded-full bg-primary transition-all'
                    style={{ width: progressWidth }}
                  />
                </div>
              </div>
            </CardHeader>
            <CardContent className='space-y-4 pt-6'>
              <div className='grid gap-3 sm:grid-cols-4'>
                <StatChip
                  label='Pass'
                  value={execution.success_cases_count}
                  tone='success'
                />
                <StatChip
                  label='Fail'
                  value={execution.failure_cases_count}
                  tone='failed'
                />
                <StatChip
                  label='Skipped'
                  value={execution.skipped_cases_count}
                  tone='skipped'
                />
                <StatChip
                  label='Running'
                  value={execution.running_cases_count}
                  tone='running'
                />
              </div>

              <div className='grid gap-3 sm:grid-cols-2'>
                <InfoLine
                  label='Elapsed'
                  value={formatDuration(execution.duration)}
                />
                <InfoLine
                  label='Last Snapshot'
                  value={formatDateTime(execution.updated_at)}
                />
              </div>

              <div className='space-y-2'>
                <div className='text-sm font-medium'>
                  Currently Running Cases
                </div>
                {execution.running_case_names.length ? (
                  <div className='flex flex-wrap gap-2'>
                    {execution.running_case_names.map((name) => (
                      <Badge key={name} variant='secondary'>
                        <PlayCircle className='mr-1 size-3.5' />
                        {name}
                      </Badge>
                    ))}
                  </div>
                ) : (
                  <div className='text-sm text-muted-foreground'>
                    No running case names available in the latest snapshot.
                  </div>
                )}
              </div>

              <div className='flex flex-wrap justify-end gap-2'>
                {execution.build_uid ? (
                  <Button size='sm' asChild>
                    <Link
                      to='/automation/executions/$buildUid'
                      params={{ buildUid: execution.build_uid }}
                    >
                      Open session
                    </Link>
                  </Button>
                ) : (
                  <Button
                    variant='outline'
                    size='sm'
                    onClick={() => onOpen(execution)}
                  >
                    View Cases
                  </Button>
                )}
                <Button variant='ghost' size='sm' asChild>
                  <a href={execution.job_url} target='_blank' rel='noreferrer'>
                    <ExternalLink className='size-4' />
                    Jenkins
                  </a>
                </Button>
              </div>
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}

function StatChip({
  label,
  value,
  tone,
}: {
  label: string
  value: number
  tone: string
}) {
  return (
    <div className='rounded-lg border bg-muted/20 p-3'>
      <div className='text-xs text-muted-foreground'>{label}</div>
      <div className={cn('mt-1 text-xl font-semibold', statusColor(tone))}>
        {value}
      </div>
    </div>
  )
}

function InfoLine({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border bg-background p-3'>
      <div className='text-xs text-muted-foreground'>{label}</div>
      <div className='mt-1 text-sm font-medium'>{value}</div>
    </div>
  )
}
