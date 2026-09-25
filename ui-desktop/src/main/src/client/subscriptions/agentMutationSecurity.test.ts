import { describe, expect, it } from 'vitest'
import {
  validateAgentDecision,
  validateAgentToken,
  validateAgentUsername
} from './agentMutationSecurity'

describe('agent mutation IPC validation', () => {
  it('accepts bounded agent identifiers without changing their identity', () => {
    expect(validateAgentUsername('research-agent_1@example.com')).toBe(
      'research-agent_1@example.com'
    )
    expect(validateAgentUsername('a'.repeat(128))).toBe('a'.repeat(128))
    expect(validateAgentToken('eip155:8453')).toBe('eip155:8453')
    expect(validateAgentToken(`0x${'a'.repeat(40)}`)).toBe(`0x${'a'.repeat(40)}`)
  })

  it.each([
    undefined,
    null,
    42,
    '',
    ' agent',
    'agent ',
    'agent:name',
    'agent\nname',
    'admin',
    'ADMIN',
    'a'.repeat(129)
  ])('rejects unsafe agent usernames (%j)', (value) => {
    expect(() => validateAgentUsername(value)).toThrow('Agent username is invalid.')
  })

  it.each([undefined, null, 42, '', ' eth', 'eth ', 'eth/token', 'eth\ntoken', 'a'.repeat(129)])(
    'rejects unsafe token identifiers (%j)',
    (value) => {
      expect(() => validateAgentToken(value)).toThrow('Agent token is invalid.')
    }
  )

  it('accepts only literal boolean decisions', () => {
    expect(validateAgentDecision(true)).toBe(true)
    expect(validateAgentDecision(false)).toBe(false)
    for (const value of ['true', 'false', 1, 0, null, undefined]) {
      expect(() => validateAgentDecision(value)).toThrow('Agent confirmation decision is invalid.')
    }
  })
})
