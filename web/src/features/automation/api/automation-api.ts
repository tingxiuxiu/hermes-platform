import { api } from '@/lib/api'
import { API_V1 } from '@/lib/api-prefix'
import { useAuthStore } from '@/stores/auth-store'
import type {
  CaseItem,
  ExecutionItem,
  ExecutionStatus,
  ItemDetail,
  JobItem,
  JobStatus,
  LiveEvent,
  LiveSnapshot,
  PipelineLastBuildStatus,
} from '../data/schema'

type RowsData<T> = {
  total: number
  page: number
  page_size: number
  rows: T[]
}

type ResponseWithData<T> = {
  data: T
}

export type JobsQuery = {
  page: number
  pageSize: number
  jobName?: string
  status?: JobStatus
  lastBuildStatus?: PipelineLastBuildStatus
}

export type ExecutionsQuery = {
  page: number
  pageSize: number
  jobName?: string
  status?: ExecutionStatus
}

export type ItemsQuery = {
  page: number
  pageSize: number
}

export type JobsPage = {
  items: JobItem[]
  total: number
  page: number
  pageSize: number
}

export type ExecutionsPage = {
  items: ExecutionItem[]
  total: number
  page: number
  pageSize: number
}

export type CasesPage = {
  items: CaseItem[]
  total: number
  page: number
  pageSize: number
}

function mapRows<T>(data: RowsData<T>) {
  return {
    items: data.rows ?? [],
    total: data.total,
    page: data.page,
    pageSize: data.page_size,
  }
}

export async function getAutomationJobs(query: JobsQuery): Promise<JobsPage> {
  const { data } = await api.get<ResponseWithData<RowsData<JobItem>>>(
    `${API_V1}/automation/pipelines`,
    {
      params: {
        page: query.page,
        page_size: query.pageSize,
        job_name: query.jobName || undefined,
        status: query.status,
      },
    }
  )

  return mapRows(data.data)
}

export async function getExecutions(
  query: ExecutionsQuery
): Promise<ExecutionsPage> {
  const { data } = await api.get<ResponseWithData<RowsData<ExecutionItem>>>(
    `${API_V1}/automation/executions`,
    {
      params: {
        page: query.page,
        page_size: query.pageSize,
        job_name: query.jobName || undefined,
        status: query.status,
      },
    }
  )

  return mapRows(data.data)
}

export async function getExecutionDetail(buildUid: string) {
  const { data } = await api.get<ResponseWithData<ExecutionItem>>(
    `${API_V1}/automation/executions/${buildUid}`
  )
  return data.data
}

export async function getLiveSnapshot(buildUid: string) {
  const { data } = await api.get<ResponseWithData<LiveSnapshot>>(
    `${API_V1}/automation/executions/${buildUid}/live`
  )
  return data.data
}

export async function getExecutionItems(
  buildUid: string,
  query: ItemsQuery
): Promise<CasesPage> {
  const { data } = await api.get<ResponseWithData<RowsData<CaseItem>>>(
    `${API_V1}/automation/executions/${buildUid}/items`,
    {
      params: {
        page: query.page,
        page_size: query.pageSize,
      },
    }
  )

  return mapRows(data.data)
}

export async function getCaseAttempts(buildUid: string, caseKey: string) {
  const { data } = await api.get<ResponseWithData<RowsData<CaseItem>>>(
    `${API_V1}/automation/executions/${buildUid}/items/attempts`,
    {
      params: { case_key: caseKey },
    }
  )

  return data.data.rows
}

export async function getItemDetail(caseUid: string) {
  const { data } = await api.get<ResponseWithData<ItemDetail>>(
    `${API_V1}/automation/items/${caseUid}`
  )
  return data.data
}

export async function subscribeLiveEvents(
  buildUid: string,
  onEvent: (event: LiveEvent) => void,
  signal: AbortSignal
) {
  const base = String(api.defaults.baseURL ?? '').replace(/\/$/, '')
  const token = useAuthStore.getState().auth.accessToken
  const response = await fetch(
    `${base}${API_V1}/automation/executions/${buildUid}/events`,
    {
      headers: token ? { Authorization: `Bearer ${token}` } : undefined,
      signal,
    }
  )
  if (!response.ok || !response.body) {
    throw new Error(`live events failed: ${response.status}`)
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    buffer = consumeSse(buffer, onEvent)
  }
}

function consumeSse(buffer: string, onEvent: (event: LiveEvent) => void) {
  let rest = buffer
  let idx = rest.indexOf('\n\n')
  while (idx >= 0) {
    const chunk = rest.slice(0, idx)
    rest = rest.slice(idx + 2)
    let data = ''
    for (const line of chunk.split('\n')) {
      if (line.startsWith('data:')) data += line.slice(5).trim()
    }
    if (data) {
      try {
        const parsed = JSON.parse(data) as LiveEvent
        if (parsed?.type) onEvent(parsed)
      } catch {
        // ignore malformed frames
      }
    }
    idx = rest.indexOf('\n\n')
  }
  return rest
}
