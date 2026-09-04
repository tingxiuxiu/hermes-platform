import { Link } from '@tanstack/react-router'
import { LoaderCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { useExecutionSession } from './hooks/use-execution-session'
import { StepTree } from './components/step-tree'
import { AutomationPageShell } from './index'
import { formatDateTime, formatDuration, shortUid, statusColor } from './data/utils'

export function ExecutionSessionPage({
  buildUid,
  caseUid,
}: {
  buildUid: string
  caseUid?: string
}) {
  const session = useExecutionSession(buildUid, caseUid)
  const execution = session.snapshot?.execution
  const tree = session.selectedCaseUid
    ? (session.trees[session.selectedCaseUid] ?? [])
    : []

  return (
    <AutomationPageShell
      title={execution?.job_name ?? 'Execution session'}
      description={
        execution
          ? `${shortUid(execution.build_uid)} · ${execution.status}`
          : 'Loading live session…'
      }
    >
      {execution?.status === 'running' && (
        <p className='-mt-2 text-sm text-muted-foreground'>
          以插件心跳为准；进程被杀约 60 秒后标记中止。
        </p>
      )}

      {session.isLoading ? (
        <div className='flex items-center justify-center py-16 text-sm text-muted-foreground'>
          <LoaderCircle className='mr-2 size-4 animate-spin' />
          Loading session…
        </div>
      ) : execution ? (
        <div className='grid min-h-[32rem] gap-4 lg:grid-cols-[minmax(18rem,24rem)_1fr]'>
          <section className='rounded-[18px] border border-[#e0e0e0] bg-white dark:bg-background'>
            <header className='border-b border-[#e0e0e0] px-4 py-3'>
              <div className='text-[17px] font-medium tracking-tight'>Cases</div>
              <div className='text-sm text-muted-foreground'>
                {session.items.length} in this run
              </div>
            </header>
            <div className='space-y-2 p-3'>
              {session.items.length ? (
                session.items.map((item) => {
                  const selected = item.case_uid === session.selectedCaseUid
                  return (
                    <button
                      key={item.case_uid}
                      type='button'
                      onClick={() => session.selectCase(item.case_uid)}
                      className={cn(
                        'flex w-full items-start justify-between gap-3 rounded-[18px] border px-3 py-3 text-left',
                        selected
                          ? 'border-[#0066cc] bg-[#0066cc]/5'
                          : 'border-[#e0e0e0] hover:bg-muted/40'
                      )}
                    >
                      <div className='min-w-0 space-y-1'>
                        <Badge
                          variant='outline'
                          className={cn('capitalize', statusColor(item.status))}
                        >
                          {item.status}
                        </Badge>
                        <div className='truncate text-[17px] font-medium tracking-tight'>
                          {item.case_name}
                        </div>
                        <div className='truncate text-xs text-muted-foreground'>
                          {item.case_key}
                        </div>
                      </div>
                      <div className='shrink-0 text-right text-xs text-muted-foreground'>
                        <div>{formatDuration(item.duration)}</div>
                      </div>
                    </button>
                  )
                })
              ) : (
                <div className='py-10 text-center text-sm text-muted-foreground'>
                  No cases reported yet.
                </div>
              )}
            </div>
          </section>

          <section className='rounded-[18px] border border-[#e0e0e0] bg-white dark:bg-background'>
            <header className='border-b border-[#e0e0e0] px-4 py-3'>
              <div className='text-[17px] font-medium tracking-tight'>
                {session.selected?.case_name ?? 'Step tree'}
              </div>
              <div className='text-sm text-muted-foreground'>
                {session.selected
                  ? `${session.selected.status} · attempt #${session.selected.attempt_number}`
                  : 'Select a case to inspect steps.'}
              </div>
            </header>
            <div className='space-y-4 p-4'>
              {session.selected ? (
                <>
                  <div className='grid gap-3 sm:grid-cols-3'>
                    <Info label='Started' value={formatDateTime(session.selected.start_time)} />
                    <Info label='Duration' value={formatDuration(session.selected.duration)} />
                    <Info label='Case UID' value={shortUid(session.selected.case_uid)} />
                  </div>
                  <StepTree nodes={tree} />
                </>
              ) : (
                <div className='py-16 text-center text-sm text-muted-foreground'>
                  Waiting for the first case.
                </div>
              )}
            </div>
          </section>
        </div>
      ) : (
        <div className='rounded-[18px] border border-[#e0e0e0] py-16 text-center text-sm text-muted-foreground'>
          Execution not found.{' '}
          <Link to='/automation/executions' className='text-[#0066cc]'>
            Back to list
          </Link>
        </div>
      )}
    </AutomationPageShell>
  )
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-[18px] border border-[#e0e0e0] px-3 py-2'>
      <div className='text-xs text-muted-foreground'>{label}</div>
      <div className='mt-1 text-sm font-medium break-all'>{value}</div>
    </div>
  )
}
