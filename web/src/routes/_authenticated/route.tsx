import { useEffect } from 'react'
import {
  createFileRoute,
  useLocation,
  useNavigate,
} from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'
import { resolvePostLoginHref } from '@/lib/post-login-redirect'

function AuthenticatedRoute() {
  const navigate = useNavigate()
  const location = useLocation()
  const { auth } = useAuthStore()

  useEffect(() => {
    if (!auth.accessToken || !auth.user) {
      navigate({
        to: '/sign-in',
        search: { redirect: resolvePostLoginHref(location.href) },
        replace: true,
      })
    }
  }, [auth.accessToken, auth.user, location.href, navigate])

  if (!auth.accessToken || !auth.user) {
    return null
  }

  return <AuthenticatedLayout />
}

export const Route = createFileRoute('/_authenticated')({
  component: AuthenticatedRoute,
})
