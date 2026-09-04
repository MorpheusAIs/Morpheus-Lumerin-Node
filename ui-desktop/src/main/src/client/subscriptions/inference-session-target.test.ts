import { describe, expect, it } from 'vitest'
import { sessionInferenceHeaders } from './inference-session-target'

describe('sessionInferenceHeaders', () => {
  it('builds headers only for a marketplace session', () => {
    expect(sessionInferenceHeaders({ sessionId: 'session-1', chatId: 'chat-1' })).toEqual({
      session_id: 'session-1',
      chat_id: 'chat-1'
    })
  })

  it('rejects the local model target even when it is paired with a session', () => {
    expect(() => sessionInferenceHeaders({ modelId: 'tinyllama' })).toThrow(
      'active Morpheus marketplace session is required'
    )
    expect(() => sessionInferenceHeaders({ modelId: 'tinyllama', sessionId: 'session-1' })).toThrow(
      'active Morpheus marketplace session is required'
    )
  })

  it('rejects missing and malformed session identifiers', () => {
    expect(() => sessionInferenceHeaders({})).toThrow('Session ID is invalid.')
    expect(() => sessionInferenceHeaders({ sessionId: 'bad\nheader' })).toThrow(
      'Session ID is invalid.'
    )
  })
})
