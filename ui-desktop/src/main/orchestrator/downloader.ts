import { LogFunctions } from 'electron-log'
import { stat, writeFile } from 'node:fs/promises'
import throttle from 'lodash/throttle'
import fs from 'fs-extra'
import path from 'node:path'

interface DownloadProgress {
  bytesDownloaded: number
  totalBytes: number | null
  // TODO: Add percent
  status: 'downloading' | 'complete' | 'error'
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
        status: 'complete'
      })
      throttledOnProgress?.flush()
      return
    }

    // TODO: Verify if file updated (store metadata)
    const tempDestinationPath = getTempFilePath(destinationPath)
    await fs.ensureDir(path.dirname(tempDestinationPath))
    await writeFile(tempDestinationPath, '', { flag: 'w' })

    const response = await fetchWithTimeout(url, {}, 30000)

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    totalBytes = Number(response.headers.get('content-length')) || null

    if (!response.body) {
      throw new Error('No response body')
    }

    const reader = response.body.getReader()

    while (true) {
      const { done, value } = await reader.read()

      if (done) {
        break
      }

      bytesDownloaded += value.length

      try {
        await writeFile(tempDestinationPath, value, { flag: 'a' })
      } catch (err: any) {
        if (err.code === 'ENOSPC') {
          throw new Error('Not enough space on disk')
        }
        throw err
      }

      throttledOnProgress?.({
        bytesDownloaded,
        totalBytes,
        status: 'downloading'
      })
    }

    // copy temp file to destination path
    await fs.move(tempDestinationPath, destinationPath)
    throttledOnProgress?.({
      bytesDownloaded,
      totalBytes,
      status: 'complete'
    })
    throttledOnProgress?.flush()
  } catch (error: any) {
    logger?.error('Download failed:', url, error)

    throttledOnProgress?.({
      bytesDownloaded: 0,
      totalBytes: null,
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
