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

import { chatStreamLimits, pumpChatResponse, validateChatStreamRequestId } from './chat-stream-ipc'

describe('chat stream IPC bounds', () => {
  beforeEach(() => vi.clearAllMocks())

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

  it('cancels and rejects responses over the total byte limit', async () => {
    let cancelled = false
    const body = new ReadableStream<Uint8Array>({
      start(controller) {
        controller.enqueue(new Uint8Array(chatStreamLimits.responseBytes + 1))
      },
      cancel() {
        cancelled = true
      }
    })

    await expect(
      pumpChatResponse(body, new AbortController().signal, () => undefined)
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

  it('bounds the number of IPC events from pathological tiny chunks', async () => {
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
      pumpChatResponse(body, new AbortController().signal, () => {
        emitted += 1
      })
    ).rejects.toThrow('Chat stream emitted too many chunks.')
    expect(emitted).toBe(chatStreamLimits.ipcEvents)
    expect(cancelled).toBe(true)
  })
})
