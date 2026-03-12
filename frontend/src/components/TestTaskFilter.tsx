import { useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { RotateCcw } from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'

export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed'

export interface FilterParams {
  status?: TaskStatus
  worker_name?: string
  plan_key?: string
}

interface TestTaskFilterProps {
  onFilterChange: (filters: FilterParams) => void
  workers?: string[]
  planKeys?: string[]
}

type SelectStatus = 'all' | TaskStatus

const STATUS_OPTIONS: SelectStatus[] = ['all', 'pending', 'running', 'completed', 'failed']

export function TestTaskFilter({
  onFilterChange,
  workers = [],
  planKeys = [],
}: TestTaskFilterProps) {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<FilterParams>({})

  const handleStatusChange = useCallback(
    (value: string) => {
      const newFilters: FilterParams = {
        ...filters,
        status: value === 'all' ? undefined : (value as TaskStatus),
      }
      setFilters(newFilters)
      onFilterChange(newFilters)
    },
    [filters, onFilterChange]
  )

  const handleWorkerChange = useCallback(
    (value: string) => {
      const newFilters: FilterParams = {
        ...filters,
        worker_name: value === 'all' ? undefined : value,
      }
      setFilters(newFilters)
      onFilterChange(newFilters)
    },
    [filters, onFilterChange]
  )

  const handlePlanKeyChange = useCallback(
    (value: string) => {
      const newFilters: FilterParams = {
        ...filters,
        plan_key: value === 'all' ? undefined : value,
      }
      setFilters(newFilters)
      onFilterChange(newFilters)
    },
    [filters, onFilterChange]
  )

  const handleReset = useCallback(() => {
    setFilters({})
    onFilterChange({})
  }, [onFilterChange])

  const hasActiveFilters = filters.status || filters.worker_name || filters.plan_key

  return (
    <div className="flex flex-wrap items-center gap-3 p-4 bg-white dark:bg-slate-800 rounded-lg border border-slate-200 dark:border-slate-700">
      <div className="flex items-center gap-2">
        <label className="text-sm font-medium text-slate-700 dark:text-slate-300 whitespace-nowrap">
          {t('testTaskFilter.status')}:
        </label>
        <Select
          value={filters.status || 'all'}
          onValueChange={handleStatusChange}
        >
          <SelectTrigger className="w-32">
            <SelectValue placeholder={t('testTaskFilter.selectStatus')} />
          </SelectTrigger>
          <SelectContent>
            {STATUS_OPTIONS.map((status) => (
              <SelectItem key={status} value={status}>
                {t(`testTaskFilter.statusOptions.${status}`)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {workers.length > 0 && (
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-slate-700 dark:text-slate-300 whitespace-nowrap">
            {t('testTaskFilter.worker')}:
          </label>
          <Select
            value={filters.worker_name || 'all'}
            onValueChange={handleWorkerChange}
          >
            <SelectTrigger className="w-40">
              <SelectValue placeholder={t('testTaskFilter.selectWorker')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">
                {t('testTaskFilter.all')}
              </SelectItem>
              {workers.map((worker) => (
                <SelectItem key={worker} value={worker}>
                  {worker}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      {planKeys.length > 0 && (
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-slate-700 dark:text-slate-300 whitespace-nowrap">
            {t('testTaskFilter.planKey')}:
          </label>
          <Select
            value={filters.plan_key || 'all'}
            onValueChange={handlePlanKeyChange}
          >
            <SelectTrigger className="w-40">
              <SelectValue placeholder={t('testTaskFilter.selectPlanKey')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">
                {t('testTaskFilter.all')}
              </SelectItem>
              {planKeys.map((planKey) => (
                <SelectItem key={planKey} value={planKey}>
                  {planKey}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      {hasActiveFilters && (
        <Button
          variant="outline"
          size="sm"
          onClick={handleReset}
          className="gap-1.5"
        >
          <RotateCcw className="h-3.5 w-3.5" />
          {t('testTaskFilter.reset')}
        </Button>
      )}
    </div>
  )
}
