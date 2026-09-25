import { afterEach, describe, expect, it, vi } from 'vitest'

// A sandboxed preload has IPC and contextBridge, but no clipboard module.
vi.mock('electron', () => ({
  ipcRenderer: { invoke: vi.fn(), on: vi.fn() },
  contextBridge: { exposeInMainWorld: vi.fn() }
}))

import { contextBridge, ipcRenderer } from 'electron'

const originalIsolation = Object.getOwnPropertyDescriptor(process, 'contextIsolated')

afterEach(() => {
  if (originalIsolation) Object.defineProperty(process, 'contextIsolated', originalIsolation)
  else Reflect.deleteProperty(process, 'contextIsolated')
})

async function loadCopyBridge() {
  vi.resetModules()
  Object.defineProperty(process, 'contextIsolated', { configurable: true, value: true })
  await import('./index')
  const exposed = vi
    .mocked(contextBridge.exposeInMainWorld)
    .mock.calls.find(([name]) => name === 'copyToClipboard')
  expect(exposed).toBeDefined()
  return exposed![1] as (text: string) => Promise<void>
}

describe('sandboxed clipboard bridge', () => {
  it('copies through the narrow main-process bridge without a preload clipboard module', async () => {
    vi.mocked(ipcRenderer.invoke).mockResolvedValueOnce(undefined)
    const copy = await loadCopyBridge()
    await expect(copy('0x1234567890')).resolves.toBeUndefined()
    expect(ipcRenderer.invoke).toHaveBeenCalledWith('clipboard:write-text', '0x1234567890')
  })

  it('preserves a rejected IPC result for the copy control to handle', async () => {
    vi.mocked(ipcRenderer.invoke).mockRejectedValueOnce(new Error('Clipboard unavailable'))
    const copy = await loadCopyBridge()
    await expect(copy('address')).rejects.toThrow('Clipboard unavailable')
  })
})
