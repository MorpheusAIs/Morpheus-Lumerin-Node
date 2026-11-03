import { BrowserWindow, dialog } from 'electron'
import { lookup as dnsLookup } from 'node:dns/promises'
import { isIP } from 'node:net'
import fs from 'node:fs'
import { Readable } from 'node:stream'
import { pipeline } from 'node:stream/promises'
import { isPublicCoworkWebAddress } from '../cowork-web'
import { proxyFetch } from './handlers'

// Mirrors videojob.MaxPromptLength in proxy-router, which counts bytes.
const MAX_PROMPT_BYTES = 4000
// A 10 MiB first frame as base64, plus room for a data-URI prefix.
const MAX_IMAGE_B64_CHARS = Math.ceil((10 * 1024 * 1024 * 4) / 3) + 64
// A played clip crosses IPC whole, so it is held well under Chromium's message
// limit; a saved clip streams to disk.
const MAX_PLAY_BYTES = 100 * 1024 * 1024
const MAX_SAVE_BYTES = 1024 * 1024 * 1024
const MAX_REDIRECTS = 5
const PING_TIMEOUT_MS = 15_000
// A download fails when it stops making progress, however long it takes.
const CLIP_STALL_MS = 60_000
const MAX_GUIDE_OFFERS = 20

const BLOCKED_HOST_SUFFIXES = ['.localhost', '.local', '.internal', '.home.arpa', '.localdomain']

export type VideoJobRequest = {
  prompt: string
  duration_seconds?: number
  aspect_ratio?: string
  resolution?: string
  generate_audio?: boolean
  image_b64?: string
}

const boundedId = (value: unknown, label: string): string => {
  const result = String(value ?? '').trim()
  if (!/^0x[0-9a-fA-F]{1,64}$/u.test(result)) throw new Error(`${label} is invalid.`)
  return result
}

const jobId = (value: unknown): string => {
  const result = String(value ?? '').trim()
  if (!result || result.length > 128 || !/^[\w.:-]+$/u.test(result)) {
    throw new Error('Video job ID is invalid.')
  }
  return result
}

const optionWord = (value: unknown, label: string): string | undefined => {
  if (value === undefined || value === null || value === '') return undefined
  const result = String(value).trim()
  if (result.length > 32 || !/^[\w:.-]+$/u.test(result)) throw new Error(`${label} is invalid.`)
  return result
}

/** The job as the renderer described it, checked before it reaches the node. */
export function videoJobRequest(raw: unknown): VideoJobRequest {
  const job = (raw && typeof raw === 'object' ? raw : {}) as Record<string, unknown>
  const prompt = typeof job.prompt === 'string' ? job.prompt.trim() : ''
  if (!prompt) throw new Error('Describe the clip first.')
  if (Buffer.byteLength(prompt, 'utf8') > MAX_PROMPT_BYTES) {
    throw new Error(`The prompt is longer than ${MAX_PROMPT_BYTES} bytes.`)
  }
  const request: VideoJobRequest = { prompt }

  if (job.duration_seconds !== undefined && job.duration_seconds !== null) {
    const duration = Number(job.duration_seconds)
    if (!Number.isSafeInteger(duration) || duration < 1 || duration > 600) {
      throw new Error('Clip length is invalid.')
    }
    request.duration_seconds = duration
  }
  const aspectRatio = optionWord(job.aspect_ratio, 'Aspect ratio')
  if (aspectRatio) request.aspect_ratio = aspectRatio
  const resolution = optionWord(job.resolution, 'Resolution')
  if (resolution) request.resolution = resolution
  if (job.generate_audio !== undefined && job.generate_audio !== null) {
    if (typeof job.generate_audio !== 'boolean') throw new Error('Audio choice is invalid.')
    request.generate_audio = job.generate_audio
  }
  if (job.image_b64 !== undefined && job.image_b64 !== null && job.image_b64 !== '') {
    const image = job.image_b64
    if (
      typeof image !== 'string' ||
      image.length > MAX_IMAGE_B64_CHARS ||
      !/^data:image\/(jpeg|png|webp);base64,[A-Za-z0-9+/]+=*$/u.test(image)
    ) {
      throw new Error('The first frame must be a JPEG, PNG or WebP image up to 10 MB.')
    }
    request.image_b64 = image
  }
  return request
}

type RatedBid = { Bid?: { Id?: string; Provider?: string } }
type ProviderRow = { Address?: string; Endpoint?: string; IsDeleted?: boolean }
type HealthReport = { modelId?: string; status?: string; video?: unknown }

export type VideoGuideOffer = {
  provider: string
  bidId: string
  status?: string
  guide?: unknown
  error?: string
}

/**
 * Each provider's job guide for a video model: the fields a job may set, their
 * limits and defaults. Providers publish it in their healthcheck pong, signed,
 * and the node drops a guide the provider did not sign, so a guide here is the
 * one the provider prices the session by.
 */
export const getVideoGuide = async (payload: {
  modelId: string
}): Promise<{ offers: VideoGuideOffer[] }> => {
  const modelId = boundedId(payload?.modelId, 'Model ID')
  const [rated, providers] = await Promise.all([
    proxyFetch<{ bids: RatedBid[] }>(
      `/blockchain/models/${encodeURIComponent(modelId)}/bids/rated`,
      {},
      'rated model bids'
    ),
    proxyFetch<{ providers: ProviderRow[] }>('/blockchain/providers', {}, 'providers')
  ])

  const endpoints = new Map<string, string>()
  for (const row of providers.providers ?? []) {
    if (row?.Address && row.Endpoint && !row.IsDeleted) {
      endpoints.set(row.Address.toLowerCase(), row.Endpoint)
    }
  }

  const bids = (rated.bids ?? []).filter((entry) => entry?.Bid?.Id && entry.Bid.Provider)
  const offers = await Promise.all(
    bids.slice(0, MAX_GUIDE_OFFERS).map(async ({ Bid: bid }): Promise<VideoGuideOffer> => {
      const provider = String(bid!.Provider)
      const offer: VideoGuideOffer = { provider, bidId: String(bid!.Id) }
      const endpoint = endpoints.get(provider.toLowerCase())
      if (!endpoint) return { ...offer, error: 'The provider has no registered endpoint.' }
      const signal = AbortSignal.timeout(PING_TIMEOUT_MS)
      try {
        const pong = await proxyFetch<{ models?: HealthReport[] }>(
          '/proxy/provider/ping',
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ providerAddr: provider, providerUrl: endpoint }),
            signal
          },
          'provider ping'
        )
        const report = (pong.models ?? []).find(
          (model) => model?.modelId?.toLowerCase() === modelId.toLowerCase()
        )
        if (!report?.video) {
          return {
            ...offer,
            status: report?.status,
            error: 'The provider publishes no video guide for this model.'
          }
        }
        return { ...offer, status: report.status, guide: report.video }
      } catch (error: any) {
        // proxyFetch reports an aborted request as an unreachable node.
        const reason = signal.aborted ? '' : error?.message
        return { ...offer, error: reason || 'The provider did not answer.' }
      }
    })
  )
  return { offers }
}

/** What the job would cost at every provider of the model. */
export const quoteVideoJob = async (payload: {
  modelId: string
  job: unknown
  hasImage?: boolean
}): Promise<unknown[]> => {
  const modelId = boundedId(payload?.modelId, 'Model ID')
  const request = videoJobRequest(payload?.job)
  // The quote only asks whether the job has a first frame; the frame itself
  // goes with the job.
  if (payload?.hasImage === true && !request.image_b64) request.image_b64 = 'first-frame'
  const data = await proxyFetch<{ data?: unknown[] }>(
    '/v1/video/quote',
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...request, model: modelId })
    },
    'video quote'
  )
  return data.data ?? []
}

/** Submits a job on a session the user already opened for it. */
export const createVideoJob = async (payload: {
  sessionId: string
  job: unknown
}): Promise<unknown> => {
  const sessionId = boundedId(payload?.sessionId, 'Session ID')
  const request = videoJobRequest(payload?.job)
  return proxyFetch(
    '/v1/video/jobs',
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', session_id: sessionId },
      body: JSON.stringify(request)
    },
    'video job'
  )
}

export const getVideoJobs = async (): Promise<unknown[]> => {
  const data = await proxyFetch<{ data?: unknown[] }>('/v1/video/jobs', {}, 'video jobs')
  return data.data ?? []
}

type VideoJob = { id?: string; status?: string; video_url?: string }

const deliveredClipUrl = async (id: string): Promise<URL> => {
  const job = await proxyFetch<VideoJob>(
    `/v1/video/jobs/${encodeURIComponent(id)}`,
    {},
    'video job'
  )
  if (job?.status !== 'succeeded' || !job.video_url) {
    throw new Error('This clip has not been delivered.')
  }
  return new URL(job.video_url)
}

/**
 * The node only relays a public https link, but the vendor's CDN may redirect,
 * so every hop is checked again: https, a public hostname and public addresses.
 */
const checkPublicHttps = async (url: URL): Promise<void> => {
  if (url.protocol !== 'https:' || url.username || url.password) {
    throw new Error('The clip link is not a public https link.')
  }
  const host = url.hostname
    .replace(/^\[|\]$/g, '')
    .replace(/\.$/, '')
    .toLowerCase()
  if (
    !host ||
    host === 'localhost' ||
    (!host.includes('.') && !isIP(host)) ||
    BLOCKED_HOST_SUFFIXES.some((suffix) => host.endsWith(suffix))
  ) {
    throw new Error('The clip link is not a public https link.')
  }
  const addresses = isIP(host)
    ? [host]
    : (await dnsLookup(host, { all: true, verbatim: true })).map((answer) => answer.address)
  if (!addresses.length || !addresses.every(isPublicCoworkWebAddress)) {
    throw new Error('The clip link is not a public https link.')
  }
}

const fetchClip = async (start: URL, signal: AbortSignal, maxBytes: number): Promise<Response> => {
  let url = start
  for (let hop = 0; hop <= MAX_REDIRECTS; hop += 1) {
    await checkPublicHttps(url)
    const response = await fetch(url, { redirect: 'manual', signal })
    if (response.status >= 300 && response.status < 400) {
      const location = response.headers.get('location')
      await response.body?.cancel().catch(() => undefined)
      if (!location) throw new Error('The clip link redirected nowhere.')
      url = new URL(location, url)
      continue
    }
    if (!response.ok) {
      await response.body?.cancel().catch(() => undefined)
      throw new Error(
        response.status === 403 || response.status === 404 || response.status === 410
          ? 'The clip link has expired.'
          : `The vendor answered HTTP ${response.status}.`
      )
    }
    const declared = Number(response.headers.get('content-length') ?? 0)
    if (Number.isFinite(declared) && declared > maxBytes) {
      await response.body?.cancel().catch(() => undefined)
      throw new Error('The clip is too large to load.')
    }
    return response
  }
  throw new Error('The clip link redirected too many times.')
}

/** An abort signal that fires once no progress was reported for `ms`. */
const stallGuard = (ms: number) => {
  const controller = new AbortController()
  let timer: ReturnType<typeof setTimeout> | undefined
  const progress = () => {
    clearTimeout(timer)
    timer = setTimeout(() => controller.abort(new Error('The clip download stalled.')), ms)
  }
  progress()
  return { signal: controller.signal, progress, done: () => clearTimeout(timer) }
}

const clipMimeType = (response: Response): string => {
  const type = (response.headers.get('content-type') ?? '').split(';')[0].trim().toLowerCase()
  return type.startsWith('video/') ? type : 'video/mp4'
}

/**
 * Loads a delivered clip so the renderer can play it from a blob. The link
 * comes from the node for this job id, never from the renderer.
 */
export const loadVideoClip = async (payload: {
  id: string
}): Promise<{ mimeType: string; data: ArrayBuffer }> => {
  const url = await deliveredClipUrl(jobId(payload?.id))
  const guard = stallGuard(CLIP_STALL_MS)
  const chunks: Buffer[] = []
  let total = 0
  let response: Response
  try {
    response = await fetchClip(url, guard.signal, MAX_PLAY_BYTES)
    const reader = response.body!.getReader()
    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      guard.progress()
      total += value.byteLength
      if (total > MAX_PLAY_BYTES) {
        await reader.cancel().catch(() => undefined)
        throw new Error('The clip is too large to play here. Save it instead.')
      }
      chunks.push(Buffer.from(value))
    }
  } finally {
    guard.done()
  }
  const clip = Buffer.concat(chunks, total)
  return {
    mimeType: clipMimeType(response),
    data: clip.buffer.slice(clip.byteOffset, clip.byteOffset + clip.byteLength)
  }
}

/** Saves a delivered clip to a file the user picks. */
export const saveVideoClip = async (payload: {
  id: string
}): Promise<{ saved: boolean; path?: string }> => {
  const id = jobId(payload?.id)
  const url = await deliveredClipUrl(id)
  const owner =
    BrowserWindow.getFocusedWindow() ??
    BrowserWindow.getAllWindows().find((window) => !window.isDestroyed())
  const options = {
    title: 'Save video',
    defaultPath: `morpheus-video-${id.replace(/[^\w-]/g, '').slice(0, 16)}.mp4`,
    filters: [{ name: 'Video', extensions: ['mp4'] }]
  }
  const choice = owner
    ? await dialog.showSaveDialog(owner, options)
    : await dialog.showSaveDialog(options)
  if (choice.canceled || !choice.filePath) return { saved: false }

  // Written beside the target and moved over it once complete, so a failed
  // download never leaves a file the user chose to overwrite half written.
  const partial = `${choice.filePath}.part`
  const guard = stallGuard(CLIP_STALL_MS)
  try {
    const response = await fetchClip(url, guard.signal, MAX_SAVE_BYTES)
    let total = 0
    const limited = Readable.fromWeb(response.body as any).map((chunk: Buffer) => {
      guard.progress()
      total += chunk.byteLength
      if (total > MAX_SAVE_BYTES) throw new Error('The clip is too large to save.')
      return chunk
    })
    await pipeline(limited, fs.createWriteStream(partial))
    await fs.promises.rename(partial, choice.filePath)
  } catch (error) {
    await fs.promises.rm(partial, { force: true }).catch(() => undefined)
    throw error
  } finally {
    guard.done()
  }
  return { saved: true, path: choice.filePath }
}
