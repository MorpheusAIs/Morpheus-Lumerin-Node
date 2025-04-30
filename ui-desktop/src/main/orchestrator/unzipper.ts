import yauzl from 'yauzl'
import fs from 'fs-extra'
import path from 'node:path'
import os from 'node:os'
import { Transform } from 'node:stream'
import throttle from 'lodash/throttle'
import zlib from 'node:zlib'
import tar from 'tar-stream'

interface ExtractProgress {
  progress: number
  status: 'extracting' | 'complete' | 'error'
  error?: string
}

// Number of parallel extractions (tune for your system).
const MAX_CONCURRENT_EXTRACTS = os.cpus().length * 2
const OnProgressUpdateRateMs = 300

const getTempPath = (p: string) => {
  // append .temp to the path folder name
  const dir = path.dirname(p)
  const base = path.basename(p)
  return path.join(dir, `${base}.temp`)
}

export async function extractFile(
  source: string,
  destination: string,
  onProgress?: (progress: ExtractProgress) => void
): Promise<void> {
  // check if the file is a zip or a tar.gz
  const ext = path.extname(source)
  if (ext === '.zip') {
    return extractZip(source, destination, onProgress)
  } else if (ext === '.gz') {
    return extractTarGz(source, destination, onProgress)
  }
  throw new Error(`Unsupported file extension: ${ext}`)
}

/**
 * Extracts a ZIP archive with optimal memory handling and parallelism.
 * @param {string} zipPath - Path to the ZIP file.
 * @param {string} outputDir - Destination folder for extraction.
 */
export async function extractZip(
  zipPath: string,
  outputDir: string,
  onProgress?: (progress: ExtractProgress) => void
): Promise<void> {
  const throttledOnProgress = onProgress
    ? throttle((progress: ExtractProgress) => onProgress(progress), OnProgressUpdateRateMs)
    : undefined

  if (fs.existsSync(outputDir)) {
    throttledOnProgress?.({
      progress: 0,
      status: 'complete'
    })
    throttledOnProgress?.flush()
    return Promise.resolve()
  }

  const outputDirTemp = getTempPath(outputDir)

  return new Promise((resolve, reject) => {
    let totalEntries = 0
    const entryQueue: yauzl.Entry[] = []

    const onError = (err: Error) => {
      const processedEntries = totalEntries - entryQueue.length
      throttledOnProgress?.({
        progress: processedEntries / totalEntries,
        status: 'error',
        error: err.message
      })
      throttledOnProgress?.flush()
      return reject(err)
    }

    yauzl.open(zipPath, { lazyEntries: true, autoClose: true }, (err, zipfile) => {
      if (err) {
        onError(err)
      }

      let activeExtracts = 0
      let done = false

      const processNextEntry = async () => {
        try {
          while (activeExtracts < MAX_CONCURRENT_EXTRACTS && entryQueue.length > 0) {
            const entry = entryQueue.shift()!
            await extractEntry(entry)
          }

          if (done && activeExtracts === 0 && entryQueue.length === 0) {
            await fs.move(outputDirTemp, outputDir)
            await fs.remove(zipPath)
            throttledOnProgress?.({ progress: 1, status: 'complete' })
            throttledOnProgress?.flush()
            resolve()
          }
        } catch (err) {
          return onError(err as Error)
        }
      }

      const extractEntry = (entry: yauzl.Entry): Promise<void> => {
        const outputPath = path.join(outputDirTemp, entry.fileName)

        if (entry.fileName.endsWith('/')) {
          return fs
            .ensureDir(outputPath)
            .then(() => {
              zipfile.readEntry()
              processNextEntry()
            })
            .catch(onError)
        }

        activeExtracts++

        zipfile.openReadStream(entry, (err, readStream) => {
          if (err) {
            return onError(err)
          }

          fs.ensureDir(path.dirname(outputPath))
            .then(() => {
              const writeStream = fs.createWriteStream(outputPath)
              const speedLimiter = new Transform({
                transform(chunk, _, callback) {
                  callback(null, chunk)
                }
              })

              readStream.pipe(speedLimiter).pipe(writeStream)

              writeStream.on('finish', () => {
                activeExtracts--
                const processedEntries = totalEntries - entryQueue.length
                throttledOnProgress?.({
                  progress: processedEntries / totalEntries,
                  status: 'extracting'
                })
                zipfile.readEntry()
                processNextEntry()
              })

              writeStream.on('error', onError)
              readStream.on('error', onError)
            })
            .catch(onError)
        })
        return Promise.resolve()
      }

      zipfile.on('entry', (entry) => {
        entryQueue.push(entry)
        totalEntries++
        processNextEntry()
      })

      zipfile.on('end', () => {
        done = true
        processNextEntry()
      })

      zipfile.on('error', onError)

      zipfile.readEntry() // Start processing entries.
    })
  })
}

// Function to extract .tar.gz
export async function extractTarGz(
  source: string,
  destination: string,
  onProgress?: (progress: ExtractProgress) => void
): Promise<void> {
  const throttledOnProgress = onProgress
    ? throttle((progress: ExtractProgress) => onProgress(progress), OnProgressUpdateRateMs)
    : undefined

  if (fs.existsSync(destination)) {
    throttledOnProgress?.({
      progress: 0,
      status: 'complete'
    })
    throttledOnProgress?.flush()
    return Promise.resolve()
  }

  const tempDestination = getTempPath(destination)

  // Get total file size for progress tracking
  const stats = await fs.stat(source)
  const totalSize = stats.size
  let processedSize = 0

  return new Promise((resolve, reject) => {
    const extract = tar.extract()
    const input = fs.createReadStream(source)
    const unzip = zlib.createGunzip()

    const onError = (err: Error) => {
      throttledOnProgress?.({
        progress: processedSize / totalSize,
        status: 'error',
        error: err.message
      })
      throttledOnProgress?.flush()
      reject(err)
    }

    input.on('error', onError)
    unzip.on('error', onError)
    extract.on('error', onError)

    // Track progress through the gunzip stream
    unzip.on('data', (chunk) => {
      processedSize += chunk.length
      throttledOnProgress?.({
        progress: processedSize / totalSize,
        status: 'extracting'
      })
    })

    input.pipe(unzip).pipe(extract)

    extract.on('entry', (header, stream, next) => {
      const outputPath = path.join(tempDestination, header.name)

      if (header.type === 'directory') {
        try {
          fs.mkdirSync(outputPath, { recursive: true })
          stream.resume()
          next()
        } catch (err) {
          onError(err as Error)
        }
      } else {
        // Ensure directories exist
        try {
          fs.mkdirSync(path.dirname(outputPath), { recursive: true })
          const output = fs.createWriteStream(outputPath)

          stream.pipe(output)

          output.on('finish', next)
          output.on('error', onError)
        } catch (err) {
          onError(err as Error)
        }
      }
    })

    extract.on('finish', async () => {
      try {
        await fs.move(tempDestination, destination, { overwrite: true })
        await fs.remove(source)
        throttledOnProgress?.({
          progress: 1,
          status: 'complete'
        }) // Signal 100% completion
        throttledOnProgress?.flush()
        resolve()
      } catch (err) {
        onError(err as Error)
      }
    })
  })
}
