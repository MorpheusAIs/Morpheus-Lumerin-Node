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
    expect(selectedModel).toContain('Workspace requires an active Morpheus marketplace session.')
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

  it('requires a fresh session for mutations while preserving read-only Workspace access', () => {
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
      'updateSchedule',
      'configureExtensions'
    ]) {
      const start = ipc.indexOf(`handle(CHANNEL.${channel}`)
      const next = ipc.indexOf('\n  handle(CHANNEL.', start + 1)
      const handler = ipc.slice(start, next === -1 ? undefined : next)
      expect(handler, channel).toContain('await requireActiveCoworkSession()')
    }
    for (const readOnlyChannel of [
      'listProjects',
      'listTasks',
      'getTask',
      'listTaskMessages',
      'previewArtifact',
      'revealArtifact',
      'listSchedules',
      'listExtensions'
    ]) {
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
    const rebindStart = ipc.indexOf('handle(CHANNEL.rebindTask')
    const rebindEnd = ipc.indexOf('\n  handle(CHANNEL.', rebindStart + 1)
    expect(ipc.slice(rebindStart, rebindEnd)).toContain('selectedModelBinding(input.model)')
  })

  it('keeps normal Chat marketplace-only and legacy local history read-only', () => {
    const chat = source('src/renderer/src/components/chat/Chat.tsx')

    expect(chat).not.toContain('useLocalModelChat')
    expect(chat).toContain('marketplaceOnly')
    expect(chat).toContain('Legacy local-demo history is read-only.')
    expect(chat).toContain('if (isLocal) {')
    expect(chat).toContain('modelId: latestSessionModel.Id')
    expect(chat).toContain('Use this session in Workspace')
  })

  it('keeps Chat header actions contained as the route narrows', () => {
    const chat = source('src/renderer/src/components/chat/Chat.tsx')
    const styles = source('src/renderer/src/components/chat/Chat.styles.tsx')
    const css = source('src/renderer/src/components/chat/Chat.css')
    const headerStyles = styles.slice(
      styles.indexOf('export const ChatHeaderControls'),
      styles.indexOf('export const Title =')
    )

    expect(chat).toContain('<ChatHeaderControls className="chat-header-controls">')
    expect(chat).toContain('<ChatHeaderActions className="chat-header-actions">')
    expect(chat).toContain('aria-label="Use this session in Workspace"')
    expect(chat).toContain('<span>Workspace</span>')
    expect(headerStyles).toContain('grid-template-columns: minmax(0, 35rem) max-content')
    expect(headerStyles).toContain('white-space: nowrap')
    expect(css).not.toContain('.change-modal')
    expect(css).toContain('@container route-space (max-width: 900px)')
    expect(css).toContain('@container route-space (max-width: 520px)')
    expect(css).toContain('.chat-header-controls.chat-header-controls')
    expect(css).toContain('.chat-header-actions.chat-header-actions')
    expect(css).toContain('.chat-header-action.chat-header-action')
    expect(css).toContain('min-height: 4.8rem')
    expect(css).toContain('text-overflow: ellipsis')
  })

  it('keeps compact Chat warnings and session checkout controls reachable', () => {
    const chat = source('src/renderer/src/components/chat/Chat.tsx')
    const styles = source('src/renderer/src/components/chat/Chat.styles.tsx')
    const liveChat = chat.slice(chat.indexOf('return (\n    <>'))
    const introStyles = styles.slice(
      styles.indexOf('export const ChatIntroContainer'),
      styles.indexOf('export const ChatIntroInnerTitle')
    )

    expect(liveChat.indexOf('<View>')).toBeLessThan(
      liveChat.indexOf('what="the latest model data"')
    )
    expect(introStyles).toContain('align-items: flex-start')
    expect(introStyles).toContain('overflow-y: auto')
    expect(introStyles).toContain('margin: auto')
    expect(introStyles).toContain('max-width: 100%')
    expect(introStyles).toContain('padding: clamp(')
  })

  it('keeps historical model bindings in the main process', () => {
    const ipc = source('src/main/src/client/cowork-ipc.ts')
    const projection = ipc.slice(
      ipc.indexOf('function publicTask('),
      ipc.indexOf('function visionCapability')
    )

    expect(projection).toContain("| 'modelBindings'")
    expect(projection).toContain('modelBindings: _modelBindings')
  })
})
