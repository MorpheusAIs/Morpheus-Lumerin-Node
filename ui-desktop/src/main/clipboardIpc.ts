import { clipboard, ipcMain } from 'electron'
import { assertTrustedRendererEvent } from './rendererTrust'

const clipboardWriteChannel = 'clipboard:write-text'
const maxClipboardTextLength = 1_048_576

export function registerClipboardIpc(): void {
  ipcMain.handle(clipboardWriteChannel, (event, text: unknown) => {
    assertTrustedRendererEvent(event)
    if (typeof text !== 'string' || text.length > maxClipboardTextLength) {
      throw new Error('Clipboard text is invalid or too large.')
    }

    // Electron's clipboard module is unavailable in a sandboxed preload.
    // Keep OS access here and resolve only after the write has succeeded.
    clipboard.writeText(text)
  })
}
