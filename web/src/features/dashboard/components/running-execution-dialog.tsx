import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { LoaderCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  formatDateTime,
  formatDuration,
  statusColor,
} from '@/features/automation/data/utils'
import { getRunningExecutionCases } from '../api/dashboard-api'
import type { DashboardRunningExecution } from '../api/dashboard-api'

export function RunningExecutionDialog({
  execution,
  open,
  onOpenChange,
}: {
  execution: DashboardRunningExecution | null
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const casesQuery = useQuery({
    queryKey: ['dashboard', 'running-execution-cases', execution?.execution_id],
    queryFn: () => getRunningExecutionCases(execution!.execution_id),
    enabled: open && !!execution,
  })

  const sortedCases = useMemo(() => {
    return [...(casesQuery.data ?? [])].sort((left, right) => {
      const leftTime = left.start_at ? new Date(left.start_at).getTime() : 0
      const rightTime = right.start_at ? new Date(right.start_at).getTime() : 0
      return rightTime - leftTime
    })
  }, [casesQuery.data])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[85vh] max-w-4xl overflow-hidden p-0'>
        <DialogHeader className='border-b px-6 py-4'>
          <DialogTitle>
            {execution
              ? `Execution #${execution.execution_id}`
              : 'Execution Cases'}
          </DialogTitle>
          <DialogDescription>
            查看当前 execution 已运行 case
            的实时快照，包括状态、attempt、耗时和失败概要。
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className='h-[70vh]'>
          <div className='space-y-3 p-6'>
            {casesQuery.isLoading ? (
              <div className='flex items-center justify-center py-16 text-sm text-muted-foreground'>
                <LoaderCircle className='mr-2 size-4 animate-spin' />
                Loading case snapshots...
              </div>
            ) : sortedCases.length ? (
              sortedCases.map((item) => (
                <div key={item.item_id} className='rounded-lg border p-4'>
                  <div className='flex flex-wrap items-start justify-between gap-3'>
                    <div className='space-y-1'>
                      <div className='flex items-center gap-2'>
                        <Badge
                          variant='outline'
                          className={cn('capitalize', statusColor(item.status))}
                        >
                          {item.status}
                        </Badge>
                        <span className='text-xs text-muted-foreground'>
                          Attempt #{item.attempt_number}
                        </span>
                      </div>
                      <div className='font-medium'>{item.case_name}</div>
                      <div className='text-xs text-muted-foreground'>
                        {item.case_key}
                      </div>
                    </div>
                    <div className='space-y-1 text-right text-xs text-muted-foreground'>
                      <div>{formatDateTime(item.start_at)}</div>
                      <div>{formatDuration(item.duration)}</div>
                    </div>
                  </div>
                  {item.error_message ? (
                    <div className='mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-900 dark:bg-red-950/30 dark:text-red-300'>
                      {item.error_message}
                    </div>
                  ) : null}
                </div>
              ))
            ) : (
              <div className='py-16 text-center text-sm text-muted-foreground'>
                No case snapshots available yet.
              </div>
            )}
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}
