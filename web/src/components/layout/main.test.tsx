import { describe, expect, it } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'
import { Main } from './main'

const CENTERED_MAX_WIDTH = '@7xl/content:max-w-7xl'

describe('Main', () => {
  it('centers content when fluid is omitted, including <Main fixed>', () => {
    expect(renderToStaticMarkup(<Main>unset</Main>)).toContain(CENTERED_MAX_WIDTH)
    expect(renderToStaticMarkup(<Main fixed>fixed</Main>)).toContain(
      CENTERED_MAX_WIDTH
    )
  })

  it('does not center content when fluid is true', () => {
    expect(renderToStaticMarkup(<Main fluid>fluid</Main>)).not.toContain(
      CENTERED_MAX_WIDTH
    )
  })
})
