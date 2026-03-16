import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useParams } from 'react-router'
import { FileText, Search, Plus, Edit, Trash2, Loader2, X, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { projectApi, type Project, type ProjectVersion, type TestPlan, type TestCase, type DictItem } from '@/services/projectApi'

function TestCaseList() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { projectId, versionId, planId } = useParams<{ projectId: string; versionId: string; planId: string }>()
  const [project, setProject] = useState<Project | null>(null)
  const [version, setVersion] = useState<ProjectVersion | null>(null)
  const [testPlan, setTestPlan] = useState<TestPlan | null>(null)
  const [testCases, setTestCases] = useState<TestCase[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [showAddModal, setShowAddModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [addLoading, setAddLoading] = useState(false)
  const [statuses, setStatuses] = useState<DictItem[]>([])
  const [priorities, setPriorities] = useState<DictItem[]>([])
  const [selectedCase, setSelectedCase] = useState<TestCase | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    status: 'developing',
    remark: '',
    pre_condition: '',
    step_description: '',
    expected_result: '',
    priority: 3
  })

  const fetchData = async () => {
    if (!projectId || !versionId || !planId) return
    try {
      setLoading(true)
      setError(null)
      
      const projResp = await projectApi.getProject(Number(projectId))
      setProject(projResp.data.project)
      
      const verResp = await projectApi.getVersion(Number(versionId))
      setVersion(verResp.data.version)
      
      const planResp = await projectApi.getTestPlan(Number(planId))
      setTestPlan(planResp.data.test_plan)
      
      const caseResp = await projectApi.getTestCases(Number(planId))
      setTestCases(caseResp.data.test_cases || [])

      const statusResp = await projectApi.getTestCaseStatuses()
      setStatuses(statusResp.data.statuses || [])

      const priorityResp = await projectApi.getPriorities()
      setPriorities(priorityResp.data.priorities || [])
    } catch (err) {
      setError(t('project.errorFetching', '获取数据失败'))
      console.error('Error fetching data:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [projectId, versionId, planId, t])

  const filteredCases = testCases.filter(
    (tc) =>
      tc.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      tc.case_key?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      tc.remark?.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const handleAddTestCase = async () => {
    if (!formData.name.trim() || !planId) return
    try {
      setAddLoading(true)
      await projectApi.createTestCase(Number(planId), formData)
      setShowAddModal(false)
      setFormData({
        name: '',
        status: 'developing',
        remark: '',
        pre_condition: '',
        step_description: '',
        expected_result: '',
        priority: 3
      })
      fetchData()
    } catch (err) {
      console.error('Error creating test case:', err)
    } finally {
      setAddLoading(false)
    }
  }

  const handleUpdateTestCase = async () => {
    if (!selectedCase || !formData.name.trim()) return
    try {
      setAddLoading(true)
      await projectApi.updateTestCase(selectedCase.id, formData)
      setShowEditModal(false)
      setSelectedCase(null)
      setFormData({
        name: '',
        status: 'developing',
        remark: '',
        pre_condition: '',
        step_description: '',
        expected_result: '',
        priority: 3
      })
      fetchData()
    } catch (err) {
      console.error('Error updating test case:', err)
    } finally {
      setAddLoading(false)
    }
  }

  const handleDeleteTestCase = async (id: number) => {
    if (!confirm(t('project.confirmDeleteCase', '确定要删除这个测试用例吗？'))) return
    try {
      await projectApi.deleteTestCase(id)
      fetchData()
    } catch (err) {
      console.error('Error deleting test case:', err)
    }
  }

  const openEditModal = (tc: TestCase) => {
    setSelectedCase(tc)
    setFormData({
      name: tc.name,
      status: tc.status,
      remark: tc.remark || '',
      pre_condition: tc.pre_condition || '',
      step_description: tc.step_description || '',
      expected_result: tc.expected_result || '',
      priority: tc.priority
    })
    setShowEditModal(true)
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'developing':
        return <Badge variant="outline">{t('project.statusDeveloping', '开发中')}</Badge>
      case 'ready':
        return <Badge variant="default">{t('project.statusReady', '就绪')}</Badge>
      case 'skip':
        return <Badge variant="secondary">{t('project.statusSkip', '跳过')}</Badge>
      default:
        return <Badge>{status}</Badge>
    }
  }

  const getPriorityBadge = (priority: number) => {
    switch (priority) {
      case 1:
        return <Badge variant="destructive">{t('project.priorityHigh', '高')}</Badge>
      case 2:
        return <Badge variant="default">{t('project.priorityMedium', '中')}</Badge>
      case 3:
        return <Badge variant="outline">{t('project.priorityLow', '低')}</Badge>
      default:
        return <Badge>{priority}</Badge>
    }
  }

  return (
    <div className="container mx-auto py-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="sm" onClick={() => navigate(`/projects/${projectId}/versions/${versionId}/test-plans`)}>
              <ArrowLeft className="h-4 w-4 mr-1" />
            </Button>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              {project?.name} - v{version?.version} - {testPlan?.name} {t('project.testCases', '测试用例')}
            </CardTitle>
          </div>
          <Button onClick={() => setShowAddModal(true)}>
            <Plus className="h-4 w-4 mr-2" />
            {t('project.addTestCase', '添加测试用例')}
          </Button>
        </CardHeader>
        <CardContent>
          <div className="mb-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                placeholder={t('project.searchCase', '搜索测试用例...')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
            </div>
          ) : error ? (
            <div className="text-center py-8 text-red-500">{error}</div>
          ) : filteredCases.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              {t('project.noTestCases', '暂无测试用例')}
            </div>
          ) : (
            <div className="space-y-4">
              {filteredCases.map((tc) => (
                <Card key={tc.id} className="hover:shadow-md transition-shadow">
                  <CardContent className="pt-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-sm text-gray-500">{tc.case_key}</span>
                          <h3 className="font-semibold text-lg">{tc.name}</h3>
                          {getStatusBadge(tc.status)}
                          {getPriorityBadge(tc.priority)}
                        </div>
                        {tc.remark && (
                          <p className="text-sm text-gray-500 mt-1">
                            {t('project.remark', '备注')}: {tc.remark}
                          </p>
                        )}
                        {tc.pre_condition && (
                          <p className="text-sm text-gray-500 mt-1">
                            {t('project.preCondition', '前置条件')}: {tc.pre_condition}
                          </p>
                        )}
                        {tc.step_description && (
                          <p className="text-sm text-gray-500 mt-1">
                            {t('project.steps', '步骤')}: {tc.step_description}
                          </p>
                        )}
                        {tc.expected_result && (
                          <p className="text-sm text-gray-500 mt-1">
                            {t('project.expectedResult', '预期结果')}: {tc.expected_result}
                          </p>
                        )}
                      </div>
                      <div className="flex items-center gap-2">
                        <Button variant="outline" size="sm" onClick={() => openEditModal(tc)}>
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button variant="outline" size="sm" onClick={() => handleDeleteTestCase(tc.id)}>
                          <Trash2 className="h-4 w-4 text-red-500" />
                        </Button>
                      </div>
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
          <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <CardHeader>
              <CardTitle>{t('project.addTestCase', '添加测试用例')}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <Label htmlFor="name">{t('project.caseName', '用例名称')} *</Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder={t('project.caseNamePlaceholder', '请输入用例名称')}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label>{t('project.status', '状态')}</Label>
                    <Select value={formData.status} onValueChange={(v) => setFormData({ ...formData, status: v })}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {statuses.map((s) => (
                          <SelectItem key={s.value} value={String(s.value)}>{s.label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label>{t('project.priority', '优先级')}</Label>
                    <Select value={String(formData.priority)} onValueChange={(v) => setFormData({ ...formData, priority: Number(v) })}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {priorities.map((p) => (
                          <SelectItem key={p.value} value={String(p.value)}>{p.label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div>
                  <Label htmlFor="remark">{t('project.remark', '备注')}</Label>
                  <Input
                    id="remark"
                    value={formData.remark}
                    onChange={(e) => setFormData({ ...formData, remark: e.target.value })}
                    placeholder={t('project.remarkPlaceholder', '请输入备注')}
                  />
                </div>
                <div>
                  <Label htmlFor="pre_condition">{t('project.preCondition', '前置条件')}</Label>
                  <Input
                    id="pre_condition"
                    value={formData.pre_condition}
                    onChange={(e) => setFormData({ ...formData, pre_condition: e.target.value })}
                    placeholder={t('project.preConditionPlaceholder', '请输入前置条件')}
                  />
                </div>
                <div>
                  <Label htmlFor="step_description">{t('project.steps', '步骤描述')}</Label>
                  <Input
                    id="step_description"
                    value={formData.step_description}
                    onChange={(e) => setFormData({ ...formData, step_description: e.target.value })}
                    placeholder={t('project.stepsPlaceholder', '请输入步骤描述')}
                  />
                </div>
                <div>
                  <Label htmlFor="expected_result">{t('project.expectedResult', '预期结果')}</Label>
                  <Input
                    id="expected_result"
                    value={formData.expected_result}
                    onChange={(e) => setFormData({ ...formData, expected_result: e.target.value })}
                    placeholder={t('project.expectedResultPlaceholder', '请输入预期结果')}
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <Button variant="outline" onClick={() => setShowAddModal(false)}>
                  <X className="h-4 w-4 mr-1" />
                  {t('common.cancel', '取消')}
                </Button>
                <Button onClick={handleAddTestCase} disabled={addLoading || !formData.name.trim()}>
                  {addLoading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                  {t('common.save', '保存')}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {showEditModal && selectedCase && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <CardHeader>
              <CardTitle>{t('project.editTestCase', '编辑测试用例')}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <Label htmlFor="editName">{t('project.caseName', '用例名称')} *</Label>
                  <Input
                    id="editName"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder={t('project.caseNamePlaceholder', '请输入用例名称')}
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label>{t('project.status', '状态')}</Label>
                    <Select value={formData.status} onValueChange={(v) => setFormData({ ...formData, status: v })}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {statuses.map((s) => (
                          <SelectItem key={s.value} value={String(s.value)}>{s.label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label>{t('project.priority', '优先级')}</Label>
                    <Select value={String(formData.priority)} onValueChange={(v) => setFormData({ ...formData, priority: Number(v) })}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {priorities.map((p) => (
                          <SelectItem key={p.value} value={String(p.value)}>{p.label}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div>
                  <Label htmlFor="editRemark">{t('project.remark', '备注')}</Label>
                  <Input
                    id="editRemark"
                    value={formData.remark}
                    onChange={(e) => setFormData({ ...formData, remark: e.target.value })}
                    placeholder={t('project.remarkPlaceholder', '请输入备注')}
                  />
                </div>
                <div>
                  <Label htmlFor="editPreCondition">{t('project.preCondition', '前置条件')}</Label>
                  <Input
                    id="editPreCondition"
                    value={formData.pre_condition}
                    onChange={(e) => setFormData({ ...formData, pre_condition: e.target.value })}
                    placeholder={t('project.preConditionPlaceholder', '请输入前置条件')}
                  />
                </div>
                <div>
                  <Label htmlFor="editStepDescription">{t('project.steps', '步骤描述')}</Label>
                  <Input
                    id="editStepDescription"
                    value={formData.step_description}
                    onChange={(e) => setFormData({ ...formData, step_description: e.target.value })}
                    placeholder={t('project.stepsPlaceholder', '请输入步骤描述')}
                  />
                </div>
                <div>
                  <Label htmlFor="editExpectedResult">{t('project.expectedResult', '预期结果')}</Label>
                  <Input
                    id="editExpectedResult"
                    value={formData.expected_result}
                    onChange={(e) => setFormData({ ...formData, expected_result: e.target.value })}
                    placeholder={t('project.expectedResultPlaceholder', '请输入预期结果')}
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <Button variant="outline" onClick={() => {
                  setShowEditModal(false)
                  setSelectedCase(null)
                }}>
                  <X className="h-4 w-4 mr-1" />
                  {t('common.cancel', '取消')}
                </Button>
                <Button onClick={handleUpdateTestCase} disabled={addLoading || !formData.name.trim()}>
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

export default TestCaseList
