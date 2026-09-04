import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { FilterX, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  type DashboardRunningExecution,
  getDashboardOverview,
} from './api/dashboard-api'
import { AutomationOverviewChart } from './components/automation-overview-chart'
import { AutomationSummaryCards } from './components/automation-summary-cards'
import { RunningExecutionDialog } from './components/running-execution-dialog'
import { RunningExecutions } from './components/running-executions'

type DashboardSearch = {
  days?: number
  jobId?: number
  jobName?: string
}

type DashboardNavigate = (options: {
  search: (prev: DashboardSearch) => DashboardSearch
}) => Promise<void>

const DAY_OPTIONS = [7, 14, 30] as const

export function Dashboard({
  search,
  navigate,
}: {
  search: DashboardSearch
  navigate: DashboardNavigate
}) {
  const [selectedExecution, setSelectedExecution] =
    useState<DashboardRunningExecution | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [jobNameInput, setJobNameInput] = useState(search.jobName ?? '')
  const selectedDays = search.days ?? 7

  useEffect(() => {
    setJobNameInput(search.jobName ?? '')
  }, [search.jobName])

  const overviewQuery = useQuery({
    queryKey: ['dashboard', 'automation-overview', search],
    queryFn: () =>
      getDashboardOverview({
        days: selectedDays,
        jobId: search.jobId,
        jobName: search.jobName,
      }),
  })

  const data = overviewQuery.data

  const hasFilters = Boolean(
    search.jobId || search.jobName || search.days !== undefined
  )

  async function applyFilters() {
    const normalizedJobName = jobNameInput.trim()
    await navigate({
      search: (prev) => ({
        ...prev,
        days: selectedDays,
        jobId: prev.jobId,
        jobName: normalizedJobName || undefined,
      }),
    })
  }

  async function resetFilters() {
    setJobNameInput('')
    await navigate({
      search: () => ({
        days: undefined,
        jobId: undefined,
        jobName: undefined,
      }),
    })
  }

  const executionListSearch = {
    jobId: search.jobId,
    jobName: search.jobName ?? '',
    page: 1,
    pageSize: 10,
    status: [],
  }

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-3'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>
              Automation Dashboard
            </h1>
            <p className='text-muted-foreground'>
              聚合展示自动化任务健康度、近 7 天趋势，以及当前运行中的 execution
              实时进度。
            </p>
          </div>
          <div className='flex flex-wrap items-center gap-2'>
            <Button variant='outline' onClick={() => overviewQuery.refetch()}>
              <RefreshCw className='size-4' />
              Refresh
            </Button>
            <Button asChild>
              <Link
                to='/automation/executions'
                search={() => executionListSearch}
              >
                Open Execution List
              </Link>
            </Button>
          </div>
        </div>

        <Card>
          <CardHeader className='gap-2'>
            <CardTitle>Dashboard Filters</CardTitle>
            <CardDescription>
              通过时间窗口和任务名称快速收窄趋势图与运行卡片，并保留默认快照视图作为快速入口。
            </CardDescription>
          </CardHeader>
          <CardContent className='flex flex-col gap-3 lg:flex-row lg:items-end'>
            <label className='space-y-2'>
              <span className='text-sm font-medium'>Trend Window</span>
              <Select
                value={String(selectedDays)}
                onValueChange={(value) => {
                  void navigate({
                    search: (prev) => ({
                      ...prev,
                      days: Number(value),
                    }),
                  })
                }}
              >
                <SelectTrigger className='w-full lg:w-40'>
                  <SelectValue placeholder='Select days' />
                </SelectTrigger>
                <SelectContent>
                  {DAY_OPTIONS.map((days) => (
                    <SelectItem key={days} value={String(days)}>
                      Last {days} days
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </label>

            <label className='space-y-2 lg:min-w-80'>
              <span className='text-sm font-medium'>Job Name</span>
              <Input
                value={jobNameInput}
                onChange={(event) => setJobNameInput(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    event.preventDefault()
                    void applyFilters()
                  }
                }}
                placeholder='Search by Jenkins job name'
              />
            </label>

            <div className='flex flex-wrap gap-2'>
              <Button onClick={() => void applyFilters()}>Apply Filters</Button>
              <Button
                variant='outline'
                onClick={() => void resetFilters()}
                disabled={!hasFilters}
              >
                <FilterX className='size-4' />
                Reset
              </Button>
              <Button variant='ghost' asChild>
                <Link
                  to='/automation/executions'
                  search={() => executionListSearch}
                >
                  Drill Into Execution List
                </Link>
              </Button>
            </div>
          </CardContent>
        </Card>

        {overviewQuery.isLoading ? (
          <Card>
            <CardContent className='py-16 text-center text-sm text-muted-foreground'>
              Loading automation dashboard...
            </CardContent>
          </Card>
        ) : data ? (
          <>
            <AutomationSummaryCards summary={data.summary} />

            <div className='grid gap-4 xl:grid-cols-[minmax(0,1.3fr)_minmax(0,1fr)]'>
              <Card>
                <CardHeader>
                  <CardTitle>Execution Trend</CardTitle>
                  <CardDescription>
                    最近 {selectedDays} 天 execution 数量与 pass/fail 用例趋势。
                  </CardDescription>
                </CardHeader>
                <CardContent className='pl-2'>
                  <AutomationOverviewChart trends={data.trends} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Trend Notes</CardTitle>
                  <CardDescription>
                    用于快速判断自动化健康度和运行负载的关键读数。
                  </CardDescription>
                </CardHeader>
                <CardContent className='space-y-4 text-sm'>
                  <InsightLine
                    label={`Passed Cases · ${selectedDays}d`}
                    value={String(data.summary.success_cases_7d)}
                  />
                  <InsightLine
                    label={`Failed Cases · ${selectedDays}d`}
                    value={String(data.summary.failure_cases_7d)}
                  />
                  <InsightLine
                    label={`Skipped Cases · ${selectedDays}d`}
                    value={String(data.summary.skipped_cases_7d)}
                  />
                  <InsightLine
                    label={`Avg Duration · ${selectedDays}d`}
                    value={
                      data.summary.avg_execution_duration_7d == null
                        ? '--'
                        : `${data.summary.avg_execution_duration_7d.toFixed(2)} s`
                    }
                  />
                </CardContent>
              </Card>
            </div>

            <div className='space-y-4'>
              <div>
                <h2 className='text-xl font-semibold tracking-tight'>
                  Running Executions
                </h2>
                <p className='text-sm text-muted-foreground'>
                  展示当前正在运行的 execution 进度、pass/fail 情况和正在执行的
                  case。
                </p>
              </div>
              <RunningExecutions
                items={data.running_executions}
                onOpen={(execution) => {
                  setSelectedExecution(execution)
                  setDialogOpen(true)
                }}
              />
            </div>
          </>
        ) : (
          <Card>
            <CardContent className='py-16 text-center text-sm text-muted-foreground'>
              Failed to load dashboard data.
            </CardContent>
          </Card>
        )}
      </Main>

      <RunningExecutionDialog
        execution={selectedExecution}
        open={dialogOpen}
        onOpenChange={(open) => {
          setDialogOpen(open)
          if (!open) {
            setSelectedExecution(null)
          }
        }}
      />
    </>
  )
}

function InsightLine({ label, value }: { label: string; value: string }) {
  return (
    <div className='flex items-center justify-between rounded-lg border px-4 py-3'>
      <span className='text-muted-foreground'>{label}</span>
      <span className='font-semibold'>{value}</span>
    </div>
  )
}
