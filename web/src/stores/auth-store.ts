import { create } from 'zustand'
import { getCookie, setCookie, removeCookie } from '@/lib/cookies'
import type { UserItem } from '@/features/auth/api/auth-api'

const ACCESS_TOKEN_KEY = 'hermes_access_token'
const USER_INFO_KEY = 'hermes_user_info'

interface AuthState {
  auth: {
    user: UserItem | null
    setUser: (user: UserItem | null) => void
    accessToken: string
    setAccessToken: (accessToken: string) => void
    resetAccessToken: () => void
    reset: () => void
  }
}

export const useAuthStore = create<AuthState>()((set) => {
  const savedToken = getCookie(ACCESS_TOKEN_KEY)
  const initToken = savedToken ? JSON.parse(savedToken) : ''

  const savedUser = getCookie(USER_INFO_KEY)
  const initUser: UserItem | null = savedUser ? JSON.parse(savedUser) : null

  return {
    auth: {
      user: initUser,
      setUser: (user) => {
        if (user) {
          setCookie(USER_INFO_KEY, JSON.stringify(user))
        } else {
          removeCookie(USER_INFO_KEY)
        }
        set((state) => ({ ...state, auth: { ...state.auth, user } }))
      },
      accessToken: initToken,
      setAccessToken: (accessToken) =>
        set((state) => {
          setCookie(ACCESS_TOKEN_KEY, JSON.stringify(accessToken))
          return { ...state, auth: { ...state.auth, accessToken } }
        }),
      resetAccessToken: () =>
        set((state) => {
          removeCookie(ACCESS_TOKEN_KEY)
          return { ...state, auth: { ...state.auth, accessToken: '' } }
        }),
      reset: () =>
        set((state) => {
          removeCookie(ACCESS_TOKEN_KEY)
          removeCookie(USER_INFO_KEY)
          return {
            ...state,
            auth: { ...state.auth, user: null, accessToken: '' },
          }
        }),
    },
  }
})
