import { describe, expect, it, vi } from 'vitest'
import type { CoworkAgentMessage, CoworkToolCall } from './cowork.types'
import {
  compactCoworkModelHistory,
  containsOmittedExecutionMarker,
  isOmittedExecutionMarker,
  materialiseCoworkImages,
  messageTextContent
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
    expect(messageTextContent(compacted[0].content).match(/^- write_file/gm)).toHaveLength(3)
    expect(
      messageTextContent(compacted[0].content).match(/Inspect the current file/g)
    ).toHaveLength(1)
    expect(compacted[0].content).not.toContain('Let me create')
  })

  it('preserves opaque provider reasoning state on assistant messages it retains', () => {
    const history: CoworkAgentMessage[] = [
      {
        role: 'assistant',
        content: null,
        reasoning_content: 'sentinel-thinking',
        tool_calls: [
          call('keep-1', 'read_file', { path: 'notes.txt' }),
          call('drop-1', 'write_file', {
            path: 'notes.txt',
            content: '[omitted after execution: 6 characters]'
          })
        ]
      },
      result('keep-1', { ok: true, result: { path: 'notes.txt' } }),
      result('drop-1', { ok: true, result: { path: 'notes.txt', bytes: 6 } }),
      { role: 'user', content: 'Continue.' },
      { role: 'assistant', content: 'Done.', reasoning_content: 'sentinel-final' }
    ]

    const compacted = compactCoworkModelHistory(history)

    expect(compacted[0]).toMatchObject({ reasoning_content: 'sentinel-thinking' })
    expect(compacted[0].tool_calls?.map((item) => item.id)).toEqual(['keep-1'])
    expect(compacted.at(-1)).toMatchObject({
      content: 'Done.',
      reasoning_content: 'sentinel-final'
    })
  })

  it('carries reasoning state onto a record that fully replaces an assistant turn', () => {
    const history: CoworkAgentMessage[] = [
      {
        role: 'assistant',
        content: 'Writing the file.',
        reasoning_content: 'first-thought',
        tool_calls: [
          call('drop-1', 'write_file', {
            path: 'notes.txt',
            content: '[omitted after execution: 6 characters]'
          })
        ]
      },
      result('drop-1', { ok: true, result: { path: 'notes.txt', bytes: 6 } }),
      {
        role: 'assistant',
        content: 'Writing it again.',
        reasoning_content: 'latest-thought',
        tool_calls: [
          call('drop-2', 'write_file', {
            path: 'notes.txt',
            content: '[omitted after execution: 6 characters]'
          })
        ]
      },
      result('drop-2', { ok: true, result: { path: 'notes.txt', bytes: 6 } })
    ]

    const compacted = compactCoworkModelHistory(history)

    expect(compacted).toHaveLength(1)
    expect(compacted[0].tool_calls).toBeUndefined()
    expect(compacted[0].reasoning_content).toBe('latest-thought')
  })
})

describe('compacted failure reasons', () => {
  it('keeps the reason a generated-file action failed', () => {
    const compacted = compactCoworkModelHistory([
      { role: 'user', content: 'Build the manifest.' },
      {
        role: 'assistant',
        content: null,
        tool_calls: [
          {
            id: 'c1',
            type: 'function',
            function: {
              name: 'create_xlsx',
              arguments: JSON.stringify({
                path: 'docs/FILE_MANIFEST.xlsx',
                content: '[omitted after execution: 40 characters]'
              })
            }
          }
        ]
      },
      {
        role: 'tool',
        tool_call_id: 'c1',
        content: JSON.stringify({
          ok: false,
          error: 'request.sheets[0].rows[3] has 5 cells but the table has 4 headers.'
        })
      },
      { role: 'user', content: 'why is it failing?' }
    ] as any)

    const serialized = JSON.stringify(compacted)
    expect(serialized).toContain('rows[3] has 5 cells')
    expect(serialized).not.toContain('[omitted after execution')
  })

  it('falls back to a bare failure when no reason was recorded', () => {
    const compacted = compactCoworkModelHistory([
      { role: 'user', content: 'Build it.' },
      {
        role: 'assistant',
        content: null,
        tool_calls: [
          {
            id: 'c1',
            type: 'function',
            function: {
              name: 'create_xlsx',
              arguments: JSON.stringify({
                path: 'a.xlsx',
                content: '[omitted after execution: 40 characters]'
              })
            }
          }
        ]
      },
      { role: 'tool', tool_call_id: 'c1', content: JSON.stringify({ ok: false }) },
      { role: 'user', content: 'again' }
    ] as any)

    expect(JSON.stringify(compacted)).toContain('reported failure')
  })
})

describe('materialiseCoworkImages', () => {
  const withImage = (): CoworkAgentMessage[] => [
    {
      role: 'user',
      content: [
        { type: 'text', text: 'Look at this.' },
        {
          type: 'image',
          image: {
            source: 'project',
            path: 'shot.png',
            mediaType: 'image/png',
            bytes: 6,
            width: 2,
            height: 3
          }
        }
      ]
    }
  ]

  it('sends pixels by default', async () => {
    const [message] = await materialiseCoworkImages(withImage(), async () => Buffer.from('pixels'))
    expect(Array.isArray(message.content)).toBe(true)
    expect(JSON.stringify(message.content)).toContain('image_url')
  })

  it('collapses to a plain string when the endpoint cannot take pictures', async () => {
    const load = vi.fn()
    const [message] = await materialiseCoworkImages(withImage(), load, { pixels: false })
    // A plain string is the one content shape no endpoint rejects.
    expect(typeof message.content).toBe('string')
    expect(message.content).toContain('Look at this.')
    expect(message.content).toContain('shot.png')
    expect(message.content).toContain('cannot receive pictures')
    // Bytes are never read for an endpoint that would only reject them.
    expect(load).not.toHaveBeenCalled()
  })
})
