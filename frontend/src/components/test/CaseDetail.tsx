import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { Clock, AlertCircle, CheckCircle, XCircle } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ScreenshotPreview } from './ScreenshotPreview'
import { testDetailApi, type TestDetail, type TestStepDetail } from '@/services/testDetailApi'

interface CaseDetailProps {
  testCase: TestDetail | null
  steps?: TestStepDetail[]
}

const STATUS_CONFIG: Record<TestDetail['test_status'], { variant: 'default' | 'secondary' | 'destructive' | 'outline'; className: string }> = {
  passed: { variant: 'default', className: 'bg-green-500 hover:bg-green-600 text-white' },
  failed: { variant: 'destructive', className: 'bg-red-500 hover:bg-red-600 text-white' },
  skipped: { variant: 'secondary', className: 'bg-slate-500 hover:bg-slate-600 text-white' },
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60
  if (hours > 0) {
    return `${hours}h ${minutes}m ${secs}s`
  }
  if (minutes > 0) {
    return `${minutes}m ${secs}s`
  }
  return `${secs}s`
}

function formatTimestamp(timestamp: number): string {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString()
}

export function CaseDetail({ testCase, steps: externalSteps }: CaseDetailProps) {
  const { t } = useTranslation()

  const { data: stepsData, isLoading: stepsLoading } = useQuery({
    queryKey: ['testSteps', testCase?.id],
    queryFn: async () => {
      if (!testCase) return null
      const response = await testDetailApi.getTestStepsByDetailId(testCase.id)
      if (!response.success) {
        throw new Error(response.error?.message || 'Failed to fetch steps')
      }
      return response.data
    },
    enabled: !!testCase && !externalSteps,
  })

  const steps = externalSteps || stepsData?.steps || []

  if (!testCase) {
    return (
      <Card className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardContent className="flex items-center justify-center py-12">
          <p className="text-slate-500 dark:text-slate-400">
            {t('caseDetail.noCaseSelected')}
          </p>
        </CardContent>
      </Card>
    )
  }

  const statusConfig = STATUS_CONFIG[testCase.test_status]

  return (
    <div className="space-y-6">
      <Card className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardHeader className="pb-4">
          <CardTitle className="text-lg font-semibold text-slate-900 dark:text-slate-100">
            {t('caseDetail.basicInfo')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="text-sm font-medium text-slate-500 dark:text-slate-400">
                {t('caseDetail.caseKey')}
              </label>
              <p className="text-slate-900 dark:text-slate-100 font-mono text-sm">
                {testCase.test_name}
              </p>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium text-slate-500 dark:text-slate-400">
                {t('caseDetail.status')}
              </label>
              <div>
                <Badge className={statusConfig.className}>
                  {testCase.test_status === 'passed' && <CheckCircle className="w-3.5 h-3.5 mr-1" />}
                  {testCase.test_status === 'failed' && <XCircle className="w-3.5 h-3.5 mr-1" />}
                  {testCase.test_status === 'skipped' && <AlertCircle className="w-3.5 h-3.5 mr-1" />}
                  {t(`caseDetail.statusOptions.${testCase.test_status}`)}
                </Badge>
              </div>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium text-slate-500 dark:text-slate-400">
                {t('caseDetail.startTime')}
              </label>
              <div className="flex items-center gap-2 text-slate-900 dark:text-slate-100">
                <Clock className="w-4 h-4 text-slate-400" />
                <span className="text-sm">{formatTimestamp(testCase.test_start_time)}</span>
              </div>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium text-slate-500 dark:text-slate-400">
                {t('caseDetail.duration')}
              </label>
              <div className="flex items-center gap-2 text-slate-900 dark:text-slate-100">
                <Clock className="w-4 h-4 text-slate-400" />
                <span className="text-sm font-mono">{formatDuration(testCase.duration)}</span>
              </div>
            </div>
          </div>

          {testCase.error_message && (
            <div className="space-y-1">
              <label className="text-sm font-medium text-red-500 dark:text-red-400">
                {t('caseDetail.errorMessage')}
              </label>
              <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
                <pre className="text-sm text-red-700 dark:text-red-300 whitespace-pre-wrap break-words font-mono">
                  {testCase.error_message}
                </pre>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
        <CardHeader className="pb-4">
          <CardTitle className="text-lg font-semibold text-slate-900 dark:text-slate-100">
            {t('caseDetail.steps')} ({steps.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {stepsLoading ? (
            <div className="flex items-center justify-center py-8">
              <p className="text-slate-500 dark:text-slate-400">
                {t('caseDetail.loadingSteps')}
              </p>
            </div>
          ) : steps.length === 0 ? (
            <div className="flex items-center justify-center py-8">
              <p className="text-slate-500 dark:text-slate-400">
                {t('caseDetail.noSteps')}
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {steps.map((step, index) => (
                <Card
                  key={step.id}
                  className="bg-slate-50 dark:bg-slate-900/50 border-slate-200 dark:border-slate-700"
                >
                  <CardContent className="p-4">
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex-1 space-y-3">
                        <div className="flex items-center gap-3">
                          <span className="flex-shrink-0 w-7 h-7 rounded-full bg-slate-200 dark:bg-slate-700 flex items-center justify-center text-sm font-medium text-slate-600 dark:text-slate-300">
                            {index + 1}
                          </span>
                          <h4 className="text-sm font-medium text-slate-900 dark:text-slate-100">
                            {step.step_name}
                          </h4>
                          <Badge
                            className={
                              step.passed
                                ? 'bg-green-500 hover:bg-green-600 text-white'
                                : 'bg-red-500 hover:bg-red-600 text-white'
                            }
                          >
                            {step.passed ? (
                              <>
                                <CheckCircle className="w-3 h-3 mr-1" />
                                {t('caseDetail.stepPassed')}
                              </>
                            ) : (
                              <>
                                <XCircle className="w-3 h-3 mr-1" />
                                {t('caseDetail.stepFailed')}
                              </>
                            )}
                          </Badge>
                        </div>

                        <div className="flex items-center gap-4 text-sm text-slate-500 dark:text-slate-400 ml-10">
                          <div className="flex items-center gap-1.5">
                            <Clock className="w-3.5 h-3.5" />
                            <span className="font-mono">{formatDuration(step.duration)}</span>
                          </div>
                        </div>
                      </div>

                      {step.screenshot && (
                        <div className="flex-shrink-0">
                          <ScreenshotPreview
                            src={step.screenshot}
                            alt={step.step_name}
                            className="shadow-md"
                          />
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
