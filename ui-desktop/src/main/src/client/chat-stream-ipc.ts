import { ipcMain, type WebContents } from 'electron'
import { isTrustedRendererEvent, isTrustedRendererUrl } from '../../rendererTrust'
import { openChatCompletionStream } from './subscriptions/handlers'

export const chatStreamChannels = {
  start: 'chat-stream:start',
  cancel: 'chat-stream:cancel',
  event: 'chat-stream:event'
} as const

export const chatStreamLimits = {
  activePerRenderer: 4,
  chunkBytes: 64 * 1024,
  ipcEvents: 16_384,
  responseBytes: 8 * 1024 * 1024,
  timeoutMs: 5 * 60_000
} as const

type StreamEvent =
  | { requestId: string; kind: 'chunk'; dataBase64: string }
  | { requestId: string; kind: 'end' }
  | { requestId: string; kind: 'error'; message: string }

type ActiveStream = {
  controller: AbortController
  sender: WebContents
  timeout: ReturnType<typeof setTimeout>
}

const activeStreams = new Map<string, ActiveStream>()

export function validateChatStreamRequestId(value: unknown): string {
  if (
    typeof value !== 'string' ||
    !/^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$/iu.test(value)
  ) {
    throw new Error('Chat stream request ID is invalid.')
  }
  return value
}

const streamKey = (senderId: number, requestId: string): string => `${senderId}:${requestId}`

function sendStreamEvent(sender: WebContents, payload: StreamEvent): void {
  if (sender.isDestroyed() || !isTrustedRendererUrl(sender.getURL())) return
  sender.send(chatStreamChannels.event, payload)
}

function cleanupStream(key: string): void {
  const active = activeStreams.get(key)
  if (!active) return
  clearTimeout(active.timeout)
  activeStreams.delete(key)
}

export async function pumpChatResponse(
  body: ReadableStream<Uint8Array> | null,
  signal: AbortSignal,
  onChunk: (chunk: Uint8Array) => void
): Promise<void> {
  if (!body) return
  const reader = body.getReader()
  let total = 0
  let events = 0
  try {
    while (true) {
      if (signal.aborted) throw new Error('Chat stream cancelled.')
      const { value, done } = await reader.read()
      if (done) return
      if (!value?.byteLength) continue
      total += value.byteLength
      if (total > chatStreamLimits.responseBytes) {
        await reader.cancel('Chat stream response exceeded the size limit.').catch(() => undefined)
        throw new Error('Chat stream response exceeded the 8 MB limit.')
      }
      for (let offset = 0; offset < value.byteLength; offset += chatStreamLimits.chunkBytes) {
        events += 1
        if (events > chatStreamLimits.ipcEvents) {
          await reader.cancel('Chat stream emitted too many chunks.').catch(() => undefined)
          throw new Error('Chat stream emitted too many chunks.')
        }
        onChunk(value.subarray(offset, offset + chatStreamLimits.chunkBytes))
      }
    }
  } finally {
    reader.releaseLock()
  }
}

function activeCountForSender(senderId: number): number {
  const prefix = `${senderId}:`
  let count = 0
  for (const key of activeStreams.keys()) if (key.startsWith(prefix)) count += 1
  return count
}

export function registerChatStreamIpc(): void {
  ipcMain.handle(chatStreamChannels.start, async (event, input: any) => {
    if (!isTrustedRendererEvent(event)) throw new Error('Untrusted chat stream request.')
    const requestId = validateChatStreamRequestId(input?.requestId)
    const key = streamKey(event.sender.id, requestId)
    if (activeStreams.has(key)) throw new Error('Chat stream request ID is already active.')
    if (activeCountForSender(event.sender.id) >= chatStreamLimits.activePerRenderer) {
      throw new Error('Too many chat streams are active.')
    }

    const controller = new AbortController()
    let responseStarted = false
    const timeout = setTimeout(() => {
      if (responseStarted) {
        sendStreamEvent(event.sender, {
          requestId,
          kind: 'error',
          message: 'The chat stream timed out.'
        })
      }
      controller.abort('Chat stream timed out.')
    }, chatStreamLimits.timeoutMs)
    activeStreams.set(key, { controller, sender: event.sender, timeout })
    const abortIfDestroyed = (): void => controller.abort('Renderer closed.')
    const abortIfNavigating = (
      _navigationEvent: Electron.Event,
      _url: string,
      _isInPlace: boolean,
      isMainFrame: boolean
    ): void => {
      if (isMainFrame) controller.abort('Renderer navigated.')
    }
    event.sender.once('destroyed', abortIfDestroyed)
    event.sender.on('did-start-navigation', abortIfNavigating)

    const detachRendererLifecycle = (): void => {
      event.sender.removeListener('destroyed', abortIfDestroyed)
      event.sender.removeListener('did-start-navigation', abortIfNavigating)
    }

    try {
      const response = await openChatCompletionStream(input?.payload, controller.signal)
      responseStarted = true
      void pumpChatResponse(response.body, controller.signal, (chunk) => {
        sendStreamEvent(event.sender, {
          requestId,
          kind: 'chunk',
          dataBase64: Buffer.from(chunk).toString('base64')
        })
      })
        .then(() => {
          if (!controller.signal.aborted) sendStreamEvent(event.sender, { requestId, kind: 'end' })
        })
        .catch((error: any) => {
          if (!controller.signal.aborted) {
            sendStreamEvent(event.sender, {
              requestId,
              kind: 'error',
              message: String(error?.message ?? 'Chat stream failed.').slice(0, 2_000)
            })
          }
        })
        .finally(() => {
          detachRendererLifecycle()
          cleanupStream(key)
        })

      return {
        ok: response.ok,
        status: response.status,
        contentType: response.headers.get('content-type') ?? 'text/plain'
      }
    } catch (error) {
      detachRendererLifecycle()
      cleanupStream(key)
      throw error
    }
  })

  ipcMain.on(chatStreamChannels.cancel, (event, input: any) => {
    if (!isTrustedRendererEvent(event)) return
    let requestId: string
    try {
      requestId = validateChatStreamRequestId(input?.requestId)
    } catch {
      return
    }
    const key = streamKey(event.sender.id, requestId)
    activeStreams.get(key)?.controller.abort('Chat stream cancelled.')
  })
}
