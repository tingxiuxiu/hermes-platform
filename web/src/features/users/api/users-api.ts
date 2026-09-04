import { api } from '@/lib/api'
import { API_V1 } from '@/lib/api-prefix'
import {
  type PermissionItem,
  type RoleItem,
  type User,
  type UserStatus,
} from '../data/schema'

export type BackendRole = RoleItem
export type BackendPermission = PermissionItem

type UserListResponse = {
  data: { total: number; page: number; page_size: number; items: User[] }
}
type RoleWithPermissions = BackendRole & { permissions: BackendPermission[] }
type RoleListResponse = { data: RoleWithPermissions[] }
type RoleOptionsResponse = { data: BackendRole[] }
type PermissionTreeResponse = { data: BackendPermission[] }

export type UsersQuery = {
  page: number
  pageSize: number
  username?: string
  email?: string
  status?: UserStatus
  roleId?: number
}

export type UsersPage = {
  items: User[]
  total: number
  page: number
  pageSize: number
}

export async function getUsers(query: UsersQuery): Promise<UsersPage> {
  const { data } = await api.get<UserListResponse>(`${API_V1}/users`, {
    params: {
      page: query.page,
      page_size: query.pageSize,
      username: query.username || undefined,
      email: query.email || undefined,
      status: query.status,
      role_id: query.roleId,
    },
  })

  return {
    items: data.data.items,
    total: data.data.total,
    page: data.data.page,
    pageSize: data.data.page_size,
  }
}

export async function getRoleOptions(): Promise<BackendRole[]> {
  const { data } = await api.get<RoleOptionsResponse>(
    `${API_V1}/users/role-options`
  )
  return data.data
}

export async function getRoles(): Promise<RoleWithPermissions[]> {
  const { data } = await api.get<RoleListResponse>(`${API_V1}/roles`)
  return data.data
}

export async function getPermissionTree(): Promise<BackendPermission[]> {
  const { data } = await api.get<PermissionTreeResponse>(
    `${API_V1}/permissions/tree`
  )
  return data.data
}

export async function updateUserStatus(userId: number, status: UserStatus) {
  await api.put(`${API_V1}/users/${userId}/status`, {
    user_id: userId,
    status,
  })
}

export async function updateUserRoles(userId: number, roleIds: number[]) {
  await api.put(`${API_V1}/users/${userId}/roles`, {
    user_id: userId,
    role_ids: roleIds,
  })
}

export async function deleteUsers(userIds: number[]) {
  await api.post(`${API_V1}/users/batch-delete`, { user_ids: userIds })
}

export type CreateUserPayload = {
  username: string
  password: string
  email?: string
  role_ids: number[]
}

export type UpdateUserPayload = {
  username?: string
  email?: string
}

export async function createUser(payload: CreateUserPayload) {
  await api.post(`${API_V1}/users`, payload)
}

export async function updateUser(userId: number, payload: UpdateUserPayload) {
  await api.put(`${API_V1}/users/${userId}`, { user_id: userId, ...payload })
}

export async function updateRolePermissions(
  roleId: number,
  permissionIds: number[]
) {
  await api.put(`${API_V1}/roles/${roleId}/permissions`, {
    role_id: roleId,
    permission_ids: permissionIds,
  })
}
