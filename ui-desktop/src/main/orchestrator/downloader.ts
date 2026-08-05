import { LogFunctions } from 'electron-log'
import { stat, open as openFile } from 'node:fs/promises'
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
const TempFileSuffix = '.temp'

const getTempFilePath = (filePath: string) => {
  return filePath + TempFileSuffix
}

export async function downloadFile(
  url: string,
  destinationPath: string,
  onProgress?: (progress: DownloadProgress) => void,
  logger?: LogFunctions
): Promise<void> {
  const throttledOnProgress = onProgress
    ? throttle((progress: DownloadProgress) => {
        onProgress(progress)
      }, OnProgressUpdateRateMs)
    : undefined

  let bytesDownloaded: number = 0
  let totalBytes: number | null = null
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

    if (fileExists) {
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

    // TODO: Verify if file updated (store metadata)
    const tempDestinationPath = getTempFilePath(destinationPath)
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

    // Stream to a single open handle. The previous implementation called
    // writeFile(..., { flag: 'a' }) for EVERY chunk, i.e. open + append + close
    // per chunk — hundreds of thousands of syscalls on a multi-hundred-MB model
    // download, which made downloads crawl and pegged the main process.
    const fileHandle = await openFile(tempDestinationPath, 'w')
    try {
      while (true) {
        const { done, value } = await reader.read()

        if (done) {
          break
        }

        bytesDownloaded += value.length

        try {
          await fileHandle.write(value)
        } catch (err: any) {
          if (err.code === 'ENOSPC') {
            throw new Error('Not enough space on disk')
          }
          throw err
        }

        throttledOnProgress?.({
          bytesDownloaded,
          totalBytes,
          progress: totalBytes ? bytesDownloaded / totalBytes : 0,
          status: 'downloading'
        })
      }
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

    // copy temp file to destination path
    await fs.move(tempDestinationPath, destinationPath)
    throttledOnProgress?.({
      bytesDownloaded,
      totalBytes,
      progress: 1,
      status: 'downloading'
    })
    throttledOnProgress?.flush()
  } catch (error: any) {
    logger?.error('Download failed:', url, error)

    // Never leave a partial .temp behind — it would be resumed-as-complete or
    // confuse the next run's disk-space accounting.
    await fs.remove(getTempFilePath(destinationPath)).catch(() => undefined)

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
