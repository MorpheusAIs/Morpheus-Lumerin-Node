import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'

const source = (relativePath: string): string =>
  readFileSync(path.join(process.cwd(), relativePath), 'utf8')

describe('session-first Cowork and Chat wiring', () => {
  it('revalidates task creation against the active wallet and never discovers local Cowork models', () => {
    const ipc = source('src/main/src/client/cowork-ipc.ts')
    const selectedModel = ipc.slice(
      ipc.indexOf('async function selectedModel'),
      ipc.indexOf('async function refreshModelTarget')
    )

    expect(ipc).toContain('const cachedModelOptions = new Map')
    expect(ipc).toContain('cachedModelOptions.get(walletAddress)')
    expect(ipc).not.toContain("proxyFetch<any[]>('/v1/models'")
    expect(ipc).toContain('loadForStableCoworkWallet(')
    expect(ipc).toContain(
      'activeCoworkMarketplaceSessions(sessions, models, Date.now(), walletAddress)'
    )
    expect(selectedModel).toContain('modelOptions(true)')
    expect(selectedModel).toContain('Cowork requires an active Morpheus marketplace session.')
  })

  it('preflights a schedule before creating its task', () => {
    const scheduledRun = source('src/main/src/client/cowork-scheduled-occurrence.ts')

    expect(
      scheduledRun.indexOf('dependencies.refreshModelTarget(schedule.task.model)')
    ).toBeGreaterThan(-1)
    expect(
      scheduledRun.indexOf('dependencies.refreshModelTarget(schedule.task.model)')
    ).toBeLessThan(scheduledRun.indexOf('const task = await dependencies.createTask'))
    expect(scheduledRun).toContain('dependencies.pauseSchedule(schedule.id)')
    expect(scheduledRun).toContain('removeUntouchedScheduledTask(task, dependencies)')
  })

  it('requires a freshly validated active session before Cowork project and file access', () => {
    const ipc = source('src/main/src/client/cowork-ipc.ts')
    const gate = ipc.slice(
      ipc.indexOf('async function requireActiveCoworkSession'),
      ipc.indexOf('function modelTarget')
    )

    expect(gate).toContain('modelOptions(true)')
    expect(gate).toContain("candidate.source === 'marketplace'")
    for (const channel of [
      'createProject',
      'updateProject',
      'deleteProject',
      'previewArtifact',
      'revealArtifact',
      'updateSchedule',
      'listExtensions',
      'configureExtensions'
    ]) {
      const start = ipc.indexOf(`handle(CHANNEL.${channel}`)
      const next = ipc.indexOf('\n  handle(CHANNEL.', start + 1)
      const handler = ipc.slice(start, next === -1 ? undefined : next)
      expect(handler, channel).toContain('await requireActiveCoworkSession()')
    }
    for (const readOnlyChannel of ['listProjects', 'listTasks', 'getTask', 'listSchedules']) {
      const start = ipc.indexOf(`handle(CHANNEL.${readOnlyChannel}`)
      const next = ipc.indexOf('\n  handle(CHANNEL.', start + 1)
      const handler = ipc.slice(start, next === -1 ? undefined : next)
      expect(handler, readOnlyChannel).not.toContain('requireActiveCoworkSession')
    }

    for (const sessionBoundChannel of ['steerTask', 'resolveApproval']) {
      const start = ipc.indexOf(`handle(CHANNEL.${sessionBoundChannel}`)
      const next = ipc.indexOf('\n  handle(CHANNEL.', start + 1)
      const handler = ipc.slice(start, next === -1 ? undefined : next)
      expect(handler, sessionBoundChannel).toContain('refreshModelTarget(task.model)')
    }
  })

  it('keeps normal Chat marketplace-only and legacy local history read-only', () => {
    const chat = source('src/renderer/src/components/chat/Chat.tsx')

    expect(chat).not.toContain('useLocalModelChat')
    expect(chat).toContain('marketplaceOnly')
    expect(chat).toContain('Legacy local-demo history is read-only.')
    expect(chat).toContain('if (isLocal) {')
    expect(chat).toContain('modelId: latestSessionModel.Id')
    expect(chat).toContain('Use this session in Cowork')
  })
})
