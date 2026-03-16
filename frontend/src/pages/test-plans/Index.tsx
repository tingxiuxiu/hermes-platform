import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router'
import { Folder, ChevronRight, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { projectApi, type Project, type ProjectVersion } from '@/services/projectApi'

function TestPlansIndex() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedProject, setSelectedProject] = useState<Project | null>(null)
  const [versions, setVersions] = useState<ProjectVersion[]>([])
  const [versionsLoading, setVersionsLoading] = useState(false)

  useEffect(() => {
    fetchProjects()
  }, [])

  const fetchProjects = async () => {
    try {
      setLoading(true)
      const response = await projectApi.getProjects()
      setProjects(response.data.projects || [])
    } catch (err) {
      console.error('Error fetching projects:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchVersions = async (projectId: number) => {
    try {
      setVersionsLoading(true)
      const response = await projectApi.getVersions(projectId)
      setVersions(response.data?.versions || [])
    } catch (err) {
      console.error('Error fetching versions:', err)
    } finally {
      setVersionsLoading(false)
    }
  }

  const handleProjectSelect = (project: Project) => {
    setSelectedProject(project)
    setVersions([])
    fetchVersions(project.id)
  }

  const handleVersionSelect = (version: ProjectVersion) => {
    if (selectedProject) {
      navigate(`/projects/${selectedProject.id}/versions/${version.id}/test-plans`)
    }
  }

  const handleBack = () => {
    setSelectedProject(null)
    setVersions([])
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6">
      <div className="flex items-center gap-2 mb-6">
        <Folder className="w-6 h-6" />
        <h1 className="text-2xl font-bold">{t('sidebar.testPlans', '测试计划')}</h1>
      </div>

      {!selectedProject ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {projects.map((project) => (
            <Card 
              key={project.id} 
              className="cursor-pointer hover:shadow-md transition-shadow"
              onClick={() => handleProjectSelect(project)}
            >
              <CardHeader className="pb-2">
                <CardTitle className="text-lg">{project.name}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground line-clamp-2">
                  {project.description || t('project.noDescription', '暂无描述')}
                </p>
              </CardContent>
            </Card>
          ))}
          {projects.length === 0 && (
            <p className="text-muted-foreground col-span-full text-center py-8">
              {t('project.noProjects', '暂无项目')}
            </p>
          )}
        </div>
      ) : (
        <div>
          <div className="flex items-center gap-2 mb-4">
            <Button variant="ghost" size="sm" onClick={handleBack}>
              {t('common.back', '返回')}
            </Button>
            <ChevronRight className="w-4 h-4 text-muted-foreground" />
            <span className="font-medium">{selectedProject.name}</span>
          </div>

          <h2 className="text-lg font-semibold mb-4">
            {t('project.selectVersion', '选择版本')}
          </h2>

          {versionsLoading ? (
            <div className="flex items-center justify-center p-8">
              <Loader2 className="w-8 h-8 animate-spin text-primary" />
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {versions.map((version) => (
                <Card 
                  key={version.id} 
                  className="cursor-pointer hover:shadow-md transition-shadow"
                  onClick={() => handleVersionSelect(version)}
                >
                  <CardHeader className="pb-2">
                    <CardTitle className="text-lg">{version.version}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {version.description || t('project.noDescription', '暂无描述')}
                    </p>
                    <div className="mt-2">
                      <span className={`text-xs px-2 py-1 rounded ${
                        version.status === 'active' 
                          ? 'bg-green-100 text-green-800' 
                          : 'bg-gray-100 text-gray-800'
                      }`}>
                        {version.status}
                      </span>
                    </div>
                  </CardContent>
                </Card>
              ))}
              {versions.length === 0 && (
                <p className="text-muted-foreground col-span-full text-center py-8">
                  {t('project.noVersions', '暂无版本')}
                </p>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default TestPlansIndex
