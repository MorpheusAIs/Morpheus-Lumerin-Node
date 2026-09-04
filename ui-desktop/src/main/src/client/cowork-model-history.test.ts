import { describe, expect, it } from 'vitest'
import type { CoworkAgentMessage, CoworkToolCall } from './cowork.types'
import {
  compactCoworkModelHistory,
  containsOmittedExecutionMarker,
  isOmittedExecutionMarker
} from './cowork-model-history'

function call(id: string, name: string, input: Record<string, unknown>): CoworkToolCall {
  return {
    id,
    type: 'function',
    function: { name, arguments: JSON.stringify(input) }
  }
}

function result(id: string, value: Record<string, unknown>): CoworkAgentMessage {
  return { role: 'tool', tool_call_id: id, content: JSON.stringify(value) }
}

function expectValidToolOrdering(messages: CoworkAgentMessage[]): void {
  for (let index = 0; index < messages.length; index++) {
    const calls = messages[index].tool_calls ?? []
    if (!calls.length) continue
    const expected = new Set(calls.map((item) => item.id))
    let cursor = index + 1
    while (cursor < messages.length && messages[cursor].role === 'tool') {
      expected.delete(messages[cursor].tool_call_id ?? '')
      cursor++
    }
    expect(expected).toEqual(new Set())
  }
}

describe('cowork model history', () => {
  it('detects only internal omission sentinels, including nested generated payloads', () => {
    expect(isOmittedExecutionMarker('[omitted after execution: 40 characters]')).toBe(true)
    expect(isOmittedExecutionMarker('[omitted after execution]')).toBe(true)
    expect(isOmittedExecutionMarker('prefix [omitted after execution: 40 characters]')).toBe(false)
    expect(
      containsOmittedExecutionMarker({ blocks: [{ text: '[omitted after execution: 8 bytes]' }] })
    ).toBe(true)
    expect(containsOmittedExecutionMarker({ content: 'ordinary file content' })).toBe(false)
  })

  it('replaces the text-analyzer sentinel write pair with a non-tool historical record', () => {
    const history: CoworkAgentMessage[] = [
      { role: 'user', content: 'Create the analyzer.' },
      {
        role: 'assistant',
        content: 'Let me write it properly.',
        tool_calls: [
          call('write-1', 'write_file', {
            path: 'text_analyzer.py',
            content: '[omitted after execution: 40 characters]'
          })
        ]
      },
      result('write-1', { ok: true, result: { path: 'text_analyzer.py', bytes: 40 } }),
      { role: 'assistant', content: 'Continuing.' }
    ]

    const compacted = compactCoworkModelHistory(history)
    const serialized = JSON.stringify(compacted)

    expect(serialized).not.toContain('[omitted after execution')
    expect(compacted).toHaveLength(3)
    expect(compacted[1]).toMatchObject({ role: 'assistant' })
    expect(compacted[1].content).toContain('write_file')
    expect(compacted[1].content).toContain('text_analyzer.py')
    expect(compacted[1].content).toContain('reported success')
    expect(compacted[1].content).toContain('Inspect the current file')
    expect(compacted[1].tool_calls).toBeUndefined()
    expectValidToolOrdering(compacted)
  })

  it('retains native ordering for a mixed generated-write and verification-read batch', () => {
    const history: CoworkAgentMessage[] = [
      {
        role: 'assistant',
        content: null,
        tool_calls: [
          call('write-1', 'write_file', {
            path: 'calculator.py',
            content: '[omitted after execution: 4449 characters]'
          }),
          call('read-1', 'read_file', { path: 'calculator.py', startLine: 1, endLine: 20 })
        ]
      },
      result('write-1', { ok: true, result: { path: 'calculator.py', bytes: 4457 } }),
      result('read-1', { ok: true, result: { path: 'calculator.py', content: 'verified' } }),
      { role: 'user', content: 'Continue.' }
    ]

    const compacted = compactCoworkModelHistory(history)

    expect(compacted[0].tool_calls?.map((item) => item.id)).toEqual(['read-1'])
    expect(compacted[1]).toMatchObject({ role: 'tool', tool_call_id: 'read-1' })
    expect(compacted[2].content).toContain('write_file')
    expect(compacted[2].content).toContain('Inspect the current file')
    expect(compacted[3]).toMatchObject({ role: 'user' })
    expectValidToolOrdering(compacted)
  })

  it('keeps unresolved calls intact instead of manufacturing a historical outcome', () => {
    const history: CoworkAgentMessage[] = [
      {
        role: 'assistant',
        content: null,
        tool_calls: [
          call('pending', 'write_file', {
            path: 'pending.py',
            content: '[omitted after execution: 500 characters]'
          })
        ]
      }
    ]

    expect(compactCoworkModelHistory(history)).toEqual(history)
  })

  it('coalesces adjacent generated action pairs into one concise history message', () => {
    const history: CoworkAgentMessage[] = []
    for (const [index, path] of ['simple_calculator.py', 'calc.py', 'calc.py'].entries()) {
      const id = `write-${index}`
      history.push({
        role: 'assistant',
        content: 'Let me create the calculator properly.',
        tool_calls: [
          call(id, 'write_file', {
            path,
            content: '[omitted after execution: 40 characters]'
          })
        ]
      })
      history.push(result(id, { ok: true, result: { path, bytes: 40 } }))
    }

    const compacted = compactCoworkModelHistory(history)

    expect(compacted).toHaveLength(1)
    expect(compacted[0].content?.match(/^- write_file/gm)).toHaveLength(3)
    expect(compacted[0].content?.match(/Inspect the current file/g)).toHaveLength(1)
    expect(compacted[0].content).not.toContain('Let me create')
  })
})
