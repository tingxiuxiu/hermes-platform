import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useParams } from 'react-router'
import { ListTodo, Plus, Trash2, Loader2, ChevronRight, X, ArrowLeft, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { projectApi, type Project, type ProjectVersion, type TestPlan } from '@/services/projectApi'

function TestPlanList() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { projectId, versionId } = useParams<{ projectId: string; versionId: string }>()
  const [project, setProject] = useState<Project | null>(null)
  const [version, setVersion] = useState<ProjectVersion | null>(null)
  const [testPlans, setTestPlans] = useState<TestPlan[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showAddModal, setShowAddModal] = useState(false)
  const [addLoading, setAddLoading] = useState(false)
  const [formData, setFormData] = useState({ name: '', description: '', status: 'draft' })

  const fetchData = async () => {
    if (!projectId || !versionId) return
    try {
      setLoading(true)
      setError(null)
      
      const projResp = await projectApi.getProject(Number(projectId))
      setProject(projResp.data.project)
      
      const verResp = await projectApi.getVersion(Number(versionId))
      setVersion(verResp.data.version)
      
      const planResp = await projectApi.getTestPlans(Number(versionId))
      setTestPlans(planResp.data.test_plans || [])
    } catch (err) {
      setError(t('project.errorFetching', '获取数据失败'))
      console.error('Error fetching data:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [projectId, versionId, t])

  const handleAddTestPlan = async () => {
    if (!formData.name.trim() || !versionId) return
    try {
      setAddLoading(true)
      await projectApi.createTestPlan(Number(versionId), formData)
      setShowAddModal(false)
      setFormData({ name: '', description: '', status: 'draft' })
      fetchData()
    } catch (err) {
      console.error('Error creating test plan:', err)
    } finally {
      setAddLoading(false)
    }
  }

  const handleDeleteTestPlan = async (id: number) => {
    if (!confirm(t('project.confirmDeletePlan', '确定要删除这个测试计划吗？'))) return
    try {
      await projectApi.deleteTestPlan(id)
      fetchData()
    } catch (err) {
      console.error('Error deleting test plan:', err)
    }
  }

  const navigateToTestCases = (plan: TestPlan) => {
    navigate(`/projects/${projectId}/versions/${versionId}/test-plans/${plan.id}/test-cases`)
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'draft':
        return <Badge variant="outline">{t('project.statusDraft', '草稿')}</Badge>
      case 'active':
        return <Badge variant="default">{t('project.statusActive', '进行中')}</Badge>
      case 'completed':
        return <Badge variant="secondary">{t('project.statusCompleted', '已完成')}</Badge>
      default:
        return <Badge>{status}</Badge>
    }
  }

  return (
    <div className="container mx-auto py-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="sm" onClick={() => navigate(`/projects/${projectId}/versions`)}>
              <ArrowLeft className="h-4 w-4 mr-1" />
            </Button>
            <CardTitle className="flex items-center gap-2">
              <ListTodo className="h-5 w-5" />
              {project?.name} - v{version?.version} {t('project.testPlans', '测试计划')}
            </CardTitle>
          </div>
          <Button onClick={() => setShowAddModal(true)}>
            <Plus className="h-4 w-4 mr-2" />
            {t('project.addTestPlan', '添加测试计划')}
          </Button>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
            </div>
          ) : error ? (
            <div className="text-center py-8 text-red-500">{error}</div>
          ) : testPlans.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              {t('project.noTestPlans', '暂无测试计划')}
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {testPlans.map((plan) => (
                <Card key={plan.id} className="hover:shadow-md transition-shadow">
                  <CardContent className="pt-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <h3 className="font-semibold text-lg">{plan.name}</h3>
                          {getStatusBadge(plan.status)}
                        </div>
                        <p className="text-sm text-gray-500 mt-1 line-clamp-2">
                          {plan.description || t('project.noDescription', '暂无描述')}
                        </p>
                      </div>
                    </div>

                    <div className="mt-4 flex items-center gap-2">
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => navigateToTestCases(plan)}
                      >
                        <FileText className="h-4 w-4 mr-1" />
                        {t('project.testCases', '测试用例')}
                        <ChevronRight className="h-4 w-4" />
                      </Button>
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => handleDeleteTestPlan(plan.id)}
                      >
                        <Trash2 className="h-4 w-4 text-red-500" />
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>{t('project.addTestPlan', '添加测试计划')}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <Label htmlFor="name">{t('project.planName', '计划名称')}</Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder={t('project.planNamePlaceholder', '请输入计划名称')}
                  />
                </div>
                <div>
                  <Label htmlFor="description">{t('project.planDescription', '计划描述')}</Label>
                  <Input
                    id="description"
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    placeholder={t('project.planDescPlaceholder', '请输入计划描述')}
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <Button variant="outline" onClick={() => setShowAddModal(false)}>
                  <X className="h-4 w-4 mr-1" />
                  {t('common.cancel', '取消')}
                </Button>
                <Button onClick={handleAddTestPlan} disabled={addLoading || !formData.name.trim()}>
                  {addLoading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                  {t('common.save', '保存')}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}

export default TestPlanList
