import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { CoworkProject, CoworkTask, CoworkToolCall } from './cowork.types'
import { mutationArgumentsHash } from './cowork-mutation-journal'
import { messageTextContent } from './cowork-model-history'

const state = vi.hoisted(() => ({
  project: undefined as CoworkProject | undefined,
  task: undefined as CoworkTask | undefined,
  replaceCalls: 0,
  failReplaceCall: undefined as number | undefined
}))

const toolMocks = vi.hoisted(() => ({
  execute: vi.fn(),
  loadImage: vi.fn(async (): Promise<Buffer | null> => Buffer.from('pixels'))
}))

const webMocks = vi.hoisted(() => ({
  retrieve: vi.fn()
}))

const visionMocks = vi.hoisted(() => ({
  verdict: vi.fn((_modelId: string) => null as { sees: boolean } | null),
  record: vi.fn(),
  run: vi.fn(async (modelId: string) => ({
    modelId,
    sees: true,
    probedAt: Date.now(),
    answer: 'red'
  }))
}))

const extensionState = vi.hoisted(() => ({
  catalog: { projectInstructions: null, skills: [] } as any
}))

const clone = <T>(value: T): T => JSON.parse(JSON.stringify(value))

const scrubToolArguments = (task: CoworkTask, toolCall: CoworkToolCall) => {
  if (toolCall.function.name !== 'write_file') return
  for (const message of task.agentMessages) {
    const stored = message.tool_calls?.find((call) => call.id === toolCall.id)
    if (!stored) continue
    const parsed = JSON.parse(stored.function.arguments)
    const length = typeof parsed.content === 'string' ? parsed.content.length : 0
    stored.function.arguments = JSON.stringify({
      ...parsed,
      content: `[omitted after execution: ${length} characters]`
    })
  }
}

const reconcileToolCalls = (task: CoworkTask) => {
  let ambiguousMutations = 0
  for (const execution of task.toolExecutions ?? []) {
    if (execution.status !== 'prepared') continue
    ambiguousMutations++
    execution.status = 'ambiguous'
    execution.completedAt = Date.now()
    execution.resultMessage = JSON.stringify({
      ok: false,
      error:
        `A previous ${execution.toolName.replaceAll('_', ' ')} action was interrupted after it was prepared. ` +
        'Cowork did not repeat it because its outcome may be ambiguous. Inspect the connected project before issuing a new instruction.'
    })
  }
  const unresolved = task.agentMessages.reduce<CoworkToolCall[]>((pending, message) => {
    if (message.role === 'assistant') pending.push(...(message.tool_calls ?? []))
    if (message.role === 'tool') {
      const index = pending.findIndex((call) => call.id === message.tool_call_id)
      if (index >= 0) pending.splice(index, 1)
    }
    return pending
  }, [])
  for (const call of unresolved) {
    const execution = task.toolExecutions?.find((item) => item.toolCallId === call.id)
    task.agentMessages.push({
      role: 'tool',
      tool_call_id: call.id,
      content:
        execution?.resultMessage ??
        JSON.stringify({ ok: false, error: 'Cowork did not replay this action.' })
    })
    scrubToolArguments(task, call)
  }
  return { ambiguousMutations, unresolvedCalls: unresolved.length }
}

vi.mock('../../config', () => ({
  default: { chain: { localProxyRouterUrl: 'http://router.test' } }
}))

vi.mock('../../logger', () => ({
  default: { error: vi.fn() }
}))

vi.mock('./cowork-extension-catalog', () => ({
  discoverCoworkExtensionCatalog: vi.fn(async () =>
    JSON.parse(JSON.stringify(extensionState.catalog))
  )
}))

vi.mock('./cowork-store', () => ({
  getProject: vi.fn(async (id: string) => (state.project?.id === id ? clone(state.project) : null)),
  getTask: vi.fn(async (id: string) => (state.task?.id === id ? clone(state.task) : null)),
  deleteTask: vi.fn(async (id: string) => {
    if (state.task?.id === id) state.task = undefined
  }),
  listRecentCompletedTaskMemories: vi.fn(async () => []),
  reconcileInterruptedToolCalls: vi.fn((task: CoworkTask) => reconcileToolCalls(task)),
  replaceTask: vi.fn(async (task: CoworkTask) => {
    state.replaceCalls++
    if (state.failReplaceCall === state.replaceCalls) {
      state.failReplaceCall = undefined
      throw new Error('Simulated durable task-save failure.')
    }
    if (!state.task || state.task.id !== task.id || state.task.revision !== task.revision) {
      throw new Error('This Cowork task changed in another operation. Refresh it and try again.')
    }
    const next = clone({ ...task, revision: task.revision + 1, updatedAt: Date.now() })
    state.task = clone(next)
    Object.assign(task, next)
    return clone(next)
  }),
  appendActivity: vi.fn((task: CoworkTask, activity: Record<string, unknown>) => {
    const value = {
      ...activity,
      id: `activity-${task.activities.length + 1}`,
      createdAt: Date.now()
    }
    task.activities.push(value as CoworkTask['activities'][number])
    return value
  }),
  appendAgentMessage: vi.fn((task: CoworkTask, message: CoworkTask['agentMessages'][number]) => {
    task.agentMessages.push(clone(message))
  }),
  appendDisplayMessage: vi.fn(
    (
      task: CoworkTask,
      role: 'user' | 'assistant',
      content: string,
      author?: CoworkTask['messages'][number]['author']
    ) => {
      task.messages.push({
        id: `message-${task.messages.length + 1}`,
        role,
        content,
        createdAt: Date.now(),
        ...(author ? { author: clone(author) } : {})
      })
    }
  ),
  setPlan: vi.fn((task: CoworkTask, plan: CoworkTask['plan']) => {
    task.plan = clone(plan)
  }),
  scrubCoworkToolArguments: vi.fn((task: CoworkTask, call: CoworkToolCall) =>
    scrubToolArguments(task, call)
  ),
  upsertArtifact: vi.fn()
}))

vi.mock('./cowork-tools', () => ({
  approvalRequirement: vi.fn(async () => null),
  executeCoworkTool: toolMocks.execute,
  loadCoworkImageBytes: toolMocks.loadImage,
  isCoworkMutationTool: vi.fn((name: string) =>
    [
      'write_file',
      'make_directory',
      'copy_file',
      'move_file',
      'delete_file',
      'create_docx',
      'create_xlsx',
      'create_pptx',
      'create_pdf'
    ].includes(name)
  ),
  toolArguments: vi.fn((raw: string) => JSON.parse(raw)),
  withCoworkMutationLock: vi.fn(async (_projectId: string, operation: () => Promise<unknown>) =>
    operation()
  )
}))

vi.mock('./cowork-web', () => ({
  retrieveCoworkWebPage: webMocks.retrieve
}))

vi.mock('./cowork-vision-cache', () => ({
  loadCoworkVisionProbes: vi.fn(async () => undefined),
  coworkVisionVerdict: visionMocks.verdict,
  recordCoworkVisionProbe: visionMocks.record
}))

vi.mock('./cowork-vision-probe', () => ({
  runCoworkVisionProbe: visionMocks.run
}))

import {
  MAX_TOOL_CALLS_PER_TURN,
  MAX_TOOL_PROTOCOL_CORRECTIONS,
  MAX_CONSECUTIVE_REJECTED_TURNS,
  MAX_TRUNCATED_TURN_CONTINUATIONS,
  MAX_UNFINISHED_PLAN_NUDGES,
  cancelCoworkRun,
  coworkRunActive,
  deleteCoworkTask,
  pauseCoworkRun,
  rebindCoworkTask,
  requestCoworkStart,
  resetCoworkImageRefusals,
  resolveCoworkApproval,
  startCoworkRun,
  steerCoworkRun
} from './cowork-runner'

const makeProject = (): CoworkProject => ({
  schemaVersion: 1,
  id: 'project-1',
  name: 'Runner regression project',
  rootPath: '/connected/project',
  instructions: '',
  approvalMode: 'auto',
  createdAt: 1,
  updatedAt: 1
})

const makeTask = (status: CoworkTask['status'] = 'queued'): CoworkTask => ({
  schemaVersion: 1,
  revision: 1,
  id: 'task-1',
  projectId: 'project-1',
  title: 'Runner regression task',
  goal: 'Create two files.',
  status,
  model: {
    modelId: 'local-test',
    modelName: 'Local test model',
    isLocal: true,
    dataBoundary: 'on-device'
  },
  plan: [],
  messages: [{ id: 'message-1', role: 'user', content: 'Create two files.', createdAt: 1 }],
  agentMessages: [{ role: 'user', content: 'Create two files.' }],
  activities: [],
  artifacts: [],
  createdAt: 1,
  updatedAt: 1
})

const completionResponse = (message: Record<string, unknown>): Response =>
  new Response(JSON.stringify({ choices: [{ message }] }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' }
  })

const errorResponse = (status: number, body: string): Response =>
  new Response(body, { status, headers: { 'Content-Type': 'application/json' } })

const textEnvelopeResponse = (value: Record<string, unknown>): Response =>
  completionResponse({ content: JSON.stringify({ protocol: 'morpheus-cowork-v1', ...value }) })

const unsupportedNativeToolsResponse = (): Response =>
  new Response(
    JSON.stringify({
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
    }),
    { status: 500, headers: { 'Content-Type': 'application/json' } }
  )

describe.sequential('Cowork runner interruption safety', () => {
  beforeEach(() => {
    state.project = makeProject()
    state.task = makeTask()
    state.replaceCalls = 0
    state.failReplaceCall = undefined
    extensionState.catalog = { projectInstructions: null, skills: [] }
    toolMocks.execute.mockReset()
    toolMocks.loadImage.mockReset()
    resetCoworkImageRefusals()
    toolMocks.loadImage.mockResolvedValue(Buffer.from('pixels'))
    visionMocks.verdict.mockReset()
    visionMocks.verdict.mockReturnValue(null)
    visionMocks.record.mockReset()
    visionMocks.run.mockClear()
    webMocks.retrieve.mockReset()
    vi.unstubAllGlobals()
  })

  it('falls back once to a strict text tool protocol when native tools are rejected', async () => {
    toolMocks.execute.mockResolvedValue({
      result: { path: 'result.txt', bytes: 4 },
      artifact: {
        path: 'result.txt',
        name: 'result.txt',
        kind: 'file',
        createdAt: 2,
        updatedAt: 2
      }
    })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(unsupportedNativeToolsResponse())
      .mockResolvedValueOnce(
        textEnvelopeResponse({
          type: 'tool_call',
          name: 'write_file',
          arguments: { path: 'result.txt', content: 'done' }
        })
      )
      .mockResolvedValueOnce(
        textEnvelopeResponse({ type: 'final', content: 'Created result.txt.' })
      )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('completed')
    expect(state.task!.toolProtocol).toBe('text-v1')
    expect(toolMocks.execute).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledTimes(3)
    for (const call of fetchMock.mock.calls) {
      expect(call[1]?.headers).toMatchObject({ 'x-morpheus-history': 'off' })
    }
    expect(state.task!.toolExecutions).toEqual([
      expect.objectContaining({
        toolName: 'write_file',
        status: 'succeeded',
        argumentsHash: expect.stringMatching(/^[a-f0-9]{64}$/),
        resultMessage: expect.stringContaining('"ok":true')
      })
    ])

    const nativeBody = JSON.parse(String(fetchMock.mock.calls[0][1]?.body))
    expect(nativeBody.tools.length).toBeGreaterThan(0)
    expect(nativeBody).not.toHaveProperty('tool_choice')
    expect(nativeBody).not.toHaveProperty('parallel_tool_calls')

    for (const call of fetchMock.mock.calls.slice(1)) {
      const body = JSON.parse(String(call[1]?.body))
      expect(body).not.toHaveProperty('tools')
      expect(body).not.toHaveProperty('tool_choice')
      expect(body).not.toHaveProperty('parallel_tool_calls')
      expect(body.messages.every((message: any) => message.role !== 'tool')).toBe(true)
      expect(body.messages.every((message: any) => !message.tool_calls)).toBe(true)
    }
    const fallbackBody = JSON.parse(String(fetchMock.mock.calls[1][1]?.body))
    expect(fallbackBody.messages[0].content).toContain('morpheus-cowork-v1')
    const storedCall = state.task!.agentMessages.find(
      (message) => message.role === 'assistant' && message.tool_calls?.length
    )?.tool_calls?.[0]
    expect(storedCall?.id).toMatch(/^text-[a-f0-9]{40}$/)
    expect(state.task!.activities).toContainEqual(
      expect.objectContaining({ label: 'Using Workspace tool compatibility mode' })
    )
  })

  it('does not retry a generic model 400 or execute anything', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response('{"error":"invalid prompt"}', {
        status: 400,
        headers: { 'Content-Type': 'application/json' }
      })
    )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('failed')
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(toolMocks.execute).not.toHaveBeenCalled()
  })

  it('rejects duplicate native tool call IDs before any action can run', async () => {
    const duplicateId = 'duplicate-write'
    const fetchMock = vi.fn().mockResolvedValue(
      completionResponse({
        content: null,
        tool_calls: [
          {
            id: duplicateId,
            type: 'function',
            function: {
              name: 'write_file',
              arguments: JSON.stringify({ path: 'first.txt', content: 'first' })
            }
          },
          {
            id: duplicateId,
            type: 'function',
            function: {
              name: 'write_file',
              arguments: JSON.stringify({ path: 'second.txt', content: 'second' })
            }
          }
        ]
      })
    )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('failed')
    expect(state.task!.error).toMatch(/duplicate tool call IDs/i)
    expect(toolMocks.execute).not.toHaveBeenCalled()
  })

  it('replays a durable mutation result without executing the same call ID again', async () => {
    const callId = 'write-once'
    const argumentsText = JSON.stringify({ path: 'once.txt', content: 'once' })
    const resultMessage = JSON.stringify({
      ok: true,
      result: { path: 'once.txt', bytes: 4 }
    })
    state.task!.toolExecutions = [
      {
        toolCallId: callId,
        toolName: 'write_file',
        status: 'succeeded',
        argumentsHash: mutationArgumentsHash('write_file', {
          content: 'once',
          path: 'once.txt'
        }),
        resultMessage,
        preparedAt: 2,
        completedAt: 3
      }
    ]
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [
            {
              id: callId,
              type: 'function',
              function: { name: 'write_file', arguments: argumentsText }
            }
          ]
        })
      )
      .mockResolvedValueOnce(completionResponse({ content: 'The existing result was retained.' }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('completed')
    expect(toolMocks.execute).not.toHaveBeenCalled()
    expect(
      state
        .task!.agentMessages.filter(
          (message) => message.role === 'tool' && message.tool_call_id === callId
        )
        .at(-1)?.content
    ).toBe(resultMessage)
    expect(state.task!.activities).toContainEqual(
      expect.objectContaining({ label: 'Skipped duplicate write file', status: 'success' })
    )
  })

  it('keeps ambiguous tasks approval-gated after a follow-up and equivalent new ID', async () => {
    state.project!.approvalMode = 'skip'
    // The retry uses an equivalent path spelling and a new ID. The task-wide
    // ambiguity gate must still prevent Skip mode from running it silently.
    const argumentsText = JSON.stringify({ path: './uncertain.txt', content: 'same content' })
    state.task!.toolExecutions = [
      {
        toolCallId: 'old-ambiguous-write',
        toolName: 'write_file',
        status: 'ambiguous',
        argumentsHash: mutationArgumentsHash('write_file', {
          content: 'same content',
          path: 'uncertain.txt'
        }),
        resultMessage: JSON.stringify({ ok: false, error: 'Outcome is ambiguous.' }),
        preparedAt: 2,
        completedAt: 3
      }
    ]
    state.task!.messages.push({
      id: 'message-after-ambiguity',
      role: 'user',
      content: 'Continue after I inspected the folder.',
      createdAt: 100
    })
    state.task!.agentMessages.push({
      role: 'user',
      content: 'Continue after I inspected the folder.'
    })
    const fetchMock = vi.fn().mockResolvedValue(
      completionResponse({
        content: null,
        tool_calls: [
          {
            id: 'new-write-id',
            type: 'function',
            function: { name: 'write_file', arguments: argumentsText }
          }
        ]
      })
    )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('waiting_approval')
    expect(state.task!.pendingApproval).toMatchObject({
      risk: 'write',
      toolCall: { id: 'new-write-id' },
      reason: expect.stringMatching(/ambiguous outcome/i)
    })
    expect(toolMocks.execute).not.toHaveBeenCalled()
  })

  it('does not let ignored arguments bypass an ambiguous-mutation identity', async () => {
    // A fresh Response per turn: the rejection is now handed back to the model,
    // so this run takes several turns instead of ending on the first one.
    const fetchMock = vi.fn().mockImplementation(async () =>
      completionResponse({
        content: null,
        tool_calls: [
          {
            id: 'write-with-extra-field',
            type: 'function',
            function: {
              name: 'write_file',
              arguments: JSON.stringify({
                path: 'uncertain.txt',
                content: 'same content',
                ignored: 'change-the-hash'
              })
            }
          }
        ]
      })
    )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('failed')
    expect(state.task!.error).toMatch(/unsupported argument/i)
    // Nothing ran on any of the attempts, and the model was told why each time.
    expect(toolMocks.execute).not.toHaveBeenCalled()
    expect(fetchMock).toHaveBeenCalledTimes(MAX_CONSECUTIVE_REJECTED_TURNS)
    expect(
      state.task!.agentMessages.filter(
        (message) =>
          message.role === 'tool' &&
          typeof message.content === 'string' &&
          message.content.includes('unsupported argument')
      )
    ).toHaveLength(MAX_CONSECUTIVE_REJECTED_TURNS)
  })

  it('reconciles a prepared mutation when saving its executed result fails', async () => {
    toolMocks.execute.mockResolvedValue({ result: { path: 'uncertain.txt', bytes: 4 } })
    state.failReplaceCall = 5
    const callId = 'write-before-save-failure'
    const fetchMock = vi.fn().mockResolvedValue(
      completionResponse({
        content: null,
        tool_calls: [
          {
            id: callId,
            type: 'function',
            function: {
              name: 'write_file',
              arguments: JSON.stringify({ path: 'uncertain.txt', content: 'data' })
            }
          }
        ]
      })
    )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(toolMocks.execute).toHaveBeenCalledTimes(1)
    expect(state.task!.status).toBe('failed')
    expect(state.task!.error).toMatch(/may have completed/i)
    expect(state.task!.toolExecutions).toEqual([
      expect.objectContaining({
        toolCallId: callId,
        status: 'ambiguous',
        resultMessage: expect.stringContaining('did not repeat it')
      })
    ])
    expect(
      state.task!.agentMessages.find(
        (message) => message.role === 'tool' && message.tool_call_id === callId
      )?.content
    ).toContain('may be ambiguous')
    const storedCall = state
      .task!.agentMessages.find((message) => message.role === 'assistant')
      ?.tool_calls?.find((call) => call.id === callId)
    expect(JSON.parse(storedCall!.function.arguments).content).toMatch(
      /^\[omitted after execution: \d+ characters\]$/
    )
  })

  const fencedEnvelope = (): Response =>
    completionResponse({
      content:
        '```json\n{"protocol":"morpheus-cowork-v1","type":"tool_call","name":"write_file","arguments":{"path":"unsafe.txt","content":"no"}}\n```'
    })

  it('never authorizes a tool from fenced JSON, and asks the model to resend', async () => {
    // Only an exact whole-response envelope may authorize a tool, so the fenced
    // form must not execute. Ending the task over it was the harsher half of
    // that rule: a fence is a formatting slip a correction turn fixes.
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(unsupportedNativeToolsResponse())
      .mockResolvedValueOnce(fencedEnvelope())
      .mockResolvedValueOnce(textEnvelopeResponse({ type: 'final', content: 'Resent properly.' }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(toolMocks.execute).not.toHaveBeenCalled()
    expect(state.task!.status).toBe('completed')
    expect(fetchMock).toHaveBeenCalledTimes(3)
  })

  it('still fails closed when the model only ever sends fenced JSON', async () => {
    const fetchMock = vi.fn().mockImplementation(async (_url: string, init: RequestInit) => {
      const body = JSON.parse(String(init.body))
      return body.tools ? unsupportedNativeToolsResponse() : fencedEnvelope()
    })
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })

    expect(state.task!.status).toBe('failed')
    expect(state.task!.error).toMatch(/compatibility protocol/i)
    expect(toolMocks.execute).not.toHaveBeenCalled()
  }, 20_000)

  it('keeps text mode for the same fingerprint and resets it for a new endpoint', async () => {
    state.task = makeTask('paused')
    state.task.modelFingerprint = 'session-a'
    state.task.toolProtocol = 'text-v1'
    const sameFingerprint = vi.fn(async () => ({
      model: state.task!.model,
      fingerprint: 'session-a'
    }))
    const fetchMock = vi
      .fn()
      .mockResolvedValue(textEnvelopeResponse({ type: 'final', content: 'Same session.' }))
    vi.stubGlobal('fetch', fetchMock)

    await requestCoworkStart(state.task.id, async () => ({}), vi.fn(), sameFingerprint, true)
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body))).not.toHaveProperty('tools')

    state.task!.status = 'paused'
    const changedFingerprint = vi.fn(async () => ({
      model: state.task!.model,
      fingerprint: 'session-b'
    }))
    fetchMock.mockReset().mockResolvedValue(completionResponse({ content: 'New session.' }))

    await requestCoworkStart(state.task!.id, async () => ({}), vi.fn(), changedFingerprint, true)
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))
    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body)).tools.length).toBeGreaterThan(0)
    expect(state.task!.toolProtocol).toBeUndefined()
  })

  it('does not carry compatibility mode from a legacy task with no fingerprint', async () => {
    state.task = makeTask('paused')
    state.task.toolProtocol = 'text-v1'
    const refreshModelTarget = vi.fn(async () => ({
      model: state.task!.model,
      fingerprint: 'resolved-endpoint'
    }))
    const fetchMock = vi.fn().mockResolvedValue(completionResponse({ content: 'Native mode.' }))
    vi.stubGlobal('fetch', fetchMock)

    await requestCoworkStart(state.task.id, async () => ({}), vi.fn(), refreshModelTarget, true)
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(JSON.parse(String(fetchMock.mock.calls[0][1]?.body)).tools.length).toBeGreaterThan(0)
    expect(state.task!.toolProtocol).toBeUndefined()
  })

  it('closes every remaining tool call when a running batch is paused', async () => {
    let markExecutionStarted!: () => void
    const executionStarted = new Promise<void>((resolve) => {
      markExecutionStarted = resolve
    })
    let releaseExecution!: () => void
    const executionGate = new Promise<void>((resolve) => {
      releaseExecution = resolve
    })

    toolMocks.execute.mockImplementationOnce(async () => {
      markExecutionStarted()
      await executionGate
      return { result: { path: 'first.txt', bytes: 5 } }
    })

    const firstCall = {
      id: 'write-first',
      type: 'function',
      function: {
        name: 'write_file',
        arguments: JSON.stringify({ path: 'first.txt', content: 'first' })
      }
    }
    const secondCall = {
      id: 'write-second',
      type: 'function',
      function: {
        name: 'write_file',
        arguments: JSON.stringify({ path: 'second.txt', content: 'second' })
      }
    }
    const fetchMock = vi
      .fn()
      .mockResolvedValue(completionResponse({ content: null, tool_calls: [firstCall, secondCall] }))
    vi.stubGlobal('fetch', fetchMock)

    const emit = vi.fn()
    startCoworkRun(state.task!.id, async () => ({}), emit)
    await executionStarted

    const pause = pauseCoworkRun(state.task!.id, emit)
    releaseExecution()
    const paused = await pause

    expect(paused.status).toBe('paused')
    expect(coworkRunActive(paused.id)).toBe(false)
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(toolMocks.execute).toHaveBeenCalledTimes(1)

    const assistantBatch = paused.agentMessages.find((message) => message.role === 'assistant')
    expect(assistantBatch?.tool_calls?.map((call) => call.id)).toEqual([
      firstCall.id,
      secondCall.id
    ])

    const toolResponses = paused.agentMessages.filter((message) => message.role === 'tool')
    expect(toolResponses).toHaveLength(2)
    expect(toolResponses.map((message) => message.tool_call_id)).toEqual([
      firstCall.id,
      secondCall.id
    ])
    expect(
      JSON.parse(
        messageTextContent(
          toolResponses.find((message) => message.tool_call_id === secondCall.id)!.content
        )
      )
    ).toEqual({
      ok: false,
      error: 'The task was interrupted before this action ran.'
    })

    for (const call of assistantBatch!.tool_calls!) {
      const storedArguments = JSON.parse(call.function.arguments)
      expect(storedArguments.content).toMatch(/^\[omitted after execution: \d+ characters\]$/)
    }
  })

  it('loads only the exact folder and skill guidance hashes that were enabled', async () => {
    const folderHash = 'a'.repeat(64)
    const skillHash = 'b'.repeat(64)
    state.project!.extensionSettings = {
      folderInstructionsEnabled: true,
      enabledSkillIds: ['review-checklist'],
      folderInstructionsHash: folderHash,
      skillInstructionHashes: { 'review-checklist': skillHash }
    }
    extensionState.catalog = {
      projectInstructions: {
        source: 'project:.morpheus/cowork/instructions.md',
        content: 'Cite each factual statement.',
        contentHash: folderHash
      },
      skills: [
        {
          id: 'review-checklist',
          name: 'Review checklist',
          instructions: 'Verify the final artifact before completion.',
          instructionsHash: skillHash
        }
      ]
    }
    const fetchMock = vi
      .fn()
      .mockResolvedValue(completionResponse({ content: 'The reviewed deliverable is ready.' }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('completed')
    expect(fetchMock).toHaveBeenCalledTimes(1)
    const request = JSON.parse(String(fetchMock.mock.calls[0][1]?.body))
    expect(request.messages[0].content).toContain(
      '[Folder guidance: project:.morpheus/cowork/instructions.md]\nCite each factual statement.'
    )
    expect(request.messages[0].content).toContain(
      '[Skill: Review checklist (review-checklist)]\nVerify the final artifact before completion.'
    )
    expect(state.task!.activities).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          label: 'Loaded enabled project guidance',
          status: 'success'
        })
      ])
    )
  })

  it.each([
    {
      label: 'folder instructions',
      settings: {
        folderInstructionsEnabled: true,
        enabledSkillIds: [],
        folderInstructionsHash: 'a'.repeat(64)
      },
      catalog: {
        projectInstructions: {
          source: 'project:.morpheus/cowork/instructions.md',
          content: 'Changed after the user reviewed it.',
          contentHash: 'c'.repeat(64)
        },
        skills: []
      },
      expectedError: /Project instructions changed after activation/
    },
    {
      label: 'skill instructions',
      settings: {
        folderInstructionsEnabled: false,
        enabledSkillIds: ['review-checklist'],
        skillInstructionHashes: { 'review-checklist': 'b'.repeat(64) }
      },
      catalog: {
        projectInstructions: null,
        skills: [
          {
            id: 'review-checklist',
            name: 'Review checklist',
            instructions: 'Changed after the user reviewed it.',
            instructionsHash: 'd'.repeat(64)
          }
        ]
      },
      expectedError: /Project skill “Review checklist” changed after activation/
    }
  ])(
    'fails closed when enabled $label no longer matches its reviewed hash',
    async ({ settings, catalog, expectedError }) => {
      state.project!.extensionSettings = settings
      extensionState.catalog = catalog
      const fetchMock = vi.fn()
      vi.stubGlobal('fetch', fetchMock)

      startCoworkRun(state.task!.id, async () => ({}), vi.fn())
      await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

      expect(state.task!.status).toBe('failed')
      expect(state.task!.error).toMatch(expectedError)
      expect(fetchMock).not.toHaveBeenCalled()
    }
  )

  it('does not implicitly restart a cancelled task', async () => {
    state.task = makeTask('cancelled')
    const refreshModelTarget = vi.fn(async () => ({
      model: state.task!.model,
      fingerprint: 'local:test'
    }))
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)

    await expect(
      requestCoworkStart(state.task.id, async () => ({}), vi.fn(), refreshModelTarget)
    ).rejects.toThrow(/Start it again explicitly/)

    expect(refreshModelTarget).not.toHaveBeenCalled()
    expect(fetchMock).not.toHaveBeenCalled()
    expect(state.task.status).toBe('cancelled')
    expect(coworkRunActive(state.task.id)).toBe(false)
  })

  it('serializes start with a concurrent cancellation so no run survives cancellation', async () => {
    state.task = makeTask('paused')
    let releaseRefresh!: () => void
    const refreshGate = new Promise<void>((resolve) => {
      releaseRefresh = resolve
    })
    const refreshModelTarget = vi.fn(async () => {
      await refreshGate
      return { model: state.task!.model, fingerprint: 'local:test' }
    })
    vi.stubGlobal(
      'fetch',
      vi.fn(
        (_url: string, init?: RequestInit) =>
          new Promise<Response>((_resolve, reject) => {
            init?.signal?.addEventListener(
              'abort',
              () => reject(new DOMException('Aborted', 'AbortError')),
              { once: true }
            )
          })
      )
    )
    const emit = vi.fn()

    const starting = requestCoworkStart(
      state.task.id,
      async () => ({}),
      emit,
      refreshModelTarget,
      true
    )
    await vi.waitFor(() => expect(refreshModelTarget).toHaveBeenCalledTimes(1))
    const cancelling = cancelCoworkRun(state.task.id, emit)
    await Promise.resolve()

    expect(state.task.status).toBe('paused')
    releaseRefresh()
    await starting
    const cancelled = await cancelling

    expect(cancelled.status).toBe('cancelled')
    expect(coworkRunActive(cancelled.id)).toBe(false)
  })

  it('serializes deletion with a concurrent start so a deleted task cannot run or reappear', async () => {
    state.task = makeTask('paused')
    const taskId = state.task.id
    let releaseRefresh!: () => void
    const refreshGate = new Promise<void>((resolve) => {
      releaseRefresh = resolve
    })
    const refreshModelTarget = vi.fn(async () => {
      await refreshGate
      return { model: state.task!.model, fingerprint: 'local:test' }
    })
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)

    const starting = requestCoworkStart(taskId, async () => ({}), vi.fn(), refreshModelTarget, true)
    await vi.waitFor(() => expect(refreshModelTarget).toHaveBeenCalledTimes(1))
    const deleting = deleteCoworkTask(taskId)

    releaseRefresh()
    await starting
    await deleting

    expect(state.task).toBeUndefined()
    expect(coworkRunActive(taskId)).toBe(false)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('serializes steering with cancellation and suppresses a late replacement run', async () => {
    state.task = makeTask('completed')
    state.task.completedAt = Date.now()
    let releaseRefresh!: () => void
    const refreshGate = new Promise<void>((resolve) => {
      releaseRefresh = resolve
    })
    const refreshModelTarget = vi.fn(async () => {
      await refreshGate
      return { model: state.task!.model, fingerprint: 'local:test' }
    })
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    const emit = vi.fn()

    const steering = steerCoworkRun(
      state.task.id,
      'Continue with a safer approach.',
      async () => ({}),
      emit,
      refreshModelTarget
    )
    await vi.waitFor(() => expect(refreshModelTarget).toHaveBeenCalledTimes(1))
    const cancelling = cancelCoworkRun(state.task.id, emit)

    releaseRefresh()
    await steering
    const cancelled = await cancelling

    expect(cancelled.status).toBe('cancelled')
    expect(cancelled.messages.at(-1)).toMatchObject({
      role: 'user',
      content: 'Continue with a safer approach.'
    })
    expect(coworkRunActive(cancelled.id)).toBe(false)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('treats an old approval token as an idempotent no-op', async () => {
    const freshCall = {
      id: 'fresh-write',
      type: 'function' as const,
      function: {
        name: 'write_file',
        arguments: JSON.stringify({ path: 'fresh.txt', content: 'fresh' })
      }
    }
    state.task = makeTask('waiting_approval')
    state.task.pendingApproval = {
      id: 'fresh-approval',
      toolCall: freshCall,
      remainingToolCalls: [],
      reason: 'Review the fresh action.',
      risk: 'write',
      createdAt: Date.now()
    }
    const refreshModelTarget = vi.fn()

    const authoritative = await resolveCoworkApproval(
      state.task.id,
      'old-approval',
      true,
      async () => ({}),
      vi.fn(),
      refreshModelTarget
    )

    expect(authoritative).toMatchObject({
      status: 'waiting_approval',
      pendingApproval: { id: 'fresh-approval' }
    })
    expect(refreshModelTarget).not.toHaveBeenCalled()
    expect(toolMocks.execute).not.toHaveBeenCalled()
  })

  it('allows remote data sharing to be denied after its session expires', async () => {
    const authorizationCall = {
      id: 'authorize-remote',
      type: 'function' as const,
      function: { name: 'authorize_remote_model', arguments: '{}' }
    }
    state.task = makeTask('waiting_approval')
    state.task.model = {
      modelId: 'remote-model',
      modelName: 'Remote model',
      isLocal: false,
      dataBoundary: 'independent-provider',
      sessionId: 'expired-session',
      sessionEndsAt: 1
    }
    state.task.modelFingerprint = 'remote:expired-session'
    state.task.pendingApproval = {
      id: 'remote-approval',
      toolCall: authorizationCall,
      remainingToolCalls: [],
      reason: 'Share the task with the independent provider.',
      risk: 'read',
      createdAt: Date.now()
    }
    const refreshModelTarget = vi.fn(async () => {
      throw new Error('The selected marketplace session has expired.')
    })

    const denied = await resolveCoworkApproval(
      state.task.id,
      state.task.pendingApproval.id,
      false,
      async () => ({}),
      vi.fn(),
      refreshModelTarget
    )

    expect(refreshModelTarget).not.toHaveBeenCalled()
    expect(denied.status).toBe('paused')
    expect(denied.pendingApproval).toBeUndefined()
    expect(denied.activities.at(-1)).toMatchObject({
      label: 'Remote model data sharing denied',
      status: 'error'
    })
  })

  it.each([
    { label: 'cancellation', stop: cancelCoworkRun, expectedStatus: 'cancelled' as const },
    { label: 'pause', stop: pauseCoworkRun, expectedStatus: 'paused' as const }
  ])(
    'aborts an approved continuation immediately while $label waits for its lock',
    async ({ stop, expectedStatus }) => {
      const webCall = {
        id: 'fetch-cancellable-docs',
        type: 'function' as const,
        function: {
          name: 'fetch_web_page',
          arguments: JSON.stringify({ url: 'https://example.com/reference' })
        }
      }
      state.task = makeTask('waiting_approval')
      state.task.agentMessages.push({ role: 'assistant', content: null, tool_calls: [webCall] })
      state.task.pendingApproval = {
        id: 'approval-cancellable-web',
        toolCall: webCall,
        remainingToolCalls: [],
        reason: 'Allow a bounded public web request.',
        risk: 'network',
        createdAt: Date.now()
      }
      let markRetrievalStarted!: () => void
      const retrievalStarted = new Promise<void>((resolve) => {
        markRetrievalStarted = resolve
      })
      let observedSignal: AbortSignal | undefined
      webMocks.retrieve.mockImplementation(
        async (_url: string, options: { signal?: AbortSignal } = {}) => {
          observedSignal = options.signal
          markRetrievalStarted()
          await new Promise<void>((_resolve, reject) => {
            options.signal?.addEventListener(
              'abort',
              () => reject(new DOMException('Aborted', 'AbortError')),
              { once: true }
            )
          })
          throw new Error('Unreachable')
        }
      )
      const refreshModelTarget = vi.fn(async () => ({
        model: state.task!.model,
        fingerprint: 'local:test'
      }))
      const emit = vi.fn()

      const resolving = resolveCoworkApproval(
        state.task.id,
        state.task.pendingApproval.id,
        true,
        async () => ({}),
        emit,
        refreshModelTarget
      )
      await retrievalStarted
      expect(coworkRunActive(state.task.id)).toBe(true)

      const stopping = stop(state.task.id, emit)
      await vi.waitFor(() => expect(observedSignal?.aborted).toBe(true))
      await resolving
      const stopped = await stopping

      expect(stopped.status).toBe(expectedStatus)
      expect(coworkRunActive(stopped.id)).toBe(false)
      expect(
        stopped.agentMessages.filter(
          (message) => message.role === 'tool' && message.tool_call_id === webCall.id
        )
      ).toHaveLength(1)
    }
  )

  it('requires explicit network approval before fetching a public HTTPS page', async () => {
    const webCall = {
      id: 'fetch-docs',
      type: 'function',
      function: {
        name: 'fetch_web_page',
        arguments: JSON.stringify({ url: 'https://example.com/reference' })
      }
    }
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(completionResponse({ content: null, tool_calls: [webCall] }))
      .mockResolvedValueOnce(completionResponse({ content: 'Research complete.' }))
    vi.stubGlobal('fetch', fetchMock)
    webMocks.retrieve.mockResolvedValue({
      finalUrl: 'https://example.com/reference',
      title: 'Reference',
      text: 'Bounded public source text.',
      contentType: 'text/html',
      truncated: false
    })
    const refreshModelTarget = vi.fn(async () => ({
      model: state.task!.model,
      fingerprint: 'local:test'
    }))
    const emit = vi.fn()

    startCoworkRun(state.task!.id, async () => ({}), emit, state.task!.projectId)
    await vi.waitFor(() => expect(state.task!.status).toBe('waiting_approval'))

    expect(state.task!.pendingApproval).toEqual(
      expect.objectContaining({
        risk: 'network',
        toolCall: expect.objectContaining({ id: webCall.id })
      })
    )
    expect(state.task!.pendingApproval?.reason).toContain('exactly “https://example.com/reference”')
    expect(webMocks.retrieve).not.toHaveBeenCalled()

    await resolveCoworkApproval(
      state.task!.id,
      state.task!.pendingApproval!.id,
      true,
      async () => ({}),
      emit,
      refreshModelTarget
    )
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(webMocks.retrieve).toHaveBeenCalledWith(
      'https://example.com/reference',
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    )
    expect(state.task!.status).toBe('completed')
  })

  it('does not fetch a model-supplied URL whose query could conceal unreviewed data', async () => {
    const webCall = {
      id: 'fetch-hidden-query',
      type: 'function',
      function: {
        name: 'fetch_web_page',
        arguments: JSON.stringify({ url: 'https://example.com/reference?payload=hidden' })
      }
    }
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(completionResponse({ content: null, tool_calls: [webCall] }))
      .mockResolvedValueOnce(completionResponse({ content: 'Skipped the unsafe URL.' }))
    vi.stubGlobal('fetch', fetchMock)
    const emit = vi.fn()

    startCoworkRun(state.task!.id, async () => ({}), emit, state.task!.projectId)
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(webMocks.retrieve).not.toHaveBeenCalled()
    expect(state.task!.pendingApproval).toBeUndefined()
    expect(state.task!.activities).toContainEqual(
      expect.objectContaining({
        label: 'fetch web page',
        status: 'error',
        detail: expect.stringContaining('query parameters are not allowed')
      })
    )
  })

  it('executes an equivalent mutation once across new call IDs, then pauses the loop', async () => {
    const argumentsText = JSON.stringify({ path: 'calc.py', content: 'print(2 + 2)' })
    toolMocks.execute.mockResolvedValue({ result: { path: 'calc.py', bytes: 12 } })
    const responseFor = (id: string) =>
      completionResponse({
        content: null,
        tool_calls: [
          {
            id,
            type: 'function',
            function: { name: 'write_file', arguments: argumentsText }
          }
        ]
      })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(responseFor('write-calc-1'))
      .mockResolvedValueOnce(responseFor('write-calc-2'))
      .mockResolvedValueOnce(responseFor('write-calc-3'))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(toolMocks.execute).toHaveBeenCalledTimes(1)
    expect(state.task!.toolExecutions).toHaveLength(1)
    expect(state.task!.status).toBe('paused')
    expect(state.task!.pauseReason).toBe('repetition_guard')
    expect(state.task!.error).toMatch(/same file action|repetition/i)
    await expect(
      requestCoworkStart(
        state.task!.id,
        async () => ({}),
        vi.fn(),
        async () => ({ model: state.task!.model, fingerprint: 'same-session' }),
        true
      )
    ).rejects.toThrow(/new instruction/i)
  })

  it('allows one equivalent restore in a fresh instruction but suppresses repeats in that instruction', async () => {
    const argumentsText = JSON.stringify({ path: 'restorable.py', content: 'print("restored")' })
    const responseFor = (id: string) =>
      completionResponse({
        content: null,
        tool_calls: [
          {
            id,
            type: 'function',
            function: { name: 'write_file', arguments: argumentsText }
          }
        ]
      })
    toolMocks.execute.mockResolvedValue({ result: { path: 'restorable.py', bytes: 17 } })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(responseFor('initial-write'))
      .mockResolvedValueOnce(completionResponse({ content: 'Initial version created.' }))
      .mockResolvedValueOnce(responseFor('restore-write'))
      .mockResolvedValueOnce(responseFor('same-instruction-repeat'))
      .mockResolvedValueOnce(completionResponse({ content: 'Restored once without looping.' }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))
    expect(toolMocks.execute).toHaveBeenCalledTimes(1)

    await steerCoworkRun(
      state.task!.id,
      'The file changed outside Workspace. Restore the requested version.',
      async () => ({}),
      vi.fn(),
      async () => ({ model: state.task!.model, fingerprint: 'same-session' })
    )
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(toolMocks.execute).toHaveBeenCalledTimes(2)
    expect(state.task!.toolExecutions).toHaveLength(2)
    expect(state.task!.toolExecutions?.map((execution) => execution.instructionId)).toEqual([
      expect.any(String),
      expect.any(String)
    ])
    expect(state.task!.toolExecutions?.[0].instructionId).not.toBe(
      state.task!.toolExecutions?.[1].instructionId
    )
    expect(
      state.task!.agentMessages.find(
        (message) => message.role === 'tool' && message.tool_call_id === 'same-instruction-repeat'
      )?.content
    ).toContain('"duplicateOf":"restore-write"')
    expect(state.task!.status).toBe('completed')
  })

  it('pauses instead of writing a persisted omitted-payload marker', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      completionResponse({
        content: null,
        tool_calls: [
          {
            id: 'placeholder-write',
            type: 'function',
            function: {
              name: 'write_file',
              arguments: JSON.stringify({
                path: 'text_analyzer.py',
                content: '[omitted after execution: 4457 characters]'
              })
            }
          }
        ]
      })
    )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(toolMocks.execute).not.toHaveBeenCalled()
    expect(state.task!.status).toBe('paused')
    expect(state.task!.error).toMatch(/omitted-payload marker/i)
  })

  it('rebinds an old task with a bounded handoff and excludes prior tool protocol', async () => {
    state.task = makeTask('paused')
    state.task.modelFingerprint = 'old-session'
    state.task.dataAccessApproved = true
    state.task.dataAccessApprovedFingerprint = 'old-session'
    state.task.toolProtocol = 'text-v1'
    state.task.agentMessages.push(
      {
        role: 'assistant',
        content: null,
        tool_calls: [
          {
            id: 'old-write',
            type: 'function',
            function: {
              name: 'write_file',
              arguments: JSON.stringify({ path: 'old.txt', content: 'old generated content' })
            }
          }
        ]
      },
      { role: 'tool', tool_call_id: 'old-write', content: JSON.stringify({ ok: true }) }
    )
    state.task.messages.push({
      id: 'old-answer',
      role: 'assistant',
      content: 'Earlier output is ready, but verify it.',
      createdAt: 2,
      author: { kind: 'model', modelName: 'Local test model' }
    })
    const nextModel = {
      modelId: 'replacement-model',
      modelName: 'Replacement model',
      isLocal: true,
      dataBoundary: 'on-device' as const
    }

    const rebound = await rebindCoworkTask(
      state.task.id,
      { model: nextModel, fingerprint: 'new-session' },
      vi.fn()
    )

    expect(rebound.model).toEqual(nextModel)
    expect(rebound.toolProtocol).toBeUndefined()
    expect(rebound.dataAccessApproved).toBeUndefined()
    expect(rebound.handoff).toMatchObject({
      previousModelName: 'Local test model',
      goal: state.task.goal
    })
    expect(rebound.modelBindings).toHaveLength(2)
    expect(rebound.messages.at(-1)).toMatchObject({
      author: { kind: 'workspace' },
      content: expect.stringContaining('Model changed')
    })

    const fetchMock = vi
      .fn()
      .mockResolvedValue(completionResponse({ content: 'Continued safely.' }))
    vi.stubGlobal('fetch', fetchMock)
    await requestCoworkStart(
      state.task.id,
      async () => ({}),
      vi.fn(),
      async () => ({ model: nextModel, fingerprint: 'new-session' }),
      true
    )
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))
    const requestBody = JSON.parse(String(fetchMock.mock.calls[0][1]?.body))
    expect(requestBody.messages.slice(1)).toEqual([
      expect.objectContaining({
        role: 'user',
        content: expect.stringContaining('explicitly labelled handoff')
      })
    ])
    expect(JSON.stringify(requestBody.messages.slice(1))).not.toContain('old generated content')
    expect(state.task!.messages.at(-1)?.author).toMatchObject({
      kind: 'model',
      modelName: 'Replacement model'
    })
  })

  it('preserves a repetition pause across rebind until the user sends a new instruction', async () => {
    state.task = makeTask('paused')
    state.task.modelFingerprint = 'old-session'
    state.task.pauseReason = 'repetition_guard'
    state.task.error =
      'The model repeated the same file action. Review the project and send a new instruction.'
    const nextModel = {
      modelId: 'replacement-model',
      modelName: 'Replacement model',
      isLocal: true,
      dataBoundary: 'on-device' as const
    }
    const refresh = async () => ({ model: nextModel, fingerprint: 'new-session' })

    const rebound = await rebindCoworkTask(
      state.task.id,
      { model: nextModel, fingerprint: 'new-session' },
      vi.fn()
    )

    expect(rebound.pauseReason).toBe('repetition_guard')
    expect(rebound.error).toMatch(/new instruction/i)
    await expect(
      requestCoworkStart(state.task.id, async () => ({}), vi.fn(), refresh, true)
    ).rejects.toThrow(/new instruction/i)

    const fetchMock = vi
      .fn()
      .mockResolvedValue(
        completionResponse({ content: 'Continued after the user reviewed the project.' })
      )
    vi.stubGlobal('fetch', fetchMock)
    await steerCoworkRun(
      state.task.id,
      'I reviewed the project. Continue without repeating the earlier write.',
      async () => ({}),
      vi.fn(),
      refresh
    )
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.pauseReason).toBeUndefined()
    expect(state.task!.status).toBe('completed')
    expect(state.task!.agentMessages).toContainEqual(
      expect.objectContaining({
        role: 'user',
        content: 'I reviewed the project. Continue without repeating the earlier write.'
      })
    )
  })
})

describe.sequential('Cowork runner provider thinking-state continuity', () => {
  const REASONING = 'plan: inspect notes.txt, then answer without repeating the write'
  const FINAL_REASONING = 'the file is already correct; report and stop'

  const toolCall = (name: string, input: Record<string, unknown>) => ({
    id: `reasoning-${name}-1`,
    type: 'function',
    function: { name, arguments: JSON.stringify(input) }
  })

  const writeSucceeds = () => {
    toolMocks.execute.mockResolvedValue({
      result: { path: 'notes.txt', bytes: 6 },
      artifact: {
        path: 'notes.txt',
        name: 'notes.txt',
        kind: 'file',
        createdAt: 2,
        updatedAt: 2
      }
    })
  }

  const bodies = (fetchMock: ReturnType<typeof vi.fn>): any[] =>
    fetchMock.mock.calls.map((call) => JSON.parse(String(call[1]?.body)))

  beforeEach(() => {
    state.project = makeProject()
    state.task = makeTask()
    state.replaceCalls = 0
    state.failReplaceCall = undefined
    extensionState.catalog = { projectInstructions: null, skills: [] }
    toolMocks.execute.mockReset()
    toolMocks.loadImage.mockReset()
    resetCoworkImageRefusals()
    toolMocks.loadImage.mockResolvedValue(Buffer.from('pixels'))
    visionMocks.verdict.mockReset()
    visionMocks.verdict.mockReturnValue(null)
    visionMocks.record.mockReset()
    visionMocks.run.mockClear()
    webMocks.retrieve.mockReset()
    vi.unstubAllGlobals()
  })

  it('replays provider reasoning state unchanged on the next native tool turn', async () => {
    toolMocks.execute.mockResolvedValue({ result: { path: 'notes.txt', bytes: 6 } })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          reasoning_content: REASONING,
          tool_calls: [toolCall('read_file', { path: 'notes.txt' })]
        })
      )
      .mockResolvedValueOnce(
        completionResponse({
          content: 'notes.txt is unchanged.',
          reasoning_content: FINAL_REASONING
        })
      )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('completed')
    expect(toolMocks.execute).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledTimes(2)

    const [first, second] = bodies(fetchMock)
    expect(first.messages.some((message: any) => 'reasoning_content' in message)).toBe(false)
    expect(second.tools.length).toBeGreaterThan(0)
    const replayed = second.messages.find((message: any) => message.tool_calls?.length)
    expect(replayed.reasoning_content).toBe(REASONING)

    const stored = state.task!.agentMessages.find(
      (message) => message.role === 'assistant' && message.tool_calls?.length
    )
    expect(stored!.reasoning_content).toBe(REASONING)
    expect(state.task!.agentMessages.at(-1)).toMatchObject({
      role: 'assistant',
      reasoning_content: FINAL_REASONING
    })
  })

  it('preserves reasoning state through generated-payload history compaction', async () => {
    writeSucceeds()
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          reasoning_content: REASONING,
          tool_calls: [toolCall('write_file', { path: 'notes.txt', content: 'Hello.' })]
        })
      )
      .mockResolvedValueOnce(completionResponse({ content: 'Wrote notes.txt.' }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('completed')
    const second = bodies(fetchMock)[1]
    const carried = second.messages.filter(
      (message: any) => message.reasoning_content !== undefined
    )
    expect(carried).toHaveLength(1)
    expect(carried[0].reasoning_content).toBe(REASONING)
    expect(carried[0].tool_calls).toBeUndefined()
  })

  it('keeps replaying reasoning state after a follow-up instruction in the same task', async () => {
    const refreshModelTarget = vi.fn(async () => ({
      model: state.task!.model,
      fingerprint: 'local:test'
    }))
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({ content: 'Ready.', reasoning_content: REASONING })
      )
      .mockResolvedValueOnce(completionResponse({ content: 'Still ready.' }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))
    await steerCoworkRun(
      state.task!.id,
      'Anything else?',
      async () => ({}),
      vi.fn(),
      refreshModelTarget
    )
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(fetchMock).toHaveBeenCalledTimes(2)
    const carried = bodies(fetchMock)[1].messages.filter(
      (message: any) => message.reasoning_content !== undefined
    )
    expect(carried).toHaveLength(1)
    expect(carried[0].reasoning_content).toBe(REASONING)
  })

  it('rejects malformed reasoning state before any action runs', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      completionResponse({
        content: null,
        reasoning_content: 42,
        tool_calls: [toolCall('write_file', { path: 'notes.txt', content: 'Hello.' })]
      })
    )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('failed')
    expect(state.task!.error).toMatch(/invalid assistant reasoning state/i)
    expect(toolMocks.execute).not.toHaveBeenCalled()
    expect(
      state.task!.agentMessages.some((message) => message.reasoning_content !== undefined)
    ).toBe(false)
  })

  it('never sends reasoning state on the text tool compatibility protocol', async () => {
    writeSucceeds()
    const envelope = (value: Record<string, unknown>): Response =>
      completionResponse({
        content: JSON.stringify({ protocol: 'morpheus-cowork-v1', ...value }),
        reasoning_content: REASONING
      })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(unsupportedNativeToolsResponse())
      .mockResolvedValueOnce(
        envelope({
          type: 'tool_call',
          name: 'write_file',
          arguments: { path: 'notes.txt', content: 'Hello.' }
        })
      )
      .mockResolvedValueOnce(envelope({ type: 'final', content: 'Wrote notes.txt.' }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('completed')
    expect(state.task!.toolProtocol).toBe('text-v1')
    for (const call of fetchMock.mock.calls) {
      expect(String(call[1]?.body)).not.toContain('reasoning_content')
    }
    expect(
      state.task!.agentMessages.some((message) => message.reasoning_content !== undefined)
    ).toBe(false)
  })
})

describe.sequential('Cowork runner transient upstream failures', () => {
  beforeEach(() => {
    state.project = makeProject()
    state.task = makeTask()
    state.replaceCalls = 0
    state.failReplaceCall = undefined
    extensionState.catalog = { projectInstructions: null, skills: [] }
    toolMocks.execute.mockReset()
    toolMocks.loadImage.mockReset()
    resetCoworkImageRefusals()
    toolMocks.loadImage.mockResolvedValue(Buffer.from('pixels'))
    visionMocks.verdict.mockReset()
    visionMocks.verdict.mockReturnValue(null)
    visionMocks.record.mockReset()
    visionMocks.run.mockClear()
    webMocks.retrieve.mockReset()
    vi.unstubAllGlobals()
  })

  const gatewayResponse = (status: number): Response =>
    new Response(
      JSON.stringify({
        providerModelError: { error: { message: 'Upstream request timed out' } },
        statusCode: status
      }),
      { status, headers: { 'Content-Type': 'application/json' } }
    )

  it.each([408, 429, 502, 503, 504])(
    'replays a completion the gateway rejected with HTTP %i',
    async (status) => {
      const fetchMock = vi
        .fn()
        .mockResolvedValueOnce(gatewayResponse(status))
        .mockResolvedValueOnce(completionResponse({ content: 'Recovered.' }))
      vi.stubGlobal('fetch', fetchMock)

      startCoworkRun(state.task!.id, async () => ({}), vi.fn())
      await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), {
        timeout: 15_000
      })

      expect(state.task!.status).toBe('completed')
      expect(fetchMock).toHaveBeenCalledTimes(2)
    },
    20_000
  )

  it('recovers a task whose work is already underway rather than discarding it', async () => {
    toolMocks.execute.mockResolvedValue({
      result: { path: 'notes.md', bytes: 4 },
      artifact: {
        path: 'notes.md',
        name: 'notes.md',
        kind: 'file',
        createdAt: 2,
        updatedAt: 2
      }
    })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [
            {
              id: 'call-1',
              type: 'function',
              function: {
                name: 'write_file',
                arguments: JSON.stringify({ path: 'notes.md', content: 'done' })
              }
            }
          ]
        })
      )
      .mockResolvedValueOnce(gatewayResponse(504))
      .mockResolvedValueOnce(completionResponse({ content: 'Wrote notes.md.' }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })

    // The tool ran before the gateway failed; the retry must not run it again.
    expect(state.task!.status).toBe('completed')
    expect(toolMocks.execute).toHaveBeenCalledTimes(1)
  }, 20_000)

  it('surfaces a deterministic rejection without replaying it', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(
        new Response(JSON.stringify({ error: 'Model not found' }), { status: 404 })
      )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })

    expect(state.task!.status).toBe('failed')
    expect(fetchMock).toHaveBeenCalledTimes(1)
  }, 20_000)

  // proxy-router reports a provider-side fault as HTTP 500 with the cause in the
  // body, so status alone cannot classify it. A single upstream stall used to
  // discard an entire multi-phase task.
  const providerFaultResponse = (message: string): Response =>
    new Response(JSON.stringify({ error: message }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' }
    })

  it.each([
    'provider request failed: provider error: failed to prompt: failed to send request: Post "http://127.0.0.1:8317/v1/chat/completions": context deadline exceeded',
    'provider request failed: dial tcp 10.0.0.4:8317: connect: connection refused',
    'provider request failed: read tcp 10.0.0.4:8317: i/o timeout'
  ])(
    'replays an HTTP 500 whose body names a transport fault: %s',
    async (message) => {
      const fetchMock = vi
        .fn()
        .mockResolvedValueOnce(providerFaultResponse(message))
        .mockResolvedValueOnce(completionResponse({ content: 'Recovered.' }))
      vi.stubGlobal('fetch', fetchMock)

      startCoworkRun(state.task!.id, async () => ({}), vi.fn())
      await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), {
        timeout: 15_000
      })

      expect(state.task!.status).toBe('completed')
      expect(fetchMock).toHaveBeenCalledTimes(2)
    },
    20_000
  )

  it('still fails fast on an HTTP 500 that describes the request itself', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(providerFaultResponse('model not found for session'))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })

    expect(state.task!.status).toBe('failed')
    expect(fetchMock).toHaveBeenCalledTimes(1)
  }, 20_000)

  it('never widens a 4xx, however its body reads', async () => {
    // A model that quotes the phrase back would otherwise buy itself retries.
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ error: 'context deadline exceeded' }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' }
      })
    )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })

    expect(state.task!.status).toBe('failed')
    expect(fetchMock).toHaveBeenCalledTimes(1)
  }, 20_000)

  it('gives up after a bounded number of attempts', async () => {
    const fetchMock = vi.fn().mockResolvedValue(gatewayResponse(504))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 20_000 })

    expect(state.task!.status).toBe('failed')
    expect(fetchMock).toHaveBeenCalledTimes(3)
  }, 25_000)

  it('replays a dead transport but never a stop request', async () => {
    const fetchMock = vi
      .fn()
      .mockRejectedValueOnce(new TypeError('fetch failed'))
      .mockResolvedValueOnce(completionResponse({ content: 'Recovered.' }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })

    expect(state.task!.status).toBe('completed')
    expect(fetchMock).toHaveBeenCalledTimes(2)

    const abortError = new Error('Aborted')
    abortError.name = 'AbortError'
    const abortingMock = vi.fn().mockRejectedValue(abortError)
    vi.stubGlobal('fetch', abortingMock)
    state.task = makeTask()

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })

    expect(abortingMock).toHaveBeenCalledTimes(1)
  }, 25_000)
})

describe.sequential('Cowork runner malformed tool turns', () => {
  beforeEach(() => {
    state.project = makeProject()
    state.task = makeTask()
    state.replaceCalls = 0
    state.failReplaceCall = undefined
    extensionState.catalog = { projectInstructions: null, skills: [] }
    toolMocks.execute.mockReset()
    toolMocks.loadImage.mockReset()
    resetCoworkImageRefusals()
    toolMocks.loadImage.mockResolvedValue(Buffer.from('pixels'))
    visionMocks.verdict.mockReset()
    visionMocks.verdict.mockReturnValue(null)
    visionMocks.record.mockReset()
    visionMocks.run.mockClear()
    webMocks.retrieve.mockReset()
    vi.unstubAllGlobals()
  })

  const readCalls = (count: number, name = 'read_file'): unknown[] =>
    Array.from({ length: count }, (_, index) => ({
      id: `call-${index + 1}`,
      type: 'function',
      function: { name, arguments: JSON.stringify({ path: `file-${index + 1}.md` }) }
    }))

  const runToCompletion = async (): Promise<void> => {
    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })
  }

  it('accepts a wide reconnaissance turn rather than capping it at a handful', async () => {
    toolMocks.execute.mockResolvedValue({ result: { content: 'x' } })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(completionResponse({ content: null, tool_calls: readCalls(16) }))
      .mockResolvedValueOnce(completionResponse({ content: 'Read them all.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    expect(toolMocks.execute).toHaveBeenCalledTimes(16)
  }, 20_000)

  it('asks the model to retry an over-wide turn instead of ending the task', async () => {
    toolMocks.execute.mockResolvedValue({ result: { content: 'x' } })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({ content: null, tool_calls: readCalls(MAX_TOOL_CALLS_PER_TURN + 1) })
      )
      .mockResolvedValueOnce(completionResponse({ content: null, tool_calls: readCalls(2) }))
      .mockResolvedValueOnce(completionResponse({ content: 'Done.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    // Nothing from the rejected turn ran, and the correction reached the model.
    expect(toolMocks.execute).toHaveBeenCalledTimes(2)
    expect(String(fetchMock.mock.calls[1]?.[1]?.body)).toContain('at most')
  }, 20_000)

  it('leaves no unanswered tool calls in the history it replays', async () => {
    toolMocks.execute.mockResolvedValue({ result: { content: 'x' } })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({ content: null, tool_calls: readCalls(MAX_TOOL_CALLS_PER_TURN + 1) })
      )
      .mockResolvedValueOnce(completionResponse({ content: 'Done.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    const replayed = JSON.parse(String(fetchMock.mock.calls[1]?.[1]?.body)).messages as Array<{
      role: string
      tool_calls?: unknown[]
    }>
    expect(replayed.some((message) => message.tool_calls?.length)).toBe(false)
    expect(state.task!.status).toBe('completed')
  }, 20_000)

  it('recovers a turn that named a tool which does not exist', async () => {
    toolMocks.execute.mockResolvedValue({ result: { content: 'x' } })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({ content: null, tool_calls: readCalls(1, 'summon_daemon') })
      )
      .mockResolvedValueOnce(completionResponse({ content: 'Used a real tool instead.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    expect(toolMocks.execute).not.toHaveBeenCalled()
    expect(String(fetchMock.mock.calls[1]?.[1]?.body)).toContain('summon_daemon')
  }, 20_000)

  it('gives up when the model keeps repeating the same malformed turn', async () => {
    // A fresh response per call: a Response body can only be read once.
    const fetchMock = vi
      .fn()
      .mockImplementation(async () =>
        completionResponse({ content: null, tool_calls: readCalls(MAX_TOOL_CALLS_PER_TURN + 1) })
      )
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('failed')
    // The first turn, then one turn per correction the runner is willing to ask for.
    expect(fetchMock).toHaveBeenCalledTimes(1 + MAX_TOOL_PROTOCOL_CORRECTIONS)
  }, 20_000)
})

describe.sequential('Cowork runner plan bookkeeping', () => {
  beforeEach(() => {
    state.project = makeProject()
    state.task = makeTask()
    state.replaceCalls = 0
    state.failReplaceCall = undefined
    extensionState.catalog = { projectInstructions: null, skills: [] }
    toolMocks.execute.mockReset()
    toolMocks.loadImage.mockReset()
    resetCoworkImageRefusals()
    toolMocks.loadImage.mockResolvedValue(Buffer.from('pixels'))
    visionMocks.verdict.mockReset()
    visionMocks.verdict.mockReturnValue(null)
    visionMocks.record.mockReset()
    visionMocks.run.mockClear()
    webMocks.retrieve.mockReset()
    vi.unstubAllGlobals()
  })

  const planCall = (id: string, name: string, args: unknown): unknown => ({
    id,
    type: 'function',
    function: { name, arguments: JSON.stringify(args) }
  })

  const runToCompletion = async (): Promise<void> => {
    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })
  }

  const toolResults = (fetchMock: ReturnType<typeof vi.fn>, callIndex: number): any[] =>
    (JSON.parse(String(fetchMock.mock.calls[callIndex]?.[1]?.body)).messages as any[])
      .filter((message) => message.role === 'tool')
      .map((message) => JSON.parse(message.content))

  it('keeps finished steps finished when the model re-plans mid-task', async () => {
    const steps = [
      { id: 'a', title: 'Read the folder' },
      { id: 'b', title: 'Write the report' }
    ]
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({ content: null, tool_calls: [planCall('c1', 'set_plan', { steps })] })
      )
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [planCall('c2', 'update_plan_step', { id: 'a', status: 'completed' })]
        })
      )
      // The re-plan resubmits the finished step with no status at all, which used
      // to reset it to pending and run the progress counter backwards.
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [
            planCall('c3', 'set_plan', {
              steps: [...steps, { id: 'c', title: 'Check the numbers' }]
            })
          ]
        })
      )
      // Saying "Done." with steps still open earns a nudge rather than an ending,
      // so the model has to keep answering until the runner gives up asking. A
      // fresh Response per call: a body can only be read once.
      .mockImplementation(async () => completionResponse({ content: 'Done.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    expect(state.task!.plan.map((step) => [step.id, step.status])).toEqual([
      ['a', 'completed'],
      ['b', 'pending'],
      ['c', 'pending']
    ])
  }, 20_000)

  it('honours an explicit status when the model deliberately reopens a step', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [planCall('c1', 'set_plan', { steps: [{ id: 'a', title: 'Draft' }] })]
        })
      )
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [planCall('c2', 'update_plan_step', { id: 'a', status: 'completed' })]
        })
      )
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [
            planCall('c3', 'set_plan', {
              steps: [{ id: 'a', title: 'Draft', status: 'in_progress' }]
            })
          ]
        })
      )
      .mockImplementation(async () => completionResponse({ content: 'Done.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.plan[0]?.status).toBe('in_progress')
  }, 20_000)

  it('reports an unknown step id back to the model instead of ending the task', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [planCall('c1', 'set_plan', { steps: [{ id: 'a', title: 'Draft' }] })]
        })
      )
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [planCall('c2', 'update_plan_step', { id: 'ghost', status: 'completed' })]
        })
      )
      .mockImplementation(async () => completionResponse({ content: 'Recovered.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    const result = toolResults(fetchMock, 2).at(-1)
    expect(result.ok).toBe(false)
    expect(result.availableStepIds).toEqual(['a'])
  }, 20_000)

  it('reports a malformed plan back to the model instead of ending the task', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [planCall('c1', 'set_plan', { steps: 'everything' })]
        })
      )
      .mockImplementation(async () => completionResponse({ content: 'Recovered.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    expect(toolResults(fetchMock, 1).at(-1).ok).toBe(false)
  }, 20_000)

  it('drops duplicate step ids so update_plan_step stays unambiguous', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [
            planCall('c1', 'set_plan', {
              steps: [
                { id: 'a', title: 'First' },
                { id: 'a', title: 'Also first' },
                { id: 'b', title: 'Second' }
              ]
            })
          ]
        })
      )
      .mockImplementation(async () => completionResponse({ content: 'Done.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.plan.map((step) => step.id)).toEqual(['a', 'b'])
    expect(state.task!.plan[0]?.title).toBe('First')
  }, 20_000)
})

describe.sequential('Cowork runner unfinished plan continuation', () => {
  beforeEach(() => {
    state.project = makeProject()
    state.task = makeTask()
    state.replaceCalls = 0
    state.failReplaceCall = undefined
    extensionState.catalog = { projectInstructions: null, skills: [] }
    toolMocks.execute.mockReset()
    toolMocks.loadImage.mockReset()
    resetCoworkImageRefusals()
    toolMocks.loadImage.mockResolvedValue(Buffer.from('pixels'))
    visionMocks.verdict.mockReset()
    visionMocks.verdict.mockReturnValue(null)
    visionMocks.record.mockReset()
    visionMocks.run.mockClear()
    webMocks.retrieve.mockReset()
    vi.unstubAllGlobals()
  })

  const planCall = (id: string, name: string, args: unknown): unknown => ({
    id,
    type: 'function',
    function: { name, arguments: JSON.stringify(args) }
  })

  const runToCompletion = async (): Promise<void> => {
    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })
  }

  const setPlanTurn = (steps: unknown[]): unknown =>
    completionResponse({ content: null, tool_calls: [planCall('c1', 'set_plan', { steps })] })

  it('asks the model to carry on when it narrates an action it never took', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        setPlanTurn([
          { id: 'a', title: 'Write the manifest' },
          { id: 'b', title: 'Check it' }
        ]) as any
      )
      // The model announces the write and then calls nothing, which used to end
      // the task and leave the user typing "continue".
      .mockResolvedValueOnce(completionResponse({ content: 'Creating it now.' }))
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [planCall('c3', 'update_plan_step', { id: 'a', status: 'completed' })]
        })
      )
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [planCall('c4', 'update_plan_step', { id: 'b', status: 'completed' })]
        })
      )
      .mockResolvedValueOnce(completionResponse({ content: 'All done.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    expect(state.task!.summary).toContain('All done.')
    expect(String(fetchMock.mock.calls[2]?.[1]?.body)).toContain('Write the manifest')
  }, 20_000)

  it('still finishes once every plan step is closed', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        setPlanTurn([{ id: 'a', title: 'Only step', status: 'completed' }]) as any
      )
      .mockResolvedValueOnce(completionResponse({ content: 'Nothing left.' }))
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    expect(fetchMock).toHaveBeenCalledTimes(2)
  }, 20_000)

  it('stops nudging a model that will not move rather than looping forever', async () => {
    let call = 0
    const fetchMock = vi.fn().mockImplementation(async () => {
      call += 1
      if (call === 1) return setPlanTurn([{ id: 'a', title: 'Stuck step' }])
      return completionResponse({ content: 'Creating it now.' })
    })
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    // One planning turn, the turn that stalled, then one turn per nudge.
    expect(fetchMock).toHaveBeenCalledTimes(2 + MAX_UNFINISHED_PLAN_NUDGES)
  }, 20_000)
})

describe.sequential('Cowork runner resilience over a long task', () => {
  beforeEach(() => {
    state.project = makeProject()
    state.task = makeTask()
    state.replaceCalls = 0
    state.failReplaceCall = undefined
    extensionState.catalog = { projectInstructions: null, skills: [] }
    toolMocks.execute.mockReset()
    toolMocks.loadImage.mockReset()
    resetCoworkImageRefusals()
    toolMocks.loadImage.mockResolvedValue(Buffer.from('pixels'))
    visionMocks.verdict.mockReset()
    visionMocks.verdict.mockReturnValue(null)
    visionMocks.record.mockReset()
    visionMocks.run.mockClear()
    webMocks.retrieve.mockReset()
    vi.unstubAllGlobals()
  })

  const call = (id: string, name: string, args: unknown): unknown => ({
    id,
    type: 'function',
    function: { name, arguments: JSON.stringify(args) }
  })

  const runToRest = async (): Promise<void> => {
    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 20_000 })
  }

  it('hands a rejected tool call back to the model instead of discarding the task', async () => {
    // Over a long run a model will eventually invent an argument. That used to
    // escape the tool handler and fail the whole task, throwing away every phase
    // already completed.
    toolMocks.execute.mockResolvedValue({ result: { path: 'a.txt', bytes: 1 } })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [call('bad', 'read_file', { path: 'a.txt', encoding: 'utf-9' })]
        }) as any
      )
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [call('good', 'read_file', { path: 'a.txt' })]
        }) as any
      )
      .mockResolvedValueOnce(completionResponse({ content: 'Recovered and finished.' }) as any)
    vi.stubGlobal('fetch', fetchMock)

    await runToRest()

    expect(state.task!.status).toBe('completed')
    // The rejection was reported as a tool result, so the turn still closed.
    const answered = state.task!.agentMessages.filter((message) => message.role === 'tool')
    expect(answered.some((message) => String(message.content).includes('"ok":false'))).toBe(true)
    expect(fetchMock).toHaveBeenCalledTimes(3)
  }, 25_000)

  it('answers every call in a batch even when one of them is rejected', async () => {
    // An unanswered tool_calls block makes every later turn unsendable.
    toolMocks.execute.mockResolvedValue({ result: { path: 'ok.txt', bytes: 1 } })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [
            call('one', 'read_file', { path: 'ok.txt' }),
            call('two', 'read_file', { path: 'ok.txt', nonsense: true }),
            call('three', 'read_file', { path: 'ok.txt' })
          ]
        }) as any
      )
      .mockResolvedValueOnce(completionResponse({ content: 'Done.' }) as any)
    vi.stubGlobal('fetch', fetchMock)

    await runToRest()

    const answeredIds = new Set(
      state
        .task!.agentMessages.filter((message) => message.role === 'tool')
        .map((message) => message.tool_call_id)
    )
    expect(answeredIds).toEqual(new Set(['one', 'two', 'three']))
    expect(state.task!.status).toBe('completed')
  }, 25_000)

  it('asks for the rest of a reply the provider cut off rather than filing the fragment', async () => {
    // finish_reason 'length' carries no tool calls because the model never got
    // to emit them. Treating that as a finished answer stored a sentence
    // fragment as the summary and stopped a task that was mid-flight.
    const truncated = (content: string): Response =>
      new Response(
        JSON.stringify({ choices: [{ message: { content }, finish_reason: 'length' }] }),
        {
          status: 200,
          headers: { 'Content-Type': 'application/json' }
        }
      )
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(truncated('I will now write the report, starting with the'))
      .mockResolvedValueOnce(completionResponse({ content: 'Finished the report.' }) as any)
    vi.stubGlobal('fetch', fetchMock)

    await runToRest()

    expect(state.task!.status).toBe('completed')
    expect(state.task!.summary).toBe('Finished the report.')
    expect(String(fetchMock.mock.calls[1]?.[1]?.body)).toContain('cut off')
  }, 25_000)

  it('gives up with a clear reason when the model never stops overrunning the limit', async () => {
    const fetchMock = vi.fn().mockImplementation(
      async () =>
        new Response(
          JSON.stringify({
            choices: [{ message: { content: 'and then' }, finish_reason: 'length' }]
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        )
    )
    vi.stubGlobal('fetch', fetchMock)

    await runToRest()

    expect(state.task!.status).toBe('failed')
    expect(state.task!.error).toMatch(/length limit/i)
    expect(fetchMock).toHaveBeenCalledTimes(1 + MAX_TRUNCATED_TURN_CONTINUATIONS)
  }, 25_000)

  it('lets a compatibility-mode model recover from answering in prose', async () => {
    // A text-v1 model that replies with prose has made the same class of mistake
    // as a native model naming a tool that does not exist, and used to be the one
    // case that ended the task outright.
    toolMocks.execute.mockResolvedValue({ result: { path: 'result.txt', bytes: 4 } })
    state.task!.toolProtocol = 'text-v1'
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(completionResponse({ content: 'Sure! Let me help with that.' }) as any)
      .mockResolvedValueOnce(
        textEnvelopeResponse({ type: 'final', content: 'Understood, and finished.' }) as any
      )
    vi.stubGlobal('fetch', fetchMock)

    await runToRest()

    expect(state.task!.status).toBe('completed')
    expect(fetchMock).toHaveBeenCalledTimes(2)
    // The correction told it what shape to reply in.
    expect(String(fetchMock.mock.calls[1]?.[1]?.body)).toContain('morpheus-cowork-v1')
  }, 25_000)

  it('carries a five-phase session through every recoverable fault to completion', async () => {
    // The individual recoveries above are each proven in isolation. This is the
    // case they exist for: one instruction, five phases, and every fault a long
    // run actually hits arriving in the same session. Each one used to end the
    // task outright, discarding every phase already on disk.
    toolMocks.execute.mockImplementation(async (_project: unknown, toolName: string) => {
      if (toolName === 'read_file') return { result: { path: 'source.csv', content: 'a,b\n1,2' } }
      if (toolName === 'create_xlsx') return { result: { path: 'report.xlsx', bytes: 4096 } }
      return { result: { path: 'notes.md', bytes: 128 } }
    })

    const truncatedTurn = (): Response =>
      new Response(
        JSON.stringify({
          choices: [
            {
              message: { content: 'Next I will build the workbook, which' },
              finish_reason: 'length'
            }
          ]
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      )

    const script: Array<() => Response> = [
      // Phase 0: plan the whole job up front.
      () =>
        completionResponse({
          content: null,
          tool_calls: [
            call('p1', 'set_plan', {
              steps: [
                { id: 's1', title: 'Read the source' },
                { id: 's2', title: 'Write notes' },
                { id: 's3', title: 'Build the workbook' },
                { id: 's4', title: 'Verify' },
                { id: 's5', title: 'Summarise' }
              ]
            })
          ]
        }) as Response,
      // Phase 1: a real read, alongside an invented argument that must not end the run.
      () =>
        completionResponse({
          content: null,
          tool_calls: [
            call('r1', 'read_file', { path: 'source.csv' }),
            call('r2', 'read_file', { path: 'source.csv', encoding: 'utf-9' })
          ]
        }) as Response,
      () =>
        completionResponse({
          content: null,
          tool_calls: [call('u1', 'update_plan_step', { id: 's1', status: 'completed' })]
        }) as Response,
      // Phase 2: a write, then a turn the provider cuts off mid-sentence.
      () =>
        completionResponse({
          content: null,
          tool_calls: [call('w1', 'write_file', { path: 'notes.md', content: '# Notes' })]
        }) as Response,
      truncatedTurn,
      () =>
        completionResponse({
          content: null,
          tool_calls: [call('u2', 'update_plan_step', { id: 's2', status: 'completed' })]
        }) as Response,
      // Phase 3: the artifact, then a turn that narrates instead of acting.
      () =>
        completionResponse({
          content: null,
          tool_calls: [
            call('x1', 'create_xlsx', {
              path: 'report.xlsx',
              title: 'Report',
              sheets: [{ name: 'Data', headers: ['a'], rows: [[1]] }]
            })
          ]
        }) as Response,
      () => completionResponse({ content: 'Building the workbook now.' }) as Response,
      () =>
        completionResponse({
          content: null,
          tool_calls: [call('u3', 'update_plan_step', { id: 's3', status: 'completed' })]
        }) as Response,
      // Phase 4: verify by reading back what was written.
      () =>
        completionResponse({
          content: null,
          tool_calls: [call('r3', 'read_file', { path: 'notes.md' })]
        }) as Response,
      () =>
        completionResponse({
          content: null,
          tool_calls: [
            call('u4', 'update_plan_step', { id: 's4', status: 'completed' }),
            call('u5', 'update_plan_step', { id: 's5', status: 'completed' })
          ]
        }) as Response,
      // Phase 5: the summary the user actually reads.
      () => completionResponse({ content: 'All five phases are complete.' }) as Response
    ]

    let turn = 0
    const fetchMock = vi
      .fn()
      .mockImplementation(async () => (script[turn++] ?? script[script.length - 1])())
    vi.stubGlobal('fetch', fetchMock)

    await runToRest()

    expect(state.task!.status).toBe('completed')
    expect(state.task!.summary).toBe('All five phases are complete.')
    // Every phase finished, and the plan says so.
    expect(state.task!.plan.map((step) => step.status)).toEqual([
      'completed',
      'completed',
      'completed',
      'completed',
      'completed'
    ])
    // The work really ran: a read, a write, and the workbook.
    const ranTools = toolMocks.execute.mock.calls.map((args: unknown[]) => args[1] as string)
    expect(ranTools).toContain('read_file')
    expect(ranTools).toContain('write_file')
    expect(ranTools).toContain('create_xlsx')
    // The one invented argument was answered rather than executed.
    expect(ranTools.filter((name: string) => name === 'read_file')).toHaveLength(2)
    // Nothing was left owing a tool result, which is what a provider rejects.
    const pending = new Set<string>()
    for (const message of state.task!.agentMessages) {
      for (const toolCall of message.tool_calls ?? []) pending.add(toolCall.id)
      if (message.role === 'tool') pending.delete(message.tool_call_id!)
    }
    expect(pending.size).toBe(0)
    expect(fetchMock).toHaveBeenCalledTimes(script.length)
  }, 30_000)
})

describe.sequential('Cowork runner unwritten-file reporting', () => {
  beforeEach(() => {
    state.project = makeProject()
    state.task = makeTask()
    state.replaceCalls = 0
    state.failReplaceCall = undefined
    extensionState.catalog = { projectInstructions: null, skills: [] }
    toolMocks.execute.mockReset()
    toolMocks.loadImage.mockReset()
    resetCoworkImageRefusals()
    toolMocks.loadImage.mockResolvedValue(Buffer.from('pixels'))
    visionMocks.verdict.mockReset()
    visionMocks.verdict.mockReturnValue(null)
    visionMocks.record.mockReset()
    visionMocks.run.mockClear()
    webMocks.retrieve.mockReset()
    vi.unstubAllGlobals()
  })

  const call = (id: string, name: string, args: unknown): unknown => ({
    id,
    type: 'function',
    function: { name, arguments: JSON.stringify(args) }
  })

  const runToCompletion = async (): Promise<void> => {
    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false), { timeout: 15_000 })
  }

  const workspaceNotices = (): string[] =>
    state
      .task!.messages.filter((message) => message.author?.kind === 'workspace')
      .map((message) => message.content)

  it('names a file the model claimed to create after the action failed', async () => {
    toolMocks.execute.mockRejectedValue(new Error('slide 3 contains unsupported key "notes"'))
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [call('c1', 'create_pptx', { path: 'reports/board-deck.pptx', slides: [] })]
        }) as any
      )
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [call('c2', 'finish_task', { summary: 'The deck is in reports/.' })]
        }) as any
      )
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    // The summary still says what the model said; the notice contradicts it.
    expect(state.task!.summary).toContain('The deck is in reports/.')
    const notices = workspaceNotices()
    expect(notices.length).toBe(1)
    expect(notices[0]).toContain('reports/board-deck.pptx')
    expect(notices[0]).toContain('unsupported key')
    expect(notices[0]).toContain('not created')
  }, 20_000)

  it('stays quiet when the model recovers and writes the same file', async () => {
    let attempt = 0
    toolMocks.execute.mockImplementation(async () => {
      attempt += 1
      if (attempt === 1) throw new Error('slide 3 contains unsupported key "notes"')
      return { result: { path: 'reports/board-deck.pptx', bytes: 4096 } }
    })
    const deckCall = (id: string): unknown =>
      call(id, 'create_pptx', { path: 'reports/board-deck.pptx', slides: [{ title: id }] })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({ content: null, tool_calls: [deckCall('c1')] }) as any
      )
      .mockResolvedValueOnce(
        completionResponse({ content: null, tool_calls: [deckCall('c2')] }) as any
      )
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [call('c3', 'finish_task', { summary: 'The deck is in reports/.' })]
        }) as any
      )
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    expect(workspaceNotices()).toEqual([])
  }, 20_000)

  it('reports the failure on a task the model ends without finish_task', async () => {
    toolMocks.execute.mockRejectedValue(new Error('the destination folder is unavailable'))
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        completionResponse({
          content: null,
          tool_calls: [call('c1', 'write_file', { path: 'notes.md', content: 'hello' })]
        }) as any
      )
      .mockResolvedValueOnce(completionResponse({ content: 'Saved the notes.' }) as any)
    vi.stubGlobal('fetch', fetchMock)

    await runToCompletion()

    expect(state.task!.status).toBe('completed')
    const notices = workspaceNotices()
    expect(notices.length).toBe(1)
    expect(notices[0]).toContain('notes.md')
    expect(notices[0]).toContain('1 file action failed')
  }, 20_000)
})

describe.sequential('Cowork runner image handling', () => {
  beforeEach(() => {
    state.project = makeProject()
    state.task = makeTask()
    state.replaceCalls = 0
    state.failReplaceCall = undefined
    extensionState.catalog = { projectInstructions: null, skills: [] }
    toolMocks.execute.mockReset()
    toolMocks.loadImage.mockReset()
    resetCoworkImageRefusals()
    toolMocks.loadImage.mockResolvedValue(Buffer.from('pixels'))
    visionMocks.verdict.mockReset()
    visionMocks.verdict.mockReturnValue(null)
    visionMocks.record.mockReset()
    visionMocks.run.mockClear()
    webMocks.retrieve.mockReset()
    vi.unstubAllGlobals()
  })

  const imageOutput = () => ({
    result: { path: 'shot.png', mediaType: 'image/png', bytes: 6, width: 2, height: 3 },
    image: {
      source: 'project' as const,
      path: 'shot.png',
      mediaType: 'image/png',
      bytes: 6,
      width: 2,
      height: 3
    }
  })

  const readImageCall = (id: string): Record<string, unknown> => ({
    id,
    type: 'function',
    function: { name: 'read_image', arguments: JSON.stringify({ path: 'shot.png' }) }
  })

  const finishCall = (id: string): Record<string, unknown> => ({
    id,
    type: 'function',
    function: { name: 'finish_task', arguments: JSON.stringify({ summary: 'Looked at it.' }) }
  })

  it('sends the pixels as a user message and never stores base64 in the transcript', async () => {
    toolMocks.execute.mockResolvedValue(imageOutput())
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(completionResponse({ tool_calls: [readImageCall('img-1')] }))
      .mockResolvedValueOnce(completionResponse({ tool_calls: [finishCall('done-1')] }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    // The stored turn keeps a reference; only the outgoing request carries bytes.
    const stored = state.task!.agentMessages.find(
      (message) => Array.isArray(message.content) && message.content.some((p) => p.type === 'image')
    )
    expect(stored).toMatchObject({ role: 'user' })
    expect(JSON.stringify(state.task!.agentMessages)).not.toContain('base64')

    const secondBody = JSON.parse(fetchMock.mock.calls[1][1].body)
    const withImage = secondBody.messages.find(
      (message: any) =>
        Array.isArray(message.content) &&
        message.content.some((part: any) => part.type === 'image_url')
    )
    expect(withImage.role).toBe('user')
    expect(withImage.content.at(-1).image_url.url).toBe(
      `data:image/png;base64,${Buffer.from('pixels').toString('base64')}`
    )
    // The reference, not the picture, is what the tool result reports.
    const toolResult = state.task!.agentMessages.find((message) => message.role === 'tool')
    expect(JSON.parse(messageTextContent(toolResult!.content))).toEqual({
      ok: true,
      result: { path: 'shot.png', mediaType: 'image/png', bytes: 6, width: 2, height: 3 }
    })
  })

  it('keeps every tool result adjacent to its call when an image arrives mid-batch', async () => {
    toolMocks.execute.mockImplementation(async (_project: unknown, name: string) =>
      name === 'read_image' ? imageOutput() : { result: { ok: true } }
    )
    const fetchMock = vi.fn().mockResolvedValueOnce(
      completionResponse({
        tool_calls: [
          readImageCall('img-1'),
          {
            id: 'list-1',
            type: 'function',
            function: { name: 'list_files', arguments: JSON.stringify({ path: '.' }) }
          },
          finishCall('done-1')
        ]
      })
    )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    const messages = state.task!.agentMessages
    for (const [index, message] of messages.entries()) {
      if (!message.tool_calls?.length) continue
      const following = messages.slice(index + 1)
      const contiguous = following.slice(
        0,
        following.findIndex((candidate) => candidate.role !== 'tool') === -1
          ? following.length
          : following.findIndex((candidate) => candidate.role !== 'tool')
      )
      const answered = contiguous.map((candidate) => candidate.tool_call_id)
      for (const call of message.tool_calls) expect(answered).toContain(call.id)
    }
  })

  it('offers the image tool whatever the model is judged capable of', async () => {
    for (const capability of ['verified', 'declared', 'detected', 'none', undefined] as const) {
      state.project = makeProject()
      state.task = makeTask()
      if (capability) state.task.model.visionCapability = capability
      const fetchMock = vi
        .fn()
        .mockResolvedValueOnce(completionResponse({ tool_calls: [finishCall('done-1')] }))
      vi.stubGlobal('fetch', fetchMock)

      startCoworkRun(state.task.id, async () => ({}), vi.fn())
      await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

      const names = JSON.parse(fetchMock.mock.calls[0][1].body).tools.map(
        (tool: any) => tool.function.name
      )
      expect(names).toContain('read_image')
      expect(names).toContain('inspect_file')
    }
  })

  it('offers the image tool even when a probe says the model is blind', async () => {
    state.task!.model.visionCapability = 'none'
    visionMocks.verdict.mockReturnValue({ sees: false })
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(completionResponse({ tool_calls: [finishCall('done-1')] }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    const names = JSON.parse(fetchMock.mock.calls[0][1].body).tools.map(
      (tool: any) => tool.function.name
    )
    expect(names).toContain('read_image')
  })

  it('states the file facts alongside the pixels so a blind model cannot invent them', async () => {
    toolMocks.execute.mockResolvedValue(imageOutput())
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(completionResponse({ tool_calls: [readImageCall('img-1')] }))
      .mockResolvedValueOnce(completionResponse({ tool_calls: [finishCall('done-1')] }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    const body = JSON.parse(fetchMock.mock.calls[1][1].body)
    const withImage = body.messages.find(
      (message: any) =>
        Array.isArray(message.content) &&
        message.content.some((part: any) => part.type === 'image_url')
    )
    const preamble = withImage.content[0].text
    expect(preamble).toContain('shot.png')
    expect(preamble).toContain('image/png')
    expect(preamble).toContain('2x3')
    expect(preamble).toContain('6 bytes')
    expect(preamble).toContain('If no picture reached you, say so')
  })

  it('replays the turn in words when the endpoint refuses image content', async () => {
    toolMocks.execute.mockResolvedValue(imageOutput())
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(completionResponse({ tool_calls: [readImageCall('img-1')] }))
      .mockResolvedValueOnce(
        errorResponse(400, JSON.stringify({ error: { message: 'image_url is not supported' } }))
      )
      .mockResolvedValueOnce(completionResponse({ tool_calls: [finishCall('done-1')] }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    // The rejected attempt carried pixels; the replay carries only prose.
    expect(
      JSON.parse(fetchMock.mock.calls[1][1].body).messages.some(
        (message: any) =>
          Array.isArray(message.content) &&
          message.content.some((part: any) => part.type === 'image_url')
      )
    ).toBe(true)
    const replay = JSON.parse(fetchMock.mock.calls[2][1].body)
    expect(JSON.stringify(replay.messages)).not.toContain('image_url')
    expect(JSON.stringify(replay.messages)).toContain('cannot receive pictures')
    // The refusal is remembered rather than advertised as vision.
    expect(visionMocks.record).toHaveBeenCalledWith(
      expect.objectContaining({ sees: false, answer: expect.stringContaining('rejected') })
    )
    expect(state.task!.status).toBe('completed')
  })

  it('does not strip pixels for a rejection that says nothing about images', async () => {
    toolMocks.execute.mockResolvedValue(imageOutput())
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(completionResponse({ tool_calls: [readImageCall('img-1')] }))
      .mockResolvedValueOnce(
        errorResponse(400, JSON.stringify({ error: { message: 'temperature must be a number' } }))
      )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(state.task!.status).toBe('failed')
  })

  it('degrades an image whose file has gone rather than failing the turn', async () => {
    toolMocks.execute.mockResolvedValue(imageOutput())
    toolMocks.loadImage.mockResolvedValue(null)
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(completionResponse({ tool_calls: [readImageCall('img-1')] }))
      .mockResolvedValueOnce(completionResponse({ tool_calls: [finishCall('done-1')] }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    const body = JSON.parse(fetchMock.mock.calls[1][1].body)
    expect(JSON.stringify(body.messages)).toContain('no longer readable')
    expect(JSON.stringify(body.messages)).not.toContain('image_url')
    expect(state.task!.status).toBe('completed')
  })
})

describe.sequential('Cowork runner vision probing', () => {
  beforeEach(() => {
    state.project = makeProject()
    state.task = makeTask()
    state.replaceCalls = 0
    state.failReplaceCall = undefined
    extensionState.catalog = { projectInstructions: null, skills: [] }
    toolMocks.execute.mockReset()
    toolMocks.loadImage.mockReset()
    resetCoworkImageRefusals()
    toolMocks.loadImage.mockResolvedValue(Buffer.from('pixels'))
    visionMocks.verdict.mockReset()
    visionMocks.verdict.mockReturnValue(null)
    visionMocks.record.mockReset()
    visionMocks.run.mockClear()
    webMocks.retrieve.mockReset()
    vi.unstubAllGlobals()
  })

  const finish = (id: string): Record<string, unknown> => ({
    id,
    type: 'function',
    function: { name: 'finish_task', arguments: JSON.stringify({ summary: 'Done.' }) }
  })

  it('probes an unknown model once and records the verdict', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(completionResponse({ tool_calls: [finish('done-1')] }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))
    await vi.waitFor(() => expect(visionMocks.record).toHaveBeenCalled())

    expect(visionMocks.run).toHaveBeenCalledTimes(1)
    expect(visionMocks.run.mock.calls[0][0]).toBe('local-test')
    expect(visionMocks.record.mock.calls[0][0]).toMatchObject({
      modelId: 'local-test',
      sees: true
    })
  })

  it('never probes a model whose verdict is already known', async () => {
    visionMocks.verdict.mockReturnValue({ sees: false })
    const fetchMock = vi
      .fn()
      .mockResolvedValue(completionResponse({ tool_calls: [finish('done-1')] }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(visionMocks.run).not.toHaveBeenCalled()
  })

  it('reports a probe verdict without letting it remove the image tool', async () => {
    for (const [capability, sees] of [
      ['detected', false],
      ['none', true]
    ] as const) {
      state.project = makeProject()
      state.task = makeTask()
      state.task.model.visionCapability = capability
      visionMocks.verdict.mockReturnValue({ sees })
      const fetchMock = vi
        .fn()
        .mockResolvedValue(completionResponse({ tool_calls: [finish('done-1')] }))
      vi.stubGlobal('fetch', fetchMock)

      startCoworkRun(state.task.id, async () => ({}), vi.fn())
      await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

      const names = JSON.parse(fetchMock.mock.calls[0][1].body).tools.map(
        (tool: any) => tool.function.name
      )
      expect(names).toContain('read_image')
    }
  })

  it('finishes the task even when probing throws', async () => {
    visionMocks.run.mockRejectedValueOnce(new Error('probe exploded'))
    const fetchMock = vi
      .fn()
      .mockResolvedValue(completionResponse({ tool_calls: [finish('done-1')] }))
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('completed')
    expect(visionMocks.record).not.toHaveBeenCalled()
  })
})
