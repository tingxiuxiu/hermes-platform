import { describe, expect, it } from 'vitest'
import { resolvePostLoginHref } from './post-login-redirect'

describe('resolvePostLoginHref', () => {
  it('defaults to home when redirect is missing', () => {
    expect(resolvePostLoginHref()).toBe('/')
    expect(resolvePostLoginHref('')).toBe('/')
    expect(resolvePostLoginHref('   ')).toBe('/')
  })

  it('keeps an in-app path including search', () => {
    expect(resolvePostLoginHref('/')).toBe('/')
    expect(resolvePostLoginHref('/?days=7')).toBe('/?days=7')
    expect(resolvePostLoginHref('/automation/executions/abc')).toBe(
      '/automation/executions/abc'
    )
  })

  it('strips origin from an absolute URL so navigate does not treat it as a relative path', () => {
    expect(resolvePostLoginHref('http://localhost:5173/')).toBe('/')
    expect(
      resolvePostLoginHref('http://localhost:5173/automation/jobs?x=1')
    ).toBe('/automation/jobs?x=1')
  })

  it('rejects protocol-relative and sign-in loops', () => {
    expect(resolvePostLoginHref('//evil.example/phish')).toBe('/')
    expect(resolvePostLoginHref('/sign-in')).toBe('/')
    expect(resolvePostLoginHref('/sign-in?redirect=/')).toBe('/')
  })
})
