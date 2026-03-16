import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router'
import { Folder, Search, Plus, Edit, Trash2, Loader2, ChevronRight, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { projectApi, type Project, type ProjectVersion } from '@/services/projectApi'

function ProjectList() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [projects, setProjects] = useState<Project[]>([])
  const [versionsMap, setVersionsMap] = useState<Record<number, ProjectVersion[]>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [showAddModal, setShowAddModal] = useState(false)
  const [showVersionModal, setShowVersionModal] = useState(false)
  const [selectedProject, setSelectedProject] = useState<Project | null>(null)
  const [addLoading, setAddLoading] = useState(false)
  const [formData, setFormData] = useState({ name: '', description: '' })
  const [versionFormData, setVersionFormData] = useState({ version: '', description: '' })

  const fetchProjects = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await projectApi.getProjects()
      if (response.data) {
        const projectsWithVersions = response.data.projects || []
        setProjects(projectsWithVersions)
        
        const versions: Record<number, ProjectVersion[]> = {}
        for (const project of projectsWithVersions) {
          if (project.versions && project.versions.length > 0) {
            versions[project.id] = project.versions
          } else {
            const verResp = await projectApi.getVersions(project.id)
            versions[project.id] = verResp.data?.versions || []
          }
        }
        setVersionsMap(versions)
      }
    } catch (err: unknown) {
      console.error('Error fetching projects:', err)
      const errorMessage = err instanceof Error ? err.message : t('project.errorFetching', '获取项目列表失败')
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchProjects()
  }, [t])

  const filteredProjects = projects.filter(
    (project) =>
      project.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      project.description?.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const handleAddProject = async () => {
    if (!formData.name.trim()) return
    try {
      setAddLoading(true)
      await projectApi.createProject(formData)
      setShowAddModal(false)
      setFormData({ name: '', description: '' })
      fetchProjects()
    } catch (err) {
      console.error('Error creating project:', err)
    } finally {
      setAddLoading(false)
    }
  }

  const handleDeleteProject = async (id: number) => {
    if (!confirm(t('project.confirmDelete', '确定要删除这个项目吗？'))) return
    try {
      await projectApi.deleteProject(id)
      fetchProjects()
    } catch (err) {
      console.error('Error deleting project:', err)
    }
  }

  const handleAddVersion = async () => {
    console.log('handleAddVersion called, selectedProject:', selectedProject)
    const projectId = selectedProject?.id
    console.log('projectId:', projectId)
    if (!selectedProject || !projectId || !versionFormData.version.trim()) {
      console.log('Validation failed: selectedProject:', selectedProject, 'version:', versionFormData.version)
      return
    }
    try {
      setAddLoading(true)
      await projectApi.createVersion(projectId, versionFormData)
      setShowVersionModal(false)
      setVersionFormData({ version: '', description: '' })
      setSelectedProject(null)
      fetchProjects()
    } catch (err) {
      console.error('Error creating version:', err)
    } finally {
      setAddLoading(false)
    }
  }

  const openVersionModal = (project: Project) => {
    setSelectedProject(project)
    setShowVersionModal(true)
  }

  const navigateToVersions = (project: Project) => {
    navigate(`/projects/${project.id}/versions`)
  }

  return (
    <div className="container mx-auto py-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Folder className="h-5 w-5" />
            {t('project.title', '项目管理')}
          </CardTitle>
          <Button onClick={() => setShowAddModal(true)}>
            <Plus className="h-4 w-4 mr-2" />
            {t('project.add', '新建项目')}
          </Button>
        </CardHeader>
        <CardContent>
          <div className="mb-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                placeholder={t('project.search', '搜索项目...')}
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
          ) : filteredProjects.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              {t('project.empty', '暂无项目')}
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {filteredProjects.map((project) => (
                <Card key={project.id} className="hover:shadow-md transition-shadow">
                  <CardContent className="pt-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <h3 className="font-semibold text-lg">{project.name}</h3>
                        <p className="text-sm text-gray-500 mt-1 line-clamp-2">
                          {project.description || t('project.noDescription', '暂无描述')}
                        </p>
                      </div>
                    </div>
                    
                    <div className="mt-4 text-sm text-gray-500">
                      {versionsMap[project.id]?.length || 0} {t('project.versions', '个版本')}
                    </div>

                    <div className="mt-4 flex items-center gap-2">
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => navigateToVersions(project)}
                      >
                        <ChevronRight className="h-4 w-4 mr-1" />
                        {t('project.manageVersions', '管理版本')}
                      </Button>
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => openVersionModal(project)}
                      >
                        <Plus className="h-4 w-4" />
                      </Button>
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => {
                          setSelectedProject(project)
                          setFormData({ name: project.name, description: project.description || '' })
                          setShowAddModal(true)
                        }}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => handleDeleteProject(project.id)}
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
              <CardTitle>{t('project.addModalTitle', selectedProject ? '编辑项目' : '新建项目')}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <Label htmlFor="name">{t('project.name', '项目名称')}</Label>
                  <Input
                    id="name"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder={t('project.namePlaceholder', '请输入项目名称')}
                  />
                </div>
                <div>
                  <Label htmlFor="description">{t('project.description', '描述')}</Label>
                  <Input
                    id="description"
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    placeholder={t('project.descriptionPlaceholder', '请输入项目描述')}
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <Button variant="outline" onClick={() => {
                  setShowAddModal(false)
                  setSelectedProject(null)
                  setFormData({ name: '', description: '' })
                }}>
                  <X className="h-4 w-4 mr-1" />
                  {t('common.cancel', '取消')}
                </Button>
                <Button onClick={handleAddProject} disabled={addLoading || !formData.name.trim()}>
                  {addLoading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                  {t('common.save', '保存')}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {showVersionModal && (
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
                    value={versionFormData.version}
                    onChange={(e) => setVersionFormData({ ...versionFormData, version: e.target.value })}
                    placeholder={t('project.versionPlaceholder', '例如: v1.0.0')}
                  />
                </div>
                <div>
                  <Label htmlFor="versionDesc">{t('project.versionDescription', '版本描述')}</Label>
                  <Input
                    id="versionDesc"
                    value={versionFormData.description}
                    onChange={(e) => setVersionFormData({ ...versionFormData, description: e.target.value })}
                    placeholder={t('project.versionDescPlaceholder', '请输入版本描述')}
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <Button variant="outline" onClick={() => {
                  setShowVersionModal(false)
                  setSelectedProject(null)
                  setVersionFormData({ version: '', description: '' })
                }}>
                  <X className="h-4 w-4 mr-1" />
                  {t('common.cancel', '取消')}
                </Button>
                <Button onClick={handleAddVersion} disabled={addLoading || !versionFormData.version.trim()}>
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

export default ProjectList
