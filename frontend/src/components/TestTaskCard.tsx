import { useTranslation } from 'react-i18next'
import { Clock, Server, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { TestTask } from '@/services/testTaskApi'

interface TestTaskCardProps {
  task: TestTask
  onClick?: () => void
  onDetailClick?: (taskId: number) => void
}

const STATUS_CONFIG: Record<TestTask['status'], { bg: string; text: string; border: string }> = {
  pending: {
    bg: 'bg-slate-100 dark:bg-slate-700',
    text: 'text-slate-700 dark:text-slate-300',
    border: 'border-slate-300 dark:border-slate-600',
  },
  running: {
    bg: 'bg-blue-100 dark:bg-blue-900/30',
    text: 'text-blue-700 dark:text-blue-300',
    border: 'border-blue-300 dark:border-blue-600',
  },
  completed: {
    bg: 'bg-green-100 dark:bg-green-900/30',
    text: 'text-green-700 dark:text-green-300',
    border: 'border-green-300 dark:border-green-600',
  },
  failed: {
    bg: 'bg-red-100 dark:bg-red-900/30',
    text: 'text-red-700 dark:text-red-300',
    border: 'border-red-300 dark:border-red-600',
  },
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60
  return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

function formatTimestamp(timestamp: number): string {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString()
}

function calculateProgress(passed: number, total: number): number {
  if (total === 0) return 0
  return Math.round((passed / total) * 100)
}

export function TestTaskCard({ task, onClick, onDetailClick }: TestTaskCardProps) {
  const { t } = useTranslation()
  const statusConfig = STATUS_CONFIG[task.status]
  const progress = calculateProgress(task.passed_tests, task.total_tests)

  const handleDetailClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    onDetailClick?.(task.id)
  }

  return (
    <div
      className={`
        bg-white dark:bg-slate-800
        rounded-xl shadow-md
        border border-slate-200 dark:border-slate-700
        p-5
        hover:shadow-xl hover:-translate-y-1 transition-all duration-300
        ${onClick ? 'cursor-pointer' : ''}
      `}
      onClick={onClick}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
      onKeyDown={(e) => {
        if (onClick && (e.key === 'Enter' || e.key === ' ')) {
          e.preventDefault()
          onClick()
        }
      }}
    >
      <div className="flex items-start justify-between mb-4">
        <h3 className="text-lg font-semibold text-slate-900 dark:text-slate-100 truncate flex-1 mr-3">
          {task.task_name}
        </h3>
        <span
          className={`
            px-3 py-1 rounded-full text-xs font-medium
            border ${statusConfig.bg} ${statusConfig.text} ${statusConfig.border}
          `}
        >
          {t(`testTaskCard.status.${task.status}`)}
        </span>
      </div>

      <div className="mb-4">
        <div className="flex items-center justify-between text-sm text-slate-600 dark:text-slate-400 mb-2">
          <span>{t('testTaskCard.progress')}</span>
          <span className="font-medium">{task.passed_tests} / {task.total_tests}</span>
        </div>
        <div className="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
          <div
            className="h-full rounded-full bg-gradient-to-r from-blue-500 to-green-500 transition-all duration-500"
            style={{ width: `${progress}%` }}
          />
        </div>
        <div className="flex justify-between text-xs text-slate-500 dark:text-slate-400 mt-1">
          <span>{progress}%</span>
          <span>
            {t('testTaskCard.failed')}: {task.failed_tests}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 text-sm mb-4">
        <div className="flex items-center gap-2 text-slate-600 dark:text-slate-400">
          <Server className="w-4 h-4 flex-shrink-0" />
          <span className="truncate" title={task.worker_name}>
            {task.worker_name}
          </span>
        </div>
        <div className="flex items-center gap-2 text-slate-600 dark:text-slate-400">
          <FileText className="w-4 h-4 flex-shrink-0" />
          <span className="truncate" title={task.plan_key}>
            {task.plan_key}
          </span>
        </div>
        <div className="flex items-center gap-2 text-slate-600 dark:text-slate-400">
          <Clock className="w-4 h-4 flex-shrink-0" />
          <span className="truncate" title={formatTimestamp(task.start_time)}>
            {formatTimestamp(task.start_time)}
          </span>
        </div>
        <div className="flex items-center gap-2 text-slate-600 dark:text-slate-400">
          <Clock className="w-4 h-4 flex-shrink-0" />
          <span className="font-mono">
            {formatDuration(task.duration)}
          </span>
        </div>
      </div>

      {onDetailClick && (
        <Button
          variant="outline"
          size="sm"
          className="w-full"
          onClick={handleDetailClick}
        >
          {t('testTaskCard.viewDetail')}
        </Button>
      )}
    </div>
  )
}
