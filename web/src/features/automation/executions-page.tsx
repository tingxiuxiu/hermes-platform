import { useQuery } from '@tanstack/react-query'
import type { NavigateFn } from '@/hooks/use-table-url-state'
import { getExecutions } from './api/automation-api'
import { ExecutionsTable } from './components/executions-table'
import type { ExecutionStatus } from './data/schema'
import { AutomationPageShell } from './index'

type ExecutionsSearch = {
  page?: number
  pageSize?: number
  jobId?: number
  jobName?: string
  status?: ExecutionStatus[]
}

export function AutomationExecutionsPage({
  search,
  navigate,
}: {
  search: ExecutionsSearch
  navigate: NavigateFn
}) {
  const { data, isLoading } = useQuery({
    queryKey: ['automation', 'executions', search],
    queryFn: () =>
      getExecutions({
        page: search.page ?? 1,
        pageSize: search.pageSize ?? 10,
        jobName: search.jobName || undefined,
        status: search.status?.length === 1 ? search.status[0] : undefined,
      }),
  })

  return (
    <AutomationPageShell
      title='Job Executions'
      description='Track every automation run and open a live session by build_uid.'
    >
      <ExecutionsTable
        data={data?.items ?? []}
        total={data?.total ?? 0}
        loading={isLoading}
        search={search}
        navigate={navigate}
      />
    </AutomationPageShell>
  )
}
