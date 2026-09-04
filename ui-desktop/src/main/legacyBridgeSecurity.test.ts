import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'

const source = (relativePath: string): string =>
  readFileSync(path.join(process.cwd(), relativePath), 'utf8')

describe('legacy renderer bridge security boundaries', () => {
  it('does not auto-open DevTools and disables them in packaged builds', () => {
    const main = source('src/main/index.ts')

    expect(main).not.toContain('.openDevTools(')
    expect(main).toContain('devTools: is.dev')
  })
})
