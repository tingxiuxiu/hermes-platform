import React, { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import useDialogState from '@/hooks/use-dialog-state'
import {
  getPermissionTree,
  getRoleOptions,
  getRoles,
  updateRolePermissions,
  updateUserRoles,
  updateUserStatus as updateUserStatusApi,
} from '../api/users-api'
import { type User, type UserStatus } from '../data/schema'

type UsersDialogType = 'invite' | 'add' | 'edit' | 'delete' | 'permissions'

type UsersContextType = {
  open: UsersDialogType | null
  setOpen: (str: UsersDialogType | null) => void
  currentRow: User | null
  setCurrentRow: React.Dispatch<React.SetStateAction<User | null>>
  roleOptions: Awaited<ReturnType<typeof getRoleOptions>>
  roles: Awaited<ReturnType<typeof getRoles>>
  permissionTree: Awaited<ReturnType<typeof getPermissionTree>>
  updateUserStatus: (id: number, status: UserStatus) => Promise<void>
  updateUserRoles: (id: number, roleIds: number[]) => Promise<void>
  updateRolePermissions: (roleId: number, permissionIds: number[]) => Promise<void>
}

const UsersContext = React.createContext<UsersContextType | null>(null)

export function UsersProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useDialogState<UsersDialogType>(null)
  const [currentRow, setCurrentRow] = useState<User | null>(null)
  const queryClient = useQueryClient()
  const roleOptionsQuery = useQuery({ queryKey: ['user-role-options'], queryFn: getRoleOptions })
  const rolesQuery = useQuery({ queryKey: ['roles'], queryFn: getRoles })
  const permissionsQuery = useQuery({
    queryKey: ['permissions', 'tree'],
    queryFn: getPermissionTree,
  })

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: UserStatus }) =>
      updateUserStatusApi(id, status),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] }),
  })

  const roleMutation = useMutation({
    mutationFn: ({ id, roleIds }: { id: number; roleIds: number[] }) =>
      updateUserRoles(id, roleIds),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] }),
  })

  const permissionMutation = useMutation({
    mutationFn: ({ roleId, permissionIds }: { roleId: number; permissionIds: number[] }) =>
      updateRolePermissions(roleId, permissionIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  const updateUserStatus = (id: number, status: UserStatus) =>
    statusMutation.mutateAsync({ id, status })
  const updateUserRolesForUser = (id: number, roleIds: number[]) =>
    roleMutation.mutateAsync({ id, roleIds })
  const updateRolePermissionsForRole = (roleId: number, permissionIds: number[]) =>
    permissionMutation.mutateAsync({ roleId, permissionIds })

  return (
    <UsersContext
      value={{
        open,
        setOpen,
        currentRow,
        setCurrentRow,
        roleOptions: roleOptionsQuery.data ?? [],
        roles: rolesQuery.data ?? [],
        permissionTree: permissionsQuery.data ?? [],
        updateUserStatus,
        updateUserRoles: updateUserRolesForUser,
        updateRolePermissions: updateRolePermissionsForRole,
      }}
    >
      {children}
    </UsersContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useUsers = () => {
  const usersContext = React.useContext(UsersContext)

  if (!usersContext) {
    throw new Error('useUsers has to be used within <UsersContext>')
  }

  return usersContext
}
