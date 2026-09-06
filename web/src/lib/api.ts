import axios from 'axios'
import { useAuthStore } from '@/stores/auth-store'

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api'

export const api = axios.create({
  baseURL: BASE_URL,
  timeout: 30_000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor — attach Bearer token from store
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().auth.accessToken
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// Response interceptor — surface error messages from our backend
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error)) {
      // Backend returns { detail: "..." } for HTTP errors
      const detail = error.response?.data?.detail ?? error.response?.data?.message
      if (typeof detail === 'string') {
        error.message = detail
      }
    }
    return Promise.reject(error)
  },
)
