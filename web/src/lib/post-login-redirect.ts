/**
 * Turn the post-login `redirect` search param into a same-app href.
 *
 * `location.href` from TanStack Router is pathname + search + hash (no origin).
 * Passing that string to `navigate({ to })` is wrong: `to` is a route path, so
 * `/?days=7` or `http://localhost:5173/` is resolved as a relative segment and
 * the user never leaves `/sign-in`.
 */
export function resolvePostLoginHref(redirectTo?: string | null): string {
  if (!redirectTo) return '/'

  let path = redirectTo.trim()
  if (!path) return '/'

  if (/^[a-z][a-z0-9+.-]*:/i.test(path)) {
    try {
      const url = new URL(path)
      path = `${url.pathname}${url.search}${url.hash}`
    } catch {
      return '/'
    }
  }

  if (!path.startsWith('/') || path.startsWith('//')) return '/'

  const pathname = path.split(/[?#]/, 1)[0] ?? '/'
  if (pathname === '/sign-in' || pathname.startsWith('/sign-in/')) return '/'

  return path
}
