import { api } from '@/lib/api'
import { API_V1 } from '@/lib/api-prefix'
import type { CaseStatus, ExecutionStatus } from '@/features/automation/data/schema'

export type DashboardSummary = {
  active_jobs_count: number
  total_executions_count: number
  running_executions_count: number
  success_cases_7d: number
  failure_cases_7d: number
  skipped_cases_7d: number
  pass_rate_7d: number
  avg_execution_duration_7d: number | null
  updated_at: string
}

export type DashboardTrend = {
  stat_date: string
  execution_total: number
  success_cases: number
  failure_cases: number
  skipped_cases: number
  running_cases: number
}

export type DashboardRunningExecution = {
  execution_id: number
  build_uid: string
  job_id: number
  job_name: string
  job_url: string
  status: ExecutionStatus
  started_at: string
  duration: number | null
  pre_cases_count: number
  completed_cases_count: number
  success_cases_count: number
  failure_cases_count: number
  skipped_cases_count: number
  running_cases_count: number
  progress_percent: number
  running_case_names: string[]
  updated_at: string
}

export type DashboardRunningCase = {
  execution_id: number
  item_id: number
  case_key: string
  case_name: string
  attempt_number: number
  status: CaseStatus
  start_at: string | null
  end_at: string | null
  duration: number | null
  error_message: string | null
  updated_at: string
}

export type DashboardOverviewFilters = {
  days?: number
  jobId?: number
  jobName?: string
}

type OverviewResponse = {
  data: {
    summary: DashboardSummary
    trends: DashboardTrend[]
    running_executions: DashboardRunningExecution[]
  }
}

type RunningCasesResponse = {
  data: {
    execution_id: number
    items: DashboardRunningCase[]
  }
}

export async function getDashboardOverview(filters: DashboardOverviewFilters = {}) {
  const params = new URLSearchParams()
  if (filters.days != null) {
    params.set('days', String(filters.days))
  }
  if (filters.jobId != null) {
    params.set('job_id', String(filters.jobId))
  }
  const jobName = filters.jobName?.trim()
  if (jobName) {
    params.set('job_name', jobName)
  }

  const { data } = await api.get<OverviewResponse>(
    `${API_V1}/automation/dashboard/overview`,
    { params }
  )
  return data.data
}

export async function getRunningExecutionCases(executionId: number) {
  const { data } = await api.get<RunningCasesResponse>(
    `${API_V1}/automation/dashboard/executions/${executionId}/cases`
  )
  return data.data.items
}
