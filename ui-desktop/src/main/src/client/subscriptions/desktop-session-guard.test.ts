import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'

const handlers = readFileSync(
  path.join(process.cwd(), 'src/main/src/client/subscriptions/handlers.ts'),
  'utf8'
)

describe('desktop session-open proxy contract', () => {
  it('opts into duplicate protection and requests only active bids', () => {
    const openStart = handlers.indexOf('export const openSession')
    const openEnd = handlers.indexOf('export const getSessionsByProvider', openStart)
    const openHandler = handlers.slice(openStart, openEnd)
    const bidsStart = handlers.indexOf('export const getBidsByModel')
    const bidsEnd = handlers.indexOf('export const getBidInfo', bidsStart)
    const bidsHandler = handlers.slice(bidsStart, bidsEnd)

    expect(openHandler).toContain('rejectExisting: true')
    expect(openHandler).toContain('parseExistingSessionConflict(error)')
    expect(bidsHandler).toContain('/bids/active?')
    expect(bidsHandler).not.toMatch(/\/bids\?offset=/u)
  })
})
