import { randomUUID } from 'node:crypto'
import fs from 'node:fs/promises'
import path from 'node:path'
import { BrowserWindow, dialog, ipcMain, type WebContents } from 'electron'
import { isTrustedRendererEvent, isTrustedRendererUrl } from '../../rendererTrust'
import { configuredLoopbackProxyUrl, getAuthHeaders } from './subscriptions/handlers'

export const ipfsDownloadChannels = {
  selectFolder: 'ipfs-download:select-folder',
  start: 'ipfs-download:start',
  cancel: 'ipfs-download:cancel',
  event: 'ipfs-download:event'
} as const

export const ipfsDownloadLimits = {
  activePerRenderer: 2,
  folderGrantsPerRenderer: 4,
  folderGrantMs: 10 * 60_000,
  responseBytes: 64 * 1024 * 1024,
  responseEvents: 300_000,
  eventBytes: 16 * 1024,
  modelBytes: 256 * 1024 * 1024 * 1024,
  timeoutMs: 12 * 60 * 60_000 + 5 * 60_000,
  rendererUpdateMs: 100
} as const

export type IpfsDownloadProgress = {
  status: 'downloading' | 'completed'
  downloaded: number
  total: number
  percentage: number
  timeUpdated: number
}

type DownloadEvent =
  | { requestId: string; kind: 'progress'; progress: IpfsDownloadProgress }
  | { requestId: string; kind: 'error'; message: string }

type FolderGrant = {
  folder: string
  sender: WebContents
  expiresAt: number
  timeout: ReturnType<typeof setTimeout>
  abortIfDestroyed: () => void
  abortIfNavigating: (
    event: Electron.Event,
    url: string,
    isInPlace: boolean,
    isMainFrame: boolean
  ) => void
}

type ActiveDownload = {
  controller: AbortController
  sender: WebContents
  timeout: ReturnType<typeof setTimeout>
}

const folderGrants = new Map<string, FolderGrant>()
const activeDownloads = new Map<string, ActiveDownload>()
const requestIdPattern = /^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$/iu

export function validateIpfsDownloadRequestId(value: unknown): string {
  if (typeof value !== 'string' || !requestIdPattern.test(value)) {
    throw new Error('IPFS download request ID is invalid.')
  }
  return value
}

export function validateIpfsDownloadCid(value: unknown): string {
  if (typeof value !== 'string' || !/^0x[0-9a-f]{64}$/iu.test(value)) {
    throw new Error('IPFS metadata CID hash is invalid.')
  }
  return value
}

const scopedKey = (senderId: number, id: string): string => `${senderId}:${id}`

function sendDownloadEvent(sender: WebContents, payload: DownloadEvent): void {
  if (sender.isDestroyed() || !isTrustedRendererUrl(sender.getURL())) return
  sender.send(ipfsDownloadChannels.event, payload)
}

function detachFolderGrant(key: string): void {
  const grant = folderGrants.get(key)
  if (!grant) return
  grant.sender.removeListener('destroyed', grant.abortIfDestroyed)
  grant.sender.removeListener('did-start-navigation', grant.abortIfNavigating)
  clearTimeout(grant.timeout)
  folderGrants.delete(key)
}

function purgeExpiredFolderGrants(): void {
  const now = Date.now()
  for (const [key, grant] of folderGrants) {
    if (grant.expiresAt <= now || grant.sender.isDestroyed()) detachFolderGrant(key)
  }
}

function activeCountForSender(senderId: number): number {
  const prefix = `${senderId}:`
  let count = 0
  for (const key of activeDownloads.keys()) if (key.startsWith(prefix)) count += 1
  return count
}

function grantCountForSender(senderId: number): number {
  const prefix = `${senderId}:`
  let count = 0
  for (const key of folderGrants.keys()) if (key.startsWith(prefix)) count += 1
  return count
}

function cleanupDownload(key: string): void {
  const active = activeDownloads.get(key)
  if (!active) return
  clearTimeout(active.timeout)
  activeDownloads.delete(key)
}

function parseProgress(value: unknown): IpfsDownloadProgress {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('The proxy-router returned invalid download progress.')
  }
  const event = value as Record<string, unknown>
  if (event.status === 'error') {
    throw new Error(String(event.error ?? 'IPFS download failed.').slice(0, 2_000))
  }
  if (event.status !== 'downloading' && event.status !== 'completed') {
    throw new Error('The proxy-router returned an invalid download status.')
  }
  const downloaded = Number(event.downloaded)
  const total = Number(event.total)
  const percentage = Number(event.percentage)
  const timeUpdated = Number(event.timeUpdated)
  if (
    !Number.isSafeInteger(downloaded) ||
    downloaded < 0 ||
    downloaded > ipfsDownloadLimits.modelBytes ||
    !Number.isSafeInteger(total) ||
    total < 0 ||
    total > ipfsDownloadLimits.modelBytes ||
    !Number.isFinite(percentage) ||
    percentage < 0 ||
    percentage > 100 ||
    !Number.isSafeInteger(timeUpdated) ||
    timeUpdated < 0
  ) {
    throw new Error('The proxy-router returned out-of-range download progress.')
  }
  return { status: event.status, downloaded, total, percentage, timeUpdated }
}

async function boundedResponseText(response: Response, limit: number): Promise<string> {
  if (!response.body) return ''
  const reader = response.body.getReader()
  const chunks: Buffer[] = []
  let total = 0
  try {
    while (true) {
      const { value, done } = await reader.read()
      if (done) return Buffer.concat(chunks, total).toString('utf8')
      if (!value?.byteLength) continue
      total += value.byteLength
      if (total > limit) {
        await reader.cancel().catch(() => undefined)
        throw new Error('The IPFS error response exceeded the size limit.')
      }
      chunks.push(Buffer.from(value))
    }
  } finally {
    reader.releaseLock()
  }
}

export async function pumpIpfsProgress(
  response: Response,
  signal: AbortSignal,
  onProgress: (progress: IpfsDownloadProgress) => void,
  now: () => number = Date.now
): Promise<IpfsDownloadProgress> {
  if (!response.ok) {
    const error = await boundedResponseText(response, ipfsDownloadLimits.eventBytes)
    throw new Error(error || `IPFS download failed (HTTP ${response.status}).`)
  }
  if (!response.body) throw new Error('The IPFS download returned no progress stream.')
  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let responseBytes = 0
  let responseEvents = 0
  let lastRendererUpdate = 0
  let completed = false
  let completionProgress: IpfsDownloadProgress | null = null

  const processEvent = (rawEvent: string): void => {
    const data = rawEvent
      .split('\n')
      .filter((line) => line.startsWith('data:'))
      .map((line) => line.slice(5).trimStart())
      .join('\n')
    if (!data) return
    responseEvents += 1
    if (responseEvents > ipfsDownloadLimits.responseEvents) {
      throw new Error('The IPFS progress stream emitted too many events.')
    }
    if (Buffer.byteLength(data, 'utf8') > ipfsDownloadLimits.eventBytes) {
      throw new Error('An IPFS progress event exceeded the size limit.')
    }
    const progress = parseProgress(JSON.parse(data))
    const currentTime = now()
    if (
      progress.status === 'completed' ||
      currentTime - lastRendererUpdate >= ipfsDownloadLimits.rendererUpdateMs
    ) {
      lastRendererUpdate = currentTime
      onProgress(progress)
    }
    if (progress.status === 'completed') {
      completed = true
      completionProgress = progress
    }
  }

  try {
    while (true) {
      if (signal.aborted) throw new Error('IPFS download cancelled.')
      const { value, done } = await reader.read()
      if (done) break
      if (!value?.byteLength) continue
      responseBytes += value.byteLength
      if (responseBytes > ipfsDownloadLimits.responseBytes) {
        throw new Error('The IPFS progress stream exceeded the 64 MB limit.')
      }
      buffer += decoder.decode(value, { stream: true })
      buffer = buffer.replace(/\r\n/gu, '\n')
      let boundary = buffer.indexOf('\n\n')
      while (boundary >= 0) {
        processEvent(buffer.slice(0, boundary))
        buffer = buffer.slice(boundary + 2)
        boundary = buffer.indexOf('\n\n')
      }
      if (Buffer.byteLength(buffer, 'utf8') > ipfsDownloadLimits.eventBytes * 2) {
        throw new Error('An incomplete IPFS progress event exceeded the size limit.')
      }
    }
    buffer += decoder.decode()
    if (buffer.trim()) processEvent(buffer)
    if (!completed || !completionProgress) {
      throw new Error('The IPFS progress stream ended before completion.')
    }
    return completionProgress
  } catch (error) {
    await reader.cancel().catch(() => undefined)
    throw error
  } finally {
    reader.releaseLock()
  }
}

export async function finalizeIpfsDownload(
  partialPath: string,
  finalPath: string,
  signal: AbortSignal
): Promise<void> {
  if (signal.aborted) throw new Error('IPFS download cancelled.')
  const partialStat = await fs.lstat(partialPath)
  if (!partialStat.isFile() || partialStat.isSymbolicLink()) {
    throw new Error('The downloaded model is not a regular file.')
  }
  // link() is atomic and refuses an existing final path. It avoids the
  // overwrite behavior of rename() while keeping finalization on the selected
  // filesystem. The randomized partial name is never exposed to the renderer.
  await fs.link(partialPath, finalPath)
  await fs.unlink(partialPath).catch(() => undefined)
}

export function registerIpfsDownloadIpc(): void {
  ipcMain.handle(ipfsDownloadChannels.selectFolder, async (event) => {
    if (!isTrustedRendererEvent(event)) throw new Error('Untrusted folder selection request.')
    purgeExpiredFolderGrants()
    if (grantCountForSender(event.sender.id) >= ipfsDownloadLimits.folderGrantsPerRenderer) {
      throw new Error('Too many unused download folders are selected.')
    }
    const owner = BrowserWindow.fromWebContents(event.sender)
    const result = owner
      ? await dialog.showOpenDialog(owner, { properties: ['openDirectory'] })
      : await dialog.showOpenDialog({ properties: ['openDirectory'] })
    if (result.canceled || result.filePaths.length !== 1) {
      return { canceled: true }
    }
    if (event.sender.isDestroyed() || !isTrustedRendererUrl(event.sender.getURL())) {
      throw new Error('The requesting window is no longer available.')
    }
    const folder = await fs.realpath(result.filePaths[0])
    const folderStat = await fs.stat(folder)
    if (!folderStat.isDirectory()) throw new Error('The selected download folder is invalid.')

    const folderToken = randomUUID()
    const key = scopedKey(event.sender.id, folderToken)
    const timeout = setTimeout(() => detachFolderGrant(key), ipfsDownloadLimits.folderGrantMs)
    const abortIfDestroyed = (): void => detachFolderGrant(key)
    const abortIfNavigating = (
      _navigationEvent: Electron.Event,
      _url: string,
      _isInPlace: boolean,
      isMainFrame: boolean
    ): void => {
      if (isMainFrame) detachFolderGrant(key)
    }
    folderGrants.set(key, {
      folder,
      sender: event.sender,
      expiresAt: Date.now() + ipfsDownloadLimits.folderGrantMs,
      timeout,
      abortIfDestroyed,
      abortIfNavigating
    })
    event.sender.once('destroyed', abortIfDestroyed)
    event.sender.on('did-start-navigation', abortIfNavigating)
    return { canceled: false, folderToken }
  })

  ipcMain.handle(ipfsDownloadChannels.start, async (event, input: any) => {
    if (!isTrustedRendererEvent(event)) throw new Error('Untrusted IPFS download request.')
    purgeExpiredFolderGrants()
    const requestId = validateIpfsDownloadRequestId(input?.requestId)
    const cidHash = validateIpfsDownloadCid(input?.cidHash)
    const folderToken = validateIpfsDownloadRequestId(input?.folderToken)
    const grantKey = scopedKey(event.sender.id, folderToken)
    const grant = folderGrants.get(grantKey)
    if (!grant || grant.expiresAt <= Date.now())
      throw new Error('Select the download folder again.')

    const key = scopedKey(event.sender.id, requestId)
    if (activeDownloads.has(key)) throw new Error('IPFS download request ID is already active.')
    if (activeCountForSender(event.sender.id) >= ipfsDownloadLimits.activePerRenderer) {
      throw new Error('Too many IPFS downloads are active.')
    }
    detachFolderGrant(grantKey)
    const finalPath = path.join(grant.folder, cidHash)
    const partialPath = path.join(grant.folder, `.${cidHash}.${requestId}.part`)
    if (path.dirname(finalPath) !== grant.folder || path.dirname(partialPath) !== grant.folder) {
      throw new Error('Invalid download destination.')
    }

    const controller = new AbortController()
    let responseStarted = false
    const timeout = setTimeout(() => {
      if (responseStarted) {
        sendDownloadEvent(event.sender, {
          requestId,
          kind: 'error',
          message: 'The IPFS download timed out.'
        })
      }
      controller.abort('IPFS download timed out.')
    }, ipfsDownloadLimits.timeoutMs)
    activeDownloads.set(key, { controller, sender: event.sender, timeout })
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
      const currentFolder = await fs.realpath(grant.folder)
      const currentFolderStat = await fs.stat(currentFolder)
      if (currentFolder !== grant.folder || !currentFolderStat.isDirectory()) {
        throw new Error('The selected download folder is no longer valid.')
      }
      for (const candidate of [finalPath, partialPath]) {
        try {
          await fs.lstat(candidate)
          throw new Error('A file already exists at the model destination.')
        } catch (error: any) {
          if (error?.code !== 'ENOENT') throw error
        }
      }
      if (controller.signal.aborted) throw new Error('IPFS download cancelled.')
      const response = await fetch(
        `${configuredLoopbackProxyUrl()}/ipfs/download/stream/${encodeURIComponent(cidHash)}?dest=${encodeURIComponent(partialPath)}`,
        { headers: await getAuthHeaders(), signal: controller.signal }
      )
      responseStarted = true
      void pumpIpfsProgress(response, controller.signal, (progress) => {
        if (progress.status === 'downloading') {
          sendDownloadEvent(event.sender, { requestId, kind: 'progress', progress })
        }
      })
        .then(async (completionProgress) => {
          await finalizeIpfsDownload(partialPath, finalPath, controller.signal)
          sendDownloadEvent(event.sender, {
            requestId,
            kind: 'progress',
            progress: completionProgress
          })
        })
        .catch((error: any) => {
          if (!controller.signal.aborted) {
            sendDownloadEvent(event.sender, {
              requestId,
              kind: 'error',
              message: String(error?.message ?? 'IPFS download failed.').slice(0, 2_000)
            })
          }
        })
        .finally(() => {
          void fs.unlink(partialPath).catch(() => undefined)
          detachRendererLifecycle()
          cleanupDownload(key)
        })
      return { accepted: true }
    } catch (error) {
      void fs.unlink(partialPath).catch(() => undefined)
      detachRendererLifecycle()
      cleanupDownload(key)
      throw error
    }
  })

  ipcMain.on(ipfsDownloadChannels.cancel, (event, input: any) => {
    if (!isTrustedRendererEvent(event)) return
    try {
      const requestId = validateIpfsDownloadRequestId(input?.requestId)
      activeDownloads
        .get(scopedKey(event.sender.id, requestId))
        ?.controller.abort('IPFS download cancelled.')
    } catch {
      // Invalid cancellation messages have no authority and are ignored.
    }
  })
}
