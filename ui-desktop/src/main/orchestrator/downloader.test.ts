import { promises as fs } from 'node:fs'
import { createHash } from 'node:crypto'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { downloadFile, installBundledExecutable, writeAll } from './downloader'

let suiteDirectory: string

const responseFor = (contents: string) =>
  new Response(contents, {
    status: 200,
    headers: { 'content-length': String(Buffer.byteLength(contents)) }
  })

const createBundle = async (
  directoryName: string,
  contents: string,
  overrides: Record<string, unknown> = {}
) => {
  const bundleDirectory = path.join(suiteDirectory, directoryName)
  await fs.mkdir(bundleDirectory, { recursive: true })
  const executablePath = path.join(bundleDirectory, 'bundled-proxy-router')
  await fs.writeFile(executablePath, contents)
  await fs.writeFile(
    path.join(bundleDirectory, 'manifest.json'),
    JSON.stringify({
      schemaVersion: 1,
      platform: process.platform,
      arch: process.arch,
      size: Buffer.byteLength(contents),
      sha256: createHash('sha256').update(contents).digest('hex'),
      buildVersion: 'dev-test',
      commit: 'a'.repeat(40),
      dirty: false,
      ...overrides
    })
  )
  return { bundleDirectory, executablePath }
}

const expectNoTemporaryFilesFor = async (destination: string) => {
  const entries = await fs.readdir(path.dirname(destination))
  const prefix = `.${path.basename(destination)}.`
  expect(entries.filter((entry) => entry.startsWith(prefix) && entry.endsWith('.tmp'))).toEqual([])
}

beforeEach(async () => {
  suiteDirectory = await fs.mkdtemp(path.join(os.tmpdir(), 'morpheus-download-'))
})

afterEach(async () => {
  vi.unstubAllGlobals()
  await fs.rm(suiteDirectory, { recursive: true, force: true })
})

describe.sequential('version-aware downloads', () => {
  it('replaces a legacy cached service once and records a non-sensitive source fingerprint', async () => {
    const destination = path.join(suiteDirectory, 'proxy-router')
    const sourceUrl = 'https://example.test/proxy-router-v7.9.0?token=do-not-store'
    await fs.writeFile(destination, 'old-v7.3.0')

    const firstFetch = vi.fn(async () => responseFor('new-v7.9.0'))
    vi.stubGlobal('fetch', firstFetch)

    await downloadFile(sourceUrl, destination, undefined, undefined, {
      refreshIfSourceChanged: true
    })

    expect(firstFetch).toHaveBeenCalledTimes(1)
    expect(await fs.readFile(destination, 'utf8')).toBe('new-v7.9.0')

    const metadataText = await fs.readFile(`${destination}.download.json`, 'utf8')
    const metadata = JSON.parse(metadataText)
    expect(metadata).toEqual({
      version: 3,
      sourceUrlSha256: expect.stringMatching(/^[a-f0-9]{64}$/),
      contentSha256: expect.stringMatching(/^[a-f0-9]{64}$/)
    })
    expect(metadataText).not.toContain(sourceUrl)
    expect(metadataText).not.toContain('do-not-store')

    const secondFetch = vi.fn(async () => responseFor('must-not-download'))
    vi.stubGlobal('fetch', secondFetch)

    await downloadFile(sourceUrl, destination, undefined, undefined, {
      refreshIfSourceChanged: true
    })

    expect(secondFetch).not.toHaveBeenCalled()
    expect(await fs.readFile(destination, 'utf8')).toBe('new-v7.9.0')
  })

  it('installs a bundled executable by content and repairs a corrupted cached copy', async () => {
    const { bundleDirectory } = await createBundle('proxy-router-bundle', 'patched-router-build')
    const destination = path.join(suiteDirectory, 'services', 'proxy-router')

    await installBundledExecutable(bundleDirectory, destination)

    expect(await fs.readFile(destination, 'utf8')).toBe('patched-router-build')
    expect((await fs.stat(destination)).mode & 0o777).toBe(0o755)
    const metadataText = await fs.readFile(`${destination}.download.json`, 'utf8')
    expect(JSON.parse(metadataText)).toEqual({
      version: 2,
      contentSha256: expect.stringMatching(/^[a-f0-9]{64}$/)
    })
    expect(metadataText).not.toContain(bundleDirectory)

    await fs.writeFile(destination, 'corrupted')
    await installBundledExecutable(bundleDirectory, destination)

    expect(await fs.readFile(destination, 'utf8')).toBe('patched-router-build')
  })

  it('preserves the installed executable when a bundled source is empty', async () => {
    const { bundleDirectory } = await createBundle('empty-proxy-router-bundle', '', { size: 1 })
    const destination = path.join(suiteDirectory, 'proxy-router')
    await fs.writeFile(destination, 'working-router')

    await expect(installBundledExecutable(bundleDirectory, destination)).rejects.toThrow(
      'size does not match'
    )

    expect(await fs.readFile(destination, 'utf8')).toBe('working-router')
  })

  it('rejects a bundle for another architecture without touching the installed router', async () => {
    const { bundleDirectory } = await createBundle('wrong-architecture-bundle', 'new-router', {
      arch: process.arch === 'arm64' ? 'x64' : 'arm64'
    })
    const destination = path.join(suiteDirectory, 'proxy-router')
    await fs.writeFile(destination, 'working-router')

    await expect(installBundledExecutable(bundleDirectory, destination)).rejects.toThrow(
      'does not match'
    )
    expect(await fs.readFile(destination, 'utf8')).toBe('working-router')
  })

  it('rejects a bundle with a mismatched hash without touching the installed router', async () => {
    const { bundleDirectory } = await createBundle('corrupt-bundle', 'new-router', {
      sha256: '0'.repeat(64)
    })
    const destination = path.join(suiteDirectory, 'proxy-router')
    await fs.writeFile(destination, 'working-router')

    await expect(installBundledExecutable(bundleDirectory, destination)).rejects.toThrow(
      'hash does not match'
    )
    expect(await fs.readFile(destination, 'utf8')).toBe('working-router')
  })

  it('retries short writes until the complete chunk is persisted', async () => {
    const input = new TextEncoder().encode('complete-download')
    const calls: Array<{ offset: number; length: number }> = []
    const writer = {
      async write(_buffer: Uint8Array, offset: number, length: number) {
        calls.push({ offset, length })
        return { bytesWritten: Math.min(3, length) }
      }
    }

    await writeAll(writer, input)

    expect(calls.length).toBeGreaterThan(1)
    expect(calls.at(-1)!.offset + Math.min(3, calls.at(-1)!.length)).toBe(input.length)
  })

  it('refreshes a locally corrupted URL-cached executable', async () => {
    const destination = path.join(suiteDirectory, 'proxy-router')
    const sourceUrl = 'https://example.test/proxy-router'
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => responseFor('verified-router'))
    )
    await downloadFile(sourceUrl, destination, undefined, undefined, {
      refreshIfSourceChanged: true
    })
    await fs.writeFile(destination, 'locally-corrupted')

    const repairFetch = vi.fn(async () => responseFor('verified-router'))
    vi.stubGlobal('fetch', repairFetch)
    await downloadFile(sourceUrl, destination, undefined, undefined, {
      refreshIfSourceChanged: true
    })

    expect(repairFetch).toHaveBeenCalledTimes(1)
    expect(await fs.readFile(destination, 'utf8')).toBe('verified-router')
  })

  it('atomically refreshes when the configured artifact changes', async () => {
    const destination = path.join(suiteDirectory, 'proxy-router')

    vi.stubGlobal(
      'fetch',
      vi.fn(async () => responseFor('v7.9.0'))
    )
    await downloadFile('https://example.test/v7.9.0', destination, undefined, undefined, {
      refreshIfSourceChanged: true
    })

    const changedFetch = vi.fn(async () => responseFor('v7.10.0'))
    vi.stubGlobal('fetch', changedFetch)
    await downloadFile('https://example.test/v7.10.0', destination, undefined, undefined, {
      refreshIfSourceChanged: true
    })

    expect(changedFetch).toHaveBeenCalledTimes(1)
    expect(await fs.readFile(destination, 'utf8')).toBe('v7.10.0')
  })

  it('preserves the installed executable when an attempted refresh fails', async () => {
    const destination = path.join(suiteDirectory, 'proxy-router')
    await fs.writeFile(destination, 'working-old-version')

    // Force the streamed body to disagree with its declared length.
    const response = responseFor('truncated')
    response.headers.set('content-length', '999')
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => response)
    )

    await expect(
      downloadFile('https://example.test/new', destination, undefined, undefined, {
        refreshIfSourceChanged: true
      })
    ).rejects.toThrow('Download incomplete')

    expect(await fs.readFile(destination, 'utf8')).toBe('working-old-version')
    await expectNoTemporaryFilesFor(destination)
  })

  it('stops a stalled response body without replacing the installed executable', async () => {
    const destination = path.join(suiteDirectory, 'proxy-router')
    await fs.writeFile(destination, 'working-router')
    let bodyCancelled = false

    vi.stubGlobal(
      'fetch',
      vi.fn(
        async () =>
          new Response(
            new ReadableStream<Uint8Array>({
              start(controller) {
                controller.enqueue(new TextEncoder().encode('partial'))
              },
              cancel() {
                bodyCancelled = true
              }
            }),
            {
              status: 200,
              headers: { 'content-length': '100' }
            }
          )
      )
    )

    await expect(
      downloadFile('https://example.test/stalled', destination, undefined, undefined, {
        refreshIfSourceChanged: true,
        bodyIdleTimeoutMs: 20
      })
    ).rejects.toThrow('Download stalled')

    expect(bodyCancelled).toBe(true)
    expect(await fs.readFile(destination, 'utf8')).toBe('working-router')
    await expectNoTemporaryFilesFor(destination)
  })

  it('keeps the existing-file cache behavior unless refresh is explicitly enabled', async () => {
    const destination = path.join(suiteDirectory, 'large-optional-asset')
    await fs.writeFile(destination, 'cached')
    const fetchMock = vi.fn(async () => responseFor('replacement'))
    vi.stubGlobal('fetch', fetchMock)

    await downloadFile('https://example.test/asset', destination)

    expect(fetchMock).not.toHaveBeenCalled()
    expect(await fs.readFile(destination, 'utf8')).toBe('cached')
  })
})
