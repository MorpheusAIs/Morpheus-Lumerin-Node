import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'

const source = (relativePath: string): string =>
  readFileSync(path.join(process.cwd(), relativePath), 'utf8')

describe('legacy renderer bridge security boundaries', () => {
  it('does not auto-open DevTools and disables them in packaged builds', () => {
    const main = source('src/main/index.ts')

    expect(main).not.toContain('.openDevTools(')
    expect(main).toContain('devTools: is.dev')
  })

  it('registers bootstrap IPC before loading the renderer and follow-up IPC before replying', () => {
    const main = source('src/main/index.ts')
    const readyBlock = main.slice(main.indexOf('app\n  .whenReady()'))
    const client = source('src/main/src/client/index.js')

    expect(readyBlock.indexOf('createClient(config)')).toBeGreaterThan(-1)
    expect(readyBlock.indexOf('createClient(config)')).toBeLessThan(
      readyBlock.indexOf('\n    createWindow()')
    )
    expect(client.indexOf('subscriptions.subscribe(core)')).toBeGreaterThan(-1)
    expect(client.indexOf('subscriptions.subscribe(core)')).toBeLessThan(
      client.indexOf("webContent.sender.send('ui-ready', payload)")
    )
  })

  it('never exposes proxy-router Basic auth credentials to the renderer', () => {
    const preload = source('src/preload/index.ts')
    const listeners = source('src/main/src/client/subscriptions/index.ts')
    const rendererClient = source('src/renderer/src/client/index.ts')

    expect(preload).not.toContain('get-auth-headers')
    expect(listeners).not.toContain('get-auth-headers')
    expect(rendererClient).not.toContain('get-auth-headers')
    expect(rendererClient).not.toContain('getAuthHeaders')
    expect(rendererClient).not.toMatch(/fetch\s*\(/u)
  })

  it('streams chat through fixed, bounded IPC without exposing Electron events', () => {
    const preload = source('src/preload/index.ts')
    const listeners = source('src/main/src/client/subscriptions/index.ts')
    const rendererClient = source('src/renderer/src/client/index.ts')
    const streamMain = source('src/main/src/client/chat-stream-ipc.ts')

    expect(listeners).not.toContain("'chat-completion'")
    expect(preload).not.toContain("'chat-completion'")
    expect(preload).toContain("start: 'chat-stream:start'")
    expect(preload).toContain("cancel: 'chat-stream:cancel'")
    expect(preload).toContain("event: 'chat-stream:event'")
    expect(preload).toContain("exposeInMainWorld('chatStream', chatStream)")
    expect(preload).toContain('payload.dataBase64.length <= 87_384')
    expect(preload).toContain('if (safePayload) listener(safePayload)')
    expect(rendererClient).toContain('window.chatStream.start(requestId, payload)')
    expect(rendererClient).toContain('window.chatStream.cancel(requestId)')
    expect(streamMain).toContain('isTrustedRendererEvent(event)')
    expect(streamMain).toContain('activePerRenderer: 4')
    expect(streamMain).toContain('responseBytes: 8 * 1024 * 1024')
    expect(streamMain).toContain('ipcEvents: 16_384')
    expect(streamMain).toContain("controller.abort('Chat stream cancelled.')")
    expect(streamMain).toContain("event.sender.once('destroyed', abortIfDestroyed)")
    expect(streamMain).toContain("event.sender.on('did-start-navigation', abortIfNavigating)")
    expect(streamMain).toContain("removeListener('did-start-navigation', abortIfNavigating)")
    expect(streamMain).not.toContain('get-auth-headers')
  })

  it('uses bounded binary IPC for large audio and document payloads', () => {
    const handlers = source('src/main/src/client/subscriptions/handlers.ts')
    const attachments = source('src/main/src/client/attachments.ts')
    const chat = source('src/renderer/src/components/chat/Chat.tsx')

    expect(handlers).toContain('data: ArrayBuffer | ArrayBufferView')
    expect(handlers).toContain('audio.length > 20 * 1024 * 1024')
    expect(handlers).not.toContain("body.toString('base64')")
    expect(attachments).toContain('data: ArrayBuffer | ArrayBufferView')
    expect(attachments).toContain('buf.length > MAX_ATTACHMENT_INPUT_BYTES')
    expect(chat).toContain('data: await file.arrayBuffer()')
    expect(chat).not.toContain('dataBase64')
    expect(chat).not.toContain('window.atob(')
    expect(chat).not.toContain('window.btoa(')
  })

  it('keeps cancellable IPFS downloads behind scoped folder grants and fixed IPC', () => {
    const preload = source('src/preload/index.ts')
    const listeners = source('src/main/src/client/subscriptions/index.ts')
    const rendererClient = source('src/renderer/src/client/index.ts')
    const downloadMain = source('src/main/src/client/ipfs-download-ipc.ts')
    const modelsTable = source('src/renderer/src/components/models/ModelsTable.tsx')

    expect(listeners).not.toContain("'get-ipfs-file'")
    expect(listeners).not.toContain("'open-select-folder-dialog'")
    expect(preload).not.toContain("'get-ipfs-file'")
    expect(preload).toContain("selectFolder: 'ipfs-download:select-folder'")
    expect(preload).toContain("start: 'ipfs-download:start'")
    expect(preload).toContain("cancel: 'ipfs-download:cancel'")
    expect(preload).toContain("event: 'ipfs-download:event'")
    expect(preload).toContain("exposeInMainWorld('ipfsDownload', ipfsDownload)")
    expect(rendererClient).toContain('window.ipfsDownload.start(requestId, folderToken, cidHash)')
    expect(rendererClient).toContain('window.ipfsDownload.cancel(requestId)')
    expect(modelsTable).toContain('client.cancelIpfsDownload({ requestId })')
    expect(downloadMain).toContain('isTrustedRendererEvent(event)')
    expect(downloadMain).toContain('activePerRenderer: 2')
    expect(downloadMain).toContain('modelBytes: 256 * 1024 * 1024 * 1024')
    expect(downloadMain).toContain('for (const candidate of [finalPath, partialPath])')
    expect(downloadMain).toContain('await finalizeIpfsDownload(partialPath, finalPath')
    expect(downloadMain).toContain('void fs.unlink(partialPath)')
    expect(downloadMain).not.toContain('return { accepted: true, destinationPath }')
    expect(downloadMain).not.toContain('return { canceled: false, filePaths: [folder]')
    expect(downloadMain).toContain("event.sender.once('destroyed', abortIfDestroyed)")
    expect(downloadMain).toContain("event.sender.on('did-start-navigation', abortIfNavigating)")
    expect(downloadMain).toContain('configuredLoopbackProxyUrl()')
    expect(downloadMain).not.toContain('get-auth-headers')
  })

  it('registers every renderer-forwarded request in both preload and main', () => {
    const preload = source('src/preload/index.ts')
    const listeners = source('src/main/src/client/subscriptions/index.ts')
    const rendererClient = source('src/renderer/src/client/index.ts')
    const forwarded = [
      ...rendererClient.matchAll(/forwardToMainProcess\(\s*['"]([^'"]+)['"]/gu)
    ].map((match) => match[1])

    expect(forwarded.length).toBeGreaterThan(30)
    for (const channel of forwarded) {
      expect(preload, `${channel} is missing from the preload allowlist`).toContain(`'${channel}'`)
      expect(listeners, `${channel} is missing from the main listener map`).toMatch(
        new RegExp(`['"]?${channel.replace(/[.*+?^${}()|[\]\\]/gu, '\\$&')}['"]?\\s*:`)
      )
    }
  })

  it('keeps the raw Electron event object behind the context bridge', () => {
    const preload = source('src/preload/index.ts')

    expect(preload).toContain('listener(payload, unsubscribe)')
    expect(preload).not.toContain('listener(event, payload')
    expect(preload).not.toContain('listener(_event, payload')
    expect(preload).toContain('const safeEvent = sanitizedCoworkTaskEvent(payload)')
    expect(preload).toContain('if (safeEvent) listener(safeEvent)')
    expect(preload).toContain('payload.title.length > 240')
    expect(preload).toContain('coworkTaskStatuses.has(payload.status)')
  })

  it('uses only the configured loopback proxy for wallet onboarding', () => {
    const handlers = source('src/main/src/client/subscriptions/handlers.ts')
    const start = handlers.indexOf('export const onboardingCompleted')
    const end = handlers.indexOf('\nexport const ', start + 1)
    const onboarding = handlers.slice(start, end < 0 ? handlers.length : end)

    expect(onboarding).toContain('configuredLoopbackProxyUrl()')
    expect(onboarding).not.toContain('data.proxyUrl')
    expect(onboarding).not.toMatch(/\{\s*proxyUrl\s*\}\s*=\s*data/u)
  })

  it('keeps agent access mutations validated, loopback-only, confirmed, and fail-closed', () => {
    const handlers = source('src/main/src/client/subscriptions/handlers.ts')
    const proxyFetchStart = handlers.indexOf('export async function proxyFetch')
    const proxyFetchEnd = handlers.indexOf('\nexport const ', proxyFetchStart)
    const proxyFetch = handlers.slice(proxyFetchStart, proxyFetchEnd)

    expect(proxyFetch).toContain('configuredLoopbackProxyUrl()')
    expect(proxyFetch).toContain('if (!response.ok)')

    const handlerNames = [
      'confirmDeclineAgentUser',
      'removeAgentUser',
      'revokeAgentAllowance',
      'confirmDeclineAgentAllowanceRequest'
    ]
    for (const name of handlerNames) {
      const start = handlers.indexOf(`export const ${name}`)
      const end = handlers.indexOf('\nexport const ', start + 1)
      const handler = handlers.slice(start, end < 0 ? handlers.length : end)

      expect(start, `${name} handler is missing`).toBeGreaterThan(-1)
      expect(handler, `${name} lacks bounded username validation`).toContain(
        'validateAgentUsername'
      )
      expect(handler, `${name} lacks a native security confirmation`).toContain(
        'confirmNativeAction'
      )
      expect(handler, `${name} bypasses the fail-closed mutation helper`).toContain(
        'mutateAgentAccess'
      )
      expect(handler, `${name} uses an unvalidated configured URL`).not.toContain(
        'config.chain.localProxyRouterUrl'
      )
    }

    for (const name of ['revokeAgentAllowance', 'confirmDeclineAgentAllowanceRequest']) {
      const start = handlers.indexOf(`export const ${name}`)
      const end = handlers.indexOf('\nexport const ', start + 1)
      expect(handlers.slice(start, end)).toContain('validateAgentToken')
    }

    for (const name of ['confirmDeclineAgentUser', 'confirmDeclineAgentAllowanceRequest']) {
      const start = handlers.indexOf(`export const ${name}`)
      const end = handlers.indexOf('\nexport const ', start + 1)
      expect(handlers.slice(start, end)).toContain('validateAgentDecision')
    }
  })
})
