import { lookup as dnsLookup } from 'node:dns/promises'
import { request as httpsRequest } from 'node:https'
import { isIP } from 'node:net'
import type { LookupAddress } from 'node:dns'
import type { IncomingMessage } from 'node:http'

export const COWORK_WEB_LIMITS = Object.freeze({
  redirects: 5,
  bodyBytes: 1024 * 1024,
  textCharacters: 250_000,
  titleCharacters: 512,
  defaultTimeoutMs: 12_000,
  maximumTimeoutMs: 30_000
})

export type CoworkWebErrorCode =
  | 'invalid-url'
  | 'blocked-address'
  | 'dns-failed'
  | 'request-failed'
  | 'timeout'
  | 'aborted'
  | 'redirect-limit'
  | 'redirect-loop'
  | 'cross-origin-redirect'
  | 'http-status'
  | 'unsupported-content-type'
  | 'unsupported-content-encoding'

export class CoworkWebError extends Error {
  constructor(
    readonly code: CoworkWebErrorCode,
    message: string,
    options?: ErrorOptions
  ) {
    super(message, options)
    this.name = 'CoworkWebError'
  }
}

export interface CoworkWebResult {
  finalUrl: string
  title: string | null
  text: string
  contentType: 'text/html' | 'text/plain' | 'application/json'
  truncated: boolean
}

export interface CoworkWebOptions {
  /** Total wall-clock budget, including DNS and redirects. Capped at 30 seconds. */
  timeoutMs?: number
  signal?: AbortSignal
}

interface AddressBinding {
  address: string
  family: 4 | 6
}

interface HopResponse {
  statusCode: number
  location: string | null
  contentType: string | null
  contentEncoding: string | null
  body: Buffer
  truncated: boolean
}

const ACCEPTED_CONTENT_TYPES: ReadonlySet<string> = new Set([
  'text/html',
  'text/plain',
  'application/json'
])

const REDIRECT_STATUS_CODES = new Set([301, 302, 303, 307, 308])
const BLOCKED_HOST_SUFFIXES = ['.localhost', '.local', '.internal', '.home.arpa']

const ipv4Ranges: ReadonlyArray<readonly [number, number]> = [
  [0x00000000, 8], // Current network, unspecified, and historical special-use space.
  [0x0a000000, 8], // RFC 1918.
  [0x64400000, 10], // Carrier-grade NAT, including common metadata endpoints.
  [0x7f000000, 8], // Loopback.
  [0xa83f8110, 32], // Azure platform virtual address (168.63.129.16).
  [0xa9fe0000, 16], // Link-local and cloud instance metadata.
  [0xac100000, 12], // RFC 1918.
  [0xc0000000, 24], // IETF protocol assignments and special service endpoints.
  [0xc0000200, 24], // TEST-NET-1.
  [0xc0586300, 24], // Deprecated 6to4 relay anycast.
  [0xc0a80000, 16], // RFC 1918.
  [0xc6120000, 15], // Benchmarking.
  [0xc6336400, 24], // TEST-NET-2.
  [0xcb007100, 24], // TEST-NET-3.
  [0xe0000000, 4], // Multicast.
  [0xf0000000, 4] // Reserved and broadcast.
]

const ipv6Ranges: ReadonlyArray<readonly [bigint, number]> = [
  [0n, 96], // Unspecified, IPv4-compatible, and other low-address aliases.
  [1n, 128], // Loopback.
  [0x00000000000000000000ffff00000000n, 96], // IPv4-mapped aliases.
  [0x0064ff9b000000000000000000000000n, 96], // Well-known NAT64 aliases.
  [0x0064ff9b000100000000000000000000n, 48], // Local-use NAT64 aliases.
  [0x01000000000000000000000000000000n, 64], // Discard-only prefix.
  [0x20010000000000000000000000000000n, 23], // IETF special-purpose allocations.
  [0x20010db8000000000000000000000000n, 32], // Documentation.
  [0x20020000000000000000000000000000n, 16], // 6to4 aliases can encode private IPv4.
  [0x3fff0000000000000000000000000000n, 20], // Documentation.
  [0xfc000000000000000000000000000000n, 7], // Unique-local.
  [0xfe800000000000000000000000000000n, 10], // Link-local.
  [0xfec00000000000000000000000000000n, 10], // Deprecated site-local.
  [0xff000000000000000000000000000000n, 8] // Multicast.
]

const inNumberCidr = (value: number, network: number, prefix: number): boolean => {
  if (prefix === 0) return true
  const mask = (0xffffffff << (32 - prefix)) >>> 0
  return (value & mask) >>> 0 === (network & mask) >>> 0
}

const inBigIntCidr = (value: bigint, network: bigint, prefix: number): boolean => {
  if (prefix === 0) return true
  const shift = BigInt(128 - prefix)
  return value >> shift === network >> shift
}

const ipv4ToNumber = (address: string): number | null => {
  const parts = address.split('.')
  if (parts.length !== 4) return null
  let value = 0
  for (const part of parts) {
    if (!/^\d{1,3}$/.test(part)) return null
    const octet = Number(part)
    if (octet > 255) return null
    value = (value * 256 + octet) >>> 0
  }
  return value
}

const ipv6ToBigInt = (rawAddress: string): bigint | null => {
  if (rawAddress.includes('%')) return null
  let address = rawAddress.toLowerCase()

  if (address.includes('.')) {
    const lastColon = address.lastIndexOf(':')
    if (lastColon < 0) return null
    const ipv4 = ipv4ToNumber(address.slice(lastColon + 1))
    if (ipv4 === null) return null
    address = `${address.slice(0, lastColon)}:${(ipv4 >>> 16).toString(16)}:${(
      ipv4 & 0xffff
    ).toString(16)}`
  }

  if ((address.match(/::/g) ?? []).length > 1) return null
  const [leftText, rightText] = address.split('::')
  const left = leftText ? leftText.split(':') : []
  const right = rightText === undefined || rightText === '' ? [] : rightText.split(':')

  if (address.includes('::')) {
    const missing = 8 - left.length - right.length
    if (missing < 1) return null
    left.push(...Array(missing).fill('0'))
  } else if (left.length !== 8) {
    return null
  }

  const groups = [...left, ...right]
  if (groups.length !== 8) return null
  let value = 0n
  for (const group of groups) {
    if (!/^[0-9a-f]{1,4}$/.test(group)) return null
    value = (value << 16n) | BigInt(`0x${group}`)
  }
  return value
}

/** Returns true only for ordinary, globally routable unicast addresses. */
export const isPublicCoworkWebAddress = (address: string): boolean => {
  const version = isIP(address)
  if (version === 4) {
    const value = ipv4ToNumber(address)
    return (
      value !== null &&
      !ipv4Ranges.some(([network, prefix]) => inNumberCidr(value, network, prefix))
    )
  }
  if (version === 6) {
    const value = ipv6ToBigInt(address)
    return (
      value !== null &&
      !ipv6Ranges.some(([network, prefix]) => inBigIntCidr(value, network, prefix))
    )
  }
  return false
}

const validateUrl = (rawUrl: string): URL => {
  if (typeof rawUrl !== 'string' || rawUrl.length === 0 || rawUrl.length > 8192) {
    throw new CoworkWebError('invalid-url', 'A bounded HTTPS URL is required')
  }
  // URL.hash cannot distinguish an absent fragment from an empty trailing '#'.
  if (rawUrl.includes('#')) {
    throw new CoworkWebError('invalid-url', 'URL fragments are not allowed')
  }

  let url: URL
  try {
    url = new URL(rawUrl)
  } catch (error) {
    throw new CoworkWebError('invalid-url', 'The web URL is invalid', { cause: error })
  }

  if (url.protocol !== 'https:') {
    throw new CoworkWebError('invalid-url', 'Only HTTPS web URLs are allowed')
  }
  if (url.username || url.password) {
    throw new CoworkWebError('invalid-url', 'Credentials are not allowed in web URLs')
  }
  if (url.hash) {
    throw new CoworkWebError('invalid-url', 'URL fragments are not allowed')
  }

  const hostname = url.hostname
    .replace(/^\[|\]$/g, '')
    .replace(/\.$/, '')
    .toLowerCase()
  if (
    !hostname ||
    hostname === 'localhost' ||
    BLOCKED_HOST_SUFFIXES.some((suffix) => hostname.endsWith(suffix))
  ) {
    throw new CoworkWebError('blocked-address', 'Local and private web hosts are not allowed')
  }
  return url
}

const abortError = (): CoworkWebError =>
  new CoworkWebError('aborted', 'The web request was aborted')

const raceWithAbort = <T>(operation: Promise<T>, signal: AbortSignal): Promise<T> => {
  if (signal.aborted) return Promise.reject(abortError())
  return new Promise<T>((resolve, reject) => {
    const onAbort = (): void => reject(abortError())
    signal.addEventListener('abort', onAbort, { once: true })
    operation.then(
      (value) => {
        signal.removeEventListener('abort', onAbort)
        resolve(value)
      },
      (error) => {
        signal.removeEventListener('abort', onAbort)
        reject(error)
      }
    )
  })
}

const resolvePublicAddress = async (url: URL, signal: AbortSignal): Promise<AddressBinding> => {
  const hostname = url.hostname.replace(/^\[|\]$/g, '').replace(/\.$/, '')
  const literalFamily = isIP(hostname)
  if (literalFamily) {
    if (!isPublicCoworkWebAddress(hostname)) {
      throw new CoworkWebError('blocked-address', 'The web host resolves to a blocked address')
    }
    return { address: hostname, family: literalFamily as 4 | 6 }
  }

  let answers: LookupAddress[]
  try {
    answers = await raceWithAbort(dnsLookup(hostname, { all: true, verbatim: true }), signal)
  } catch (error) {
    if (error instanceof CoworkWebError) throw error
    throw new CoworkWebError('dns-failed', 'The web host could not be resolved', { cause: error })
  }

  if (answers.length === 0) {
    throw new CoworkWebError('dns-failed', 'The web host returned no addresses')
  }
  if (answers.some(({ address }) => !isPublicCoworkWebAddress(address))) {
    throw new CoworkWebError(
      'blocked-address',
      'The web host resolves to a local, private, special-use, or metadata address'
    )
  }

  const selected = answers[0]
  if (selected.family !== 4 && selected.family !== 6) {
    throw new CoworkWebError('dns-failed', 'The web host returned an unsupported address family')
  }
  return { address: selected.address, family: selected.family }
}

const firstHeader = (header: string | string[] | undefined): string | null => {
  if (typeof header === 'string') return header
  if (Array.isArray(header)) return header[0] ?? null
  return null
}

const readResponse = (
  response: IncomingMessage,
  statusCode: number,
  signal: AbortSignal,
  settle: (response: HopResponse) => void,
  fail: (error: Error) => void
): void => {
  const location = firstHeader(response.headers.location)
  if (REDIRECT_STATUS_CODES.has(statusCode)) {
    settle({
      statusCode,
      location,
      contentType: null,
      contentEncoding: null,
      body: Buffer.alloc(0),
      truncated: false
    })
    response.destroy()
    return
  }

  if (statusCode < 200 || statusCode >= 300) {
    fail(new CoworkWebError('http-status', `The web server returned HTTP ${statusCode}`))
    response.destroy()
    return
  }

  const rawContentType = firstHeader(response.headers['content-type'])
  const contentType = rawContentType?.split(';', 1)[0]?.trim().toLowerCase() ?? null
  if (!contentType || !ACCEPTED_CONTENT_TYPES.has(contentType)) {
    fail(
      new CoworkWebError(
        'unsupported-content-type',
        'The web response is not HTML, plain text, or JSON'
      )
    )
    response.destroy()
    return
  }

  const contentEncoding =
    firstHeader(response.headers['content-encoding'])?.trim().toLowerCase() ?? null
  if (contentEncoding && contentEncoding !== 'identity') {
    fail(
      new CoworkWebError(
        'unsupported-content-encoding',
        'Compressed web responses are not accepted'
      )
    )
    response.destroy()
    return
  }

  const chunks: Buffer[] = []
  let bytes = 0
  let truncated = false

  response.on('data', (chunk: Buffer | string) => {
    if (signal.aborted || truncated) return
    const buffer = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk)
    const remaining = COWORK_WEB_LIMITS.bodyBytes - bytes
    if (buffer.length > remaining) {
      if (remaining > 0) chunks.push(buffer.subarray(0, remaining))
      bytes += Math.max(remaining, 0)
      truncated = true
      settle({
        statusCode,
        location: null,
        contentType,
        contentEncoding,
        body: Buffer.concat(chunks, bytes),
        truncated: true
      })
      response.destroy()
      return
    }
    chunks.push(buffer)
    bytes += buffer.length
  })

  response.once('end', () => {
    if (truncated) return
    settle({
      statusCode,
      location: null,
      contentType,
      contentEncoding,
      body: Buffer.concat(chunks, bytes),
      truncated: false
    })
  })
  response.once('aborted', () => {
    if (!truncated)
      fail(new CoworkWebError('request-failed', 'The web response ended unexpectedly'))
  })
  response.once('error', (error) => {
    if (!truncated)
      fail(new CoworkWebError('request-failed', 'The web response failed', { cause: error }))
  })
}

const requestHop = (url: URL, binding: AddressBinding, signal: AbortSignal): Promise<HopResponse> =>
  new Promise<HopResponse>((resolve, reject) => {
    if (signal.aborted) {
      reject(abortError())
      return
    }

    let settled = false
    const settle = (response: HopResponse): void => {
      if (settled) return
      settled = true
      signal.removeEventListener('abort', onAbort)
      resolve(response)
    }
    const fail = (error: Error): void => {
      if (settled) return
      settled = true
      signal.removeEventListener('abort', onAbort)
      reject(error)
    }

    const request = httpsRequest(
      {
        protocol: 'https:',
        hostname: url.hostname.replace(/^\[|\]$/g, ''),
        port: url.port ? Number(url.port) : 443,
        path: `${url.pathname}${url.search}`,
        method: 'GET',
        agent: false,
        headers: {
          Accept: 'text/html, text/plain, application/json;q=0.9',
          'Accept-Encoding': 'identity',
          'User-Agent': 'Morpheus-Workspace-Web/1.0'
        },
        lookup: (_hostname, _options, callback) => callback(null, binding.address, binding.family)
      },
      (response) => readResponse(response, response.statusCode ?? 0, signal, settle, fail)
    )

    const onAbort = (): void => {
      const error = abortError()
      request.destroy(error)
      fail(error)
    }
    signal.addEventListener('abort', onAbort, { once: true })
    request.once('error', (error) => {
      if (error instanceof CoworkWebError) fail(error)
      else fail(new CoworkWebError('request-failed', 'The HTTPS request failed', { cause: error }))
    })
    request.end()
  })

const decodeEntities = (text: string): string => {
  const named: Record<string, string> = {
    amp: '&',
    apos: "'",
    bull: '•',
    copy: '©',
    gt: '>',
    hellip: '…',
    laquo: '«',
    lt: '<',
    mdash: '—',
    middot: '·',
    nbsp: ' ',
    ndash: '–',
    quot: '"',
    raquo: '»',
    reg: '®',
    trade: '™'
  }
  return text.replace(
    /&(?:#(\d{1,7})|#x([0-9a-f]{1,6})|([a-z][a-z0-9]{1,31}));/gi,
    (match, dec, hex, name) => {
      if (name) return named[String(name).toLowerCase()] ?? match
      const codePoint = Number.parseInt(dec ?? hex, dec ? 10 : 16)
      if (!Number.isFinite(codePoint) || codePoint <= 0 || codePoint > 0x10ffff) return '�'
      try {
        return String.fromCodePoint(codePoint)
      } catch {
        return '�'
      }
    }
  )
}

const normalizeReadableText = (text: string): string =>
  decodeEntities(text)
    .replace(/\r\n?/g, '\n')
    .replace(/[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f]/g, '')
    .replace(/[\t ]+/g, ' ')
    .replace(/ *\n */g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()

const findTagEnd = (html: string, start: number): number => {
  let quote: '"' | "'" | null = null
  for (let index = start + 1; index < html.length; index += 1) {
    const character = html[index]
    if (quote) {
      if (character === quote) quote = null
    } else if (character === '"' || character === "'") {
      quote = character
    } else if (character === '>') {
      return index
    }
  }
  return -1
}

const plainTitle = (html: string): string | null => {
  const match = /<title\b[^>]*>([\s\S]*?)<\/title\s*>/i.exec(html)
  if (!match) return null
  const title = normalizeReadableText(match[1].replace(/<[^>]*>/g, ' '))
  return title ? title.slice(0, COWORK_WEB_LIMITS.titleCharacters) : null
}

const htmlToReadableText = (html: string): { title: string | null; text: string } => {
  const blockedTags = new Set(['head', 'script', 'style', 'noscript', 'template', 'svg'])
  const breakTags = new Set([
    'address',
    'article',
    'aside',
    'blockquote',
    'br',
    'dd',
    'div',
    'dl',
    'dt',
    'fieldset',
    'figcaption',
    'figure',
    'footer',
    'form',
    'h1',
    'h2',
    'h3',
    'h4',
    'h5',
    'h6',
    'header',
    'hr',
    'li',
    'main',
    'nav',
    'ol',
    'p',
    'pre',
    'section',
    'table',
    'tbody',
    'td',
    'tfoot',
    'th',
    'thead',
    'tr',
    'ul'
  ])
  const blockedStack: string[] = []
  const parts: string[] = []
  let cursor = 0

  while (cursor < html.length) {
    const tagStart = html.indexOf('<', cursor)
    if (tagStart < 0) {
      if (blockedStack.length === 0) parts.push(html.slice(cursor))
      break
    }
    if (tagStart > cursor && blockedStack.length === 0) {
      parts.push(html.slice(cursor, tagStart))
    }

    if (html.startsWith('<!--', tagStart)) {
      const commentEnd = html.indexOf('-->', tagStart + 4)
      cursor = commentEnd < 0 ? html.length : commentEnd + 3
      continue
    }

    const tagEnd = findTagEnd(html, tagStart)
    if (tagEnd < 0) {
      if (blockedStack.length === 0) parts.push(html.slice(tagStart))
      break
    }
    const source = html.slice(tagStart + 1, tagEnd).trim()
    const parsed = /^(\/)?\s*([a-z][a-z0-9:-]*)/i.exec(source)
    if (!parsed) {
      if (blockedStack.length === 0 && !source.startsWith('!') && !source.startsWith('?')) {
        parts.push('<')
      }
      cursor = tagEnd + 1
      continue
    }

    const closing = Boolean(parsed[1])
    const tagName = parsed[2].toLowerCase()
    if (blockedTags.has(tagName)) {
      if (closing) {
        const matchingIndex = blockedStack.lastIndexOf(tagName)
        if (matchingIndex >= 0) blockedStack.splice(matchingIndex)
      } else {
        // In text/html, a trailing slash does not self-close script/style/head.
        // Conservatively suppress until a real closing tag is encountered.
        blockedStack.push(tagName)
      }
      cursor = tagEnd + 1
      continue
    }

    if (blockedStack.length === 0 && breakTags.has(tagName)) {
      if (!closing && tagName === 'li') parts.push('\n• ')
      else parts.push('\n')
    }
    cursor = tagEnd + 1
  }

  return { title: plainTitle(html), text: normalizeReadableText(parts.join('')) }
}

const clipText = (text: string): { text: string; truncated: boolean } => {
  if (text.length <= COWORK_WEB_LIMITS.textCharacters) return { text, truncated: false }
  let clipped = text.slice(0, COWORK_WEB_LIMITS.textCharacters)
  const last = clipped.charCodeAt(clipped.length - 1)
  if (last >= 0xd800 && last <= 0xdbff) clipped = clipped.slice(0, -1)
  return { text: clipped.trimEnd(), truncated: true }
}

const decodeBody = (body: Buffer, rawContentType: string): string => {
  const charset = /charset\s*=\s*["']?([^;\s"']+)/i.exec(rawContentType)?.[1]?.toLowerCase()
  if (charset === 'iso-8859-1' || charset === 'latin1' || charset === 'windows-1252') {
    return body.toString('latin1')
  }
  return body.toString('utf8')
}

const presentResponse = (url: URL, response: HopResponse): CoworkWebResult => {
  const contentType = response.contentType as CoworkWebResult['contentType']
  const decoded = decodeBody(response.body, response.contentType ?? '')
  let title: string | null = null
  let readable = decoded

  if (contentType === 'text/html') {
    const converted = htmlToReadableText(decoded)
    title = converted.title
    readable = converted.text
  } else if (contentType === 'text/plain') {
    readable = normalizeReadableText(decoded)
  } else {
    readable = decoded.trim()
  }

  const clipped = clipText(readable)
  return {
    finalUrl: url.toString(),
    title,
    text: clipped.text,
    contentType,
    truncated: response.truncated || clipped.truncated
  }
}

const normalizedTimeout = (timeoutMs: number | undefined): number => {
  if (timeoutMs === undefined || !Number.isFinite(timeoutMs)) {
    return COWORK_WEB_LIMITS.defaultTimeoutMs
  }
  return Math.min(Math.max(Math.floor(timeoutMs), 1), COWORK_WEB_LIMITS.maximumTimeoutMs)
}

/**
 * Retrieves one public HTTPS document without ambient credentials, cookies,
 * referrers, automatic redirects, decompression, script execution, or DNS rebinding.
 */
export const retrieveCoworkWebPage = async (
  rawUrl: string,
  options: CoworkWebOptions = {}
): Promise<CoworkWebResult> => {
  const timeoutMs = normalizedTimeout(options.timeoutMs)
  const controller = new AbortController()
  let timedOut = false
  let externallyAborted = false

  const onExternalAbort = (): void => {
    externallyAborted = true
    controller.abort()
  }
  if (options.signal?.aborted) onExternalAbort()
  else options.signal?.addEventListener('abort', onExternalAbort, { once: true })

  const timer = setTimeout(() => {
    timedOut = true
    controller.abort()
  }, timeoutMs)
  timer.unref?.()

  try {
    let current = validateUrl(rawUrl)
    const visited = new Set<string>()

    for (let redirects = 0; ; redirects += 1) {
      if (controller.signal.aborted) throw abortError()
      const canonical = current.toString()
      if (visited.has(canonical)) {
        throw new CoworkWebError('redirect-loop', 'The web request entered a redirect loop')
      }
      visited.add(canonical)

      const binding = await resolvePublicAddress(current, controller.signal)
      const response = await requestHop(current, binding, controller.signal)
      if (!REDIRECT_STATUS_CODES.has(response.statusCode)) {
        return presentResponse(current, response)
      }

      if (!response.location) {
        throw new CoworkWebError('request-failed', 'The web redirect did not include a location')
      }
      if (redirects >= COWORK_WEB_LIMITS.redirects) {
        throw new CoworkWebError('redirect-limit', 'The web request exceeded the redirect limit')
      }

      let redirected: string
      try {
        redirected = new URL(response.location, current).toString()
      } catch (error) {
        throw new CoworkWebError('invalid-url', 'The web redirect location is invalid', {
          cause: error
        })
      }
      const next = validateUrl(redirected)
      if (next.origin !== current.origin) {
        throw new CoworkWebError(
          'cross-origin-redirect',
          'The web request refused a redirect to a different origin'
        )
      }
      current = next
    }
  } catch (error) {
    if (timedOut) throw new CoworkWebError('timeout', 'The web request timed out')
    if (externallyAborted) throw abortError()
    throw error
  } finally {
    clearTimeout(timer)
    options.signal?.removeEventListener('abort', onExternalAbort)
  }
}
