import { useQuery } from '@tanstack/react-query'
import type { NavigateFn } from '@/hooks/use-table-url-state'
import { getAutomationJobs } from './api/automation-api'
import { JobsTable } from './components/jobs-table'
import type { ExecutionStatus, JobStatus } from './data/schema'
import { AutomationPageShell } from './index'

type JobsSearch = {
  page?: number
  pageSize?: number
  jobName?: string
  status?: JobStatus[]
  lastBuildStatus?: ExecutionStatus[]
}

export function AutomationJobsPage({
  search,
  navigate,
}: {
  search: JobsSearch
  navigate: NavigateFn
}) {
  const { data, isLoading } = useQuery({
    queryKey: ['automation', 'jobs', search],
    queryFn: () =>
      getAutomationJobs({
        page: search.page ?? 1,
        pageSize: search.pageSize ?? 10,
        jobName: search.jobName || undefined,
        status: search.status?.length === 1 ? search.status[0] : undefined,
        lastBuildStatus:
          search.lastBuildStatus?.length === 1
            ? search.lastBuildStatus[0]
            : undefined,
      }),
  })

  return (
    <AutomationPageShell
      title='Automation Jobs'
      description='Browse Jenkins jobs. Running status comes from plugin reports, not Jenkins trigger.'
    >
      <JobsTable
        data={data?.items ?? []}
        total={data?.total ?? 0}
        loading={isLoading}
        search={search}
        navigate={navigate}
      />
    </AutomationPageShell>
  )
}
