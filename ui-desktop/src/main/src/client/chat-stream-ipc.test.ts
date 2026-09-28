import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('electron', () => ({
  ipcMain: { handle: vi.fn(), on: vi.fn() }
}))
vi.mock('../../rendererTrust', () => ({
  isTrustedRendererEvent: vi.fn(() => true),
  isTrustedRendererUrl: vi.fn(() => true)
}))
vi.mock('./subscriptions/handlers', () => ({
  openChatCompletionStream: vi.fn()
}))

import {
  chatStreamLimits,
  pumpChatResponse,
  validateChatStreamRequestId,
  type ChatStreamLimits
} from './chat-stream-ipc'

// Production runs with the size, event and time ceilings switched off, because
// a reasoning model crosses all three while it is still answering. The guards
// themselves still work when a ceiling is configured, so they are exercised
// here with explicit limits rather than deleted along with the defaults.
const withLimits = (overrides: Partial<ChatStreamLimits>): ChatStreamLimits => ({
  ...chatStreamLimits,
  ...overrides
})

describe('chat stream IPC bounds', () => {
  beforeEach(() => vi.clearAllMocks())

  it('applies no size, event or time ceiling by default', () => {
    expect(chatStreamLimits.responseBytes).toBe(0)
    expect(chatStreamLimits.ipcEvents).toBe(0)
    expect(chatStreamLimits.timeoutMs).toBe(0)
    // Splitting is not refusing, and concurrency is not length. Both stay.
    expect(chatStreamLimits.chunkBytes).toBeGreaterThan(0)
    expect(chatStreamLimits.activePerRenderer).toBeGreaterThan(0)
  })

  it('carries a response past the byte count that used to end it', async () => {
    let delivered = 0
    const oldCeiling = 8 * 1024 * 1024
    const body = new ReadableStream<Uint8Array>({
      start(controller) {
        controller.enqueue(new Uint8Array(oldCeiling + 1))
        controller.close()
      }
    })

    await pumpChatResponse(body, new AbortController().signal, (chunk) => {
      delivered += chunk.byteLength
    })

    expect(delivered).toBe(oldCeiling + 1)
  })

  it('accepts only UUID v4 request identifiers', () => {
    const id = '123e4567-e89b-42d3-a456-426614174000'
    expect(validateChatStreamRequestId(id)).toBe(id)
    for (const value of [
      undefined,
      '',
      '123e4567-e89b-12d3-a456-426614174000',
      '123e4567-e89b-42d3-c456-426614174000',
      `${id}-suffix`
    ]) {
      expect(() => validateChatStreamRequestId(value)).toThrow('Chat stream request ID is invalid.')
    }
  })

  it('splits oversized network chunks into bounded IPC messages', async () => {
    const input = new Uint8Array(chatStreamLimits.chunkBytes * 2 + 7).fill(5)
    const chunks: Uint8Array[] = []
    const body = new ReadableStream<Uint8Array>({
      start(controller) {
        controller.enqueue(input)
        controller.close()
      }
    })

    await pumpChatResponse(body, new AbortController().signal, (chunk) => chunks.push(chunk))

    expect(chunks.map((chunk) => chunk.byteLength)).toEqual([
      chatStreamLimits.chunkBytes,
      chatStreamLimits.chunkBytes,
      7
    ])
  })

  it('cancels and rejects responses over a configured byte limit', async () => {
    const limits = withLimits({ responseBytes: 8 * 1024 * 1024 })
    let cancelled = false
    const body = new ReadableStream<Uint8Array>({
      start(controller) {
        controller.enqueue(new Uint8Array(limits.responseBytes + 1))
      },
      cancel() {
        cancelled = true
      }
    })

    await expect(
      pumpChatResponse(body, new AbortController().signal, () => undefined, limits)
    ).rejects.toThrow('Chat stream response exceeded the 8 MB limit.')
    expect(cancelled).toBe(true)
  })

  it('does not read after cancellation', async () => {
    const controller = new AbortController()
    controller.abort()
    const body = new ReadableStream<Uint8Array>({
      start(stream) {
        stream.enqueue(new Uint8Array([1]))
      }
    })

    await expect(pumpChatResponse(body, controller.signal, () => undefined)).rejects.toThrow(
      'Chat stream cancelled.'
    )
  })

  it('bounds IPC events from pathological tiny chunks when a limit is configured', async () => {
    const limits = withLimits({ ipcEvents: 32 })
    let emitted = 0
    let cancelled = false
    const body = new ReadableStream<Uint8Array>({
      pull(controller) {
        controller.enqueue(new Uint8Array([1]))
      },
      cancel() {
        cancelled = true
      }
    })

    await expect(
      pumpChatResponse(
        body,
        new AbortController().signal,
        () => {
          emitted += 1
        },
        limits
      )
    ).rejects.toThrow('Chat stream emitted too many chunks.')
    expect(emitted).toBe(limits.ipcEvents)
    expect(cancelled).toBe(true)
  })

  it('emits far past the old event ceiling when none is configured', async () => {
    const oldCeiling = 16_384
    let emitted = 0
    const body = new ReadableStream<Uint8Array>({
      pull(controller) {
        if (emitted > oldCeiling) {
          controller.close()
          return
        }
        controller.enqueue(new Uint8Array([1]))
      }
    })

    await pumpChatResponse(body, new AbortController().signal, () => {
      emitted += 1
    })

    expect(emitted).toBeGreaterThan(oldCeiling)
  })
})
