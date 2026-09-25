import { LogFunctions } from 'electron-log'
import { createReadStream } from 'node:fs'
import { open as openFile, rename, stat } from 'node:fs/promises'
import { createHash, randomUUID } from 'node:crypto'
import throttle from 'lodash/throttle'
import fs from 'fs-extra'
import path from 'node:path'

interface DownloadProgress {
  bytesDownloaded: number
  totalBytes: number | null
  progress: number
  // TODO: Add percent
  status: 'downloading' | 'error'
  error?: string
}

const OnProgressUpdateRateMs = 300
const DownloadMetadataSuffix = '.download.json'

export interface DownloadOptions {
  /**
   * Refresh an existing file when it came from a different artifact URL.
   *
   * This is intended for versioned service executables. Large optional assets
   * keep the legacy "existing file wins" behavior unless explicitly enabled.
   */
  refreshIfSourceChanged?: boolean
  /** Maximum time to wait between streamed response chunks. */
  bodyIdleTimeoutMs?: number
}

interface DownloadMetadata {
  version: 3
  sourceUrlSha256: string
  contentSha256: string
}

interface BundledArtifactMetadata {
  version: 2
  contentSha256: string
}

interface BundledArtifactManifest {
  schemaVersion: 1
  platform: NodeJS.Platform
  arch: string
  size: number
  sha256: string
  buildVersion: string
  commit: string
  dirty: boolean
}

const BundledExecutableName = 'bundled-proxy-router'
const BundledManifestName = 'manifest.json'

const getMetadataFilePath = (filePath: string) => {
  return filePath + DownloadMetadataSuffix
}

const getUniqueTempFilePath = (filePath: string) => {
  return path.join(
    path.dirname(filePath),
    `.${path.basename(filePath)}.${process.pid}.${randomUUID()}.tmp`
  )
}

const replaceFileAtomically = async (sourcePath: string, destinationPath: string) => {
  // Both paths are in the same directory, so native rename is atomic on the
  // supported filesystems. fs-extra move(overwrite) first deletes the working
  // destination and can leave no executable if replacement then fails.
  await rename(sourcePath, destinationPath)
}

const redactUrlForLog = (rawUrl: string) => {
  try {
    const parsed = new URL(rawUrl)
    parsed.username = ''
    parsed.password = ''
    if (parsed.search) parsed.search = '?[redacted]'
    return parsed.toString()
  } catch {
    return '[invalid URL]'
  }
}

const getSourceUrlSha256 = (url: string) => {
  // Store only a digest. Download URLs may contain temporary credentials or
  // signed query parameters that must never be copied into user-data files.
  return createHash('sha256').update(url).digest('hex')
}

async function isCurrentDownload(destinationPath: string, url: string): Promise<boolean> {
  try {
    const metadata = (await fs.readJson(getMetadataFilePath(destinationPath))) as DownloadMetadata
    return (
      metadata.version === 3 &&
      metadata.sourceUrlSha256 === getSourceUrlSha256(url) &&
      metadata.contentSha256 === (await fileSha256(destinationPath))
    )
  } catch {
    // Existing installs predate metadata. Refresh them once so a stale service
    // binary cannot survive every desktop-app upgrade indefinitely.
    return false
  }
}

async function writeDownloadMetadata(
  destinationPath: string,
  url: string,
  contentSha256: string
): Promise<void> {
  const metadataPath = getMetadataFilePath(destinationPath)
  const tempMetadataPath = getUniqueTempFilePath(metadataPath)
  const metadata: DownloadMetadata = {
    version: 3,
    sourceUrlSha256: getSourceUrlSha256(url),
    contentSha256
  }

  try {
    await fs.writeJson(tempMetadataPath, metadata, { spaces: 2 })
    await replaceFileAtomically(tempMetadataPath, metadataPath)
  } finally {
    await fs.remove(tempMetadataPath).catch(() => undefined)
  }
}

async function fileSha256(filePath: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const hash = createHash('sha256')
    const stream = createReadStream(filePath)
    stream.on('error', reject)
    stream.on('data', (chunk) => hash.update(chunk))
    stream.on('end', () => resolve(hash.digest('hex')))
  })
}

async function readBundledArtifactManifest(
  bundleDirectory: string
): Promise<BundledArtifactManifest> {
  const manifestPath = path.join(bundleDirectory, BundledManifestName)
  let manifest: BundledArtifactManifest
  try {
    manifest = (await fs.readJson(manifestPath)) as BundledArtifactManifest
  } catch (error) {
    throw new Error(`Bundled proxy-router manifest is missing or invalid: ${String(error)}`)
  }

  if (
    manifest.schemaVersion !== 1 ||
    manifest.platform !== process.platform ||
    manifest.arch !== process.arch ||
    !Number.isSafeInteger(manifest.size) ||
    manifest.size <= 0 ||
    !/^[a-f0-9]{64}$/.test(manifest.sha256) ||
    typeof manifest.buildVersion !== 'string' ||
    !manifest.buildVersion ||
    !/^[a-f0-9]{40}$/.test(manifest.commit) ||
    typeof manifest.dirty !== 'boolean'
  ) {
    throw new Error(
      `Bundled proxy-router manifest does not match ${process.platform}/${process.arch}`
    )
  }
  return manifest
}

async function isCurrentBundledArtifact(
  destinationPath: string,
  contentSha256: string
): Promise<boolean> {
  try {
    const metadata = (await fs.readJson(
      getMetadataFilePath(destinationPath)
    )) as BundledArtifactMetadata
    if (metadata.version !== 2 || metadata.contentSha256 !== contentSha256) {
      return false
    }
    return (await fileSha256(destinationPath)) === contentSha256
  } catch {
    return false
  }
}

async function writeBundledArtifactMetadata(
  destinationPath: string,
  contentSha256: string
): Promise<void> {
  const metadataPath = getMetadataFilePath(destinationPath)
  const tempMetadataPath = getUniqueTempFilePath(metadataPath)
  const metadata: BundledArtifactMetadata = {
    version: 2,
    contentSha256
  }

  try {
    await fs.writeJson(tempMetadataPath, metadata, { spaces: 2 })
    await replaceFileAtomically(tempMetadataPath, metadataPath)
  } finally {
    await fs.remove(tempMetadataPath).catch(() => undefined)
  }
}

async function readWithIdleTimeout(
  reader: ReadableStreamDefaultReader<Uint8Array>,
  timeoutMs: number
): Promise<ReadableStreamReadResult<Uint8Array>> {
  return new Promise((resolve, reject) => {
    const timeoutId = setTimeout(() => {
      reject(new Error('Download stalled while waiting for data'))
      void reader.cancel('download body idle timeout').catch(() => undefined)
    }, timeoutMs)

    reader.read().then(
      (result) => {
        clearTimeout(timeoutId)
        resolve(result)
      },
      (error) => {
        clearTimeout(timeoutId)
        reject(error)
      }
    )
  })
}

interface WritableFileHandle {
  write(buffer: Uint8Array, offset: number, length: number): Promise<{ bytesWritten: number }>
}

export async function writeAll(fileHandle: WritableFileHandle, value: Uint8Array): Promise<void> {
  let writeOffset = 0
  while (writeOffset < value.length) {
    const { bytesWritten } = await fileHandle.write(value, writeOffset, value.length - writeOffset)
    if (bytesWritten === 0) {
      throw new Error('Unable to make progress while writing download')
    }
    writeOffset += bytesWritten
  }
}

/**
 * Installs a proxy-router shipped inside a local desktop package.
 *
 * Release CI normally injects a URL for the router built from the same commit.
 * Local packages have no such artifact URL, so without this path they silently
 * fall back to an older upstream router. Content-addressed metadata avoids an
 * 80 MB copy on every launch while still repairing a corrupted cached binary.
 */
export async function installBundledExecutable(
  bundleDirectory: string,
  destinationPath: string,
  onProgress?: (progress: DownloadProgress) => void,
  logger?: LogFunctions
): Promise<void> {
  const sourcePath = path.join(bundleDirectory, BundledExecutableName)
  const tempDestinationPath = getUniqueTempFilePath(destinationPath)
  try {
    const manifest = await readBundledArtifactManifest(bundleDirectory)
    const sourceInfo = await stat(sourcePath)
    if (!sourceInfo.isFile() || sourceInfo.size !== manifest.size) {
      throw new Error('Bundled proxy-router size does not match its manifest')
    }

    const contentSha256 = await fileSha256(sourcePath)
    if (contentSha256 !== manifest.sha256) {
      throw new Error('Bundled proxy-router hash does not match its manifest')
    }
    const destinationExists = await stat(destinationPath)
      .then((info) => info.isFile())
      .catch((error: NodeJS.ErrnoException) => {
        if (error.code === 'ENOENT') return false
        throw error
      })

    if (destinationExists && (await isCurrentBundledArtifact(destinationPath, contentSha256))) {
      logger?.info(`Bundled executable already installed at ${destinationPath}`)
      onProgress?.({
        bytesDownloaded: sourceInfo.size,
        totalBytes: sourceInfo.size,
        progress: 1,
        status: 'downloading'
      })
      return
    }

    await fs.ensureDir(path.dirname(tempDestinationPath))
    await fs.copyFile(sourcePath, tempDestinationPath)

    if ((await fileSha256(tempDestinationPath)) !== contentSha256) {
      throw new Error('Bundled proxy-router failed integrity verification')
    }

    await fs.chmod(tempDestinationPath, 0o755)
    const tempHandle = await openFile(tempDestinationPath, 'r+')
    try {
      await tempHandle.sync()
    } finally {
      await tempHandle.close()
    }
    await replaceFileAtomically(tempDestinationPath, destinationPath)
    await writeBundledArtifactMetadata(destinationPath, contentSha256)
    logger?.info(`Installed bundled executable at ${destinationPath}`)
    onProgress?.({
      bytesDownloaded: sourceInfo.size,
      totalBytes: sourceInfo.size,
      progress: 1,
      status: 'downloading'
    })
  } catch (error: any) {
    await fs.remove(tempDestinationPath).catch(() => undefined)
    onProgress?.({
      bytesDownloaded: 0,
      totalBytes: null,
      progress: 0,
      status: 'error',
      error: error.message
    })
    throw error
  }
}

export async function downloadFile(
  url: string,
  destinationPath: string,
  onProgress?: (progress: DownloadProgress) => void,
  logger?: LogFunctions,
  options: DownloadOptions = {}
): Promise<void> {
  const throttledOnProgress = onProgress
    ? throttle((progress: DownloadProgress) => {
        onProgress(progress)
      }, OnProgressUpdateRateMs)
    : undefined

  let bytesDownloaded: number = 0
  let totalBytes: number | null = null
  let tempDestinationPath: string | null = null
  try {
    // check if file exists
    const fileExists = await stat(destinationPath)
      .then(() => true)
      .catch((err: any) => {
        if (err.code === 'ENOENT') {
          return false
        }
        throw err
      })

    const sourceIsCurrent =
      fileExists &&
      options.refreshIfSourceChanged &&
      (await isCurrentDownload(destinationPath, url))

    if (fileExists && (!options.refreshIfSourceChanged || sourceIsCurrent)) {
      logger?.info(`File already exists at ${destinationPath}, skipping download`)
      throttledOnProgress?.({
        bytesDownloaded: bytesDownloaded,
        totalBytes: totalBytes,
        progress: 1,
        status: 'downloading'
      })
      throttledOnProgress?.flush()
      return
    }

    if (fileExists) {
      logger?.info(`A newer download source is configured for ${destinationPath}; refreshing`)
    }

    tempDestinationPath = getUniqueTempFilePath(destinationPath)
    await fs.ensureDir(path.dirname(tempDestinationPath))

    const response = await fetchWithTimeout(url, {}, 30000)

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    totalBytes = Number(response.headers.get('content-length')) || null

    if (!response.body) {
      throw new Error('No response body')
    }

    const reader = response.body.getReader()
    const bodyIdleTimeoutMs = Math.max(1, options.bodyIdleTimeoutMs ?? 30_000)

    // Stream to a single open handle. The previous implementation called
    // writeFile(..., { flag: 'a' }) for EVERY chunk, i.e. open + append + close
    // per chunk — hundreds of thousands of syscalls on a multi-hundred-MB model
    // download, which made downloads crawl and pegged the main process.
    const fileHandle = await openFile(tempDestinationPath, 'w')
    try {
      while (true) {
        const { done, value } = await readWithIdleTimeout(reader, bodyIdleTimeoutMs)

        if (done) {
          break
        }

        try {
          await writeAll(fileHandle, value)
        } catch (err: any) {
          if (err.code === 'ENOSPC') {
            throw new Error('Not enough space on disk')
          }
          throw err
        }

        bytesDownloaded += value.length

        throttledOnProgress?.({
          bytesDownloaded,
          totalBytes,
          progress: totalBytes ? bytesDownloaded / totalBytes : 0,
          status: 'downloading'
        })
      }
      await fileHandle.sync()
    } finally {
      await fileHandle.close().catch(() => undefined)
    }

    // A truncated download that still produced a file is worse than no file at
    // all: it gets moved into place, passes the "already exists" check forever,
    // and the service then fails to start with an unrelated-looking error.
    if (totalBytes && bytesDownloaded !== totalBytes) {
      throw new Error(
        `Download incomplete: got ${bytesDownloaded} of ${totalBytes} bytes. Please try again.`
      )
    }

    // Replace only after a complete download. If the request fails, the
    // existing executable remains untouched and can still be used/retried.
    const contentSha256 = await fileSha256(tempDestinationPath)
    await replaceFileAtomically(tempDestinationPath, destinationPath)
    tempDestinationPath = null

    if (options.refreshIfSourceChanged) {
      await writeDownloadMetadata(destinationPath, url, contentSha256)
    }
    throttledOnProgress?.({
      bytesDownloaded,
      totalBytes,
      progress: 1,
      status: 'downloading'
    })
    throttledOnProgress?.flush()
  } catch (error: any) {
    logger?.error('Download failed:', redactUrlForLog(url), error)

    // Never leave a partial .temp behind — it would be resumed-as-complete or
    // confuse the next run's disk-space accounting.
    if (tempDestinationPath) {
      await fs.remove(tempDestinationPath).catch(() => undefined)
    }

    throttledOnProgress?.({
      bytesDownloaded: 0,
      totalBytes: null,
      progress: 0,
      status: 'error',
      error: error.message
    })
    throttledOnProgress?.flush()

    throw error
  }
}

function fetchWithTimeout(url: string, options = {}, timeout = 5000) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), timeout)

  return fetch(url, { ...options, signal: controller.signal }).finally(() =>
    clearTimeout(timeoutId)
  )
}
