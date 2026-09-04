import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { CoworkProject, CoworkTask, CoworkToolCall } from './cowork.types'
import { mutationArgumentsHash } from './cowork-mutation-journal'

const state = vi.hoisted(() => ({
  project: undefined as CoworkProject | undefined,
  task: undefined as CoworkTask | undefined,
  replaceCalls: 0,
  failReplaceCall: undefined as number | undefined
}))

const toolMocks = vi.hoisted(() => ({
  execute: vi.fn()
}))

const webMocks = vi.hoisted(() => ({
  retrieve: vi.fn()
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

import {
  cancelCoworkRun,
  coworkRunActive,
  deleteCoworkTask,
  pauseCoworkRun,
  rebindCoworkTask,
  requestCoworkStart,
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
    const fetchMock = vi.fn().mockResolvedValue(
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
    expect(toolMocks.execute).not.toHaveBeenCalled()
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

  it('fails closed when the compatibility response contains fenced JSON', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(unsupportedNativeToolsResponse())
      .mockResolvedValueOnce(
        completionResponse({
          content:
            '```json\n{"protocol":"morpheus-cowork-v1","type":"tool_call","name":"write_file","arguments":{"path":"unsafe.txt","content":"no"}}\n```'
        })
      )
    vi.stubGlobal('fetch', fetchMock)

    startCoworkRun(state.task!.id, async () => ({}), vi.fn())
    await vi.waitFor(() => expect(coworkRunActive(state.task!.id)).toBe(false))

    expect(state.task!.status).toBe('failed')
    expect(state.task!.error).toMatch(/compatibility protocol/i)
    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(toolMocks.execute).not.toHaveBeenCalled()
  })

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
      JSON.parse(toolResponses.find((message) => message.tool_call_id === secondCall.id)!.content!)
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
