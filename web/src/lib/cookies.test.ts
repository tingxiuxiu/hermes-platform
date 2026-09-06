import { clearCookies } from '@/test-utils/cookies'
import { beforeEach, describe, expect, it } from 'vitest'
import { getCookie, removeCookie, setCookie } from './cookies'

const COOKIE_PREFIX = 'test_cookie_'

describe('cookies', () => {
  const uniqueName = () =>
    `${COOKIE_PREFIX}${Math.random().toString(36).slice(2)}`

  beforeEach(() => {
    clearCookies(COOKIE_PREFIX)
  })

  it('stores a value that can be read back', () => {
    const name = uniqueName()
    const value = 'hello-world'

    setCookie(name, value)

    expect(getCookie(name)).toBe(value)
  })

  it('round-trips JSON that contains commas, quotes, and non-ascii text', () => {
    const name = uniqueName()
    const value = JSON.stringify({
      username: 'admin',
      role: '管理员',
    })

    setCookie(name, value)

    expect(getCookie(name)).toBe(value)
  })

  it('clears a value so it is no longer readable', () => {
    const name = uniqueName()

    setCookie(name, 'x')
    expect(getCookie(name)).toBe('x')

    removeCookie(name)

    expect(getCookie(name)).toBeUndefined()
  })
})
