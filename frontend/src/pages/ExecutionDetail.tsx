import { useState } from 'react'
import { useParams, useNavigate } from 'react-router'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { CaseList } from '@/components/test/CaseList'
import { CaseDetail } from '@/components/test/CaseDetail'
import { testDetailApi, type TestDetail } from '@/services/testDetailApi'

function ExecutionDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [selectedCase, setSelectedCase] = useState<TestDetail | null>(null)

  const { data: detailsData, isLoading: detailsLoading, error: detailsError } = useQuery({
    queryKey: ['testDetails', id],
    queryFn: async () => {
      if (!id) return null
      const response = await testDetailApi.getTestDetailsByTaskId(Number(id))
      if (!response.success) {
        throw new Error(response.error?.message || 'Failed to fetch test details')
      }
      return response.data
    },
    enabled: !!id,
  })

  const cases = detailsData?.details || []

  const handleSelectCase = (testCase: TestDetail) => {
    setSelectedCase(testCase)
  }

  const handleBackToHome = () => {
    navigate('/')
  }

  if (detailsLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-slate-500 dark:text-slate-400">{t('testTaskDashboard.loading')}</div>
      </div>
    )
  }

  if (detailsError || !detailsData) {
    return (
      <div className="space-y-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={handleBackToHome}
          className="gap-2"
        >
          <ArrowLeft className="h-4 w-4" />
          {t('executionDetail.backToHome')}
        </Button>
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-red-500 dark:text-red-400">{t('executionDetail.notFound')}</div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={handleBackToHome}
          className="gap-2"
        >
          <ArrowLeft className="h-4 w-4" />
          {t('executionDetail.backToHome')}
        </Button>
      </div>

      <div className="flex flex-col lg:flex-row gap-4 h-[calc(100vh-180px)]">
        <div className="w-full lg:w-[320px] flex-shrink-0 bg-white dark:bg-slate-800 rounded-xl shadow-md border border-slate-200 dark:border-slate-700 overflow-hidden">
          <div className="p-4 border-b border-slate-200 dark:border-slate-700">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">
              {t('executionDetail.testResults')} ({cases.length})
            </h2>
          </div>
          <div className="h-[calc(100%-60px)]">
            <CaseList
              cases={cases}
              selectedId={selectedCase?.id ?? null}
              onSelect={handleSelectCase}
            />
          </div>
        </div>

        <div className="flex-1 overflow-auto">
          <CaseDetail testCase={selectedCase} />
        </div>
      </div>
    </div>
  )
}

export default ExecutionDetail
