import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest'
import { mutationArgumentsHash } from './cowork-mutation-journal'
import { messageTextContent } from './cowork-model-history'
import type { CoworkAgentMessage } from './cowork.types'

const electron = vi.hoisted(() => ({ userData: '' }))

vi.mock('electron', () => ({
  app: {
    getPath: (name: string) => {
      if (name !== 'userData') throw new Error(`Unexpected Electron path request: ${name}`)
      return electron.userData
    }
  }
}))

let suiteDirectory: string
let projectRoot: string

beforeAll(async () => {
  suiteDirectory = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-cowork-store-'))
  electron.userData = path.join(suiteDirectory, 'user-data')
  projectRoot = path.join(suiteDirectory, 'connected-project')
  await fs.mkdir(projectRoot, { recursive: true })
})

afterAll(async () => {
  await fs.rm(suiteDirectory, { recursive: true, force: true })
})

describe.sequential('Cowork store persistence', () => {
  it('canonicalizes projects and creates a durable task record', async () => {
    const store = await import('./cowork-store')
    const initialPolicy = await store.getCoworkApprovalPolicy()
    expect(initialPolicy).toMatchObject({
      schemaVersion: 1,
      id: 'workspace',
      mode: 'manual',
      revision: 1
    })
    expect(initialPolicy).not.toHaveProperty('_id')
    const project = await store.createProject({
      name: '  Documentation  ',
      rootPath: projectRoot,
      instructions: '  Keep examples concise.  ',
      approvalMode: 'auto'
    })

    expect(project).toMatchObject({
      schemaVersion: 1,
      name: 'Documentation',
      rootPath: await fs.realpath(projectRoot),
      instructions: 'Keep examples concise.',
      approvalMode: 'manual'
    })

    const globalPolicy = await store.updateCoworkApprovalPolicy('auto', initialPolicy.revision)
    expect(globalPolicy).toMatchObject({ mode: 'auto', revision: 2 })
    expect(globalPolicy).not.toHaveProperty('_id')
    expect(await store.getProject(project.id)).toMatchObject({ approvalMode: 'auto' })
    const { coworkCollection } = await import('./cowork-database')
    expect(await coworkCollection('projects').findOneAsync({ id: project.id })).toMatchObject({
      approvalMode: 'auto'
    })
    await expect(store.updateCoworkApprovalPolicy('skip', initialPolicy.revision)).rejects.toThrow(
      /policy changed/i
    )

    const task = await store.createTask({
      projectId: project.id,
      title: '',
      goal: '  Prepare a release checklist for the desktop application.  ',
      model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
    })

    expect(task.title).toBe('Prepare a release checklist for the desktop application.')
    expect(task.status).toBe('queued')
    expect(task.revision).toBe(1)
    expect(task.messages).toHaveLength(1)
    expect(task.agentMessages).toEqual([
      { role: 'user', content: 'Prepare a release checklist for the desktop application.' }
    ])

    const displayHistory = task.messages
    const agentHistory = task.agentMessages
    store.appendDisplayMessage(task, 'assistant', 'I will prepare the checklist.')
    store.appendAgentMessage(task, {
      role: 'assistant',
      content: 'I will prepare the checklist.'
    })
    await store.replaceTask(task)
    expect(task.messages).toBe(displayHistory)
    expect(task.agentMessages).toBe(agentHistory)
    expect(task.revision).toBe(2)

    const [summary] = await store.listTaskSummaries(project.id)
    expect(summary).toEqual({
      id: task.id,
      projectId: project.id,
      title: task.title,
      status: 'queued',
      createdAt: task.createdAt,
      updatedAt: task.updatedAt
    })
    expect(summary).not.toHaveProperty('messages')
    expect(summary).not.toHaveProperty('activities')
    expect(summary).not.toHaveProperty('artifacts')

    await store.setTaskStatus(task.id, 'completed', { summary: 'Completed release checklist.' })
    expect(await store.listRecentCompletedTaskMemories(project.id, 'another-task')).toEqual([
      expect.objectContaining({
        id: task.id,
        title: task.title,
        summary: 'Completed release checklist.'
      })
    ])

    if (process.platform !== 'win32') {
      const dataDirectoryMode = (await fs.stat(path.join(electron.userData, 'Cowork'))).mode & 0o777
      const projectsFileMode =
        (await fs.stat(path.join(electron.userData, 'Cowork', 'projects.db'))).mode & 0o777
      const preferencesFileMode =
        (await fs.stat(path.join(electron.userData, 'Cowork', 'preferences.db'))).mode & 0o777
      expect(dataDirectoryMode).toBe(0o700)
      expect(projectsFileMode).toBe(0o600)
      expect(preferencesFileMode).toBe(0o600)
    }

    await store.setTaskStatus(task.id, 'paused', { summary: 'Waiting for the next app launch.' })

    // A fresh module graph simulates a new main process reading the NeDB files.
    vi.resetModules()
    const reloadedStore = await import('./cowork-store')
    const recoveredProject = await reloadedStore.getProject(project.id)
    const recoveredTask = await reloadedStore.getTask(task.id)
    const recoveredPolicy = await reloadedStore.getCoworkApprovalPolicy()

    expect(recoveredProject).toMatchObject({
      id: project.id,
      name: 'Documentation',
      approvalMode: 'auto'
    })
    expect(recoveredPolicy).toMatchObject({ mode: 'auto', revision: 2 })
    expect(recoveredPolicy).not.toHaveProperty('_id')
    expect(recoveredTask).toMatchObject({
      id: task.id,
      projectId: project.id,
      status: 'paused',
      summary: 'Waiting for the next app launch.',
      revision: 4
    })
  })

  it('keeps the complete display transcript while the task record stays bounded', async () => {
    const store = await import('./cowork-store')
    const [project] = await store.listProjects()
    const task = await store.createTask({
      projectId: project.id,
      title: 'Long-lived workspace history',
      goal: 'Keep every visible turn available.',
      model: { modelId: 'model-a', modelName: 'Model A', isLocal: true }
    })

    for (let index = 0; index < 130; index++) {
      store.appendDisplayMessage(task, 'assistant', `Visible answer ${index + 1}`, {
        kind: 'model',
        modelId: 'model-a',
        modelName: 'Model A'
      })
    }
    await store.replaceTask(task)

    expect(task.messages).toHaveLength(100)
    expect(task.hasEarlierMessages).toBe(true)
    const latest = await store.listTaskMessages(task.id, { limit: 100 })
    const earlier = await store.listTaskMessages(task.id, {
      beforeSequence: latest.nextBeforeSequence,
      limit: 100
    })
    expect(latest.hasMore).toBe(true)
    expect(earlier.hasMore).toBe(false)
    expect([...earlier.messages, ...latest.messages]).toHaveLength(131)
    expect(latest.messages.at(-1)).toMatchObject({
      content: 'Visible answer 130',
      author: { kind: 'model', modelName: 'Model A' }
    })

    await store.deleteTask(task.id)
    expect(await store.listTaskMessages(task.id)).toMatchObject({ messages: [], hasMore: false })
  })

  it('does not persist transcript messages from a task revision that loses the save CAS', async () => {
    const store = await import('./cowork-store')
    const project = await store.createProject({
      name: 'Concurrent transcript saves',
      rootPath: projectRoot,
      approvalMode: 'auto'
    })
    const created = await store.createTask({
      projectId: project.id,
      title: 'Keep the winning transcript only',
      goal: 'Resolve two concurrent task saves.',
      model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
    })
    const first = (await store.getTask(created.id))!
    const stale = (await store.getTask(created.id))!
    store.appendDisplayMessage(first, 'assistant', 'accepted revision message')
    store.appendDisplayMessage(stale, 'assistant', 'rejected stale revision message')

    const [accepted, rejected] = await Promise.allSettled([
      store.replaceTask(first),
      store.replaceTask(stale)
    ])

    expect(accepted.status).toBe('fulfilled')
    expect(rejected).toMatchObject({
      status: 'rejected',
      reason: expect.objectContaining({
        message: expect.stringMatching(/changed in another operation/)
      })
    })
    const transcript = await store.listTaskMessages(created.id, { limit: 100 })
    expect(transcript.messages.map((message) => message.content)).toEqual([
      'Resolve two concurrent task saves.',
      'accepted revision message'
    ])

    await store.deleteTask(created.id)
    await store.deleteProject(project.id)
  })

  it('soft-archives projects without deleting their task history', async () => {
    const store = await import('./cowork-store')
    const [project] = await store.listProjects()
    const [task] = await store.listTasks(project.id)

    await store.deleteProject(project.id)

    expect(await store.listProjects()).toEqual([])
    expect(await store.getProject(project.id)).toMatchObject({
      id: project.id,
      archivedAt: expect.any(Number)
    })
    expect(await store.getTask(task.id)).toMatchObject({ id: task.id, projectId: project.id })
  })

  it('recomputes bounded history size after large write arguments are scrubbed', async () => {
    const store = await import('./cowork-store')
    const project = await store.createProject({
      name: 'History cache',
      rootPath: projectRoot,
      approvalMode: 'auto'
    })
    const task = await store.createTask({
      projectId: project.id,
      title: 'History accounting',
      goal: 'Create several bounded files.',
      model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
    })

    for (let index = 0; index < 7; index++) {
      const call = {
        id: `large-write-${index}`,
        type: 'function' as const,
        function: {
          name: 'write_file',
          arguments: JSON.stringify({
            path: `file-${index}.txt`,
            content: 'x'.repeat(1_900_000)
          })
        }
      }
      expect(() =>
        store.appendAgentMessage(task, { role: 'assistant', content: null, tool_calls: [call] })
      ).not.toThrow()
      store.scrubCoworkToolArguments(task, call)
    }

    expect(task.agentMessages).toHaveLength(8)
    expect(JSON.parse(task.agentMessages.at(-1)!.tool_calls![0].function.arguments).content).toBe(
      '[omitted after execution: 1900000 characters]'
    )
  })

  it('keeps a rebound model context boundary aligned when old history is trimmed', async () => {
    const store = await import('./cowork-store')
    const project = await store.createProject({
      name: 'Rebound context boundary',
      rootPath: projectRoot,
      approvalMode: 'auto'
    })
    const task = await store.createTask({
      projectId: project.id,
      title: 'Continue with a replacement model',
      goal: 'Preserve only the replacement model context.',
      model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
    })
    task.agentMessages = Array.from(
      { length: 500 },
      (_, index): CoworkAgentMessage => ({
        role: index === 0 || index === 250 ? 'user' : 'assistant',
        content: `old-context-${index}`
      })
    )
    task.modelContextStart = task.agentMessages.length

    store.appendAgentMessage(task, {
      role: 'user',
      content: 'Continue from the bounded replacement-model handoff.'
    })

    expect(task.agentMessages).toHaveLength(251)
    expect(task.modelContextStart).toBe(250)
    expect(task.agentMessages.slice(task.modelContextStart)).toEqual([
      { role: 'user', content: 'Continue from the bounded replacement-model handoff.' }
    ])
  })

  it('fails closed instead of replaying a prepared mutation after app restart', async () => {
    const store = await import('./cowork-store')
    const project = await store.createProject({
      name: 'Recovery safety',
      rootPath: projectRoot,
      approvalMode: 'auto'
    })
    const task = await store.createTask({
      projectId: project.id,
      title: 'Interrupted write',
      goal: 'Write one file.',
      model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
    })
    const callId = 'prepared-write'
    // A runner error can mark the task failed while the last durable journal
    // entry is still prepared; startup recovery must scan beyond running tasks.
    task.status = 'failed'
    task.agentMessages.push({
      role: 'assistant',
      content: null,
      tool_calls: [
        {
          id: callId,
          type: 'function',
          function: {
            name: 'write_file',
            arguments: JSON.stringify({
              path: 'uncertain.txt',
              content: 'sensitive generated body'
            })
          }
        }
      ]
    })
    task.toolExecutions = [
      {
        toolCallId: callId,
        toolName: 'write_file',
        status: 'prepared',
        argumentsHash: 'a'.repeat(64),
        preparedAt: Date.now()
      }
    ]
    await store.replaceTask(task)

    expect(await store.recoverInterruptedTasks()).toBe(1)
    const recovered = await store.getTask(task.id)

    expect(recovered).toMatchObject({
      status: 'paused',
      error: expect.stringMatching(/did not repeat potentially ambiguous actions/i),
      toolExecutions: [
        expect.objectContaining({
          toolCallId: callId,
          status: 'ambiguous',
          resultMessage: expect.stringContaining('did not repeat it')
        })
      ]
    })
    expect(
      recovered!.agentMessages.filter(
        (message) => message.role === 'tool' && message.tool_call_id === callId
      )
    ).toHaveLength(1)
    const recoveredCall = recovered!.agentMessages
      .find((message) => message.role === 'assistant')
      ?.tool_calls?.find((call) => call.id === callId)
    expect(JSON.parse(recoveredCall!.function.arguments)).toEqual({
      path: 'uncertain.txt',
      content: '[omitted after execution: 24 characters]'
    })
  })

  it('does not attach an old success result to a changed call that reused its ID', async () => {
    const store = await import('./cowork-store')
    const project = await store.createProject({
      name: 'Reused ID recovery',
      rootPath: projectRoot,
      approvalMode: 'skip'
    })
    const task = await store.createTask({
      projectId: project.id,
      title: 'Changed replay',
      goal: 'Create a file.',
      model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
    })
    const callId = 'reused-write-id'
    const oldInput = { path: 'old.txt', content: 'old' }
    const newInput = { path: 'new.txt', content: 'new' }
    task.status = 'running'
    task.agentMessages.push(
      {
        role: 'assistant',
        content: null,
        tool_calls: [
          {
            id: callId,
            type: 'function',
            function: { name: 'write_file', arguments: JSON.stringify(oldInput) }
          }
        ]
      },
      { role: 'tool', tool_call_id: callId, content: JSON.stringify({ ok: true }) },
      {
        role: 'assistant',
        content: null,
        tool_calls: [
          {
            id: callId,
            type: 'function',
            function: { name: 'write_file', arguments: JSON.stringify(newInput) }
          }
        ]
      }
    )
    task.toolExecutions = [
      {
        toolCallId: callId,
        toolName: 'write_file',
        status: 'succeeded',
        argumentsHash: mutationArgumentsHash('write_file', oldInput),
        resultMessage: JSON.stringify({ ok: true, result: { path: 'old.txt' } }),
        preparedAt: 2,
        completedAt: 3
      }
    ]
    await store.replaceTask(task)

    expect(await store.recoverInterruptedTasks()).toBe(1)
    const recovered = (await store.getTask(task.id))!
    const results = recovered.agentMessages.filter(
      (message) => message.role === 'tool' && message.tool_call_id === callId
    )

    expect(results).toHaveLength(2)
    expect(JSON.parse(messageTextContent(results.at(-1)!.content))).toEqual({
      ok: false,
      error:
        'Workspace blocked a reused tool call ID whose action name or arguments had changed. No file action ran.'
    })
  })

  it('rejects invalid project folders, missing projects, and blank goals', async () => {
    const store = await import('./cowork-store')
    const missingPath = path.join(suiteDirectory, 'missing')

    await expect(store.createProject({ name: 'Missing', rootPath: missingPath })).rejects.toThrow()
    await expect(store.createProject({ name: '   ', rootPath: projectRoot })).rejects.toThrow(
      /name is required/
    )
    await expect(
      store.createTask({
        projectId: 'not-a-project',
        title: 'No project',
        goal: 'Do something',
        model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
      })
    ).rejects.toThrow(/project not found/)

    const archivedProject = (await store.getProject((await store.listTasks())[0].projectId))!
    await expect(
      store.createTask({
        projectId: archivedProject.id,
        title: 'Blank goal',
        goal: '   ',
        model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
      })
    ).rejects.toThrow(/outcome you want/)
  })

  it('fails closed when a persisted Workspace approval policy is invalid', async () => {
    const { coworkCollection } = await import('./cowork-database')
    await coworkCollection('preferences').updateAsync(
      { id: 'workspace' },
      { $set: { mode: 'unexpected-mode' } },
      {}
    )

    vi.resetModules()
    const reloadedStore = await import('./cowork-store')
    const policy = await reloadedStore.getCoworkApprovalPolicy()

    expect(policy).toMatchObject({
      schemaVersion: 1,
      id: 'workspace',
      mode: 'manual',
      revision: 3
    })
    expect(policy).not.toHaveProperty('_id')
  })

  it('normalizes duplicate policies to the strictest mode at the highest revision', async () => {
    const { coworkCollection } = await import('./cowork-database')
    await coworkCollection('preferences').insertAsync({
      schemaVersion: 1,
      id: 'workspace',
      mode: 'skip',
      revision: 7,
      updatedAt: Date.now() + 1_000
    })
    await coworkCollection('preferences').insertAsync({
      schemaVersion: 1,
      id: 'workspace',
      mode: 'manual',
      revision: 7,
      updatedAt: Date.now()
    })

    vi.resetModules()
    const reloadedStore = await import('./cowork-store')
    const policy = await reloadedStore.getCoworkApprovalPolicy()
    const reloadedDatabase = await import('./cowork-database')
    const policyRows = await reloadedDatabase.coworkCollection('preferences').findAsync({
      id: 'workspace'
    })
    const projectRows = await reloadedDatabase.coworkCollection('projects').findAsync({})

    expect(policy).toMatchObject({ mode: 'manual', revision: 7 })
    expect(policy).not.toHaveProperty('_id')
    expect(policyRows).toHaveLength(1)
    expect(projectRows.every((project) => project.approvalMode === 'manual')).toBe(true)
  })

  it('repairs a malformed higher revision without reusing CAS history', async () => {
    const { coworkCollection } = await import('./cowork-database')
    await coworkCollection('preferences').insertAsync({
      schemaVersion: 1,
      id: 'workspace',
      mode: 'invalid',
      revision: 9,
      updatedAt: Date.now()
    })

    vi.resetModules()
    const reloadedStore = await import('./cowork-store')
    const policy = await reloadedStore.getCoworkApprovalPolicy()
    const reloadedDatabase = await import('./cowork-database')
    const rows = await reloadedDatabase.coworkCollection('preferences').findAsync({
      id: 'workspace'
    })

    expect(policy).toMatchObject({ mode: 'manual', revision: 10 })
    expect(rows).toHaveLength(1)
  })

  it('preserves opaque provider reasoning state across save and reload', async () => {
    const store = await import('./cowork-store')
    const project = await store.createProject({
      name: 'Reasoning continuity',
      rootPath: projectRoot,
      approvalMode: 'auto'
    })
    const task = await store.createTask({
      projectId: project.id,
      title: 'Thinking-mode continuation',
      goal: 'Keep provider thinking state exact.',
      model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
    })
    const reasoning = 'step 1: read the file\nstep 2: write \u201cnotes.txt\u201d exactly once'
    const message: CoworkAgentMessage = {
      role: 'assistant',
      content: null,
      reasoning_content: reasoning,
      tool_calls: [
        {
          id: 'reasoning-call-1',
          type: 'function',
          function: { name: 'read_file', arguments: JSON.stringify({ path: 'notes.txt' }) }
        }
      ]
    }
    store.appendAgentMessage(task, message)
    await store.replaceTask(task)

    const recovered = (await store.getTask(task.id))!
    expect(recovered.agentMessages.at(-1)).toEqual(message)
    expect(recovered.agentMessages.at(-1)!.reasoning_content).toBe(reasoning)
  })

  it('trims a long single-instruction history instead of failing the task', async () => {
    // A multi-phase run produces one user message and then hundreds of
    // assistant/tool pairs. Trimming used to insist on a second user message to
    // cut at, so a long task threw here and lost everything it had built.
    const store = await import('./cowork-store')
    const task = await store.createTask({
      projectId: (await store.createProject({ name: 'Trim', rootPath: projectRoot })).id,
      title: 'Long build',
      goal: 'Build a multi-phase project.',
      model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
    })
    // createTask already seeds the goal as message 0. That single user message is
    // the whole point: a multi-phase run never produces a second one.
    expect(task.agentMessages).toEqual([{ role: 'user', content: 'Build a multi-phase project.' }])
    expect(() => {
      for (let turn = 0; turn < 400; turn += 1) {
        store.appendAgentMessage(task, {
          role: 'assistant',
          content: null,
          tool_calls: [
            {
              id: `call-${turn}`,
              type: 'function',
              function: { name: 'read_file', arguments: JSON.stringify({ path: `f${turn}.txt` }) }
            }
          ]
        })
        store.appendAgentMessage(task, {
          role: 'tool',
          tool_call_id: `call-${turn}`,
          content: JSON.stringify({ ok: true })
        })
      }
    }).not.toThrow()

    expect(task.agentMessages.length).toBeLessThanOrEqual(500)
    // The statement of what the task is for outlives the middle of the run.
    expect(task.agentMessages[0]).toMatchObject({
      role: 'user',
      content: 'Build a multi-phase project.'
    })
    // Every retained tool result still answers a call the model can see, which
    // is what providers reject a history for.
    const visible = new Set<string>()
    for (const message of task.agentMessages) {
      for (const call of message.tool_calls ?? []) visible.add(call.id)
      if (message.role === 'tool') expect(visible.has(message.tool_call_id!)).toBe(true)
    }
  })

  it('keeps modelContextStart pointing at a real turn boundary after a trim', async () => {
    const store = await import('./cowork-store')
    const task = await store.createTask({
      projectId: (await store.createProject({ name: 'Trim rebind', rootPath: projectRoot })).id,
      title: 'Rebound long build',
      goal: 'Build a multi-phase project.',
      model: { modelId: 'local-test', modelName: 'Local test model', isLocal: true }
    })
    task.modelContextStart = 1
    for (let turn = 0; turn < 400; turn += 1) {
      store.appendAgentMessage(task, {
        role: 'assistant',
        content: null,
        tool_calls: [
          {
            id: `c${turn}`,
            type: 'function',
            function: { name: 'read_file', arguments: '{}' }
          }
        ]
      })
      store.appendAgentMessage(task, {
        role: 'tool',
        tool_call_id: `c${turn}`,
        content: '{"ok":true}'
      })
    }

    expect(task.modelContextStart).toBeGreaterThanOrEqual(1)
    expect(task.modelContextStart).toBeLessThan(task.agentMessages.length)
    const boundary = task.agentMessages[task.modelContextStart!]
    expect(boundary.role === 'user' || boundary.role === 'assistant').toBe(true)
  })

  it('refuses to overwrite approval settings from a newer schema', async () => {
    const { coworkCollection } = await import('./cowork-database')
    const future = {
      schemaVersion: 2,
      id: 'workspace',
      mode: 'skip',
      revision: 11,
      updatedAt: Date.now()
    }
    await coworkCollection('preferences').insertAsync(future)

    vi.resetModules()
    const reloadedStore = await import('./cowork-store')
    await expect(reloadedStore.getCoworkApprovalPolicy()).rejects.toThrow(/newer app version/i)
    const reloadedDatabase = await import('./cowork-database')
    const rows = await reloadedDatabase.coworkCollection('preferences').findAsync({
      id: 'workspace',
      schemaVersion: 2
    })

    expect(rows).toHaveLength(1)
    expect(rows[0]).toMatchObject(future)
  })
})
