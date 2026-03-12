import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Users as UsersIcon, Search, Plus, MoreHorizontal, Edit, Trash2, Shield, Loader2, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { userApi, type User } from '@/services/userApi'
import { authApi } from '@/services/authApi'

function Users() {
  const { t } = useTranslation()
  const [searchQuery, setSearchQuery] = useState('')
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showAddModal, setShowAddModal] = useState(false)
  const [addLoading, setAddLoading] = useState(false)
  const [addError, setAddError] = useState<string | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    confirmPassword: ''
  })
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        setLoading(true)
        setError(null)
        const response = await userApi.getUsers()
        setUsers(response.data.users)
      } catch (err) {
        setError(t('users.errorFetching', '获取用户列表失败'))
        console.error('Error fetching users:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchUsers()
  }, [t])

  const filteredUsers = users.filter(
    (user) =>
      user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.email.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: value
    }))

    const errors: Record<string, string> = { ...fieldErrors }
    delete errors[name]

    if (name === 'name') {
      if (!value) {
        errors.name = t('users.errorNameRequired', '请输入用户名')
      } else if (value.length < 2) {
        errors.name = t('users.errorNameTooShort', '用户名至少需要 2 个字符')
      }
    }

    if (name === 'email') {
      if (!value) {
        errors.email = t('users.errorEmailRequired', '请输入邮箱')
      } else {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
        if (!emailRegex.test(value)) {
          errors.email = t('users.errorInvalidEmail', '请输入有效的邮箱地址')
        }
      }
    }

    if (name === 'password') {
      if (!value) {
        errors.password = t('users.errorPasswordRequired', '请输入密码')
      } else if (value.length < 6) {
        errors.password = t('users.errorPasswordTooShort', '密码至少需要 6 位')
      }
    }

    if (name === 'confirmPassword') {
      if (!value) {
        errors.confirmPassword = t('users.errorConfirmPasswordRequired', '请确认密码')
      } else if (value !== formData.password) {
        errors.confirmPassword = t('users.errorPasswordMismatch', '密码和确认密码不一致')
      }
    }

    setFieldErrors(errors)
  }

  const handleAddUser = async () => {
    const errors: Record<string, string> = {}

    if (!formData.name) {
      errors.name = t('users.errorNameRequired', '请输入用户名')
    } else if (formData.name.length < 2) {
      errors.name = t('users.errorNameTooShort', '用户名至少需要 2 个字符')
    }

    if (!formData.email) {
      errors.email = t('users.errorEmailRequired', '请输入邮箱')
    } else {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      if (!emailRegex.test(formData.email)) {
        errors.email = t('users.errorInvalidEmail', '请输入有效的邮箱地址')
      }
    }

    if (!formData.password) {
      errors.password = t('users.errorPasswordRequired', '请输入密码')
    } else if (formData.password.length < 6) {
      errors.password = t('users.errorPasswordTooShort', '密码至少需要 6 位')
    }

    if (!formData.confirmPassword) {
      errors.confirmPassword = t('users.errorConfirmPasswordRequired', '请确认密码')
    } else if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = t('users.errorPasswordMismatch', '密码和确认密码不一致')
    }

    setFieldErrors(errors)

    if (Object.keys(errors).length > 0) {
      return
    }

    try {
      setAddLoading(true)
      setAddError(null)
      await authApi.register({
        name: formData.name,
        email: formData.email,
        password: formData.password
      })
      setShowAddModal(false)
      resetForm()
      // 重新获取用户列表
      const response = await userApi.getUsers()
      setUsers(response.data.users)
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : (err as { response?: { data?: { message?: string } } })?.response?.data?.message
      setAddError(errorMessage || t('users.errorAdding', '添加用户失败'))
      console.error('Error adding user:', err)
    } finally {
      setAddLoading(false)
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      email: '',
      password: '',
      confirmPassword: ''
    })
    setAddError(null)
    setFieldErrors({})
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-blue-100 dark:bg-blue-900/30 rounded-lg flex items-center justify-center">
            <UsersIcon className="w-5 h-5 text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">
              {t('users.title', '用户管理')}
            </h1>
            <p className="text-sm text-slate-500 dark:text-slate-400">
              {t('users.subtitle', '管理系统用户和权限')}
            </p>
          </div>
        </div>
        <Button className="gap-2" onClick={() => setShowAddModal(true)}>
          <Plus className="w-4 h-4" />
          {t('users.addUser', '添加用户')}
        </Button>
      </div>

      <div className="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700">
        <div className="p-4 border-b border-slate-200 dark:border-slate-700">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              type="text"
              placeholder={t('users.searchPlaceholder', '搜索用户名或邮箱...')}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 text-blue-500 animate-spin" />
          </div>
        ) : error ? (
          <div className="flex items-center justify-center py-12 text-red-500">
            <p>{error}</p>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-slate-200 dark:border-slate-700">
                    <th className="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {t('users.username', '用户名')}
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {t('users.email', '邮箱')}
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {t('users.role', '角色')}
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {t('users.status', '状态')}
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {t('users.createdAt', '创建时间')}
                    </th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {t('users.actions', '操作')}
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200 dark:divide-slate-700">
                  {filteredUsers.map((user) => (
                    <tr key={user.id} className="hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors">
                      <td className="px-4 py-3">
                        <span className="font-medium text-slate-900 dark:text-slate-100">{user.name}</span>
                      </td>
                      <td className="px-4 py-3 text-slate-600 dark:text-slate-300">{user.email}</td>
                      <td className="px-4 py-3">
                        <span
                          className={cn(
                            'inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium',
                            user.roles?.includes('管理员')
                              ? 'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300'
                              : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300'
                          )}
                        >
                          {user.roles?.includes('管理员') && <Shield className="w-3 h-3" />}
                          {user.roles?.join(', ') || t('users.noRole', '无角色')}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={cn(
                            'inline-flex px-2 py-1 rounded-full text-xs font-medium',
                            'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300'
                          )}
                        >
                          {t('users.statusActive', '活跃')}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-slate-600 dark:text-slate-300">
                        {new Date(user.created_at).toLocaleString()}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center justify-end gap-1">
                          <button
                            className="p-2 rounded-lg text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
                            title={t('users.edit', '编辑')}
                          >
                            <Edit className="w-4 h-4" />
                          </button>
                          <button
                            className="p-2 rounded-lg text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
                            title={t('users.more', '更多')}
                          >
                            <MoreHorizontal className="w-4 h-4" />
                          </button>
                          <button
                            className="p-2 rounded-lg text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                            title={t('users.delete', '删除')}
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {filteredUsers.length === 0 && (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <UsersIcon className="w-12 h-12 text-slate-300 dark:text-slate-600 mb-4" />
                <p className="text-slate-500 dark:text-slate-400">
                  {t('users.noResults', '没有找到匹配的用户')}
                </p>
              </div>
            )}
          </>
        )}
      </div>

      {/* 添加用户模态框 */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-slate-800 rounded-xl shadow-lg w-full max-w-md">
            <div className="flex items-center justify-between p-6 border-b border-slate-200 dark:border-slate-700">
              <h2 className="text-xl font-bold text-slate-900 dark:text-slate-100">
                {t('users.addUser', '添加用户')}
              </h2>
              <button
                className="p-2 rounded-full text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors"
                onClick={() => {
                  setShowAddModal(false)
                  resetForm()
                }}
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="p-6 space-y-4">
              {addError && (
                <div className="p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg">
                  {addError}
                </div>
              )}
              <div className="space-y-2">
                <Label htmlFor="name">{t('users.name', '用户名')}</Label>
                <Input
                  id="name"
                  name="name"
                  value={formData.name}
                  onChange={handleInputChange}
                  placeholder={t('users.namePlaceholder', '请输入用户名')}
                />
                {fieldErrors.name && (
                  <p className="text-sm text-red-500 dark:text-red-400">{fieldErrors.name}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">{t('users.email', '邮箱')}</Label>
                <Input
                  id="email"
                  name="email"
                  type="email"
                  value={formData.email}
                  onChange={handleInputChange}
                  placeholder={t('users.emailPlaceholder', '请输入邮箱')}
                />
                {fieldErrors.email && (
                  <p className="text-sm text-red-500 dark:text-red-400">{fieldErrors.email}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">{t('users.password', '密码')}</Label>
                <Input
                  id="password"
                  name="password"
                  type="password"
                  value={formData.password}
                  onChange={handleInputChange}
                  placeholder={t('users.passwordPlaceholder', '请输入密码')}
                />
                {fieldErrors.password && (
                  <p className="text-sm text-red-500 dark:text-red-400">{fieldErrors.password}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirmPassword">{t('users.confirmPassword', '确认密码')}</Label>
                <Input
                  id="confirmPassword"
                  name="confirmPassword"
                  type="password"
                  value={formData.confirmPassword}
                  onChange={handleInputChange}
                  placeholder={t('users.confirmPasswordPlaceholder', '请确认密码')}
                />
                {fieldErrors.confirmPassword && (
                  <p className="text-sm text-red-500 dark:text-red-400">{fieldErrors.confirmPassword}</p>
                )}
              </div>
            </div>
            <div className="flex items-center justify-end gap-3 p-6 border-t border-slate-200 dark:border-slate-700">
              <Button
                type="button"
                variant="secondary"
                onClick={() => {
                  setShowAddModal(false)
                  resetForm()
                }}
              >
                {t('users.cancel', '取消')}
              </Button>
              <Button
                type="button"
                onClick={handleAddUser}
                disabled={addLoading}
              >
                {addLoading ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  t('users.save', '保存')
                )}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Users
