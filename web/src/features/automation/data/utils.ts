import { format } from 'date-fns'

export function formatDateTime(value: string | null | undefined) {
  if (!value) return '--'

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--'

  return format(date, 'yyyy-MM-dd HH:mm:ss')
}

export function formatDuration(value: number | null | undefined) {
  if (value == null) return '--'
  if (value < 1) return `${Math.round(value * 1000)} ms`
  if (value < 10) return `${value.toFixed(2)} s`
  return `${value.toFixed(1)} s`
}

/** API pass_rate is a 0–1 ratio; display as a percentage. */
export function formatPassRate(rate: number | null | undefined) {
  if (rate == null || Number.isNaN(rate)) return '--'
  return `${(rate * 100).toFixed(2)}%`
}

export function formatTrendDay(statDate: string | null | undefined) {
  if (!statDate) return '--'
  const date = new Date(statDate)
  if (Number.isNaN(date.getTime())) return statDate
  const year = date.getUTCFullYear()
  const month = String(date.getUTCMonth() + 1).padStart(2, '0')
  const day = String(date.getUTCDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function stringifyParams(value: Record<string, unknown> | null | undefined) {
  if (!value) return '--'
  const text = JSON.stringify(value)
  return text.length > 80 ? `${text.slice(0, 77)}...` : text
}

export function shortUid(value: string | null | undefined) {
  if (!value) return '--'
  return value.length <= 8 ? value : value.slice(0, 8)
}

export function statusColor(status: string | null | undefined) {
  switch (status) {
    case 'success':
    case 'completed':
    case 'active':
    case 'passed':
    case 'calculated':
      return 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-300'
    case 'running':
    case 'pending':
    case 'unstable':
      return 'border-[#0066cc]/30 bg-[#0066cc]/5 text-[#0066cc] dark:border-[#2997ff]/40 dark:bg-[#2997ff]/10 dark:text-[#2997ff]'
    case 'failure':
    case 'failed':
    case 'error':
    case 'broken':
    case 'aborted':
      return 'border-red-200 bg-red-50 text-red-700 dark:border-red-900 dark:bg-red-950/40 dark:text-red-300'
    case 'inactive':
    case 'skipped':
      return 'border-slate-200 bg-slate-50 text-slate-700 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-300'
    default:
      return 'border-border bg-muted text-muted-foreground'
  }
}
