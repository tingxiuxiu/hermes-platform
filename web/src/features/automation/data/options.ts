import type { CaseStatus, ExecutionStatus, JobStatus, StepStatus } from './schema'

export const jobStatusOptions: { label: string; value: JobStatus }[] = [
  { label: 'Active', value: 'active' },
  { label: 'Inactive', value: 'inactive' },
]

export const buildStatusOptions: { label: string; value: ExecutionStatus }[] = [
  { label: 'Running', value: 'running' },
  { label: 'Completed', value: 'completed' },
  { label: 'Failed', value: 'failed' },
  { label: 'Aborted', value: 'aborted' },
]

export const executionStatusOptions: {
  label: string
  value: ExecutionStatus
}[] = [
  { label: 'Running', value: 'running' },
  { label: 'Completed', value: 'completed' },
  { label: 'Failed', value: 'failed' },
  { label: 'Aborted', value: 'aborted' },
  { label: 'Calculated', value: 'calculated' },
]

export const caseStatusOptions: { label: string; value: CaseStatus }[] = [
  { label: 'Running', value: 'running' },
  { label: 'Passed', value: 'passed' },
  { label: 'Failed', value: 'failed' },
  { label: 'Skipped', value: 'skipped' },
  { label: 'Broken', value: 'broken' },
]

export const stepStatusOptions: { label: string; value: StepStatus }[] = [
  { label: 'Running', value: 'running' },
  { label: 'Passed', value: 'passed' },
  { label: 'Failed', value: 'failed' },
  { label: 'Skipped', value: 'skipped' },
  { label: 'Broken', value: 'broken' },
]
