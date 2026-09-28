import fs from 'node:fs/promises'
import os from 'node:os'
import path from 'node:path'
import { describe, expect, it, vi } from 'vitest'

vi.mock('electron', () => ({
  BrowserWindow: { fromWebContents: vi.fn() },
  dialog: { showOpenDialog: vi.fn() },
  ipcMain: { handle: vi.fn(), on: vi.fn() }
}))
vi.mock('../../rendererTrust', () => ({
  isTrustedRendererEvent: vi.fn(() => true),
  isTrustedRendererUrl: vi.fn(() => true)
}))
vi.mock('./subscriptions/handlers', () => ({
  configuredLoopbackProxyUrl: vi.fn(() => 'http://127.0.0.1:8082'),
  getAuthHeaders: vi.fn(async () => ({ Authorization: 'redacted-in-test' }))
}))

import {
  ipfsDownloadLimits,
  finalizeIpfsDownload,
  pumpIpfsProgress,
  validateIpfsDownloadCid,
  validateIpfsDownloadRequestId
} from './ipfs-download-ipc'

const encode = (value: string): Uint8Array => new TextEncoder().encode(value)

describe('IPFS download IPC bounds', () => {
  it('accepts only UUID v4 request IDs and bytes32 CID hashes', () => {
    const id = '123e4567-e89b-42d3-a456-426614174000'
    const cid = `0x${'ab'.repeat(32)}`
    expect(validateIpfsDownloadRequestId(id)).toBe(id)
    expect(validateIpfsDownloadCid(cid)).toBe(cid)
    expect(() => validateIpfsDownloadRequestId('not-a-uuid')).toThrow()
    expect(() => validateIpfsDownloadCid('../outside')).toThrow()
    expect(() => validateIpfsDownloadCid(`0x${'a'.repeat(63)}`)).toThrow()
  })

  it('parses split SSE frames and throttles renderer updates while preserving completion', async () => {
    const frames = [
      'data: {"status":"downloading","downloaded":1024,"total":4096,"percentage":25,"timeUpdated":1}\n\n',
      'data: {"status":"downloading","downloaded":2048,"total":4096,"percentage":50,"timeUpdated":2}\n\n',
      'data: {"status":"completed","downloaded":4096,"total":4096,"percentage":100,"timeUpdated":3}\n\n'
    ].join('')
    const midpoint = 73
    const body = new ReadableStream<Uint8Array>({
      start(controller) {
        controller.enqueue(encode(frames.slice(0, midpoint)))
        controller.enqueue(encode(frames.slice(midpoint)))
        controller.close()
      }
    })
    const updates: Array<{ status: string; percentage: number }> = []
    const times = [1_000, 1_050, 1_060]

    await pumpIpfsProgress(
      new Response(body, { status: 200 }),
      new AbortController().signal,
      (progress) => updates.push(progress),
      () => times.shift() ?? 1_060
    )

    expect(updates).toEqual([
      expect.objectContaining({ status: 'downloading', percentage: 25 }),
      expect.objectContaining({ status: 'completed', percentage: 100 })
    ])
  })

  it('rejects declared model sizes beyond the configured bound', async () => {
    const frame = `data: ${JSON.stringify({
      status: 'downloading',
      downloaded: 1,
      total: ipfsDownloadLimits.modelBytes + 1,
      percentage: 0,
      timeUpdated: 1
    })}\n\n`
    const response = new Response(frame, { status: 200 })
    await expect(
      pumpIpfsProgress(response, new AbortController().signal, () => undefined)
    ).rejects.toThrow('out-of-range download progress')
  })

  it('rejects non-2xx responses and cancelled streams', async () => {
    await expect(
      pumpIpfsProgress(
        new Response('{"error":"invalid CID"}', { status: 400 }),
        new AbortController().signal,
        () => undefined
      )
    ).rejects.toThrow('invalid CID')

    const controller = new AbortController()
    controller.abort()
    await expect(
      pumpIpfsProgress(new Response('data: {}\n\n'), controller.signal, () => undefined)
    ).rejects.toThrow('IPFS download cancelled.')
  })

  it('finalizes atomically without overwriting an existing destination', async () => {
    const directory = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-ipfs-test-'))
    const partialPath = path.join(directory, '.download.part')
    const finalPath = path.join(directory, 'model')
    try {
      await fs.writeFile(partialPath, 'first')
      await finalizeIpfsDownload(partialPath, finalPath, new AbortController().signal)
      expect(await fs.readFile(finalPath, 'utf8')).toBe('first')
      await expect(fs.lstat(partialPath)).rejects.toMatchObject({ code: 'ENOENT' })

      await fs.writeFile(partialPath, 'replacement')
      await expect(
        finalizeIpfsDownload(partialPath, finalPath, new AbortController().signal)
      ).rejects.toMatchObject({ code: 'EEXIST' })
      expect(await fs.readFile(finalPath, 'utf8')).toBe('first')
    } finally {
      await fs.rm(directory, { recursive: true, force: true })
    }
  })
})
