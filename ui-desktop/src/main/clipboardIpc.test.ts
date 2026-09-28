import type { IpcMainInvokeEvent } from 'electron'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('electron', () => ({
  clipboard: { writeText: vi.fn() },
  ipcMain: { handle: vi.fn() }
}))
vi.mock('./rendererTrust', () => ({
  assertTrustedRendererEvent: vi.fn()
}))

import { clipboard, ipcMain } from 'electron'
import { assertTrustedRendererEvent } from './rendererTrust'
import { registerClipboardIpc } from './clipboardIpc'

const event = {} as IpcMainInvokeEvent

function registerHandler() {
  registerClipboardIpc()
  expect(ipcMain.handle).toHaveBeenCalledWith('clipboard:write-text', expect.any(Function))
  return vi.mocked(ipcMain.handle).mock.calls[0][1]
}

describe('desktop clipboard IPC', () => {
  beforeEach(() => vi.resetAllMocks())

  it('writes the exact full address to the OS clipboard after validating the sender', () => {
    const address = '0x51d01234567890abcdef1234567890abcdefc25a1'
    const write = registerHandler()
    write(event, address)

    expect(assertTrustedRendererEvent).toHaveBeenCalledWith(event)
    expect(clipboard.writeText).toHaveBeenCalledTimes(1)
    expect(clipboard.writeText).toHaveBeenCalledWith(address)
  })

  it('rejects an untrusted renderer without changing the clipboard', () => {
    const write = registerHandler()
    vi.mocked(assertTrustedRendererEvent).mockImplementationOnce(() => {
      throw new Error('Rejected IPC from an untrusted renderer.')
    })

    expect(() => write(event, 'address')).toThrow('untrusted renderer')
    expect(clipboard.writeText).not.toHaveBeenCalled()
  })

  it('rejects non-text and oversized requests without coercing them', () => {
    const write = registerHandler()
    for (const text of [null, undefined, 42, {}, ['address'], 'x'.repeat(1_048_577)]) {
      expect(() => write(event, text)).toThrow('invalid or too large')
    }
    expect(clipboard.writeText).not.toHaveBeenCalled()
  })

  it('propagates clipboard failures so the renderer cannot report false success', () => {
    const write = registerHandler()
    vi.mocked(clipboard.writeText).mockImplementationOnce(() => {
      throw new Error('Clipboard unavailable')
    })

    expect(() => write(event, 'address')).toThrow('Clipboard unavailable')
  })
})
