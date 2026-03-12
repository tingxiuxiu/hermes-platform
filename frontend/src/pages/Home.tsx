import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { RefreshCw, Inbox } from 'lucide-react'
import { TestTaskCard } from '@/components/TestTaskCard'
import { TestTaskFilter, type FilterParams, type TaskStatus } from '@/components/TestTaskFilter'
import { Pagination } from '@/components/ui/pagination'
import { Button } from '@/components/ui/button'
import { testTaskApi, type TestTaskQueryParams } from '@/services/testTaskApi'

const PAGE_SIZE = 12

function TaskSkeleton() {
  return (
    <div className="bg-white dark:bg-slate-800 rounded-xl shadow-md border border-slate-200 dark:border-slate-700 p-5 animate-pulse">
      <div className="flex items-start justify-between mb-4">
        <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded w-3/4" />
        <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded-full w-16" />
      </div>
      <div className="mb-4">
        <div className="flex items-center justify-between mb-2">
          <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-12" />
          <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-16" />
        </div>
        <div className="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-full" />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-full" />
        <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-full" />
        <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-full" />
        <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-full" />
      </div>
    </div>
  )
}

function Home() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [currentPage, setCurrentPage] = useState(1)
  const [filters, setFilters] = useState<FilterParams>({})

  const queryParams: TestTaskQueryParams = useMemo(() => ({
    page: currentPage,
    page_size: PAGE_SIZE,
    status: filters.status as TaskStatus | undefined,
    worker_name: filters.worker_name,
    plan_key: filters.plan_key,
  }), [currentPage, filters])

  const { data, isLoading, isFetching, refetch } = useQuery({
    queryKey: ['testTasks', queryParams],
    queryFn: () => testTaskApi.getTestTasks(queryParams),
  })

  const tasks = data?.data?.tasks ?? []
  const total = data?.data?.total ?? 0
  const totalPages = Math.ceil(total / PAGE_SIZE)

  const workers = useMemo(() => {
    if (!tasks.length) return []
    const uniqueWorkers = [...new Set(tasks.map(task => task.worker_name))]
    return uniqueWorkers.filter(Boolean).sort()
  }, [tasks])

  const planKeys = useMemo(() => {
    if (!tasks.length) return []
    const uniquePlanKeys = [...new Set(tasks.map(task => task.plan_key))]
    return uniquePlanKeys.filter(Boolean).sort()
  }, [tasks])

  const handleFilterChange = (newFilters: FilterParams) => {
    setFilters(newFilters)
    setCurrentPage(1)
  }

  const handlePageChange = (page: number) => {
    setCurrentPage(page)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const handleRefresh = () => {
    refetch()
  }

  const handleViewDetail = (taskId: number) => {
    navigate(`/execution/${taskId}`)
  }

  const isRefreshing = isFetching && !isLoading

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">
            {t('testTaskDashboard.title')}
          </h1>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={handleRefresh}
          disabled={isRefreshing}
          className="gap-2"
        >
          <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          {t('testTaskDashboard.refresh')}
        </Button>
      </div>

      <TestTaskFilter
        onFilterChange={handleFilterChange}
        workers={workers}
        planKeys={planKeys}
      />

      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {Array.from({ length: PAGE_SIZE }).map((_, index) => (
            <TaskSkeleton key={index} />
          ))}
        </div>
      ) : tasks.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="w-16 h-16 bg-slate-100 dark:bg-slate-800 rounded-full flex items-center justify-center mb-4">
            <Inbox className="w-8 h-8 text-slate-400 dark:text-slate-500" />
          </div>
          <h3 className="text-lg font-medium text-slate-900 dark:text-slate-100 mb-2">
            {t('testTaskDashboard.noTasks')}
          </h3>
          <p className="text-sm text-slate-500 dark:text-slate-400 max-w-sm">
            {t('testTaskDashboard.noTasksDescription')}
          </p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {tasks.map((task) => (
              <TestTaskCard key={task.id} task={task} onDetailClick={handleViewDetail} />
            ))}
          </div>

          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={handlePageChange}
            totalItems={total}
          />
        </>
      )}
    </div>
  )
}

export default Home
