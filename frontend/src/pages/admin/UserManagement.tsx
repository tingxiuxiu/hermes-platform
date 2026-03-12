import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { 
  Search, 
  RefreshCw, 
  Key, 
  Shield, 
  ToggleLeft, 
  ToggleRight,
  Trash2,
  Users
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Pagination } from '@/components/ui/pagination'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { userApi, type User, type UserQueryParams } from '@/services/userApi'
import { roleApi } from '@/services/roleApi'

const PAGE_SIZE = 10

const STATUS_CONFIG = {
  active: {
    bg: 'bg-green-100 dark:bg-green-900/30',
    text: 'text-green-700 dark:text-green-300',
    label: 'userManagement.statusActive',
  },
  inactive: {
    bg: 'bg-slate-100 dark:bg-slate-700',
    text: 'text-slate-700 dark:text-slate-300',
    label: 'userManagement.statusInactive',
  },
}

function UserTableSkeleton() {
  return (
    <>
      {Array.from({ length: 5 }).map((_, index) => (
        <TableRow key={index}>
          <TableCell>
            <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-24 animate-pulse" />
          </TableCell>
          <TableCell>
            <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-32 animate-pulse" />
          </TableCell>
          <TableCell>
            <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-16 animate-pulse" />
          </TableCell>
          <TableCell>
            <div className="h-6 bg-slate-200 dark:bg-slate-700 rounded-full w-12 animate-pulse" />
          </TableCell>
          <TableCell>
            <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-20 animate-pulse" />
          </TableCell>
          <TableCell>
            <div className="flex gap-2">
              <div className="h-8 w-8 bg-slate-200 dark:bg-slate-700 rounded animate-pulse" />
              <div className="h-8 w-8 bg-slate-200 dark:bg-slate-700 rounded animate-pulse" />
              <div className="h-8 w-8 bg-slate-200 dark:bg-slate-700 rounded animate-pulse" />
            </div>
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}

function ResetPasswordDialog({
  open,
  onOpenChange,
  userId,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  userId: number | null
  onSuccess: () => void
}) {
  const { t } = useTranslation()
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const queryClient = useQueryClient()

  const resetPasswordMutation = useMutation({
    mutationFn: (id: number) => userApi.resetPassword(id, newPassword),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      onSuccess()
      setNewPassword('')
      setConfirmPassword('')
      setError('')
    },
    onError: () => {
      setError(t('userManagement.resetPasswordFailed'))
    },
  })

  const handleSubmit = () => {
    if (!newPassword) {
      setError(t('userManagement.passwordRequired'))
      return
    }
    if (newPassword.length < 6) {
      setError(t('userManagement.passwordMinLength'))
      return
    }
    if (newPassword !== confirmPassword) {
      setError(t('userManagement.passwordMismatch'))
      return
    }
    if (userId) {
      resetPasswordMutation.mutate(userId)
    }
  }

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setNewPassword('')
      setConfirmPassword('')
      setError('')
    }
    onOpenChange(open)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('userManagement.resetPasswordTitle')}</DialogTitle>
          <DialogDescription>
            {t('userManagement.resetPasswordDescription')}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="newPassword">{t('userManagement.newPassword')}</Label>
            <Input
              id="newPassword"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder={t('userManagement.passwordPlaceholder')}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirmPassword">{t('userManagement.confirmPassword')}</Label>
            <Input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder={t('userManagement.confirmPasswordPlaceholder')}
            />
          </div>
          {error && (
            <p className="text-sm text-red-500 dark:text-red-400">{error}</p>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)}>
            {t('home.cancel')}
          </Button>
          <Button 
            onClick={handleSubmit} 
            disabled={resetPasswordMutation.isPending}
          >
            {resetPasswordMutation.isPending 
              ? t('userManagement.resetting') 
              : t('userManagement.resetPassword')
            }
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function AssignRolesDialog({
  open,
  onOpenChange,
  userId,
  currentRoles,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  userId: number | null
  currentRoles: string[]
  onSuccess: () => void
}) {
  const { t } = useTranslation()
  const [selectedRoleIds, setSelectedRoleIds] = useState<number[]>([])
  const queryClient = useQueryClient()

  const { data: rolesData } = useQuery({
    queryKey: ['roles'],
    queryFn: () => roleApi.getRoles(),
  })

  const roles = rolesData?.data?.roles ?? []

  const assignRolesMutation = useMutation({
    mutationFn: (id: number) => userApi.assignRoles(id, selectedRoleIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      onSuccess()
    },
  })

  const handleOpenChange = (open: boolean) => {
    if (open && roles.length > 0) {
      const currentRoleIds = roles
        .filter((role) => currentRoles.includes(role.name))
        .map((role) => role.id)
      setSelectedRoleIds(currentRoleIds)
    }
    onOpenChange(open)
  }

  const toggleRole = (roleId: number) => {
    setSelectedRoleIds((prev) =>
      prev.includes(roleId)
        ? prev.filter((id) => id !== roleId)
        : [...prev, roleId]
    )
  }

  const handleSubmit = () => {
    if (userId) {
      assignRolesMutation.mutate(userId)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('userManagement.assignRolesTitle')}</DialogTitle>
          <DialogDescription>
            {t('userManagement.assignRolesDescription')}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3 max-h-60 overflow-y-auto">
          {roles.map((role) => (
            <label
              key={role.id}
              className="flex items-center gap-3 p-3 rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 cursor-pointer transition-colors"
            >
              <input
                type="checkbox"
                checked={selectedRoleIds.includes(role.id)}
                onChange={() => toggleRole(role.id)}
                className="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
              />
              <div className="flex-1">
                <div className="font-medium text-slate-900 dark:text-slate-100">
                  {role.name}
                </div>
                {role.description && (
                  <div className="text-sm text-slate-500 dark:text-slate-400">
                    {role.description}
                  </div>
                )}
              </div>
            </label>
          ))}
          {roles.length === 0 && (
            <div className="text-center text-slate-500 dark:text-slate-400 py-4">
              {t('userManagement.noRoles')}
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('home.cancel')}
          </Button>
          <Button 
            onClick={handleSubmit} 
            disabled={assignRolesMutation.isPending}
          >
            {assignRolesMutation.isPending 
              ? t('userManagement.assigning') 
              : t('userManagement.assignRoles')
            }
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function DeleteUserDialog({
  open,
  onOpenChange,
  userId,
  userName,
  onSuccess,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  userId: number | null
  userName: string
  onSuccess: () => void
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const deleteUserMutation = useMutation({
    mutationFn: (id: number) => userApi.deleteUser(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      onSuccess()
    },
  })

  const handleSubmit = () => {
    if (userId) {
      deleteUserMutation.mutate(userId)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('userManagement.deleteUserTitle')}</DialogTitle>
          <DialogDescription>
            {t('userManagement.deleteUserDescription', { name: userName })}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('home.cancel')}
          </Button>
          <Button 
            variant="destructive"
            onClick={handleSubmit} 
            disabled={deleteUserMutation.isPending}
          >
            {deleteUserMutation.isPending 
              ? t('userManagement.deleting') 
              : t('userManagement.deleteUser')
            }
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function UserManagement() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [currentPage, setCurrentPage] = useState(1)
  const [searchQuery, setSearchQuery] = useState('')
  const [searchInput, setSearchInput] = useState('')
  
  const [resetPasswordOpen, setResetPasswordOpen] = useState(false)
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null)
  const [assignRolesOpen, setAssignRolesOpen] = useState(false)
  const [selectedUserRoles, setSelectedUserRoles] = useState<string[]>([])
  const [deleteUserOpen, setDeleteUserOpen] = useState(false)
  const [selectedUserName, setSelectedUserName] = useState('')

  const queryParams: UserQueryParams = useMemo(() => ({
    page: currentPage,
    page_size: PAGE_SIZE,
    name: searchQuery || undefined,
    email: searchQuery || undefined,
  }), [currentPage, searchQuery])

  const { data, isLoading, isFetching, refetch } = useQuery({
    queryKey: ['users', queryParams],
    queryFn: () => userApi.getUsers(queryParams),
  })

  const users = data?.data?.users ?? []
  const total = data?.data?.total ?? 0
  const totalPages = Math.ceil(total / PAGE_SIZE)

  const toggleStatusMutation = useMutation({
    mutationFn: (user: User) => 
      userApi.updateUser(user.id, { 
        status: user.status === 'active' ? 'inactive' : 'active' 
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  const handleSearch = () => {
    setSearchQuery(searchInput)
    setCurrentPage(1)
  }

  const handlePageChange = (page: number) => {
    setCurrentPage(page)
  }

  const handleRefresh = () => {
    refetch()
  }

  const handleToggleStatus = (user: User) => {
    toggleStatusMutation.mutate(user)
  }

  const handleResetPassword = (userId: number) => {
    setSelectedUserId(userId)
    setResetPasswordOpen(true)
  }

  const handleAssignRoles = (userId: number, roles: string[]) => {
    setSelectedUserId(userId)
    setSelectedUserRoles(roles)
    setAssignRolesOpen(true)
  }

  const handleDeleteUser = (userId: number, userName: string) => {
    setSelectedUserId(userId)
    setSelectedUserName(userName)
    setDeleteUserOpen(true)
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
  }

  const isRefreshing = isFetching && !isLoading

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-3">
          <Users className="w-6 h-6 text-blue-600 dark:text-blue-400" />
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">
            {t('userManagement.title')}
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

      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <Input
            className="pl-10"
            placeholder={t('userManagement.searchPlaceholder')}
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          />
        </div>
        <Button onClick={handleSearch}>
          {t('userManagement.search')}
        </Button>
      </div>

      <div className="bg-white dark:bg-slate-800 rounded-xl shadow-md border border-slate-200 dark:border-slate-700 overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('userManagement.username')}</TableHead>
              <TableHead>{t('userManagement.email')}</TableHead>
              <TableHead>{t('userManagement.roles')}</TableHead>
              <TableHead>{t('userManagement.status')}</TableHead>
              <TableHead>{t('userManagement.createdAt')}</TableHead>
              <TableHead className="text-right">{t('userManagement.actions')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <UserTableSkeleton />
            ) : users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8">
                  <div className="flex flex-col items-center gap-2">
                    <Users className="w-12 h-12 text-slate-300 dark:text-slate-600" />
                    <p className="text-slate-500 dark:text-slate-400">
                      {t('userManagement.noUsers')}
                    </p>
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              users.map((user) => {
                const statusConfig = STATUS_CONFIG[user.status]
                return (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">{user.name}</TableCell>
                    <TableCell className="text-slate-600 dark:text-slate-400">
                      {user.email}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {user.roles.map((role) => (
                          <Badge key={role} variant="secondary" className="text-xs">
                            {role}
                          </Badge>
                        ))}
                        {user.roles.length === 0 && (
                          <span className="text-slate-400 dark:text-slate-500 text-sm">
                            -
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge 
                        variant={user.status === 'active' ? 'default' : 'outline'}
                        className={cn(
                          user.status === 'active' 
                            ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300' 
                            : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300'
                        )}
                      >
                        {t(statusConfig.label)}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-slate-600 dark:text-slate-400 text-sm">
                      {formatDate(user.created_at)}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => handleToggleStatus(user)}
                          title={user.status === 'active' 
                            ? t('userManagement.deactivate') 
                            : t('userManagement.activate')
                          }
                        >
                          {user.status === 'active' ? (
                            <ToggleRight className="h-4 w-4 text-green-600 dark:text-green-400" />
                          ) : (
                            <ToggleLeft className="h-4 w-4 text-slate-400" />
                          )}
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => handleResetPassword(user.id)}
                          title={t('userManagement.resetPassword')}
                        >
                          <Key className="h-4 w-4 text-slate-600 dark:text-slate-400" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => handleAssignRoles(user.id, user.roles)}
                          title={t('userManagement.assignRoles')}
                        >
                          <Shield className="h-4 w-4 text-slate-600 dark:text-slate-400" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => handleDeleteUser(user.id, user.name)}
                          title={t('userManagement.deleteUser')}
                        >
                          <Trash2 className="h-4 w-4 text-red-500 dark:text-red-400" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </div>

      {totalPages > 1 && (
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={handlePageChange}
          totalItems={total}
        />
      )}

      <ResetPasswordDialog
        open={resetPasswordOpen}
        onOpenChange={setResetPasswordOpen}
        userId={selectedUserId}
        onSuccess={() => setResetPasswordOpen(false)}
      />

      <AssignRolesDialog
        open={assignRolesOpen}
        onOpenChange={setAssignRolesOpen}
        userId={selectedUserId}
        currentRoles={selectedUserRoles}
        onSuccess={() => setAssignRolesOpen(false)}
      />

      <DeleteUserDialog
        open={deleteUserOpen}
        onOpenChange={setDeleteUserOpen}
        userId={selectedUserId}
        userName={selectedUserName}
        onSuccess={() => setDeleteUserOpen(false)}
      />
    </div>
  )
}

export default UserManagement
