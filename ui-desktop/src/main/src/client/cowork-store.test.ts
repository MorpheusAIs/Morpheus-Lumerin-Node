import { promises as fs } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest'
import { mutationArgumentsHash } from './cowork-mutation-journal'
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
      approvalMode: 'auto'
    })

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
      expect(dataDirectoryMode).toBe(0o700)
      expect(projectsFileMode).toBe(0o600)
    }

    await store.setTaskStatus(task.id, 'paused', { summary: 'Waiting for the next app launch.' })

    // A fresh module graph simulates a new main process reading the NeDB files.
    vi.resetModules()
    const reloadedStore = await import('./cowork-store')
    const recoveredProject = await reloadedStore.getProject(project.id)
    const recoveredTask = await reloadedStore.getTask(task.id)

    expect(recoveredProject).toMatchObject({ id: project.id, name: 'Documentation' })
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
    expect(JSON.parse(results.at(-1)!.content!)).toEqual({
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
})
