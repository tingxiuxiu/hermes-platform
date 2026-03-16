import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useParams } from 'react-router'
import { Tag, Plus, Trash2, Loader2, ChevronRight, X, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { projectApi, type Project, type ProjectVersion, type TestPlan } from '@/services/projectApi'

function VersionList() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { projectId } = useParams<{ projectId: string }>()
  const [project, setProject] = useState<Project | null>(null)
  const [versions, setVersions] = useState<ProjectVersion[]>([])
  const [testPlansMap, setTestPlansMap] = useState<Record<number, TestPlan[]>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showAddModal, setShowAddModal] = useState(false)
  const [showPlanModal, setShowPlanModal] = useState(false)
  const [selectedVersion, setSelectedVersion] = useState<ProjectVersion | null>(null)
  const [addLoading, setAddLoading] = useState(false)
  const [formData, setFormData] = useState({ version: '', description: '', status: 'active' })
  const [planFormData, setPlanFormData] = useState({ name: '', description: '' })

  const fetchData = async () => {
    if (!projectId) return
    try {
      setLoading(true)
      setError(null)
      
      const projResp = await projectApi.getProject(Number(projectId))
      setProject(projResp.data.project)
      
      const verResp = await projectApi.getVersions(Number(projectId))
      setVersions(verResp.data.versions || [])

      const plans: Record<number, TestPlan[]> = {}
      for (const ver of verResp.data.versions || []) {
        const planResp = await projectApi.getTestPlans(ver.id)
        plans[ver.id] = planResp.data.test_plans || []
      }
      setTestPlansMap(plans)
    } catch (err) {
      setError(t('project.errorFetching', '获取数据失败'))
      console.error('Error fetching data:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [projectId, t])

  const handleAddVersion = async () => {
    if (!formData.version.trim() || !projectId) return
    try {
      setAddLoading(true)
      await projectApi.createVersion(Number(projectId), formData)
      setShowAddModal(false)
      setFormData({ version: '', description: '', status: 'active' })
      fetchData()
    } catch (err) {
      console.error('Error creating version:', err)
    } finally {
      setAddLoading(false)
    }
  }

  const handleDeleteVersion = async (id: number) => {
    if (!confirm(t('project.confirmDeleteVersion', '确定要删除这个版本吗？'))) return
    try {
      await projectApi.deleteVersion(id)
      fetchData()
    } catch (err) {
      console.error('Error deleting version:', err)
    }
  }

  const handleAddTestPlan = async () => {
    if (!selectedVersion || !planFormData.name.trim()) return
    try {
      setAddLoading(true)
      await projectApi.createTestPlan(selectedVersion.id, planFormData)
      setShowPlanModal(false)
      setPlanFormData({ name: '', description: '' })
      setSelectedVersion(null)
      fetchData()
    } catch (err) {
      console.error('Error creating test plan:', err)
    } finally {
      setAddLoading(false)
    }
  }

  const navigateToTestPlans = (version: ProjectVersion) => {
    navigate(`/projects/${projectId}/versions/${version.id}/test-plans`)
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge variant="default">{t('project.statusActive', '激活')}</Badge>
      case 'archived':
        return <Badge variant="secondary">{t('project.statusArchived', '归档')}</Badge>
      default:
        return <Badge>{status}</Badge>
    }
  }

  return (
    <div className="container mx-auto py-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="sm" onClick={() => navigate('/projects')}>
              <ArrowLeft className="h-4 w-4 mr-1" />
            </Button>
            <CardTitle className="flex items-center gap-2">
              <Tag className="h-5 w-5" />
              {project?.name || t('project.versions', '版本管理')}
            </CardTitle>
          </div>
          <Button onClick={() => setShowAddModal(true)}>
            <Plus className="h-4 w-4 mr-2" />
            {t('project.addVersion', '添加版本')}
          </Button>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
            </div>
          ) : error ? (
            <div className="text-center py-8 text-red-500">{error}</div>
          ) : versions.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              {t('project.noVersions', '暂无版本')}
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {versions.map((version) => (
                <Card key={version.id} className="hover:shadow-md transition-shadow">
                  <CardContent className="pt-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <h3 className="font-semibold text-lg">v{version.version}</h3>
                          {getStatusBadge(version.status)}
                        </div>
                        <p className="text-sm text-gray-500 mt-1 line-clamp-2">
                          {version.description || t('project.noDescription', '暂无描述')}
                        </p>
                      </div>
                    </div>
                    
                    <div className="mt-4 text-sm text-gray-500">
                      {testPlansMap[version.id]?.length || 0} {t('project.testPlans', '个测试计划')}
                    </div>

                    <div className="mt-4 flex items-center gap-2">
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => navigateToTestPlans(version)}
                      >
                        <ChevronRight className="h-4 w-4 mr-1" />
                        {t('project.manageTestPlans', '管理计划')}
                      </Button>
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => {
                          setSelectedVersion(version)
                          setShowPlanModal(true)
                        }}
                      >
                        <Plus className="h-4 w-4" />
                      </Button>
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => handleDeleteVersion(version.id)}
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
              <CardTitle>{t('project.addVersion', '添加版本')}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <Label htmlFor="version">{t('project.version', '版本号')}</Label>
                  <Input
                    id="version"
                    value={formData.version}
                    onChange={(e) => setFormData({ ...formData, version: e.target.value })}
                    placeholder={t('project.versionPlaceholder', '例如: 1.0.0')}
                  />
                </div>
                <div>
                  <Label htmlFor="description">{t('project.description', '描述')}</Label>
                  <Input
                    id="description"
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    placeholder={t('project.descriptionPlaceholder', '请输入版本描述')}
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <Button variant="outline" onClick={() => setShowAddModal(false)}>
                  <X className="h-4 w-4 mr-1" />
                  {t('common.cancel', '取消')}
                </Button>
                <Button onClick={handleAddVersion} disabled={addLoading || !formData.version.trim()}>
                  {addLoading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                  {t('common.save', '保存')}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {showPlanModal && selectedVersion && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>{t('project.addTestPlan', '添加测试计划')}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <Label htmlFor="planName">{t('project.planName', '计划名称')}</Label>
                  <Input
                    id="planName"
                    value={planFormData.name}
                    onChange={(e) => setPlanFormData({ ...planFormData, name: e.target.value })}
                    placeholder={t('project.planNamePlaceholder', '请输入计划名称')}
                  />
                </div>
                <div>
                  <Label htmlFor="planDesc">{t('project.planDescription', '计划描述')}</Label>
                  <Input
                    id="planDesc"
                    value={planFormData.description}
                    onChange={(e) => setPlanFormData({ ...planFormData, description: e.target.value })}
                    placeholder={t('project.planDescPlaceholder', '请输入计划描述')}
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <Button variant="outline" onClick={() => {
                  setShowPlanModal(false)
                  setSelectedVersion(null)
                  setPlanFormData({ name: '', description: '' })
                }}>
                  <X className="h-4 w-4 mr-1" />
                  {t('common.cancel', '取消')}
                </Button>
                <Button onClick={handleAddTestPlan} disabled={addLoading || !planFormData.name.trim()}>
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

export default VersionList
