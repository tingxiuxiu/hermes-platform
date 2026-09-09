export type JobStatus = 'active' | 'inactive'
export type ExecutionStatus =
  | 'running'
  | 'completed'
  | 'failed'
  | 'aborted'
  | 'calculated'
export type CaseStatus = 'running' | 'passed' | 'failed' | 'skipped' | 'broken'
export type StepStatus = CaseStatus
export type PipelineLastBuildStatus = ExecutionStatus

export type JobItem = {
  id: number
  job_name: string
  job_url: string
  status: JobStatus
  last_build_number: number | null
  last_build_uid: string | null
  last_build_status: PipelineLastBuildStatus | null
  last_build_timestamp: string | null
  last_build_duration: number | null
  pipeline_params: Record<string, unknown> | null
  sync_at: string | null
}

export type ExecutionItem = {
  id: number
  build_uid: string
  job_name: string
  job_url: string | null
  project_name: string
  software_name: string
  software_version: string
  labels: string[] | null
  status: ExecutionStatus
  start_time: string
  end_time: string | null
  duration: number | null
  planned_cases_count: number | null
  pass_count: number | null
  failure_count: number | null
  skipped_count: number | null
  pass_rate: number | null
  last_heartbeat_at: string | null
}

export type CaseItem = {
  id: number
  case_uid: string
  case_key: string
  case_name: string
  attempt_number: number
  status: CaseStatus
  start_time: string | null
  end_time: string | null
  duration: number | null
}

export type GoStepRow = {
  id?: number
  step_path: string
  step_name: string
  status: StepStatus
  start_time?: string | null
  end_time?: string | null
  duration?: number | null
  sub_steps?: GoStepRow[] | null
}

export type LiveStepNode = {
  step_path: string
  step_name: string
  status: StepStatus
  start_time: string | null
  end_time: string | null
  duration: number | null
  children: LiveStepNode[]
}

export type LiveCurrentItem = {
  case_uid: string
  case_key: string
  case_name: string
  status: CaseStatus
  steps: GoStepRow[] | null
}

export type LiveSnapshot = {
  execution: ExecutionItem
  items: CaseItem[]
  current_item: LiveCurrentItem | null
}

export type ItemDetail = CaseItem & {
  error_message?: string | null
  error_traceback?: string | null
  steps: GoStepRow[] | null
}

export type StepDelta = {
  type: 'step.upserted'
  build_uid: string
  case_uid: string
  case_key?: string
  case_name?: string
  step_path: string
  step_name: string
  status: StepStatus
}

export type ItemUpdatedEvent = {
  type: 'item.updated'
  build_uid: string
  case_uid: string
  case_key?: string
  case_name?: string
  attempt_number?: number
  status: CaseStatus
  start_time?: string | null
  end_time?: string | null
  duration?: number | null
  error_message?: string
}

export type ExecutionUpdatedEvent = {
  type: 'execution.updated'
  build_uid: string
  status: ExecutionStatus
}

export type LiveEvent = StepDelta | ItemUpdatedEvent | ExecutionUpdatedEvent
