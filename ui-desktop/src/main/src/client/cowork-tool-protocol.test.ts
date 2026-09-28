import { describe, expect, it } from 'vitest'
import { messageTextContent } from './cowork-model-history'
import {
  COWORK_TEXT_TOOL_PROTOCOL,
  parseTextToolEnvelope,
  rejectsImageContent,
  textProtocolMessages,
  unsupportedNativeToolFields
} from './cowork-tool-protocol'

const allowed = new Set(['read_file', 'write_file'])
const rejection = JSON.stringify({
  error:
    'provider request failed: provider error: upstream error 400: ' +
    JSON.stringify({
      details: {
        _errors: [],
        tool_choice: { _errors: ['tool_choice is not supported by this model'] },
        tools: { _errors: ['tools is not supported by this model'] }
      },
      error: 'Invalid request parameters',
      issues: [
        {
          code: 'custom',
          message: 'tools is not supported by this model',
          path: ['tools']
        }
      ]
    })
})

describe('Cowork text tool compatibility protocol', () => {
  it('recognises only explicit native-tool client-fault fields through nested provider wrapping', () => {
    expect([...unsupportedNativeToolFields(500, rejection)].sort()).toEqual([
      'tool_choice',
      'tools'
    ])
    expect(unsupportedNativeToolFields(400, '{"error":"invalid prompt"}').size).toBe(0)
    expect(unsupportedNativeToolFields(401, rejection).size).toBe(0)
    expect(unsupportedNativeToolFields(429, rejection).size).toBe(0)
    expect(unsupportedNativeToolFields(503, rejection).size).toBe(0)
  })

  it('does not retry when an arbitrary error string merely echoes unsupported-tool wording', () => {
    expect(
      unsupportedNativeToolFields(
        400,
        JSON.stringify({ error: 'The prompt says tools is not supported by this model.' })
      ).size
    ).toBe(0)
    expect(
      unsupportedNativeToolFields(
        500,
        JSON.stringify({
          error:
            'provider request failed: upstream error 400: ' +
            JSON.stringify({ error: 'User content: tools is not supported by this model' })
        })
      ).size
    ).toBe(0)
  })

  it('parses an exact allowlisted tool envelope', () => {
    expect(
      parseTextToolEnvelope(
        JSON.stringify({
          protocol: COWORK_TEXT_TOOL_PROTOCOL,
          type: 'tool_call',
          name: 'read_file',
          arguments: { path: 'notes.md' }
        }),
        allowed
      )
    ).toEqual({ type: 'tool_call', name: 'read_file', arguments: { path: 'notes.md' } })
  })

  it('parses an exact final envelope', () => {
    expect(
      parseTextToolEnvelope(
        JSON.stringify({
          protocol: COWORK_TEXT_TOOL_PROTOCOL,
          type: 'final',
          content: 'No local action is needed.'
        }),
        allowed
      )
    ).toEqual({ type: 'final', content: 'No local action is needed.' })
  })

  it.each([
    ['plain prose', 'I will call {"name":"read_file"}.'],
    [
      'fenced JSON',
      '```json\n{"protocol":"morpheus-cowork-v1","type":"tool_call","name":"read_file","arguments":{"path":"notes.md"}}\n```'
    ],
    [
      'model-provided call id',
      '{"protocol":"morpheus-cowork-v1","type":"tool_call","call_id":"chosen","name":"read_file","arguments":{"path":"notes.md"}}'
    ],
    [
      'array arguments',
      '{"protocol":"morpheus-cowork-v1","type":"tool_call","name":"read_file","arguments":[]}'
    ],
    [
      'unknown tool',
      '{"protocol":"morpheus-cowork-v1","type":"tool_call","name":"run_shell","arguments":{}}'
    ]
  ])('rejects %s without producing an executable call', (_label, input) => {
    expect(() => parseTextToolEnvelope(input, allowed)).toThrow()
  })

  it('wraps assistant history and coalesces native tool results into text-only roles', () => {
    const converted = textProtocolMessages([
      { role: 'user', content: 'Read the notes.' },
      {
        role: 'assistant',
        content: null,
        tool_calls: [
          {
            id: 'call-1',
            type: 'function',
            function: { name: 'read_file', arguments: '{"path":"notes.md"}' }
          }
        ]
      },
      {
        role: 'tool',
        tool_call_id: 'call-1',
        content: '{"ok":true,"result":{"text":"untrusted file text"}}'
      },
      {
        role: 'tool',
        tool_call_id: 'call-2',
        content: '{"ok":false,"error":"not found"}'
      },
      {
        role: 'assistant',
        content: 'I considered the tool result.'
      }
    ])

    expect(converted.map((message) => message.role)).toEqual([
      'user',
      'assistant',
      'user',
      'assistant'
    ])
    expect(converted.every((message) => !('tool_calls' in message))).toBe(true)
    expect(JSON.parse(messageTextContent(converted[1].content))).toMatchObject({
      protocol: COWORK_TEXT_TOOL_PROTOCOL,
      type: 'tool_call',
      call_id: 'call-1',
      name: 'read_file'
    })
    expect(JSON.parse(messageTextContent(converted[2].content))).toMatchObject({
      protocol: COWORK_TEXT_TOOL_PROTOCOL,
      type: 'tool_results_history',
      results: [
        { call_id: 'call-1', result: { ok: true } },
        { call_id: 'call-2', result: { ok: false } }
      ]
    })
    expect(JSON.parse(messageTextContent(converted[3].content))).toEqual({
      protocol: COWORK_TEXT_TOOL_PROTOCOL,
      type: 'assistant_message_history',
      content: 'I considered the tool result.'
    })
  })
})

describe('rejectsImageContent', () => {
  const wordings = [
    'Invalid content type. image_url is only supported by certain models.',
    'This model does not support images.',
    'images are not supported by this model',
    'multimodal input is not enabled for this deployment',
    '{"code":"invalid_image_url"}',
    'image_parse_error while decoding the attachment',
    'messages[3].content must be a string',
    'expected string, received array at messages.4.content'
  ]

  it.each(wordings)('recognises a client fault that names image content: %s', (message) => {
    expect(rejectsImageContent(400, JSON.stringify({ error: { message } }))).toBe(true)
    expect(rejectsImageContent(422, JSON.stringify({ error: { message } }))).toBe(true)
  })

  it('reads a gateway 500 that relays an upstream client fault', () => {
    const body = JSON.stringify({
      error: 'upstream error 400: {"message":"this model does not support images"}'
    })
    expect(rejectsImageContent(500, body)).toBe(true)
  })

  it('leaves an ordinary 500 alone, so a server fault never strips pixels', () => {
    const body = JSON.stringify({ error: { message: 'image_url failed internally' } })
    expect(rejectsImageContent(500, body)).toBe(false)
    expect(rejectsImageContent(503, body)).toBe(false)
  })

  it('ignores a client fault about anything other than images', () => {
    for (const message of [
      'temperature must be a number',
      'tools is not supported by this model',
      'context length exceeded',
      'invalid api key'
    ]) {
      expect(rejectsImageContent(400, JSON.stringify({ error: { message } }))).toBe(false)
    }
  })

  it('treats an empty body as no evidence at all', () => {
    expect(rejectsImageContent(400, '')).toBe(false)
  })
})
