export type UserStatus = 0 | 1 | 2

export type RoleItem = {
  id: number
  code: string
  name: string
  description: string | null
}

export type PermissionItem = {
  id: number
  parent_id: number | null
  code: string
  name: string
  resource_type: number
  path: string | null
  method: string | null
  created_at?: string | null
  children?: PermissionItem[]
}

export type User = {
  id: number
  username: string
  email: string | null
  status: UserStatus
  roles: RoleItem[]
  last_login_at: string | null
  created_at: string | null
}
