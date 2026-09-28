import type { IpcMainEvent, IpcMainInvokeEvent } from 'electron'
import { is } from '@electron-toolkit/utils'
import { join } from 'node:path'
import { pathToFileURL } from 'node:url'

export function getTrustedRendererUrl(): URL {
  const developmentUrl = is.dev && process.env.ELECTRON_RENDERER_URL
  return developmentUrl
    ? new URL(developmentUrl)
    : pathToFileURL(join(__dirname, '../renderer/index.html'))
}

export function isTrustedRendererUrl(value: string): boolean {
  try {
    const candidate = new URL(value)
    const trusted = getTrustedRendererUrl()
    if (trusted.protocol === 'file:') {
      return (
        candidate.protocol === trusted.protocol &&
        candidate.host === trusted.host &&
        candidate.pathname === trusted.pathname
      )
    }
    return candidate.protocol === trusted.protocol && candidate.origin === trusted.origin
  } catch {
    return false
  }
}

type RendererEvent = Pick<IpcMainEvent | IpcMainInvokeEvent, 'sender' | 'senderFrame'>

export function isTrustedRendererEvent(event: RendererEvent): boolean {
  // Only the main frame receives the preload bridge. Refuse subframe IPC even
  // when a child happens to share the development origin.
  if (event.senderFrame?.parent) return false
  const value = event.senderFrame?.url || event.sender.getURL()
  return isTrustedRendererUrl(value)
}

export function assertTrustedRendererEvent(event: RendererEvent): void {
  if (!isTrustedRendererEvent(event)) throw new Error('Rejected IPC from an untrusted renderer.')
}
