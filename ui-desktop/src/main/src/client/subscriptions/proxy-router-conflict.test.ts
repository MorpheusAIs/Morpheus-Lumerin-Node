import { describe, expect, it } from 'vitest'
import { parseExistingSessionConflict } from './proxy-router-conflict'

const sessionID = `0x${'ab'.repeat(32)}`

describe('parseExistingSessionConflict', () => {
  it('returns a validated existing session from a proxy conflict', () => {
    expect(
      parseExistingSessionConflict({
        status: 409,
        responseBody: { existingSessionID: sessionID }
      })
    ).toEqual({ existingSessionID: sessionID })
  })

  it('does not swallow ordinary errors or malformed conflict bodies', () => {
    expect(
      parseExistingSessionConflict({
        status: 500,
        responseBody: { existingSessionID: sessionID }
      })
    ).toBeNull()
    expect(
      parseExistingSessionConflict({ status: 409, responseBody: { existingSessionID: 'bad-id' } })
    ).toBeNull()
    expect(parseExistingSessionConflict(new Error('network failed'))).toBeNull()
  })
})
