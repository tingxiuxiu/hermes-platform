import { api } from '@/lib/api'

// ─── Request Shapes ───────────────────────────────────────────────────────────

export interface LoginRequest {
  username: string
  password: string
}

export interface SignUpRequest {
  username: string
  email: string
  password: string
}

// ─── Response Shapes (mirrors backend UserLoginResponse) ──────────────────────

export interface RoleItem {
  id: number
  code: string
  name: string
  description: string | null
}

export interface UserItem {
  id: number
  username: string
  email: string | null
  status: number
  roles: RoleItem[]
  last_login_at: string | null
  created_at: string | null
}

export interface LoginData {
  access_token: string
  token_type: string
  expires_in: number | null
  user: UserItem | null
}

export interface LoginResponse {
  code: number
  message: string
  success: boolean
  data: LoginData
}

// ─── API Functions ─────────────────────────────────────────────────────────────

const API_V1 = '/hermes-platform/api/v1'

/**
 * JSON-body login — calls POST /hermes-platform/api/v1/login
 */
export async function loginApi(payload: LoginRequest): Promise<LoginResponse> {
  const { data } = await api.post<LoginResponse>(`${API_V1}/login`, payload)
  return data
}

/**
 * Logout — calls POST /hermes-platform/api/v1/logout
 * Requires the token to be present in the axios interceptor.
 */
export async function logoutApi(): Promise<void> {
  await api.post(`${API_V1}/logout`)
}

export async function signUpApi(payload: SignUpRequest): Promise<LoginResponse> {
  const { data } = await api.post<LoginResponse>(`${API_V1}/register`, payload)
  return data
}
